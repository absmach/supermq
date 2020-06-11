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
	msg1  = "message 1"
	msg2  = "message 2"
	msg3  = "message 3"
	level = 10
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
			err:  wrap(level),
			msg:  message(level),
		},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.msg, tc.err.Error(), fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, tc.err.Error()))
	}
}

func TestContains(t *testing.T) {
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
			desc:      "res of errors.Wrap(err1, errors.Wrap(err2, err3)) contains err2",
			container: errors.Wrap(err1, errors.Wrap(err2, err3)),
			contained: err2,
		},
		{
			desc:      fmt.Sprintf("level %d wrapped error contains", level),
			container: wrap(level),
			contained: errors.New(strconv.Itoa(level / 2)),
		},
	}
	for _, tc := range cases {
		contains := errors.Contains(tc.container, tc.contained)
		cntStr := strconv.FormatBool(contains)
		assert.Equal(t, true, contains, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, "true", cntStr))
	}

}

func wrap(n int) error {
	if n < 1 {
		return errors.New(strconv.Itoa(n))
	}
	return errors.Wrap(errors.New(strconv.Itoa(n)), wrap(n-1))
}

func message(level int) string {
	msg := strconv.Itoa(level)
	for i := level - 1; i >= 0; i-- {
		msg = fmt.Sprintf("%s : %s", msg, strconv.Itoa(i))
	}
	return msg
}
