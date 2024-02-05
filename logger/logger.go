// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package logger

import (
	"fmt"
	"io"
	"log/slog"
	"time"
)

// New returns wrapped slog logger
// if handler is not set slog.JsonHandler is used.
func New(w io.Writer, levelText string, handler slog.Handler) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(levelText)); err != nil {
		return &slog.Logger{}, fmt.Errorf(`{"level":"error","message":"%s: %s","ts":"%s"}`, err, levelText, time.RFC3339Nano)
	}

	if handler == nil {
		handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
			Level: level,
		})
	}

	return slog.New(handler), nil
}
