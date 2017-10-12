package counter

import (
	"context"
	"strconv"
	"time"

	"github.com/koding/ropecount/pkg"
)

// Service is the interface for counter operations.
type Service interface {
	Start(ctx context.Context, p StartRequest) (string, error)
	Stop(ctx context.Context, p StopRequest) (string, error)
}

type counterService struct {
	app *pkg.App
}

// NewService creates a Counter service backend.
func NewService(app *pkg.App) Service {
	return &counterService{
		app: app,
	}
}

func (c *counterService) Start(ctx context.Context, p StartRequest) (string, error) {
	// Create the Claims
	claims := &pkg.JWTData{
		Source:   p.Source,
		Target:   p.Target,
		FuncName: p.FuncName,
	}

	tokenString, err := pkg.SignJWT(claims)
	if err != nil {
		return "", err
	}

	c.app.Logger.Log("signedstring", tokenString)

	claims2, err := pkg.ParseJWT(c.app.Logger, tokenString)
	if err != nil {
		return "", err
	}

	c.app.Logger.Log("claims", claims2, "err", err)
	return tokenString, nil
}

func (c *counterService) Stop(ctx context.Context, p StopRequest) (string, error) {
	c.app.Logger.Log("signedstring", p.Token)

	claims, err := pkg.ParseJWT(c.app.Logger, p.Token)
	if err != nil {
		return "", err
	}

	dur := time.Now().UTC().Sub(claims.CreatedAt)
	c.app.Logger.Log("took", dur.String())

	redisConn := c.app.MustGetRedis()
	conn := redisConn.Pool().Get()
	defer conn.Close()

	redisConn.SetPrefix("ropecount")
	// We dont need to DISCARD on error cases. Conn.Close already handles them.
	// For futher info see pool.go/pooledConnection::Close()
	if _, err := conn.Do("MULTI"); err != nil {
		return "", err
	}

	d := 5 * time.Minute // we only work in around 5 mins
	segment := time.Now().UTC().Add(-(d / 2)).Round(d).Unix()
	suffix := ":" + strconv.FormatInt(segment, 10)

	if _, err := conn.Do("SADD", redisConn.AddPrefix("set:counter:src"+suffix), claims.Source); err != nil {
		return "", err
	}

	if _, err := conn.Do("SADD", redisConn.AddPrefix("set:counter:tgt"+suffix), claims.Target); err != nil {
		return "", err
	}

	if _, err := conn.Do("HINCRBY", redisConn.AddPrefix("hset:counter:src"+suffix+":"+claims.Source), claims.FuncName, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("HINCRBY", redisConn.AddPrefix("hset:counter:tgt:"+suffix+":"+claims.Target), claims.FuncName, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return "", err
	}

	return dur.String(), nil
}
