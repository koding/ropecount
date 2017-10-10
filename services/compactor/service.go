package compactor

import (
	"context"
	"strconv"
	"time"

	redigo "github.com/garyburd/redigo/redis"
	"github.com/koding/redis"
	"github.com/koding/ropecount/pkg"
)

// Service is a simple interface for compactor operations.
type Service interface {
	Process(ctx context.Context, p ProcessRequest) error
}

type compactorService struct {
	app *pkg.App
}

// NewService creates a Compator service
func NewService(app *pkg.App) Service {
	return &compactorService{
		app: app,
	}
}

// Process
func (c *compactorService) Process(ctx context.Context, p ProcessRequest) error {
	c.app.Logger.Log("starttime", p.StartAt.Format(time.RFC3339))

	d := 5 * time.Minute // we only work in around 5 mins
	t := p.StartAt

	// We compact the values that are set 2*d duration before.
	// d/2 is required for time.Round(d). It rounds up after the halfway values.
	tr := t.Add(-d * 2).Add(-(d / 2)).Round(d)
	tl := tr.Add(-time.Hour) // / process till this time

	redisConn := c.app.MustGetRedis()
	redisConn.SetPrefix("ropecount")

	for tl.UnixNano() <= tr.UnixNano() {
		c.app.InfoLog("time", tr.Format(time.RFC3339))

		if err := c.process(redisConn, "src", tr); err != nil {
			return err
		}

		if err := c.process(redisConn, "dst", tr); err != nil {
			return err
		}

		tr = tr.Add(-d)
	}

	return nil
}

func (c *compactorService) process(redisConn *redis.RedisSession, suffix string, tr time.Time) error {
	timedSuffix := suffix + ":" + strconv.FormatInt(tr.Unix(), 10)
	return c.withLock(redisConn, timedSuffix, func(srcMember string) error {
		segment := tr.Add(-(time.Hour / 2)).Round(time.Hour).Unix()

		hashSetPrefix := "hset:counter:" + timedSuffix + ":"
		hourlyDestHashSetPrefix := "hset:counter:" + suffix + ":" + strconv.FormatInt(segment, 10)

		hashSetName := hashSetPrefix + srcMember
		destHashSetName := hourlyDestHashSetPrefix + srcMember

		fns, err := redigo.Int64Map(redisConn.HashGetAll(hashSetName))
		if err == redis.ErrNil {
			c.app.ErrorLog("msg", "item was in the queue but the corresponding values does not exist as hash map")
			return nil
		}

		if err != nil {
			return err
		}

		if len(fns) == 0 {
			return nil
		}

		if err = c.merge(redisConn, fns, destHashSetName); err != nil {
			return err
		}

		res, err := redisConn.Del(hashSetName)
		if err == redis.ErrNil {
			c.app.Logger.Log("msg", "we should be able to delete the counter here, but failed.", "name", hashSetName)
			return nil
		}
		if err != nil {
			return err
		}

		if res == 0 {
			c.app.Logger.Log("msg", "we should be able to delete the counter hash map here, but failed.")
		}

		return nil
	})

}

func (c *compactorService) withLock(redisConn *redis.RedisSession, suffix string, fn func(srcMember string) error) error {
	queueName := "set:counter:" + suffix

	srcMember, err := redisConn.RandomSetMember(queueName)
	if err == redis.ErrNil {
		return nil // we dont have any, so nothing to do.
	}

	if err != nil {
		return err
	}

	processingQueueName := queueName + "_processing"

	res, err := redisConn.MoveSetMember(queueName, processingQueueName, srcMember)
	if err != nil {
		return err
	}
	// 1 if the element is moved. 0 if the element is not a member of source
	// and no operation was performed.
	if res != 1 {
		c.app.InfoLog("msg", "we tried to move a current member to processing queue but failed, someone has already moved the item in the mean time...")
		return c.withLock(redisConn, suffix, fn)
	}

	fnErr := fn(srcMember)

	res, err = redisConn.RemoveSetMembers(processingQueueName, srcMember)
	if err == redis.ErrNil {
		c.app.ErrorLog("msg", "we should be able to delete from the processing set here, but failed.")
		return fnErr
	}

	if err != nil {
		c.app.ErrorLog("msg", err.Error())
		return fnErr
	}

	if res == 0 {
		c.app.ErrorLog("msg", "we should be able to delete from the processing set here, but failed.")
	}

	return fnErr
}

func (c *compactorService) merge(redisConn *redis.RedisSession, fns map[string]int64, destHashSetName string) error {
	conn := redisConn.Pool().Get()
	defer conn.Close()

	// We dont need to DISCARD on error cases. Conn.Close already handles them.
	// For futher info see pool.go/pooledConnection::Close()
	if _, err := conn.Do("MULTI"); err != nil {
		return err
	}

	for fn, val := range fns {
		if _, err := conn.Do("HINCRBY", redisConn.AddPrefix(destHashSetName), fn, val); err != nil {
			return err
		}
	}

	_, err := conn.Do("EXEC")
	return err
}
