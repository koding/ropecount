package compactor

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Endpoints collects all of the endpoints that compose a compactor service.
type Endpoints struct {
	ProcessEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a compactor
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		ProcessEndpoint: MakeProcessEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
// Useful in a compactor client.
func MakeClientEndpoints(instance string) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	options := []httptransport.ClientOption{}

	// Note that the request encoders need to modify the request URL, changing
	// the path and method. That's fine: we simply need to provide specific
	// encoders for each endpoint.

	return Endpoints{
		ProcessEndpoint: httptransport.NewClient("POST", tgt, encodeProcessRequest, decodeProcessResponse, options...).Endpoint(),
	}, nil
}

// Process implements Service. Primarily useful in a client.
func (e Endpoints) Process(ctx context.Context, p Profile) error {
	request := ProcessRequest{Profile: p}
	response, err := e.ProcessEndpoint(ctx, request)
	if err != nil {
		return err
	}
	resp := response.(ProcessResponse)
	return resp.Err
}

// MakeProcessEndpoint returns an endpoint via the passed service.
// Primarily useful in a server.
func MakeProcessEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(ProcessRequest)
		e := s.Process(ctx, req.Profile)
		return ProcessResponse{Err: e}, nil
	}
}

// ProcessRequest holds the request data for the Process handler
type ProcessRequest struct {
	Profile Profile
}

// ProcessResponse holds the response data for the Process handler
type ProcessResponse struct {
	Err error `json:"err,omitempty"`
}

func (r ProcessResponse) error() error { return r.Err }
