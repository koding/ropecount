package main

import (
	"context"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
	"github.com/koding/ropecount/services/compactor"
)

func main() {
	name := "compactor"
	app := pkg.NewApp(name, pkg.ConfigureHTTP(), pkg.ConfigureRedis(), pkg.ConfigureMongo())

	var s compactor.Service
	{
		s = compactor.NewService(app)
		s = compactor.LoggingMiddleware(app.Logger)(s)
	}

	var h http.Handler
	{
		h = compactor.MakeHTTPHandler(s, log.With(app.Logger, "component", "HTTP"))
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := s.Process(ctx, compactor.ProcessRequest{StartAt: time.Now().UTC()}); err != nil {
					app.ErrorLog("err", err.Error())
				}
			}
		}
	}()
	app.Logger.Log("exit", <-app.Listen(h))
	cancel()
}
