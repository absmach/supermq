//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package logger_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"

	log "github.com/mainflux/mainflux/logger"
	"github.com/stretchr/testify/assert"
)

var _ io.Writer = (*mockWriter)(nil)

type mockWriter struct {
	value []byte
}

func (writer *mockWriter) Write(p []byte) (int, error) {
	writer.value = p
	return len(p), nil
}

func (writer *mockWriter) Read() (logMsg, error) {
	var output logMsg
	err := json.Unmarshal(writer.value, &output)
	return output, err
}

type logMsg struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func TestDebug(t *testing.T) {
	cases := map[string]struct {
		input  string
		output logMsg
	}{
		"info log ordinary string": {"input_string", logMsg{log.Debug.String(), "input_string"}},
		"info log empty string":    {"", logMsg{log.Debug.String(), ""}},
	}

	writer := mockWriter{}
	logger := log.New(&writer, log.Debug)

	for desc, tc := range cases {
		logger.Debug(tc.input)
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}
}

func TestInfo(t *testing.T) {
	cases := map[string]struct {
		input  string
		output logMsg
	}{
		"info log ordinary string": {"input_string", logMsg{log.Info.String(), "input_string"}},
		"info log empty string":    {"", logMsg{log.Info.String(), ""}},
	}

	writer := mockWriter{}
	logger := log.New(&writer, log.Info)

	for desc, tc := range cases {
		logger.Info(tc.input)
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}
}

func TestWarn(t *testing.T) {
	cases := map[string]struct {
		input  string
		output logMsg
	}{
		"warn log ordinary string": {"input_string", logMsg{log.Warn.String(), "input_string"}},
		"warn log empty string":    {"", logMsg{log.Warn.String(), ""}},
	}

	writer := mockWriter{}
	logger := log.New(&writer, log.Warn)

	for desc, tc := range cases {
		logger.Warn(tc.input)
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}
}

func TestError(t *testing.T) {
	cases := map[string]struct {
		input  string
		output logMsg
	}{
		"error log ordinary string": {"input_string", logMsg{log.Error.String(), "input_string"}},
		"error log empty string":    {"", logMsg{log.Error.String(), ""}},
	}

	writer := mockWriter{}
	logger := log.New(&writer, log.Error)

	for desc, tc := range cases {
		logger.Error(tc.input)
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}
}

func TestLogLevel(t *testing.T) {
	debugCases := map[string]struct {
		input  string
		output logMsg
	}{
		"debug log ordinary string": {"input_string", logMsg{"", ""}},
		"debug log empty string":    {"", logMsg{"", ""}},
	}

	infoCases := map[string]struct {
		input  string
		output logMsg
	}{
		"info log ordinary string": {"input_string", logMsg{"", ""}},
		"info log empty string":    {"", logMsg{"", ""}},
	}

	warnCases := map[string]struct {
		input  string
		output logMsg
	}{
		"warn log ordinary string": {"input_string", logMsg{log.Warn.String(), "input_string"}},
		"warn log empty string":    {"", logMsg{log.Warn.String(), ""}},
	}

	errorCases := map[string]struct {
		input  string
		output logMsg
	}{
		"error log ordinary string": {"input_string", logMsg{log.Error.String(), "input_string"}},
		"error log empty string":    {"", logMsg{log.Error.String(), ""}},
	}

	writer := mockWriter{}
	logger := log.New(&writer, log.Warn) //Set the log level to warn

	for desc, tc := range debugCases {
		logger.Debug(tc.input) //Try to log with Debug when warn is the level
		output, err := writer.Read()
		assert.Error(t, err, fmt.Sprint(`&json.SyntaxError{msg:"unexpected end of JSON input', Offset:0}`, desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}

	for desc, tc := range infoCases {
		logger.Info(tc.input) //Try to log with info when warn is the level
		output, err := writer.Read()
		assert.Error(t, err, fmt.Sprint(`&json.SyntaxError{msg:"unexpected end of JSON input', Offset:0}`, desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}

	for desc, tc := range warnCases {
		logger.Warn(tc.input) //Try to log with Warn to show we can still log with the selected level
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}

	for desc, tc := range errorCases {
		logger.Error(tc.input) //Try to log with Error to show we can still log with the higher levels
		output, err := writer.Read()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, output, fmt.Sprintf("%s: expected %s got %s", desc, tc.output, output))
	}
}
