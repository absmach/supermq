package errors

import (
	"fmt"
)

type Error struct {
	msg string
	err *Error
}

func (err Error) Error() string {
	if err.err != nil {
		return fmt.Sprintf("%s: %s", err.msg, err.err.Error())
	}

	return err.msg
}

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

func Cast(err error) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		msg: err.Error(),
	}
}

func New(text string) Error {
	return Error{
		msg: text,
		err: nil,
	}
}
