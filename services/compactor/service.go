package compactor

import (
	"context"
	"errors"
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

		for {
			var srcErr, dstErr error
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				srcErr = c.process(redisConn, "src", tr)
				dstErr = c.process(redisConn, "dst", tr)
			}
			if srcErr == dstErr && srcErr == errNotFound {
				break
			}
			if srcErr != nil && srcErr != errNotFound {
				return srcErr
			}
			if dstErr != nil && dstErr != errNotFound {
				return dstErr
			}
		}

		tr = tr.Add(-d)
	}

	return nil
}

const seperator = ":"

var errNotFound = errors.New("no item to process")

func generateSegmentSuffixes(directionSuffix string, tr time.Time) (current string, hourly string) {
	current = directionSuffix + seperator + strconv.FormatInt(tr.Unix(), 10)
	hourly = directionSuffix + seperator + strconv.FormatInt(tr.Add(-(time.Hour/2)).Round(time.Hour).Unix(), 10)
	return current, hourly
}

func generateHashKeys(currentSuffix, hourlySuffix, srcMember string) (source string, target string) {
	currentHashSetPrefix := "hset:counter:" + currentSuffix + seperator
	hourlyHashSetPrefix := "hset:counter:" + hourlySuffix + seperator

	source = currentHashSetPrefix + srcMember
	target = hourlyHashSetPrefix + srcMember
	return source, target
}

func (c *compactorService) process(redisConn *redis.RedisSession, directionSuffix string, tr time.Time) error {
	var (
		currentSegmentSuffix, hourlySegmentSuffix = generateSegmentSuffixes(directionSuffix, tr)
		queueName                                 = "set:counter:" + currentSegmentSuffix
	)
	c.app.InfoLog("queue", queueName)
	return c.withLock(redisConn, queueName, func(srcMember string) error {
		source, target := generateHashKeys(currentSegmentSuffix, hourlySegmentSuffix, srcMember)
		return c.merge(redisConn, source, target)
	})
}

// withLock gets an item from the current segment's item set and passes it to
// the given processor function. After getting a response from the processor a successfull
func (c *compactorService) withLock(redisConn *redis.RedisSession, queueName string, fn func(srcMember string) error) error {
	srcMember, err := redisConn.RandomSetMember(queueName)
	if err == redis.ErrNil {
		return errNotFound // we dont have any, so nothing to do.
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
		return c.withLock(redisConn, queueName, fn)
	}

	fnErr := fn(srcMember)

	if fnErr != nil {
		_, err = redisConn.MoveSetMember(processingQueueName, queueName, srcMember)
		if err != nil {
			c.app.ErrorLog("msg", "error while trying to put to item back to process set after an unseccesful operation", "err", err.Error())
		}

		return fnErr
	}

	res, err = redisConn.RemoveSetMembers(processingQueueName, srcMember)
	if err == redis.ErrNil {
		c.app.ErrorLog("msg", "we should be able to delete from the processing set here, but failed.")
		return err
	}

	if err != nil {
		c.app.ErrorLog("msg", err.Error())
		return err
	}

	if res == 0 {
		c.app.ErrorLog("msg", "we should be able to delete from the processing set here, but failed.")
	}

	return nil
}

// merge merges the source hash map values to the target, then deletes the
// source hash map from the server.
func (c *compactorService) merge(redisConn *redis.RedisSession, source, target string) error {
	fns, err := redigo.Int64Map(redisConn.HashGetAll(source))
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

	if err = c.incrementMapValues(redisConn, target, fns); err != nil {
		return err
	}

	res, err := redisConn.Del(source)
	if err != redis.ErrNil {
		c.app.Logger.Log("msg", "we should be able to delete the counter here, but failed.", "err", err)
		return nil
	}

	if err != nil {
		return err
	}

	if res == 0 {
		c.app.Logger.Log("msg", "we should be able to delete the counter hash map here, but failed. someone might already have deleted it..")
	}

	return nil
}

func (c *compactorService) incrementMapValues(redisConn *redis.RedisSession, target string, fns map[string]int64) error {
	conn := redisConn.Pool().Get()
	defer conn.Close()

	// We dont need to DISCARD on error cases. Conn.Close already handles them.
	// For futher info see pool.go/pooledConnection::Close()
	if _, err := conn.Do("MULTI"); err != nil {
		return err
	}

	for fn, val := range fns {
		if _, err := conn.Do("HINCRBY", redisConn.AddPrefix(target), fn, val); err != nil {
			return err
		}
	}

	_, err := conn.Do("EXEC")
	return err
}
