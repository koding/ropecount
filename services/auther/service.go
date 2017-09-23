package auther

import (
	"context"
	"sync"

	"github.com/go-kit/kit/log"
)

// Service is the interface for auther operations.
type Service interface {
	Auth(ctx context.Context, p Auth) error
}

// Auth represents a single auth request.
type Auth struct {
	ID string `json:"id"`
}

type inmemService struct {
	mtx    sync.RWMutex
	logger log.Logger
	m      map[string]Auth
}

// NewInmemService creates a Auth service backend.
func NewInmemService(logger log.Logger) Service {
	return &inmemService{
		m:      map[string]Auth{},
		logger: logger,
	}
}

func (s *inmemService) Auth(ctx context.Context, p Auth) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[p.ID]; ok {
		return nil
	}

	s.m[p.ID] = p
	return nil
}
