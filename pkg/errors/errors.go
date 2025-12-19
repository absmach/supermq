// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error specifies an API that must be fullfiled by error type.
type Error interface {
	// Error implements the error interface.
	Error() string

	// Msg returns error message.
	Msg() string

	// Err returns wrapped error.
	Err() error

	// MarshalJSON returns a marshaled error.
	MarshalJSON() ([]byte, error)
}

var _ Error = (*customError)(nil)

// customError represents a SuperMQ error.
type customError struct {
	msg string
	err error
}

// New returns an Error that formats as the given text.
func New(text string) Error {
	return &customError{
		msg: text,
		err: errors.New(text),
	}
}

func (ce *customError) Error() string {
	if ce == nil {
		return ""
	}
	if ce.err == nil {
		return ce.msg
	}
	return ce.err.Error()
}

func (ce *customError) Msg() string {
	return ce.msg
}

func (ce *customError) Err() error {
	return ce.err
}

func (ce *customError) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Msg string `json:"message"`
	}{
		Msg: ce.Msg(),
	})
}

func Contains(e1, e2 error) bool {
	if e1 == nil || e2 == nil {
		return e2 == e1
	}
	ce, ok := e1.(Error)
	if ok {
		if ce.Msg() == e2.Error() {
			return true
		}
		return Contains(ce.Err(), e2)
	}

	return errors.Is(e1, e2) || e1.Error() == e2.Error()
}

// Wrap returns an Error that wrap err with wrapper.
func Wrap(wrapper, err error) error {
	if wrapper == nil || err == nil {
		return wrapper
	}
	if ne, ok := err.(NestError); ok {
		return ne.Embed(wrapper)
	}
	if ce, ok := wrapper.(NestError); ok {
		return ce.Embed(err)
	}
	return &customError{
		msg: wrapper.Error(),
		err: fmt.Errorf("%w: %w", wrapper, err),
	}
}

func cast(err error) Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(Error); ok {
		return e
	}
	return &customError{
		msg: err.Error(),
		err: nil,
	}
}
