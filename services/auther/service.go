package auther

import (
	"context"

	"github.com/go-kit/kit/log"
	"github.com/koding/ropecount/pkg"
)

// Service is the interface for auther operations.
type Service interface {
	Auth(ctx context.Context, p Auth) (string, error)
}

// Auth represents a single auth request.
type Auth struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	FuncName string `json:"funcName"`
}

type autherService struct {
	logger log.Logger
}

// NewService creates a Auth service backend.
func NewService(logger log.Logger) Service {
	return &autherService{
		logger: logger,
	}
}

func (s *autherService) Auth(ctx context.Context, p Auth) (string, error) {
	// Create the Claims
	claims := &pkg.JWTData{
		Source: p.Source,
		Target: p.Target,
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
