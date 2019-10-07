//
// Copyright (c) Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package errors

import (
	"fmt"
)

// Error represents a Mainflux error.
type Error struct {
	msg string
	err *Error
}

// Error implements the error interface.
func (err Error) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %s", err.msg, err.err.Error())
	}

	return err.msg
}

// Msg returns error message
func (err Error) Msg() string {
	return err.msg
}

func (err Error) Contains(e error) bool {
	if e == nil {
		return false
	}

	if err.msg == e.Error() {
		return true
	}
	if err.err == nil {
		return false
	}
	return err.err.Contains(e)
}

func Wrap(wrapper Error, err *Error) Error {
	return Error{
		msg: wrapper.msg,
		err: err,
	}
}

// Cast returns pointer to Error type with message of given error
func Cast(err error) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		msg: err.Error(),
	}
}

// New returns an Error that formats as the given text.
func New(text string) Error {
	return Error{
		msg: text,
		err: nil,
	}
}
