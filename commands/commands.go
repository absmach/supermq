package commands

import (
	"context"

	"github.com/mainflux/mainflux/pkg/errors"
)

type Metadata map[string]interface{}

type Command struct {
	ID          string
	Owner       string
	Name        string
	Command     string
	ChannelID   string
	ExecuteTime string
	Metadata    Metadata
}

var (
	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")

	// ErrConflict indicates that entity already exists.
	ErrConflict = errors.New("entity already exists")

	// ErrScanMetadata indicates problem with metadata in db
	ErrScanMetadata = errors.New("failed to scan metadata in db")

	// ErrSelectEntity indicates error while reading entity from database
	ErrSelectEntity = errors.New("select entity from db error")

	// ErrEntityConnected indicates error while checking connection in database
	ErrEntityConnected = errors.New("check thing-channel connection in database error")

	ErrMalformedEntity = errors.New("")
)

type CommandPage struct {
	PageMetadata
	Commands []Command
}

type CommandRepository interface {
	Save(ctx context.Context, c Command) (string, error)

	Update(ctx context.Context, u Command) error

	RetrieveByID(ctx context.Context, id string) (Command, error)

	RetrieveAll(ctx context.Context, offset, limit uint64, commandIDs []string, m Metadata) (CommandPage, error)

	Remove(ctx context.Context, owner, id string) error
}
