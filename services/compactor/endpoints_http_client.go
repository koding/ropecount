package compactor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
)

// MakeHTTPClientEndpoints returns an Endpoints struct where each endpoint
// invokes the corresponding method on the remote instance, via a
// transport/http.Client. Useful in a compactor client.
func MakeHTTPClientEndpoints(instance string) (Endpoints, error) {
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

func decodeProcessResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response ProcessResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}
