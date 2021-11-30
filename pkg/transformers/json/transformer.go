// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package json

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/transformers"
)

const sep = "/"

var (
	keys = [...]string{"publisher", "protocol", "channel", "subtopic"}

	// ErrTransform represents an error during parsing message.
	ErrTransform = errors.New("unable to parse JSON object")
	// ErrInvalidKey represents the use of a reserved message field.
	ErrInvalidKey = errors.New("invalid object key")

	errUnknownFormat     = errors.New("unknown format of JSON message")
	errInvalidFormat     = errors.New("invalid JSON object")
	errInvalidNestedJSON = errors.New("invalid nested JSON object")

	timestamps = map[string]string{}
)

type funcTransformer func(messaging.Message) (interface{}, error)

// New returns a new JSON transformer.
func New(ts map[string]string) transformers.Transformer {
	// TODO: Improve the timestamp config
	timestamps = ts
	return funcTransformer(transformer)
}

// Transform transforms Mainflux message to a list of JSON messages.
func (fh funcTransformer) Transform(msg messaging.Message) (interface{}, error) {
	return fh(msg)
}

func transformer(msg messaging.Message) (interface{}, error) {
	ret := Message{
		Publisher: msg.Publisher,
		Created:   msg.Created,
		Protocol:  msg.Protocol,
		Channel:   msg.Channel,
		Subtopic:  msg.Subtopic,
	}

	if ret.Subtopic == "" {
		return nil, errors.Wrap(ErrTransform, errUnknownFormat)
	}

	subs := strings.Split(ret.Subtopic, ".")
	if len(subs) == 0 {
		return nil, errors.Wrap(ErrTransform, errUnknownFormat)
	}

	format := subs[len(subs)-1]
	var payload interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, errors.Wrap(ErrTransform, err)
	}

	switch p := payload.(type) {
	case map[string]interface{}:
		ret.Payload = p

		// Apply timestamnp transformation rules depending on key/unit pairs
		if len(timestamps) > 0 {
			ts, err := transformTimestamp(msg.Payload)
			if ts != 0 && err != nil {
				ret.Created = ts
			}
		}
		return Messages{[]Message{ret}, format}, nil
	case []interface{}:
		res := []Message{}
		// Make an array of messages from the root array.
		for _, val := range p {
			v, ok := val.(map[string]interface{})
			if !ok {
				return nil, errors.Wrap(ErrTransform, errInvalidNestedJSON)
			}
			newMsg := ret

			// Apply timestamnp transformation rules depending on key/unit pairs
			if len(timestamps) > 0 {
				ts, err := transformTimestamp(msg.Payload)
				if ts != 0 && err != nil {
					ret.Created = ts
				}
			}

			newMsg.Payload = v
			res = append(res, newMsg)
		}
		return Messages{res, format}, nil
	default:
		return nil, errors.Wrap(ErrTransform, errInvalidFormat)
	}
}

// ParseFlat receives flat map that represents complex JSON objects and returns
// the corresponding complex JSON object with nested maps. It's the opposite
// of the Flatten function.
func ParseFlat(flat interface{}) interface{} {
	msg := make(map[string]interface{})
	switch v := flat.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			subKeys := strings.Split(key, sep)
			n := len(subKeys)
			if n == 1 {
				msg[key] = value
				continue
			}
			current := msg
			for i, k := range subKeys {
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				if i == n-1 {
					current[k] = value
					break
				}
				current = current[k].(map[string]interface{})
			}
		}
	}
	return msg
}

// Flatten makes nested maps flat using composite keys created by concatenation of the nested keys.
func Flatten(m map[string]interface{}) (map[string]interface{}, error) {
	return flatten("", make(map[string]interface{}), m)
}

func flatten(prefix string, m, m1 map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range m1 {
		if strings.Contains(k, sep) {
			return nil, ErrInvalidKey
		}
		for _, key := range keys {
			if k == key {
				return nil, ErrInvalidKey
			}
		}
		switch val := v.(type) {
		case map[string]interface{}:
			var err error
			m, err = flatten(prefix+k+sep, m, val)
			if err != nil {
				return nil, err
			}
		default:
			m[prefix+k] = v
		}
	}
	return m, nil
}

func transformTimestamp(payload []byte) (int64, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return 0, err
	}

	for field, format := range timestamps {
		if fieldVal, ok := data[field]; ok {
			t, err := parseTimestamp(format, fieldVal, "")
			if err != nil {
				return 0, err
			}
			return t.UnixNano(), nil
		}
	}

	return 0, nil
}

// ParseTimestamp parses a Time according to the standard Telegraf options.
// These are generally displayed in the toml similar to:
//   json_time_key= "timestamp"
//   json_time_format = "2006-01-02T15:04:05Z07:00"
//   json_timezone = "America/Los_Angeles"
//
// The format can be one of "unix", "unix_ms", "unix_us", "unix_ns", or a Go
// time layout suitable for time.Parse.
//
// When using the "unix" format, a optional fractional component is allowed.
// Specific unix time precisions cannot have a fractional component.
//
// Unix times may be an int64, float64, or string.  When using a Go format
// string the timestamp must be a string.
//
// The location is a location string suitable for time.LoadLocation.  Unix
// times do not use the location string, a unix time is always return in the
// UTC location.
func parseTimestamp(format string, timestamp interface{}, location string) (time.Time, error) {
	switch format {
	case "unix", "unix_ms", "unix_us", "unix_ns":
		return parseUnix(format, timestamp)
	default:
		if location == "" {
			location = "UTC"
		}
		return parseTime(format, timestamp, location)
	}
}

func parseUnix(format string, timestamp interface{}) (time.Time, error) {
	integer, fractional, err := parseComponents(timestamp)
	if err != nil {
		return time.Unix(0, 0), err
	}

	switch strings.ToLower(format) {
	case "unix":
		return time.Unix(integer, fractional).UTC(), nil
	case "unix_ms":
		return time.Unix(0, integer*1e6).UTC(), nil
	case "unix_us":
		return time.Unix(0, integer*1e3).UTC(), nil
	case "unix_ns":
		return time.Unix(0, integer).UTC(), nil
	default:
		return time.Unix(0, 0), errors.New("unsupported type")
	}
}

// Returns the integers before and after an optional decimal point.  Both '.'
// and ',' are supported for the decimal point.  The timestamp can be an int64,
// float64, or string.
//   ex: "42.5" -> (42, 5, nil)
func parseComponents(timestamp interface{}) (int64, int64, error) {
	switch ts := timestamp.(type) {
	case string:
		parts := strings.SplitN(ts, ".", 2)
		if len(parts) == 2 {
			return parseUnixTimeComponents(parts[0], parts[1])
		}

		parts = strings.SplitN(ts, ",", 2)
		if len(parts) == 2 {
			return parseUnixTimeComponents(parts[0], parts[1])
		}

		integer, err := strconv.ParseInt(ts, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		return integer, 0, nil
	case int8:
		return int64(ts), 0, nil
	case int16:
		return int64(ts), 0, nil
	case int32:
		return int64(ts), 0, nil
	case int64:
		return ts, 0, nil
	case uint8:
		return int64(ts), 0, nil
	case uint16:
		return int64(ts), 0, nil
	case uint32:
		return int64(ts), 0, nil
	case uint64:
		return int64(ts), 0, nil
	case float32:
		integer, fractional := math.Modf(float64(ts))
		return int64(integer), int64(fractional * 1e9), nil
	case float64:
		integer, fractional := math.Modf(ts)
		return int64(integer), int64(fractional * 1e9), nil
	default:
		return 0, 0, errors.New("unsupported type")
	}
}

func parseUnixTimeComponents(first, second string) (int64, int64, error) {
	integer, err := strconv.ParseInt(first, 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// Convert to nanoseconds, dropping any greater precision.
	buf := []byte("000000000")
	copy(buf, second)

	fractional, err := strconv.ParseInt(string(buf), 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return integer, fractional, nil
}

// ParseTime parses a string timestamp according to the format string.
func parseTime(format string, timestamp interface{}, location string) (time.Time, error) {
	switch ts := timestamp.(type) {
	case string:
		loc, err := time.LoadLocation(location)
		if err != nil {
			return time.Unix(0, 0), err
		}
		switch strings.ToLower(format) {
		case "ansic":
			format = time.ANSIC
		case "unixdate":
			format = time.UnixDate
		case "rubydate":
			format = time.RubyDate
		case "rfc822":
			format = time.RFC822
		case "rfc822z":
			format = time.RFC822Z
		case "rfc850":
			format = time.RFC850
		case "rfc1123":
			format = time.RFC1123
		case "rfc1123z":
			format = time.RFC1123Z
		case "rfc3339":
			format = time.RFC3339
		case "rfc3339nano":
			format = time.RFC3339Nano
		case "stamp":
			format = time.Stamp
		case "stampmilli":
			format = time.StampMilli
		case "stampmicro":
			format = time.StampMicro
		case "stampnano":
			format = time.StampNano
		}
		return time.ParseInLocation(format, ts, loc)
	default:
		return time.Unix(0, 0), errors.New("unsupported type")
	}
}
