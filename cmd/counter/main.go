package main

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/counter"
)

func main() {
	name := "counter"
	app := pkg.NewApp(name, pkg.ConfigureHTTP(), pkg.ConfigureRedis())

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
