// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"github.com/dgraph-io/ristretto/v2"
)

type TopicParser interface {
	ParsePublishTopic(topic string, resolve bool) (domainID, chanID, subtopic string, err error)
	ParsePublishSubtopic(subtopic string) (parseSubTopic string, err error)
}

type parser struct {
	cache *ristretto.Cache[string, string]
}

func NewTopicParser(cache *ristretto.Cache[string, string]) TopicParser {
	return &parser{
		cache: cache,
	}
}

func (p *parser) ParsePublishTopic(topic string, resolve bool) (domainID, chanID, subtopic string, err error) {
	
	domainID, chanID, subtopic, err = ParseTopic(topic)
	if err != nil {
		return "", "", "", err
	}
	subtopic, err = p.ParsePublishSubtopic(subtopic)
	if err != nil {
		return "", "", "", errors.Wrap(ErrMalformedTopic, err)
	}

	return domainID, chanID, subtopic, nil
}

func (p *parser) ParsePublishSubtopic(subtopic string) (parseSubTopic string, err error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err = formatSubtopic(subtopic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedSubtopic, err)
	}

	if strings.ContainsAny(subtopic, subtopicInvalidChars+wildcards) {
		return "", ErrMalformedSubtopic
	}

	if strings.Contains(subtopic, "..") {
		return "", ErrMalformedSubtopic
	}

	if p.cache != nil {
		if cached, found := p.cache.Get(subtopic); found {
			return cached, nil
		}
	}

	formatted := pathReplacer.Replace(subtopic)
	if p.cache != nil {
		p.cache.SetWithTTL(subtopic, formatted, 1, 0)
	}

	return formatted, nil
}

func encodeValue(domainID, channelID, subtopic string) string {
	val  := domain
}