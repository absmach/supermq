// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"errors"
	"fmt"
)

type NestError interface {
	Error
	Embed(e error) error
}

var _ NestError = (*customError)(nil)

func (e *customError) Embed(err error) error {
	if err == nil {
		return e
	}

	return &customError{
		msg: e.msg,
		err: fmt.Errorf("%w: %w", e.err, err),
	}
}

type RequestError struct {
	customError
}

var _ NestError = (*RequestError)(nil)

func NewRequestError(message string) NestError {
	return &RequestError{
		customError: customError{
			msg: message,
			err: errors.New(message),
		},
	}
}

func NewRequestErrorWithErr(message string, err error) NestError {
	return &RequestError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
		},
	}
}

func (e *RequestError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &RequestError{
		customError: *embedded.(*customError),
	}
}

type AuthNError struct {
	customError
}

var _ NestError = (*AuthNError)(nil)

func NewAuthNError(message string) NestError {
	return &AuthNError{
		customError: customError{
			msg: message,
			err: errors.New(message),
		},
	}
}

func NewAuthNErrorWithErr(message string, err error) NestError {
	return &AuthNError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
		},
	}
}

func (e *AuthNError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &AuthNError{
		customError: *embedded.(*customError),
	}
}

var _ NestError = (*AuthZError)(nil)

type AuthZError struct {
	customError
}

func (e *AuthZError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &AuthZError{
		customError: *embedded.(*customError),
	}
}

func NewAuthZError(message string) NestError {
	return &AuthZError{
		customError: customError{
			msg: message,
			err: errors.New(message),
		},
	}
}

func NewAuthZErrorWithErr(message string, err error) NestError {
	return &AuthZError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
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
			err: errors.New("internal server error"),
		},
	}
}

func NewInternalErrorWithErr(err error) NestError {
	return &InternalError{
		customError: customError{
			msg: "internal server error",
			err: fmt.Errorf("%w: %w", errors.New("internal server error"), err),
		},
	}
}

func (e *InternalError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &InternalError{
		customError: *embedded.(*customError),
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
			err: errors.New(message),
		},
	}
}

func NewServiceErrorWithErr(message string, err error) NestError {
	return &ServiceError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
		},
	}
}

func (e *ServiceError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &ServiceError{
		customError: *embedded.(*customError),
	}
}

type MediaTypeError struct {
	customError
}

var _ NestError = (*MediaTypeError)(nil)

func NewMediaTypeError(message string) NestError {
	return &MediaTypeError{
		customError: customError{
			msg: message,
			err: errors.New(message),
		},
	}
}

func NewMediaTypeErrorWithErr(message string, err error) NestError {
	return &MediaTypeError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
		},
	}
}

func (e *MediaTypeError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &MediaTypeError{
		customError: *embedded.(*customError),
	}
}

type NotFoundError struct {
	customError
}

var _ NestError = (*NotFoundError)(nil)

func NewNotFoundError(message string) NestError {
	return &NotFoundError{
		customError: customError{
			msg: message,
			err: errors.New(message),
		},
	}
}

func NewNotFoundErrorWithErr(message string, err error) NestError {
	return &NotFoundError{
		customError: customError{
			msg: message,
			err: fmt.Errorf("%w: %w", errors.New(message), err),
		},
	}
}

func (e *NotFoundError) Embed(err error) error {
	embedded := e.customError.Embed(err)
	return &NotFoundError{
		customError: *embedded.(*customError),
	}
}
