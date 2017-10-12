package persister

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
)

// MakeClientEndpoints returns an Endpoints struct where each endpoint invokes
// the corresponding method on the remote instance, via a transport/http.Client.
// Useful in a persister client.
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

func encodeProcessRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "POST", "/process/"
	return encodeRequest(ctx, req, request)
}

func decodeProcessRequest(_ context.Context, r *http.Request) (request interface{}, err error) {
	var req ProcessRequest
	if e := json.NewDecoder(r.Body).Decode(&req); e != nil {
		return nil, e
	}
	return req, nil
}
