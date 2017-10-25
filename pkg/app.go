package pkg

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/koding/redis"
	"github.com/koding/ropecount/pkg/mongodb"
)

// App is the context for services.
type App struct {
	Logger log.Logger
	redis  *redis.RedisSession
	mongo  *mongodb.MongoDB

	name     string
	httpAddr *string
}

// NewApp creates a new App context for the system.
func NewApp(name string, opts ...Opts) *App {
	var logger log.Logger

	{ // initialize logger
		logger = log.NewJSONLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "service", name)
		logger = level.NewFilter(logger, level.AllowDebug()) // TODO: make this configurable
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	app := &App{
		name:   name,
		Logger: logger,
	}

	for _, opt := range opts {
		dieIfError(logger, opt(app), "configure")
	}

	return app
}

// MustGetRedis returns the redis if it is already initialized. If the config is
// not given or the connection is not established yet, panics.
func (a *App) MustGetRedis() *redis.RedisSession {
	if a.redis == nil {
		panic("redis is not initialized yet.")
	}
	return a.redis
}

// MustGetMongo returns the Mongo if it is already initialized. If the config is
// not given or the connection is not established yet, panics.
func (a *App) MustGetMongo() *mongodb.MongoDB {
	if a.mongo == nil {
		panic("mongo is not initialized yet.")
	}
	return a.mongo
}

// Opts configures the application
type Opts func(*App) error

// ConfigureRedis configures redis
func ConfigureRedis() func(*App) error {
	url := os.Getenv("REDIS_URL")
	if url == "" {
		url = "redis:6379"
	}

	return func(app *App) error {
		var err error
		app.redis, err = NewRedisPool(url)
		if err != nil {
			return fmt.Errorf("redisconn: %s", err)
		}
		return nil
	}
}

// ConfigureMongo configures Mongo
func ConfigureMongo() func(*App) error {
	url := os.Getenv("MONGO_URL")
	if url == "" {
		url = "mongodb://mongo:27017"
	}

	return func(app *App) error {
		var err error
		app.mongo, err = mongodb.New(url)
		if err != nil {
			return fmt.Errorf("mongoconn: %s", err)
		}
		return nil
	}
}

// ConfigureHTTP configures HTTP server
func ConfigureHTTP() func(*App) error {
	uri := os.Getenv("HTTP_ADDR")
	if uri == "" {
		uri = ":8080"
	}

	return func(app *App) error {
		app.httpAddr = &uri
		return nil
	}
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

	a.Logger.Log("func", "http listen")
	return errs
}

func (a *App) ErrorLog(keyvals ...interface{}) error {
	return level.Error(a.Logger).Log(keyvals...)
}

func (a *App) WarnLog(keyvals ...interface{}) error {
	return level.Warn(a.Logger).Log(keyvals...)
}

func (a *App) InfoLog(keyvals ...interface{}) error {
	return level.Info(a.Logger).Log(keyvals...)
}

func (a *App) DebugLog(keyvals ...interface{}) error {
	return level.Debug(a.Logger).Log(keyvals...)
}

func dieIfError(logger log.Logger, err error, name string) {
	if err != nil {
		level.Error(logger).Log(name, err)
		os.Exit(1)
	}
}
