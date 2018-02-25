package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/http"
)

func sendMessageEndpoint(svc http.Service) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		msg := request.(mainflux.Message)
		err := svc.Publish(msg)
		return nil, err
	}
}
