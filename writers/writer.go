// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package writers

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	pubsub "github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/transformers"
	"github.com/mainflux/mainflux/pkg/transformers/senml"
)

var (
	errOpenConfFile      = errors.New("unable to open configuration file")
	errParseConfFile     = errors.New("unable to parse configuration file")
	errMessageConversion = errors.New("error conversing transformed messages")
)

type consumer struct {
	repo        MessageRepository
	transformer transformers.Transformer
	logger      logger.Logger
}

// Start method starts consuming messages received from NATS.
// This method transforms messages to SenML format before
// using MessageRepository to store them.
func Start(sub messaging.Subscriber, repo MessageRepository, contentType string, queue string, cfgPath string, logger logger.Logger) error {
	c := consumer{
		repo:   repo,
		logger: logger,
	}

	subjects, keys, err := loadConfig(cfgPath)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to load subjects: %s", err))
	}

	switch contentType {
	case json.ContentType:
		c.transformer = json.New(keys)
	case senml.ContentTypeJSON,
		senml.ContentTypeCBOR:
		c.transformer = senml.New(contentType)
	default:
		c.transformer = senml.New(senml.ContentTypeJSON)
	}

	for _, subject := range subjects {
		if err := sub.Subscribe(subject, c.handler); err != nil {
			return err
		}
	}
	return nil
}

func (c *consumer) handler(msg messaging.Message) error {
	t, err := c.transformer.Transform(msg)
	if err != nil {
		return err
	}

	return c.repo.Save(t)
}

type filterConfig struct {
	Filter []string `toml:"filter"`
}

type writerConfig struct {
	Subjects filterConfig `toml:"subjects"`
	Keys     filterConfig `toml:"keys"`
}

func loadConfig(subjectsConfigPath string) ([]string, []string, error) {
	data, err := ioutil.ReadFile(subjectsConfigPath)
	if err != nil {
		return []string{pubsub.SubjectAllChannels}, []string{"*"}, errors.Wrap(errOpenConfFile, err)
	}

	var cfg writerConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return []string{pubsub.SubjectAllChannels}, []string{"*"}, errors.Wrap(errParseConfFile, err)
	}

	return cfg.Subjects.Filter, cfg.Keys.Filter, nil
}
