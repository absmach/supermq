// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"
)

const (
	LevelTrace slog.Level = slog.LevelDebug - 4

	traceStr      = "TRACE"
	traceLevelStr = "DEBUG-4"
)

var customLevelNames = map[slog.Leveler]string{
	LevelTrace: traceStr,
}

// New returns wrapped slog logger.
func New(w io.Writer, levelText string) (*slog.Logger, error) {
	var level slog.Level
	levelUpper := strings.ToUpper(levelText)

	switch levelUpper {
	case traceStr, traceLevelStr:
		level = LevelTrace
	default:
		if err := level.UnmarshalText([]byte(levelUpper)); err != nil {
			return &slog.Logger{}, fmt.Errorf(`{"level":"error","message":"invalid log level %s: %s","ts":"%s"}`, levelText, err, time.RFC3339Nano)
		}
	}

	logHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.LevelKey {
				lvl := a.Value.Any().(slog.Level)
				if name, ok := customLevelNames[lvl]; ok {
					return slog.Attr{Key: slog.LevelKey, Value: slog.StringValue(name)}
				}
			}

			return a
		},
	})

	return slog.New(logHandler), nil
}
