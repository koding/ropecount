package counter

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// Endpoints collects all of the endpoints that compose a counter service.
type Endpoints struct {
	StartEndpoint endpoint.Endpoint
	StopEndpoint  endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a counter
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		StartEndpoint: MakeStartEndpoint(s),
		StopEndpoint:  MakeStopEndpoint(s),
	}
}

// Start implements Service. Primarily useful in a client.
func (e Endpoints) Start(ctx context.Context, request StartRequest) (string, error) {
	response, err := e.StartEndpoint(ctx, request)
	if err != nil {
		return "", err
	}
	resp := response.(StartResponse)
	return resp.Token, resp.Err
}

// Stop implements Service. Primarily useful in a client.
func (e Endpoints) Stop(ctx context.Context, request StopRequest) (string, error) {
	response, err := e.StopEndpoint(ctx, request)
	if err != nil {
		return "", err
	}
	resp := response.(StopResponse)
	return "", resp.Err
}

// MakeStartEndpoint returns an endpoint for the server.
func MakeStartEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(StartRequest)
		token, e := s.Start(ctx, req)
		return StartResponse{Token: token, Err: e}, nil
	}
}

// MakeStopEndpoint returns an endpoint for the server.
func MakeStopEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(StopRequest)
		_, e := s.Stop(ctx, req)
		return StopResponse{Err: e}, nil
	}
}

// StartResponse holds the response data for the Start handler
type StartResponse struct {
	Token string `json:"token,omitempty"`
	Err   error  `json:"err,omitempty"`
}

func (r StartResponse) error() error { return r.Err }

// StopResponse holds the response data for the Stop handler
type StopResponse struct {
	Err error `json:"err,omitempty"`
}

func (r StopResponse) error() error { return r.Err }
