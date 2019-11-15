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

	// Err returns wrapped error
	Err() *customError

	// Contains inspects if Error's message is same as error
	// in argument. If not it continues to examin in next
	// layers of Error until it founds it or unwrap every layers
}

var _ Error = (*customError)(nil)

// customError struct represents a Mainflux error
type customError struct {
	msg string
	err *customError
}

func (ce *customError) Error() string {
	if ce.err != nil {
		return fmt.Sprintf("%s: %s", ce.msg, ce.err.Error())
	}

	return ce.msg
}

func (ce *customError) Msg() string {
	return ce.msg
}

func (ce *customError) Err() *customError {
	return ce.err
}

// Contains contains
func Contains(ce Error, e error) bool {
	// if ce != nil {
	if e == nil {
		// return false
		return ce == nil
	}
	if ce.Msg() == e.Error() {
		return true
	}
	if ce.Err() == nil {
		return false
	}

	return Contains(ce.Err(), e)
	// }
	// return e == nil
}

// Wrap returns an Error that wrap err with wrapper
func Wrap(wrapper Error, err Error) Error {
	return &customError{
		msg: wrapper.Msg(),
		err: err.(*customError),
	}
}

// Cast returns Error type with message of given error
func Cast(err error) Error {
	if err == nil {
		// return Empty()
		return nil
	}
	return New(err.Error())
}

// New returns an Error that formats as the given text.
func New(text string) Error {
	return &customError{
		msg: text,
		err: nil,
	}
}
