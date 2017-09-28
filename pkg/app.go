package pkg

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// App is the context for services.
type App struct {
	Logger log.Logger
	Redis  *redis.Pool

	redisAddr *string
	httpAddr  *string
}

const (
	// ConfHTTPAddr holds the flag name for http addres
	ConfHTTPAddr = "http.addr"

	// ConfRedisAddr holds the flag name for redis server addres
	ConfRedisAddr = "redis.addr"
)

// NewApp creates a new App context for the system.
func NewApp(name string, conf *flag.FlagSet) *App {

	app := &App{}
	var err error
	var redisPool *redis.Pool
	var logger log.Logger

	{ // initialize logger
		logger = log.NewJSONLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "service", name)
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	err = conf.Parse(os.Args[1:])
	dieIfError(logger, err, "flagparse")

	{ // initialize configs
		if redisFlag := conf.Lookup(ConfRedisAddr); redisFlag != nil {
			redisAddr := redisFlag.Value.String()
			app.redisAddr = &redisAddr
		}
		if httpFlag := conf.Lookup(ConfHTTPAddr); httpFlag != nil {
			httpAddr := httpFlag.Value.String()
			app.httpAddr = &httpAddr
		}
	}

	{ // initialize if redis is given as config
		if app.redisAddr != nil {
			redisPool, err = NewRedisPool(*app.redisAddr)
			dieIfError(logger, err, "redisconn")
		}
	}

	return &App{
		Logger: logger,
		Redis:  redisPool,
	}
}

func dieIfError(logger log.Logger, err error, name string) {
	if err != nil {
		level.Error(logger).Log(name, err)
		os.Exit(1)
	}
}

// AddRedisConf adds redis conf onto flags.
func AddRedisConf(conf *flag.FlagSet) *string {
	return conf.String(ConfRedisAddr, "redis:6379", "Redis server address")
}

// AddHTTPConf adds redis conf onto flags.
func AddHTTPConf(conf *flag.FlagSet) *string {
	return conf.String(ConfHTTPAddr, ":8080", "HTTP listen address")
}

// Listen waits for app shutdown.
func (a *App) Listen(handler http.Handler) chan error {
	errs := make(chan error)

	// TODO go func is not required here for now, added for future extensibility.
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	{ // initialize if redis is given as config
		if a.httpAddr != nil {
			go func() {
				a.Logger.Log("transport", "HTTP", "addr", *a.httpAddr)
				errs <- http.ListenAndServe(*a.httpAddr, handler)
			}()
		}
	}

	return errs
}
