package counter

import (
	"context"
	"time"

	"github.com/koding/ropecount/pkg"
)

// Service is the interface for counter operations.
type Service interface {
	Start(ctx context.Context, p StartRequest) (string, error)
	Stop(ctx context.Context, p StopRequest) (string, error)
}

// StartRequest represents a single Start request.
type StartRequest struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	FuncName string `json:"funcName"`
}

// StopRequest represents a single Stop request.
type StopRequest struct {
	Token string `json:"token"`
}

type counterService struct {
	app *pkg.App
}

// NewService creates a Start service backend.
func NewService(app *pkg.App) Service {
	return &counterService{
		app: app,
	}
}

func (s *counterService) Start(ctx context.Context, p StartRequest) (string, error) {
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

	s.app.Logger.Log("signedstring", tokenString)

	claims2, err := pkg.ParseJWT(s.app.Logger, tokenString)
	if err != nil {
		return "", err
	}

	s.app.Logger.Log("claims", claims2, "err", err)
	return tokenString, nil
}

func (s *counterService) Stop(ctx context.Context, p StopRequest) (string, error) {
	s.app.Logger.Log("signedstring", p.Token)

	claims, err := pkg.ParseJWT(s.app.Logger, p.Token)
	if err != nil {
		return "", err
	}

	dur := time.Now().UTC().Sub(claims.CreatedAt)
	s.app.Logger.Log("took", dur.String())

	redisConn := s.app.MustGetRedis()
	conn := redisConn.Pool().Get()
	defer conn.Close()

	// We dont need to DISCARD on error cases. Conn.Close already handles them.
	// For futher info see pool.go/pooledConnection::Close()
	if _, err := conn.Do("MULTI"); err != nil {
		return "", err
	}

	if _, err := conn.Do("INCRBY", "src:"+claims.Source, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("INCRBY", "tgt:"+claims.Target, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("INCRBY", "fn:"+claims.FuncName, int64(dur)); err != nil {
		return "", err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return "", err
	}

	return dur.String(), nil
}
