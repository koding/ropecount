package pkg

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

// NewRedisPool creates a redis connection pool
func NewRedisPool(server string, options ...redis.DialOption) (*redis.Pool, error) {
	if len(options) == 0 {
		options = []redis.DialOption{
			redis.DialReadTimeout(5 * time.Second),
			redis.DialWriteTimeout(time.Second),
			redis.DialConnectTimeout(time.Second),
		}
	}

	p := &redis.Pool{
		MaxIdle:     3,
		MaxActive:   1000,
		IdleTimeout: 30 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server, options...)
			if err != nil {
				return nil, err
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

	// test if the connection is working or not.
	if err := p.Get().Close(); err != nil {
		return nil, err
	}

	return p, nil
}
