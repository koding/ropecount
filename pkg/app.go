package pkg

import (
	"github.com/garyburd/redigo/redis"
	"github.com/go-kit/kit/log"
)

// App is the context for services.
type App struct {
	Logger log.Logger
	Redis  *redis.Pool
}
