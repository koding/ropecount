package main

import (
	"flag"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/auther"
)

func main() {
	name := "auther"
	conf := flag.NewFlagSet(name, flag.ExitOnError)

	pkg.AddHTTPConf(conf)
	pkg.AddRedisConf(conf)

	app := pkg.NewApp(name, conf)

	var s auther.Service
	{
		s = auther.NewInmemService(app.Logger)
		s = auther.LoggingMiddleware(app.Logger)(s)
	}

	var h http.Handler
	{
		h = auther.MakeHTTPHandler(s, log.With(app.Logger, "component", "HTTP"))
	}

	app.Logger.Log("exit", <-app.Listen(h))
}
