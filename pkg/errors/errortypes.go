// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"
)

type NewError interface {
	// Error implements the error interface.
	Error() string

	// Msg returns error message.
	Msg() string

	// Err returns wrapped error.
	Unwrap() error

	Wrap(e error) error

	// MarshalJSON returns a marshaled error.
	MarshalJSON() ([]byte, error)
}

// NewError specifies an that request could be processed and error which should be addressed by user.
type newError struct {
	Err     error  // Contains other internal details and errors as wrapped error
	Message string // Message for end users returned by API layer or other end layer
}

func (e newError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + " : " + e.Err.Error()
}

func (e newError) Unwrap() error {
	return e.Err
}

func (e newError) Msg() string {
	return e.Message
}

func (e newError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Err string `json:"error"`
	}{
		Err: e.Message,
	})
}

var _ NewError = (*RequestError)(nil)

type RequestError struct {
	newError
}

func (e RequestError) Wrap(err error) error {
	e.Err = errors.Join(err, e.Err)
	return e
}

func NewRequestError(message string) error {
	return RequestError{
		newError: newError{
			Message: message,
		},
	}
}

func NewRequestErrorWithErr(message string, err error) error {
	return RequestError{
		newError: newError{
			Message: message,
			Err:     err,
		},
	}
}

var _ NewError = (*AuthNError)(nil)

type AuthNError struct {
	newError
}

func (e AuthNError) Wrap(err error) error {
	e.Err = errors.Join(err, e.Err)
	return e
}

func NewAuthNError(message string) error {
	return AuthNError{
		newError: newError{
			Message: message,
		},
	}
}

func NewAuthNErrorWithErr(message string, err error) error {
	return AuthNError{
		newError: newError{
			Message: message,
			Err:     err,
		},
	}
}

var _ NewError = (*AuthZError)(nil)

type AuthZError struct {
	newError
}

func (e AuthZError) Wrap(err error) error {
	e.Err = errors.Join(err, e.Err)
	return e
}

func NewAuthZError(message string) error {
	return AuthZError{
		newError: newError{
			Message: message,
		},
	}
}

func NewAuthZErrorWithErr(message string, err error) error {
	return AuthZError{
		newError: newError{
			Message: message,
			Err:     err,
		},
	}
}
