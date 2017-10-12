package persister

import (
	"net/http"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// MakeHTTPHandler mounts all of the service endpoints into an http.Handler.
// Useful in a persister server.
func MakeHTTPHandler(s Service, logger log.Logger) http.Handler {
	r := mux.NewRouter()

	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("POST").Path("/process").Handler(httptransport.NewServer(
		MakeProcessEndpoint(s),
		decodeProcessRequest,
		encodeResponse,
		options...,
	))

	return r
}
