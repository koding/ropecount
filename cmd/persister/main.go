package main

import (
	"flag"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/persister"
)

func main() {
	name := "persister"
	conf := flag.NewFlagSet(name, flag.ExitOnError)

	pkg.AddHTTPConf(conf)
	pkg.AddRedisConf(conf)

	app := pkg.NewApp(name, conf)

	var s persister.Service
	{
		s = persister.NewInmemService(app.Logger)
		s = persister.LoggingMiddleware(app.Logger)(s)
	}

	var h http.Handler
	{
		h = persister.MakeHTTPHandler(s, log.With(app.Logger, "component", "HTTP"))
	}

	app.Logger.Log("exit", <-app.Listen(h))
}
