//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package logger

import (
	"io"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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

var logLevels = map[string]level.Option{
	"debug": level.AllowDebug(),
	"info":  level.AllowInfo(),
	"warn":  level.AllowWarn(),
	"error": level.AllowError(),
}

var _ Logger = (*logger)(nil)

type logger struct {
	kitLogger log.Logger
}

// New returns wrapped go kit logger.
func New(out io.Writer, logLevel string) Logger {
	l := log.NewJSONLogger(log.NewSyncWriter(out))
	if allowedLevel, ok := logLevels[strings.ToLower(logLevel)]; ok {
		l = level.NewFilter(l, allowedLevel)
	} else {
		l = level.NewFilter(l, level.AllowInfo())
	}
	l = log.With(l, "ts", log.DefaultTimestampUTC)
	return &logger{l}
}

func (l logger) Debug(msg string) {
	level.Debug(l.kitLogger).Log("level", level.DebugValue().String(), "message", msg)
}

func (l logger) Info(msg string) {
	level.Info(l.kitLogger).Log("level", level.InfoValue().String(), "message", msg)
}

func (l logger) Warn(msg string) {
	level.Warn(l.kitLogger).Log("level", level.WarnValue().String(), "message", msg)
}

func (l logger) Error(msg string) {
	level.Error(l.kitLogger).Log("level", level.ErrorValue().String(), "message", msg)

}
