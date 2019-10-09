//
// Copyright (c) Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package errors

import "fmt"

type Error interface {
	Error() string
	Msg() string
	Contains(error) bool
	IsEmpty() bool
}

var _ Error = (*customError)(nil)

// customError struct represents a Mainflux error
type customError struct {
	msg string
	err Error
}

// Error implements the error interface.
func (err customError) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %s", err.msg, err.err.Error())
	}

	return err.msg
}

// Msg returns error message
func (err customError) Msg() string {
	return err.msg
}

func (err customError) IsEmpty() bool {
	if err.Msg() == "" {
		return true
	}
	return false
}

// Contains inspects if Error's message is same as error
// in argument. If not it continues to examin in next
// layers of Error until it founds it or unwrap every layers
func (err customError) Contains(e error) bool {
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
		return New("")
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
