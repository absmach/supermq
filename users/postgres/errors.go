package postgres

import (
	"github.com/absmach/supermq/pkg/errors"
)

var _ errors.Mapper = (*duplicateErrors)(nil)

type duplicateErrors struct{}

func (d duplicateErrors) GetError(key string) (error, bool) {
	switch key {
	case "clients_email_key":
		return errors.NewRequestError("email id already exists"), true
	case "clients_username_key":
		return errors.NewRequestError("username not available"), true
	default:
		return nil, false
	}
}

func NewDuplicateErrors() errors.Mapper {
	return duplicateErrors{}
}
