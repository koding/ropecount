package counter

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service

// LoggingMiddleware logs the incoming method, id and request latency data
func LoggingMiddleware(logger log.Logger) Middleware {
	return func(next Service) Service {
		return &loggingMiddleware{
			next:   next,
			logger: logger,
		}
	}
}

type loggingMiddleware struct {
	next   Service
	logger log.Logger
}

func (mw loggingMiddleware) Start(ctx context.Context, p StartRequest) (token string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Start", "source", p.Source, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Start(ctx, p)
}

func (mw loggingMiddleware) Stop(ctx context.Context, p StopRequest) (token string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Stop", "source", p.Token, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Stop(ctx, p)
}
