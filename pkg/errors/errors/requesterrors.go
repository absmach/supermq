// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"
)

type ReturnError interface {
	Error() string
	Wrap(e error) ReturnError
	Unwrap() error
	MarshalJSON() ([]byte, error)
}

type returnError struct {
	err     error
	message string
}

func (e returnError) Error() string {
	if e.err == nil {
		return e.message
	}
	return e.message + " : " + e.err.Error()
}

func (e returnError) Wrap(err error) ReturnError {
	if err != nil {
		e.err = errors.Join(err, e.err)
	}
	return e
}

func (e returnError) Unwrap() error {
	return e.err
}

func (e returnError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Err string `json:"error"`
	}{
		Err: e.message,
	})
}

var _ ReturnError = (*RequestError)(nil)

type RequestError struct {
	returnError
}

func NewRequestError(message string) RequestError {
	return RequestError{
		returnError: returnError{
			message: message,
		},
	}
}

func NewRequestErrorWithErr(message string, err error) RequestError {
	return RequestError{
		returnError: returnError{
			err:     err,
			message: message,
		},
	}
}

var _ ReturnError = (*AuthNError)(nil)

type AuthNError struct {
	returnError
}

func NewAuthNError(message string) AuthNError {
	return AuthNError{
		returnError: returnError{
			message: message,
		},
	}
}

func NewAuthNErrorWithErr(message string, err error) AuthNError {
	return AuthNError{
		returnError: returnError{
			err:     err,
			message: message,
		},
	}
}

var _ ReturnError = (*AuthZError)(nil)

type AuthZError struct {
	returnError
}

func NewAuthZError(message string) AuthZError {
	return AuthZError{
		returnError: returnError{
			message: message,
		},
	}
}

func NewAuthZErrorWithErr(message string, err error) AuthZError {
	return AuthZError{
		returnError: returnError{
			err:     err,
			message: message,
		},
	}
}
