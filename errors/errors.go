// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors

import "fmt"

// Error specifies an API that must be fullfiled by error type
type Error interface {

	// Error implements the error interface.
	Error() string

	// Msg returns error message
	Msg() string

	// Contains inspects if Error's message is same as error
	// in argument. If not it continues to examin in next
	// layers of Error until it founds it or unwrap every layers
	Contains(error) bool

	//IsEmpty check if Error is empty
	IsEmpty() bool
}

var _ Error = (*customError)(nil)

// customError struct represents a Mainflux error
type customError struct {
	msg string
	err Error
}

func (ce customError) Error() string {
	if ce.err != nil {
		return fmt.Sprintf("%s: %s", ce.msg, ce.err.Error())
	}

	return ce.msg
}

func (ce customError) Msg() string {
	return ce.msg
}

func (ce customError) IsEmpty() bool {
	if ce.Msg() == "" {
		return true
	}
	return false
}

func (ce customError) Contains(e error) bool {
	if e == nil {
		return false
	}
	if ce.msg == e.Error() {
		return true
	}
	if ce.err == nil {
		return false
	}

	return ce.err.Contains(e)
}

// Wrap returns an Error that wrap err with wrapper
func Wrap(wrapper Error, err Error) Error {
	return customError{
		msg: wrapper.Msg(),
		err: err,
	}
}

// Cast returns Error type with message of given error
func Cast(err error) Error {
	if err == nil {
		return Empty()
	}
	return New(err.Error())
}

// New returns an Error that formats as the given text.
func New(text string) Error {
	return customError{
		msg: text,
		err: nil,
	}
}

// Empty returns a new empty Error
func Empty() Error {
	return New("")
}
