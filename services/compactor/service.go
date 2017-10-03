package compactor

import (
	"context"
	"fmt"
	"time"

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

func (c *compactorService) Process(ctx context.Context, p ProcessRequest) error {
	c.app.Logger.Log("starttime", p.StartAt.Format(time.RFC3339))

	d := 5 * time.Minute // we only work in around 5 mins
	t := p.StartAt

	// We compact the values that are set 2*d duration before.
	// d/2 is required for time.Round(d). It rounds up after the halfway values.
	tr := t.Add(-d * 2).Add(-(d / 2)).Round(d)
	tl := tr.Add(-time.Hour) // / process till this time

	for tl.UnixNano() <= tr.UnixNano() {
		fmt.Printf(" = %s\n", tr.Format(time.RFC3339))

		tr = tr.Add(-d)
	}

	// redisConn := c.app.MustGetRedis()
	// conn := redisConn.Pool().Get()
	// defer conn.Close()

	// // We dont need to DISCARD on error cases. Conn.Close already handles them.
	// // For futher info see pool.go/pooledConnection::Close()
	// if _, err := conn.Do("MULTI"); err != nil {
	// 	return err
	// }

	// if _, err := conn.Do("EXEC"); err != nil {
	// 	return err
	// }

	return nil
}
