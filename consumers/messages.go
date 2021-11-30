// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package consumers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pelletier/go-toml"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	pubsub "github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/transformers"
	"github.com/mainflux/mainflux/pkg/transformers/json"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
)

var (
	errOpenConfFile  = errors.New("unable to open configuration file")
	errParseConfFile = errors.New("unable to parse configuration file")
)

// Start method starts consuming messages received from NATS.
// This method transforms messages to SenML format before
// using MessageRepository to store them.
func Start(sub messaging.Subscriber, consumer Consumer, configPath string, logger logger.Logger) error {
	cfg, err := loadConfig(configPath)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to load consumer config: %s", err))
	}

	transformer := makeTransformer(cfg.transformer, logger)

	for _, subject := range cfg.subscriber.subjects {
		if err := sub.Subscribe(subject, handler(transformer, consumer)); err != nil {
			return err
		}
	}
	return nil
}

func handler(t transformers.Transformer, c Consumer) messaging.MessageHandler {
	return func(msg messaging.Message) error {
		m := interface{}(msg)
		var err error
		if t != nil {
			m, err = t.Transform(msg)
			if err != nil {
				return err
			}
		}
		return c.Consume(m)
	}
}

type subscriberConfig struct {
	subjects []string `toml:"subjects"`
}

type transformerConfig struct {
	format      string            `toml:"format"`
	contentType string            `toml:"content_type"`
	timestamps  map[string]string `toml:"timestamps"`
}

type config struct {
	subscriber  subscriberConfig  `toml:"subscriber"`
	transformer transformerConfig `toml:"transformer"`
}

func loadConfig(subjectsConfigPath string) (config, error) {
	cfg := config{
		subscriber: subscriberConfig{
			subjects: []string{pubsub.SubjectAllChannels},
		},
	}

	data, err := ioutil.ReadFile(subjectsConfigPath)
	if err != nil {
		return cfg, errors.Wrap(errOpenConfFile, err)
	}

	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, errors.Wrap(errParseConfFile, err)
	}

	return cfg, nil
}

func makeTransformer(cfg transformerConfig, logger logger.Logger) transformers.Transformer {
	switch strings.ToUpper(cfg.format) {
	case "SENML":
		logger.Info("Using SenML transformer")
		return senml.New(cfg.contentType)
	case "JSON":
		logger.Info("Using JSON transformer")
		return json.New(cfg.timestamps)
	default:
		logger.Error(fmt.Sprintf("Can't create transformer: unknown transformer type %s", cfg.format))
		os.Exit(1)
		return nil
	}
}
