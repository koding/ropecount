package counter

import (
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a counter server.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/start").Handler(httptransport.NewServer(
		MakeStartEndpoint(s),
		decodeStartRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/stop").Handler(httptransport.NewServer(
		MakeStopEndpoint(s),
		decodeStopRequest,
		encodeResponse,
		options...,
	))

	return r
}
