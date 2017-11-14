package clients

import (
	"io"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	"github.com/ropelive/count/services/counter"
)

// NewCounter returns a service that's load-balanced over instances of counter
// found in the provided Consul server. The mechanism of looking up counter
// instances in Consul is hard-coded into the client.
func NewCounter(consulAddr string, logger log.Logger) (counter.Service, error) {
	apiclient, err := consulapi.NewClient(&consulapi.Config{
		Address: consulAddr,
	})
	if err != nil {
		return nil, err
	}

	// As the implementer of counter, we declare and enforce these
	// parameters for all of the counter consumers.
	var (
		consulService = "counter"
		consulTags    = []string{"prod"}
		passingOnly   = true
		retryMax      = 3
		retryTimeout  = 500 * time.Millisecond
	)

	var (
		sdclient  = consul.NewClient(apiclient)
		instancer = consul.NewInstancer(sdclient, logger, consulService, consulTags, passingOnly)
		endpoints counter.Endpoints
	)
	{
		factory := factoryForCounter(counter.MakeStartEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.StartEndpoint = retry
	}
	{
		factory := factoryForCounter(counter.MakeStopEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(retryMax, retryTimeout, balancer)
		endpoints.StopEndpoint = retry
	}

	return endpoints, nil
}

func factoryForCounter(makeEndpoint func(counter.Service) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		service, err := counter.MakeHTTPClientEndpoints(instance)
		if err != nil {
			return nil, nil, err
		}
		return makeEndpoint(service), nil, nil
	}
}
