// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"fmt"

	"github.com/absmach/supermq/pkg/errors"
	repoerr "github.com/absmach/supermq/pkg/errors/repository"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ errors.Handler = (*errHandler)(nil)

type errHandler struct {
	duplicateErrors errors.Mapper
}

func WithDuplicateErrors(mapper errors.Mapper) errors.HandlerOption {
	return func(eh *errors.Handler) {
		if h, ok := (*eh).(*errHandler); ok {
			h.duplicateErrors = mapper
		}
	}
}

func NewErrorHandler(opts ...errors.HandlerOption) errors.Handler {
	var eh errors.Handler = &errHandler{}
	for _, opt := range opts {
		opt(&eh)
	}
	return eh
}

// Handle handles the error.
func (eh errHandler) HandleError(wrapper, err error) error {
	pqErr, ok := err.(*pgconn.PgError)
	if ok {
		switch pqErr.Code {
		case errDuplicate:
			if knownErr, ok := eh.duplicateErrors.GetError(pqErr.ConstraintName); ok {
				fmt.Printf("knownErr type %T\n", knownErr)
				fmt.Printf("wrapper Error type %T\n", errors.Wrap(wrapper, knownErr))
				return errors.Wrap(wrapper, knownErr)
			}
			return errors.Wrap(repoerr.ErrConflict, err)
		case errInvalid, errInvalidChar, errTruncation, errUntranslatable:
			return errors.Wrap(repoerr.ErrMalformedEntity, err)
		case errFK:
			return errors.Wrap(repoerr.ErrCreateEntity, err)
		}
	}

	return errors.Wrap(wrapper, err)
}
