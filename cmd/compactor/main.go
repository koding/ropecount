package main

import (
	"flag"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/compactor"
)

func main() {
	name := "compator"
	conf := flag.NewFlagSet(name, flag.ExitOnError)

	pkg.AddHTTPConf(conf)
	pkg.AddRedisConf(conf)
	pkg.AddMongoConf(conf)

	app := pkg.NewApp(name, conf)

	var s compactor.Service
	{
		s = compactor.NewService(app)
		s = compactor.LoggingMiddleware(app.Logger)(s)
	}

	var h http.Handler
	{
		h = compactor.MakeHTTPHandler(s, log.With(app.Logger, "component", "HTTP"))
	}

	app.Logger.Log("exit", <-app.Listen(h))
}
