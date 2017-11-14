package counter

import (
	"context"
	"time"

	"github.com/ropelive/count/pkg"
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

	segment := pkg.GetCurrentSegment()
	keyNames := pkg.GenerateKeyNames(segment)
	currentSrcHSet := keyNames.Src.HashSetName(claims.Source)
	currentDstHSet := keyNames.Dst.HashSetName(claims.Target)

	redisConn.SetPrefix("ropecount")
	// We dont need to DISCARD on error cases. Conn.Close already handles them.
	// For futher info see pool.go/pooledConnection::Close()
	if _, err := conn.Do("MULTI"); err != nil {
		return "", err
	}

	if _, err := conn.Do("SADD", redisConn.AddPrefix(keyNames.Src.CurrentCounterSet), claims.Source); err != nil {
		return "", err
	}

	if _, err := conn.Do("SADD", redisConn.AddPrefix(keyNames.Dst.CurrentCounterSet), claims.Target); err != nil {
		return "", err
	}

	if _, err := conn.Do("HINCRBY", redisConn.AddPrefix(currentSrcHSet), claims.FuncName, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("HINCRBY", redisConn.AddPrefix(currentDstHSet), claims.FuncName, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return "", err
	}

	return dur.String(), nil
}
