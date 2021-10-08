// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"errors"
	"fmt"

	"github.com/mainflux/mainflux"
)

var (
	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
	// when accessing a protected resource.
	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")

	// ErrCreateUUID indicates error in creating uuid for entity creation
	ErrCreateUUID = errors.New("uuid creation failed")

	// ErrCreateEntity indicates error in creating entity or entities
	ErrCreateEntity = errors.New("create entity failed")

	// ErrUpdateEntity indicates error in updating entity or entities
	ErrUpdateEntity = errors.New("update entity failed")

	// ErrViewEntity indicates error in viewing entity or entities
	ErrViewEntity = errors.New("view entity failed")

	// ErrRemoveEntity indicates error in removing entity
	ErrRemoveEntity = errors.New("remove entity failed")

	// ErrConnect indicates error in adding connection
	ErrConnect = errors.New("add connection failed")

	// ErrDisconnect indicates error in removing connection
	ErrDisconnect = errors.New("remove connection failed")

	// ErrFailedToRetrieveThings failed to retrieve things.
	ErrFailedToRetrieveThings = errors.New("failed to retrieve group members")
)

// var (
// 	// ErrMalformedEntity indicates malformed entity specification (e.g.
// 	// invalid username or password).
// 	ErrMalformedEntity = errors.New("malformed entity specification")

// 	// ErrUnauthorizedAccess indicates missing or invalid credentials provided
// 	// when accessing a protected resource.
// 	ErrUnauthorizedAccess = errors.New("missing or invalid credentials provided")
// )

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	CreateCommand(token string, commands Command) (string, error)
	ViewCommand(token string, id string) (Command, error)
	ListCommands(token string, filter interface{}) ([]Command, error)
	UpdateCommand(token string, commands Command) error
	RemoveCommand(token string, id string) error
}

type PageMetadata struct {
	Total        uint64
	Offset       uint64                 `json:"offset,omitempty"`
	Limit        uint64                 `json:"limit,omitempty"`
	Name         string                 `json:"name,omitempty"`
	Order        string                 `json:"order,omitempty"`
	Dir          string                 `json:"dir,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Disconnected bool                   // Used for connected or disconnected lists
}

type commandsService struct {
	repo CommandRepository
	auth mainflux.AuthServiceClient
}

var _ Service = (*commandsService)(nil)

func New(repo CommandRepository) Service {
	return commandsService{
		repo: repo,
	}
}
func (ks commandsService) CreateCommand(token string, commands Command) (string, error) {
	fmt.Println("Command Created")
	return "", nil
}

func (ks commandsService) ViewCommand(token, id string) (Command, error) {
	fmt.Println("View Command")
	return Command{}, nil
}

func (ks commandsService) ListCommands(token string, filter interface{}) ([]Command, error) {
	fmt.Println("List Command")
	return nil, nil
}

func (ks commandsService) UpdateCommand(token string, command Command) error {
	fmt.Println("Command Updated")
	return nil
}

func (ks commandsService) RemoveCommand(token, id string) error {
	fmt.Println("Command removed")
	return nil
}
