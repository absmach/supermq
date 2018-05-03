package manager

import (
	"context"
	"time"

	pb "github.com/mainflux/mainflux/users/api/grpc"
)

var _ Service = (*managerService)(nil)

type managerService struct {
	users    pb.UsersServiceClient
	clients  ClientRepository
	channels ChannelRepository
	hasher   Hasher
	idp      IdentityProvider
}

// New instantiates the domain service implementation.
func New(users pb.UsersServiceClient, clients ClientRepository, channels ChannelRepository, hasher Hasher, idp IdentityProvider) Service {
	return &managerService{
		users:    users,
		clients:  clients,
		channels: channels,
		hasher:   hasher,
		idp:      idp,
	}
}

func (ms *managerService) AddClient(key string, client Client) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return "", ErrUnauthorizedAccess
	}

	client.ID = ms.clients.Id()
	client.Owner = res.Value
	client.Key, _ = ms.idp.PermanentKey(client.ID)

	return client.ID, ms.clients.Save(client)
}

func (ms *managerService) UpdateClient(key string, client Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	client.Owner = res.Value

	return ms.clients.Update(client)
}

func (ms *managerService) ViewClient(key, id string) (Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return Client{}, ErrUnauthorizedAccess
	}

	return ms.clients.One(res.Value, id)
}

func (ms *managerService) ListClients(key string, offset, limit int) ([]Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return nil, ErrUnauthorizedAccess
	}

	return ms.clients.All(res.Value, offset, limit), nil
}

func (ms *managerService) RemoveClient(key, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	return ms.clients.Remove(res.Value, id)
}

func (ms *managerService) CreateChannel(key string, channel Channel) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return "", ErrUnauthorizedAccess
	}

	channel.Owner = res.Value
	return ms.channels.Save(channel)
}

func (ms *managerService) UpdateChannel(key string, channel Channel) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	channel.Owner = res.Value
	return ms.channels.Update(channel)
}

func (ms *managerService) ViewChannel(key, id string) (Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return Channel{}, ErrUnauthorizedAccess
	}

	return ms.channels.One(res.Value, id)
}

func (ms *managerService) ListChannels(key string, offset, limit int) ([]Channel, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return nil, ErrUnauthorizedAccess
	}

	return ms.channels.All(res.Value, offset, limit), nil
}

func (ms *managerService) RemoveChannel(key, id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	return ms.channels.Remove(res.Value, id)
}

func (ms *managerService) Connect(key, chanID, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	return ms.channels.Connect(res.Value, chanID, clientID)
}

func (ms *managerService) Disconnect(key, chanID, clientID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := ms.users.Identify(ctx, &pb.Token{key})
	if err != nil || res.Err != "" {
		return ErrUnauthorizedAccess
	}

	return ms.channels.Disconnect(res.Value, chanID, clientID)
}

func (ms *managerService) Identity(key string) (string, error) {
	client, err := ms.idp.Identity(key)
	if err != nil {
		return "", err
	}

	return client, nil
}

func (ms *managerService) CanAccess(key, channel string) (string, error) {
	client, err := ms.idp.Identity(key)
	if err != nil {
		return "", err
	}

	if !ms.channels.HasClient(channel, client) {
		return "", ErrUnauthorizedAccess
	}

	return client, nil
}
