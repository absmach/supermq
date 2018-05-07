package mocks

import (
	"context"

	"github.com/mainflux/mainflux/clients"

	clientsapi "github.com/mainflux/mainflux/clients/api/grpc"
	"google.golang.org/grpc"
)

var _ clientsapi.ClientsServiceClient = (*clientsClient)(nil)

type clientsClient struct {
	clients map[string]string
}

// NewClientsClient returns mock implementation of clients service client.
func NewClientsClient(data map[string]string) clientsapi.ClientsServiceClient {
	return &clientsClient{data}
}

func (client clientsClient) CanAccess(ctx context.Context, req *clientsapi.AccessReq, opts ...grpc.CallOption) (*clientsapi.Identity, error) {
	key := req.GetClientKey()
	if key == "" {
		return nil, clients.ErrUnauthorizedAccess
	}

	id, ok := client.clients[key]
	if !ok {
		return nil, clients.ErrUnauthorizedAccess
	}

	return &clientsapi.Identity{id}, nil
}
