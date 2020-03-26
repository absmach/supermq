// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package writers

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/gogo/protobuf/proto"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/transformers"
	"github.com/mainflux/mainflux/transformers/senml"
	nats "github.com/nats-io/nats.go"
	"io/ioutil"
)

var (
	errOpenConfFile = errors.New("Unable to open configuration file")
	errParseConfFile = errors.New("Unable to parse configuration file")
)

type consumer struct {
	nc          *nats.Conn
	subjects    []string
	repo        MessageRepository
	transformer transformers.Transformer
	logger      logger.Logger
}

// Start method starts consuming messages received from NATS.
// This method transforms messages to SenML format before
// using MessageRepository to store them.
func Start(nc *nats.Conn, repo MessageRepository, transformer transformers.Transformer, queue string, filters FiltersCfg, logger logger.Logger) error {
	c := consumer{
		nc:          nc,
		subjects:    filters.subjects,
		repo:        repo,
		transformer: transformer,
		logger:      logger,
	}

	// TODO subscribe with only selected subjects
	_, err := nc.QueueSubscribe(mainflux.InputChannels, queue, c.consume)
	return err
}

func (c *consumer) consume(m *nats.Msg) {
	var msg mainflux.Message
	if err := proto.Unmarshal(m.Data, &msg); err != nil {
		c.logger.Warn(fmt.Sprintf("Failed to unmarshal received message: %s", err))
		return
	}

	t, err := c.transformer.Transform(msg)
	if err != nil {
		c.logger.Warn(fmt.Sprintf("Failed to tranform received message: %s", err))
		return
	}
	msgs, ok := t.([]senml.Message)
	if !ok {
		c.logger.Warn("Invalid message format from the Transformer output.")
		return
	}

	if err := c.repo.Save(msgs...); err != nil {
		c.logger.Warn(fmt.Sprintf("Failed to save message: %s", err))
		return
	}
}

type filterConfig struct {
	List []string `toml:"filter"`
}

type subjectsConfig struct {
	Subjects filterConfig `toml:"subjects"`
}

type FiltersCfg struct {
	subjects []string
}

func LoadSubjectsConfig(subjectsConfigPath string) (FiltersCfg, error)  {
	data, err := ioutil.ReadFile(subjectsConfigPath)
	if err != nil {
		return FiltersCfg{}, errors.Wrap(errOpenConfFile, err)
	}

	var subjectsCfg subjectsConfig
	if err := toml.Unmarshal(data, &subjectsCfg); err != nil {
		return FiltersCfg{}, errors.Wrap(errParseConfFile, err)
	}

	return FiltersCfg{
		subjects: subjectsCfg.Subjects.List,
	}, err
}
