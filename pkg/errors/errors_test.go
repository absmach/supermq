// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/stretchr/testify/assert"
)

const (
	msg1 = "message 1"
	msg2 = "message 2"
	msg3 = "message 3"
)

var (
	err1 = errors.New(msg1)
	err2 = errors.New(msg2)
	err3 = errors.New(msg3)
)

type customError struct {
	msg string
}

func (ce *customError) Error() string {
	return ce.msg
}

func TestWrap(t *testing.T) {
	level := 10
	err, msg := wrap(level)

	cases := []struct {
		desc string
		err  error
		msg  string
	}{
		{
			desc: "level 0 wrapped error",
			err:  err1,
			msg:  msg1,
		},
		{
			desc: "level 1 wrapped error",
			err:  errors.Wrap(err1, err2),
			msg:  fmt.Sprintf("%s : %s", msg1, msg2),
		},
		{
			desc: "level 2 wrapped error",
			err:  errors.Wrap(err1, errors.Wrap(err2, err3)),
			msg:  fmt.Sprintf("%s : %s : %s", msg1, msg2, msg3),
		},
		{
			desc: fmt.Sprintf("level %d wrapped error", level),
			err:  err,
			msg:  msg,
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.msg, tc.err.Error(), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, tc.err.Error()))
	}
}

func TestContains(t *testing.T) {
	level := 10
	err, _ := wrap(level)

	cases := []struct {
		desc      string
		container error
		contained error
	}{
		{
			desc:      "res of errors.Wrap(err1, err2) contains err1",
			container: errors.Wrap(err1, err2),
			contained: err1,
		},
		{
			desc:      "res of errors.Wrap(err1, err2) contains err2",
			container: errors.Wrap(err1, err2),
			contained: err2,
		},
		{
			desc:      fmt.Sprintf("level %d wrapped error contains", level),
			container: err,
			contained: errors.New(""),
		},
	}
	for _, tc := range cases {
		contains := errors.Contains(tc.container, tc.contained)
		cntStr := strconv.FormatBool(contains)
		assert.Equal(t, true, contains, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, "true", cntStr))
	}

}

func wrap(level int) (error, string) {
	msg := "0"
	err := error(&customError{msg: msg})
	for i := 1; i < level; i++ {
		err = errors.Wrap(errors.New(err.Error()), errors.New(strconv.Itoa(i)))
		msg += fmt.Sprintf(" : %s", strconv.Itoa(i))
	}
	// msg == "0 : 1 : 2 : 3 : 4 : 5 : 6 : 7 : 8 : 9 : ... : level - 1"
	return err, msg
}
