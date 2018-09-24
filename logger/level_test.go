package logger

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestUnmarshalText(t *testing.T) {
	invalidCases := map[string]struct {
		input  string
		output Level
	}{
		"select log level Not_A_Level": {"Not_A_Level", 0},
		"select log level Bad_Input":   {"Bad_Input", 0},
	}

	validCases := map[string]struct {
		input  string
		output Level
	}{
		"select log level debug": {"debug", Debug},
		"select log level DEBUG": {"DEBUG", Debug},
		"select log level info":  {"info", Info},
		"select log level INFO":  {"INFO", Info},
		"select log level warn":  {"warn", Warn},
		"select log level WARN":  {"WARN", Warn},
		"select log level Error": {"Error", Error},
		"select log level ERROR": {"ERROR", Error},
	}

	var logLevel Level
	for desc, tc := range invalidCases {
		err := logLevel.UnmarshalText(tc.input)
		assert.Error(t, err, ErrInvalidLogLevel.Error(), desc, err)
		assert.Equal(t, tc.output, logLevel, fmt.Sprintf("%s: expected %s got %d", desc, tc.output, logLevel))
	}

	for desc, tc := range validCases {
		err := logLevel.UnmarshalText(tc.input)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))
		assert.Equal(t, tc.output, logLevel, fmt.Sprintf("%s: expected %s got %d", desc, tc.output, logLevel))
	}
}

func TestLevelIsAllowed(t *testing.T) {
	cases := map[string]struct {
		requestedLevel Level
		allowedLevel   Level
		output         bool
	}{
		"log debug when level debug": {Debug, Debug, true},
		"log info when level debug":  {Info, Debug, true},
		"log warn when level debug":  {Warn, Debug, true},
		"log error when level debug": {Error, Debug, true},
		"log warn when level info":   {Warn, Info, true},
		"log error when level warn":  {Error, Warn, true},
		"log error when level error": {Error, Error, true},

		"log debug when level error": {Debug, Error, false},
		"log info when level error":  {Info, Error, false},
		"log warn when level error":  {Warn, Error, false},
		"log debug when level warn":  {Debug, Warn, false},
		"log info when level warn":   {Info, Warn, false},
		"log debug when level info":  {Debug, Info, false},
	}
	for desc, tc := range cases {
		result := tc.requestedLevel.isAllowed(tc.allowedLevel)
		assert.Equal(t, tc.output, result, fmt.Sprintf("%s: expected %t got %t", desc, tc.output, result))
	}
}
