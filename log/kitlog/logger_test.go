package kitlog_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mainflux/mainflux/log"
	"github.com/mainflux/mainflux/log/kitlog"
	"github.com/stretchr/testify/assert"
)

func TestInfo(t *testing.T) {
	cases := map[string]struct {
		input interface{}
	}{
		"info log ordinary string": {"input_string"},
		"info log some object":     {map[string]interface{}{"field": "val"}},
	}

	out := bytes.NewBufferString("")
	logger := kitlog.New(out)

	for desc, tc := range cases {
		out.Reset()
		logger.Info(tc.input)

		var output map[string]interface{}
		err := json.Unmarshal(out.Bytes(), &output)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))

		val := output[log.Info.String()]
		assert.EqualValues(t, tc.input, val, fmt.Sprintf("%s: unexpected value %s instead of %s", desc, val, tc.input))
	}
}

func TestWarn(t *testing.T) {
	cases := map[string]struct {
		input interface{}
	}{
		"warn log ordinary string": {"input_string"},
		"warn log some object":     {map[string]interface{}{"field": "val"}},
	}

	out := bytes.NewBufferString("")
	logger := kitlog.New(out)

	for desc, tc := range cases {
		out.Reset()
		logger.Warn(tc.input)

		var output map[string]interface{}
		err := json.Unmarshal(out.Bytes(), &output)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))

		val := output[log.Warn.String()]
		assert.EqualValues(t, tc.input, val, fmt.Sprintf("%s: unexpected value %s instead of %s", desc, val, tc.input))
	}
}

func TestError(t *testing.T) {
	cases := map[string]struct {
		input interface{}
	}{
		"error log ordinary string": {"input_string"},
		"error log some object":     {map[string]interface{}{"field": "val"}},
	}

	out := bytes.NewBufferString("")
	logger := kitlog.New(out)

	for desc, tc := range cases {
		out.Reset()
		logger.Error(tc.input)

		var output map[string]interface{}
		err := json.Unmarshal(out.Bytes(), &output)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))

		val := output[log.Error.String()]
		assert.EqualValues(t, tc.input, val, fmt.Sprintf("%s: unexpected value %s instead of %s", desc, val, tc.input))
	}
}

func TestLog(t *testing.T) {
	const customLvl log.Level = 4
	log.AddLevel(customLvl, "Custom")

	cases := map[string]struct {
		lvl   log.Level
		input interface{}
	}{
		"info log ordenary string":       {log.Info, "input_string"},
		"warn log some object":           {log.Warn, map[string]interface{}{"field": "val"}},
		"error log ordinary string":      {log.Error, "error_desc"},
		"custom lvl log ordinary string": {customLvl, "custom_lvl_log"},
	}

	out := bytes.NewBufferString("")
	logger := kitlog.New(out)

	for desc, tc := range cases {
		out.Reset()
		logger.Log(tc.lvl, tc.input)

		var output map[string]interface{}
		err := json.Unmarshal(out.Bytes(), &output)
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", desc, err))

		val := output[tc.lvl.String()]
		assert.EqualValues(t, tc.input, val, fmt.Sprintf("%s: unexpected value %s instead of %s", desc, val, tc.input))
	}
}

func TestSetLevel(t *testing.T) {
	const customLvl log.Level = 4
	log.AddLevel(customLvl, "Custom")

	cases := map[string]struct {
		filterLvl log.Level
		lvl       log.Level
		input     interface{}
		output    interface{}
	}{
		"info log ordenary string":       {log.Info, log.Info, "input_string", "input_string"},
		"warn log some object":           {log.Info, log.Warn, map[string]interface{}{"field": "val"}, nil},
		"error log ordinary string":      {log.Warn, log.Error, "error_desc", nil},
		"custom lvl log ordinary string": {log.Error, customLvl, "custom_lvl_log", "custom_lvl_log"},
	}

	out := bytes.NewBufferString("")
	logger := kitlog.New(out)

	for desc, tc := range cases {
		out.Reset()
		logger.SetLevel(tc.filterLvl)
		logger.Log(tc.lvl, tc.input)

		var output map[string]interface{}
		json.Unmarshal(out.Bytes(), &output)

		val := output[tc.lvl.String()]
		assert.EqualValues(t, tc.output, val, fmt.Sprintf("%s: unexpected value %s instead of %s", desc, val, tc.output))
	}
}
