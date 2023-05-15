package env

import (
	"fmt"
	"testing"

	"github.com/caarlos0/env/v7"
	"github.com/mainflux/mainflux/internal/clients/grpc"
	"github.com/mainflux/mainflux/internal/server"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	t.Run("Parsing a grpc.Config struct", func(t *testing.T) {
		opts := []env.Options{
			{
				Environment: map[string]string{
					"URL": "value",
				},
			},
		}
		cfg := &grpc.Config{}
		err := env.Parse(cfg, opts...)
		fmt.Println(cfg.URL)
		assert.NoError(t, err)
	})

	t.Run("Parsing a server.Config struct", func(t *testing.T) {
		cfg := &server.Config{}
		err := env.Parse(cfg)
		assert.NoError(t, err)
	})

	t.Run("Parsing an unsupported struct", func(t *testing.T) {
		//var cfg struct{}
		//err := env.Parse(&cfg)
		//assert.Error(t, err)
	})

	t.Run("Parsing with custom options", func(t *testing.T) {
		cfg := &grpc.Config{}
		opts := []env.Options{
			{
				Environment: map[string]string{
					"MY_ENV_VAR": "value",
				},
			},
		}
		err := env.Parse(cfg, opts...)
		assert.NoError(t, err)
	})
}
