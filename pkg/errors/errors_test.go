// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package errors_test

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/stretchr/testify/assert"
)

const level = 10

func TestError(t *testing.T) {
	cases := []struct {
		desc string
		err  error
		msg  string
	}{
		{
			desc: "level 0 wrapped error",
			err:  errors.New("0"),
			msg:  "0",
		},
		{
			desc: "level 1 wrapped error",
			err:  wrap(1),
			msg:  message(1),
		},
		{
			desc: "level 2 wrapped error",
			err:  wrap(2),
			msg:  message(2),
		},
		{
			desc: fmt.Sprintf("level %d wrapped error", level),
			err:  wrap(level),
			msg:  message(level),
		},
	}

	for _, tc := range cases {
		errMsg := tc.err.Error()
		assert.Equal(t, tc.msg, errMsg, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.msg, errMsg))
	}
}

func TestContains(t *testing.T) {
	err0 := errors.New("0")
	err1 := errors.New("1")
	err2 := errors.New("2")

	cases := []struct {
		desc      string
		container error
		contained error
		contains  bool
	}{
		{
			desc:      "nil contains nil",
			container: nil,
			contained: nil,
			contains:  true,
		},
		{
			desc:      "nil contains non-nil",
			container: nil,
			contained: err0,
			contains:  false,
		},
		{
			desc:      "non-nil contains nil",
			container: err0,
			contained: nil,
			contains:  false,
		},
		{
			desc:      "non-nil contains non-nil",
			container: err0,
			contained: err1,
			contains:  false,
		},
		{
			desc:      "res of errors.Wrap(err1, err0) contains err0",
			container: errors.Wrap(err1, err0),
			contained: err0,
			contains:  true,
		},
		{
			desc:      "res of errors.Wrap(err1, err0) contains err1",
			container: errors.Wrap(err1, err0),
			contained: err1,
			contains:  true,
		},
		{
			desc:      "res of errors.Wrap(err2, errors.Wrap(err1, err0)) contains err1",
			container: errors.Wrap(err2, errors.Wrap(err1, err0)),
			contained: err1,
			contains:  true,
		},
		{
			desc:      fmt.Sprintf("level %d wrapped error contains", level),
			container: wrap(level),
			contained: errors.New(strconv.Itoa(level / 2)),
			contains:  true,
		},
		{
			desc:      "wrapped error containes wrapped error",
			container: wrap(level),
			contained: wrap(level / 2),
			contains:  false,
		},
	}
	for _, tc := range cases {
		contains := errors.Contains(tc.container, tc.contained)
		assert.Equal(t, tc.contains, contains, fmt.Sprintf("%s: expected %v to contain %v\n", tc.desc, tc.container, tc.contained))
	}

}

func TestWrap(t *testing.T) {
	cases := []struct {
		desc  string
		level int
	}{
		{
			desc:  "level 1 wrap",
			level: 1,
		},
		{
			desc:  "level 5 wrap",
			level: 5,
		},
		{
			desc:  "level 10 wrap",
			level: 10,
		},
	}

	for _, tc := range cases {
		err := wrap(tc.level)
		msg := message(tc.level)
		errMsg := err.Error()
		assert.Equal(t, msg, errMsg, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, msg, errMsg))
		contained := errors.New(strconv.Itoa(rand.Intn(tc.level)))
		contains := errors.Contains(err, contained)
		assert.Equal(t, true, contains, fmt.Sprintf("%s: expected %v to contain %v\n", tc.desc, err, contained))
	}
}

func wrap(level int) error {
	if level == 0 {
		return errors.New(strconv.Itoa(level))
	}
	return errors.Wrap(errors.New(strconv.Itoa(level)), wrap(level-1))
}

// message generates error message of wrap() generated wrapper error
func message(level int) string {
	if level == 0 {
		return "0"
	}
	return strconv.Itoa(level) + " : " + message(level-1)
}
