package pkg

import (
	redigo "github.com/garyburd/redigo/redis"
	"github.com/koding/redis"
)

// NewRedisPool creates a redis connection pool
func NewRedisPool(server string, options ...redigo.DialOption) (*redis.RedisSession, error) {
	r, err := redis.NewRedisSession(&redis.RedisConf{Server: server})
	if err != nil {
		return nil, err
	}

	// test if the connection is working or not.
	if err := r.Pool().Get().Close(); err != nil {
		return nil, err
	}

	return r, nil
}
