package main

import (
	"flag"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/counter"
)

func main() {
	name := "counter"
	conf := flag.NewFlagSet(name, flag.ExitOnError)

	pkg.AddHTTPConf(conf)
	pkg.AddRedisConf(conf)

	app := pkg.NewApp(name, conf)

	var s counter.Service
	{
		s = counter.NewService(app)
		s = counter.LoggingMiddleware(app.Logger)(s)
	}

	var h http.Handler
	{
		h = counter.MakeHTTPHandler(s, log.With(app.Logger, "component", "HTTP"))
	}

	app.Logger.Log("exit", <-app.Listen(h))
}
