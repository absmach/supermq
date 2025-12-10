// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import "errors"

type TypedError interface {
	Error
	Wrap(e error) error
}

var _ TypedError = (*RequestError)(nil)

type RequestError struct {
	customError
}

func wrap(wrapper, err error) Error {
	if wrapper == nil || err == nil {
		return &customError{
			msg: wrapper.Error(),
			err: wrapper,
		}
	}
	return &customError{
		msg: wrapper.Error(),
		err: errors.Join(wrapper, err),
	}
}

func (e *RequestError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &RequestError{
		customError: e.customError,
	}
}

func NewRequestError(message string) error {
	return &RequestError{
		customError: customError{
			msg: message,
		},
	}
}

func NewRequestErrorWithErr(message string, err error) error {
	return &RequestError{
		customError: customError{
			msg: message,
			err: err,
		},
	}
}

var _ TypedError = (*AuthNError)(nil)

type AuthNError struct {
	customError
}

func (e *AuthNError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &AuthNError{
		customError: e.customError,
	}
}

func NewAuthNError(message string) error {
	return &AuthNError{
		customError: customError{
			msg: message,
		},
	}
}

func NewAuthNErrorWithErr(message string, err error) error {
	return &AuthNError{
		customError: customError{
			msg: message,
			err: err,
		},
	}
}

var _ TypedError = (*AuthZError)(nil)

type AuthZError struct {
	customError
}

func (e *AuthZError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &AuthZError{
		customError: e.customError,
	}
}

func NewAuthZError(message string) error {
	return &AuthZError{
		customError: customError{
			msg: message,
		},
	}
}

func NewAuthZErrorWithErr(message string, err error) error {
	return &AuthZError{
		customError: customError{
			msg: message,
			err: cast(err),
		},
	}
}

var _ TypedError = (*InternalError)(nil)

type InternalError struct {
	customError
}

func (e *InternalError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &InternalError{
		customError: e.customError,
	}
}

func NewInternalError() error {
	return &InternalError{
		customError: customError{
			msg: "internal server error",
		},
	}
}

func NewInternalErrorWithErr(err error) error {
	return &InternalError{
		customError: customError{
			msg: "internal server error",
			err: cast(err),
		},
	}
}

type ConflictError struct {
	customError
}

func (e *ConflictError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &ConflictError{
		customError: e.customError,
	}
}

func NewConflictError(message string) error {
	return &ConflictError{
		customError: customError{
			msg: message,
		},
	}
}

func NewConflictErrorWithErr(message string, err error) error {
	return &ConflictError{
		customError: customError{
			msg: message,
			err: cast(err),
		},
	}
}

type ServiceError struct {
	customError
}

func (e *ServiceError) Wrap(err error) error {
	e.err = wrap(err, e.err)
	return &ServiceError{
		customError: e.customError,
	}
}

func NewServiceError(message string) error {
	return &ServiceError{
		customError: customError{
			msg: message,
		},
	}
}

func NewServiceErrorWithErr(message string, err error) error {
	return &ServiceError{
		customError: customError{
			msg: message,
			err: err,
		},
	}
}
