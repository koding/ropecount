package persister

import (
	"context"

	"github.com/koding/ropecount/pkg"
)

// Service is a simple interface for persister operations.
type Service interface {
	Process(ctx context.Context, req ProcessRequest) error
}

type persisterService struct {
	app *pkg.App
}

// NewService creates a Persister service
func NewService(app *pkg.App) Service {
	return &persisterService{
		app: app,
	}
}

func (s *persisterService) Process(ctx context.Context, req ProcessRequest) error {
	return nil
}
