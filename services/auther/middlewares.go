package auther

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

func (mw loggingMiddleware) Auth(ctx context.Context, p Auth) (token string, err error) {
	defer func(begin time.Time) {
		mw.logger.Log("method", "Auth", "source", p.Source, "took", time.Since(begin), "err", err)
	}(time.Now())
	return mw.next.Auth(ctx, p)
}
