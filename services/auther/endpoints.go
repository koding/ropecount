package auther

import (
	"context"
	"net/url"
	"strings"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
)

// Endpoints collects all of the endpoints that compose a auther service.
type Endpoints struct {
	AuthEndpoint endpoint.Endpoint
}

// MakeServerEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the provided service. Useful in a auther
// server.
func MakeServerEndpoints(s Service) Endpoints {
	return Endpoints{
		AuthEndpoint: MakeAuthEndpoint(s),
	}
}

// MakeClientEndpoints returns an Endpoints for auther client.
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

	return Endpoints{
		AuthEndpoint: httptransport.NewClient("POST", tgt, encodeAuthRequest, decodeAuthResponse, options...).Endpoint(),
	}, nil
}

// Auth implements Service. Primarily useful in a client.
func (e Endpoints) Auth(ctx context.Context, p Auth) (string, error) {
	request := AuthRequest{Auth: p}
	response, err := e.AuthEndpoint(ctx, request)
	if err != nil {
		return "", err
	}
	resp := response.(AuthResponse)
	return resp.Token, resp.Err
}

// MakeAuthEndpoint returns an endpoint for the server.
func MakeAuthEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(AuthRequest)
		token, e := s.Auth(ctx, req.Auth)
		return AuthResponse{Token: token, Err: e}, nil
	}
}

// AuthRequest holds the request data for the Auth handler
type AuthRequest struct {
	Auth Auth
}

// AuthResponse holds the response data for the Auth handler
type AuthResponse struct {
	Token string `json:"token,omitempty"`
	Err   error  `json:"err,omitempty"`
}

func (r AuthResponse) error() error { return r.Err }
