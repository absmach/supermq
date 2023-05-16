package env

import (
	"fmt"
	"testing"
	"time"

	"github.com/mainflux/mainflux/internal/clients/grpc"
	"github.com/mainflux/mainflux/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("Parsing a grpc.Config struct", func(t *testing.T) {
		testCfg := &grpc.Config{
			URL:     "val.com",
			Timeout: time.Second,
		}
		opts := []Options{
			{
				Environment: map[string]string{
					"URL":     testCfg.URL,
					"TIMEOUT": testCfg.Timeout.String(),
				},
			},
		}
		cfg := &grpc.Config{}
		err := Parse(cfg, opts...)
		assert.NoError(t, err)
		assert.Equal(t, *testCfg, *cfg, fmt.Sprintf("expected %v got %v", testCfg, cfg))
	})
	t.Run("Invalid type parsing", func(t *testing.T) {
		testCfg := &grpc.Config{
			URL:     "val.com",
			Timeout: time.Second,
		}
		opts := []Options{
			{
				Environment: map[string]string{
					"URL":     testCfg.URL,
					"TIMEOUT": "invalid",
				},
			},
		}
		cfg := &grpc.Config{}
		err := Parse(cfg, opts...)
		assert.Error(t, err)
	})

	t.Run("Parsing with Server Config with Alt-Prefix", func(t *testing.T) {
		testCfg := &server.Config{
			Host:     "localhost",
			Port:     "8080",
			CertFile: "cert",
			KeyFile:  "key",
		}
		opts := []Options{
			{
				Environment: map[string]string{
					"MF-HOST":        testCfg.Host,
					"MF-HTTP-PORT":   testCfg.Port,
					"MF-SERVER_CERT": testCfg.CertFile,
					"MF-SERVER_KEY":  testCfg.KeyFile,
				},
				Prefix:    "MF-",
				AltPrefix: "MF-HTTP-",
			},
		}
		cfg := &server.Config{}
		err := Parse(cfg, opts...)
		assert.NoError(t, err)
		assert.Equal(t, *testCfg, *cfg, fmt.Sprintf("expected %v got %v", testCfg, cfg))
	})
	t.Run("Parsing a grpc.Config with Alt-Prefix", func(t *testing.T) {
		testCfg := &grpc.Config{
			URL:     "val.com",
			Timeout: time.Second,
		}
		opts := []Options{
			{
				Environment: map[string]string{
					"MF-URL":          testCfg.URL,
					"MF-GRPC-TIMEOUT": testCfg.Timeout.String(),
				},
				Prefix:    "MF-",
				AltPrefix: "MF-GRPC-",
			},
		}
		cfg := &grpc.Config{}
		err := Parse(cfg, opts...)
		assert.NoError(t, err)
		assert.Equal(t, *testCfg, *cfg, fmt.Sprintf("expected %v got %v", testCfg, cfg))
	})
	t.Run("Parsing a grpc.Config with Alt-Prefix and errors", func(t *testing.T) {
		testCfg := &grpc.Config{
			URL:     "val.com",
			Timeout: time.Second,
		}
		opts := []Options{
			{
				Environment: map[string]string{
					"MF-URL":          testCfg.URL,
					"MF-GRPC-TIMEOUT": "not-duration",
				},
				Prefix:    "MF-",
				AltPrefix: "MF-GRPC-",
			},
		}
		cfg := &grpc.Config{}
		err := Parse(cfg, opts...)
		assert.Error(t, err)
	})
}
