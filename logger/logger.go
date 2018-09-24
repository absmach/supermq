//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package logger

import (
	"github.com/go-kit/kit/log"
	"io"
)

// Logger specifies logging API.
type Logger interface {
	// Debug logs any object in JSON format on debug level.
	Debug(string)
	// Info logs any object in JSON format on info level.
	Info(string)
	// Warn logs any object in JSON format on warning level.
	Warn(string)
	// Error logs any object in JSON format on error level.
	Error(string)
}

var _ Logger = (*logger)(nil)

type logger struct {
	kitLogger log.Logger
	level     Level
}

// New returns wrapped go kit logger.
func New(out io.Writer, level Level) Logger {
	l := log.NewJSONLogger(log.NewSyncWriter(out))
	l = log.With(l, "ts", log.DefaultTimestampUTC)
	return &logger{l, level}
}

func (l logger) Debug(msg string) {
	if Debug.isAllowed(l.level) {
		l.kitLogger.Log("level", Debug.String(), "message", msg)
	}
}

func (l logger) Info(msg string) {
	if Info.isAllowed(l.level) {
		l.kitLogger.Log("level", Info.String(), "message", msg)
	}
}

func (l logger) Warn(msg string) {
	if Warn.isAllowed(l.level) {
		l.kitLogger.Log("level", Warn.String(), "message", msg)
	}
}

func (l logger) Error(msg string) {
	if Error.isAllowed(l.level) {
		l.kitLogger.Log("level", Error.String(), "message", msg)
	}
}
