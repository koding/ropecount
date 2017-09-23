package persister

import (
	"context"
	"sync"

	"github.com/go-kit/kit/log"
)

// Service is a simple interface for persister operations.
type Service interface {
	Process(ctx context.Context, p Profile) error
}

// Profile represents a single user profile.
// ID should be globally unique.
type Profile struct {
	ID string `json:"id"`
}

type inmemService struct {
	mtx    sync.RWMutex
	logger log.Logger
	m      map[string]Profile
}

// NewInmemService creates a Compator service
func NewInmemService(logger log.Logger) Service {
	return &inmemService{
		m:      map[string]Profile{},
		logger: logger,
	}
}

func (s *inmemService) Process(ctx context.Context, p Profile) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if _, ok := s.m[p.ID]; ok {
		return nil
	}

	s.m[p.ID] = p
	return nil
}
