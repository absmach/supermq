// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import "errors"

type NestError interface {
	Error
	Embed(e error) error
}

var _ NestError = (*customError)(nil)

func (e *customError) Embed(err error) error {
	e.err = errors.Join(err, e.err)
	return e
}

type RequestError struct {
	customError
}

var _ NestError = (*RequestError)(nil)

func NewRequestError(message string) NestError {
	return &RequestError{
		customError: customError{
			msg: message,
		},
	}
}

func NewRequestErrorWithErr(message string, err error) NestError {
	return &RequestError{
		customError: customError{
			msg: message,
			err: err,
		},
	}
}

func (e *RequestError) Embed(err error) error {
	e.customError.Embed(err)
	return e
}

type AuthNError struct {
	customError
}

var _ NestError = (*AuthNError)(nil)

func NewAuthNError(message string) NestError {
	return &AuthNError{
		customError: customError{
			msg: message,
		},
	}
}

func NewAuthNErrorWithErr(message string, err error) NestError {
	return &AuthNError{
		customError: customError{
			msg: message,
			err: err,
		},
	}
}

func (e *AuthNError) Embed(err error) error {
	e.customError.Embed(err)
	return e
}

var _ NestError = (*AuthZError)(nil)

type AuthZError struct {
	customError
}

func (e *AuthZError) Embed(err error) error {
	e.customError.Embed(err)
	return e
}

func NewAuthZError(message string) NestError {
	return &AuthZError{
		customError: customError{
			msg: message,
		},
	}
}

func NewAuthZErrorWithErr(message string, err error) NestError {
	return &AuthZError{
		customError: customError{
			msg: message,
			err: cast(err),
		},
	}
}

type InternalError struct {
	customError
}

var _ NestError = (*InternalError)(nil)

func NewInternalError() error {
	return &InternalError{
		customError: customError{
			msg: "internal server error",
		},
	}
}

func NewInternalErrorWithErr(err error) NestError {
	return &InternalError{
		customError: customError{
			msg: "internal server error",
			err: cast(err),
		},
	}
}

type ServiceError struct {
	customError
}

var _ NestError = (*ServiceError)(nil)

func NewServiceError(message string) NestError {
	return &ServiceError{
		customError: customError{
			msg: message,
		},
	}
}

func NewServiceErrorWithErr(message string, err error) NestError {
	return &ServiceError{
		customError: customError{
			msg: message,
			err: cast(err),
		},
	}
}

func (e *ServiceError) Embed(err error) error {
	e.customError.Embed(err)
	return e
}
