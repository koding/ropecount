package counter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	httptransport "github.com/go-kit/kit/transport/http"
)

// MakeHTTPClientEndpoints returns an Endpoints for counter client.
func MakeHTTPClientEndpoints(instance string, options ...httptransport.ClientOption) (Endpoints, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	tgt, err := url.Parse(instance)
	if err != nil {
		return Endpoints{}, err
	}
	tgt.Path = ""

	return Endpoints{
		StartEndpoint: httptransport.NewClient("POST", tgt, encodeStartRequest, decodeStartResponse, options...).Endpoint(),
		StopEndpoint:  httptransport.NewClient("POST", tgt, encodeStopRequest, decodeStopResponse, options...).Endpoint(),
	}, nil
}

func encodeStartRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "POST", "/start/"
	return encodeRequest(ctx, req, request)
}

func encodeStopRequest(ctx context.Context, req *http.Request, request interface{}) error {
	req.Method, req.URL.Path = "POST", "/stop/"
	return encodeRequest(ctx, req, request)
}

func decodeStartResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response StartResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}

func decodeStopResponse(_ context.Context, resp *http.Response) (interface{}, error) {
	var response StopResponse
	err := json.NewDecoder(resp.Body).Decode(&response)
	return response, err
}
