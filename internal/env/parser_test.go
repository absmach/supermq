package env

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"

	"github.com/mainflux/mainflux/internal/clients/grpc"
	"github.com/mainflux/mainflux/internal/server"
	"github.com/stretchr/testify/assert"
)

var errNotDuration error = errors.New("unable to parse duration")

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		config         interface{}
		expectedConfig interface{}
		options        []Options
		err            error
	}{
		{
			"Parsing a grpc.Config struct",
			&grpc.Config{},
			&grpc.Config{
				URL:     "val.com",
				Timeout: time.Second,
			},
			[]Options{
				{
					Environment: map[string]string{
						"URL":     "val.com",
						"TIMEOUT": time.Second.String(),
					},
				},
			},
			nil,
		},
		{
			"Invalid type parsing",
			&grpc.Config{},
			&grpc.Config{URL: "val.com"},
			[]Options{
				{
					Environment: map[string]string{
						"URL":     "val.com",
						"TIMEOUT": "invalid",
					},
				},
			},
			errNotDuration,
		},
		{
			"Parsing with Server Config with Alt-Prefix",
			&server.Config{},
			&server.Config{
				Host:     "localhost",
				Port:     "8080",
				CertFile: "cert",
				KeyFile:  "key",
			},
			[]Options{
				{
					Environment: map[string]string{
						"MF-HOST":        "localhost",
						"MF-HTTP-PORT":   "8080",
						"MF-SERVER_CERT": "cert",
						"MF-SERVER_KEY":  "key",
					},
					Prefix:    "MF-",
					AltPrefix: "MF-HTTP-",
				},
			},
			nil,
		},
		{
			"Parsing a grpc.Config with Alt-Prefix",
			&grpc.Config{},
			&grpc.Config{
				URL:     "val.com",
				Timeout: time.Second,
			},
			[]Options{
				{
					Environment: map[string]string{
						"MF-URL":          "val.com",
						"MF-GRPC-TIMEOUT": time.Second.String(),
					},
					Prefix:    "MF-",
					AltPrefix: "MF-GRPC-",
				},
			},
			nil,
		},
		{
			"Parsing a grpc.Config with Alt-Prefix and errors",
			&grpc.Config{},
			&grpc.Config{
				URL:     "val.com",
				Timeout: time.Second,
			},
			[]Options{
				{
					Environment: map[string]string{
						"MF-URL":          "val.com",
						"MF-GRPC-TIMEOUT": "not-duration",
					},
					Prefix:    "MF-",
					AltPrefix: "MF-GRPC-",
				},
			},
			errNotDuration,
		},
	}
	for _, test := range tests {
		err := Parse(test.config, test.options...)
		if test.err != nil {
			assert.Error(t, err, "expected error but got nil")
		} else {
			assert.NoError(t, err, fmt.Sprintf("expected no error but got %v", err))
		}
		assert.Equal(t, test.expectedConfig, test.config, fmt.Sprintf("expected %v got %v", test.expectedConfig, test.config))

	}
}
