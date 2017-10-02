package counter

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
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
	logger log.Logger
}

// NewService creates a Start service backend.
func NewService(logger log.Logger) Service {
	return &counterService{
		logger: logger,
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

	s.logger.Log("signedstring", tokenString)

	claims2, err := pkg.ParseJWT(s.logger, tokenString)
	if err != nil {
		return "", err
	}

	s.logger.Log("claims", claims2, "err", err)
	return tokenString, nil
}

func (s *counterService) Stop(ctx context.Context, p StopRequest) (string, error) {
	s.logger.Log("signedstring", p.Token)

	claims, err := pkg.ParseJWT(s.logger, p.Token)
	if err != nil {
		return "", err
	}

	dur := time.Now().UTC().Sub(claims.CreatedAt)
	s.logger.Log("took", dur.String())

	return dur.String(), nil
}
