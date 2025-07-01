// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/absmach/supermq/pkg/errors"
)

const (
	MsgTopicPrefix     = 'm'
	ChannelTopicPrefix = 'c'
)

var (
	ErrMalformedTopic    = errors.New("malformed topic")
	ErrMalformedSubtopic = errors.New("malformed subtopic")
	mqWildcards          = "+#"
	wildcards            = "*>"
	subtopicInvalidChars = " #+"
	wildcardsReplacer    = strings.NewReplacer("+", "*", "#", ">")
	pathReplacer         = strings.NewReplacer("/", ".")
)

func ParsePublishTopic(topic string) (domainID, chanID, subtopic string, err error) {
	domainID, chanID, subtopic, err = ParseTopic(topic)
	if err != nil {
		return "", "", "", err
	}
	subtopic, err = ParsePublishSubtopic(subtopic)
	if err != nil {
		return "", "", "", errors.Wrap(ErrMalformedTopic, err)
	}

	return domainID, chanID, subtopic, nil
}

func ParsePublishSubtopic(subtopic string) (parseSubTopic string, err error) {
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

	return subtopic, nil
}

func ParseSubscribeTopic(topic string) (domainID string, chanID string, subtopic string, err error) {
	domainID, chanID, subtopic, err = ParseTopic(topic)
	if err != nil {
		return "", "", "", err
	}
	subtopic, err = ParseSubscribeSubtopic(subtopic)
	if err != nil {
		return "", "", "", errors.Wrap(ErrMalformedTopic, err)
	}

	return domainID, chanID, subtopic, nil
}

func ParseSubscribeSubtopic(subtopic string) (parseSubTopic string, err error) {
	if subtopic == "" {
		return "", nil
	}

	if strings.ContainsAny(subtopic, mqWildcards) {
		subtopic = wildcardsReplacer.Replace(subtopic)
	}
	subtopic, err = formatSubtopic(subtopic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedSubtopic, err)
	}

	if strings.ContainsAny(subtopic, subtopicInvalidChars) {
		return "", ErrMalformedSubtopic
	}

	if strings.Contains(subtopic, "..") {
		return "", ErrMalformedSubtopic
	}

	for _, elem := range strings.Split(subtopic, ".") {
		if len(elem) > 1 && strings.ContainsAny(elem, wildcards) {
			return "", ErrMalformedSubtopic
		}
	}
	return subtopic, nil
}

func formatSubtopic(subtopic string) (string, error) {
	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", err
	}
	subtopic = strings.TrimPrefix(subtopic, "/")
	subtopic = strings.TrimSuffix(subtopic, "/")
	subtopic = strings.TrimSpace(subtopic)
	subtopic = pathReplacer.Replace(subtopic)
	return subtopic, nil
}

func EncodeTopic(domainID string, channelID string, subtopic string) string {
	return fmt.Sprintf("%s.%s", string(MsgTopicPrefix), EncodeTopicSuffix(domainID, channelID, subtopic))
}

func EncodeTopicSuffix(domainID string, channelID string, subtopic string) string {
	subject := fmt.Sprintf("%s.%s.%s", domainID, string(ChannelTopicPrefix), channelID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func EncodeMessageTopic(m *Message) string {
	return EncodeTopicSuffix(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func EncodeMessageMQTTTopic(m *Message) string {
	topic := fmt.Sprintf("%s/%s/%s/%s", string(MsgTopicPrefix), m.GetDomain(), string(ChannelTopicPrefix), m.GetChannel())
	if m.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(m.GetSubtopic(), ".", "/")
	}
	return topic
}

func EncodeAdapterTopic(domain, channel, subtopic string) string {
	topic := fmt.Sprintf("%s/%s/%s/%s", string(MsgTopicPrefix), domain, string(ChannelTopicPrefix), channel)
	if subtopic != "" {
		topic = topic + "/" + subtopic
	}
	return topic
}

// ParseTopic parses a messaging topic string and returns the domain ID, channel ID, and subtopic.
// This is an optimized version with no regex and minimal allocations.
func ParseTopic(topic string) (domainID, chanID, subtopic string, err error) {
	// location of string "m"
	start := 0
	// Handle both formats: "/m/domain/c/channel/subtopic" and "m/domain/c/channel/subtopic".
	// If topic start with m/ then start is 0 , If topic start with /m/ then start is 1.
	n := len(topic)
	if n > 0 && topic[0] == '/' {
		start = 1
	}

	// length check - minimum: "m/<domain_id>/c/" = 5 characters if ignore <domain_id> and in this case start will be 0
	// length check - minimum: "/m/<domain_id>/c/" = 6 characters if ignore <domain_id> and in this case start will be 1
	if n < start+5 {
		return "", "", "", ErrMalformedTopic
	}
	if topic[start] != MsgTopicPrefix || topic[start+1] != '/' {
		return "", "", "", ErrMalformedTopic
	}
	pos := start + 2

	// Find "/c/" to locate domain ID
	cPos := -1
	for i := pos; i <= n-3; i++ {
		if topic[i] == '/' && topic[i+1] == ChannelTopicPrefix && topic[i+2] == '/' {
			cPos = i - pos
			break
		}
	}
	if cPos == -1 || cPos == 0 {
		return "", "", "", ErrMalformedTopic
	}
	domainID = topic[pos : pos+cPos]
	// skip "/c/"
	pos = pos + cPos + 3

	// Ensure channel exists
	if pos >= n {
		return "", "", "", ErrMalformedTopic
	}

	// Find '/' after channelID
	nextSlash := -1
	for i := pos; i < n; i++ {
		if topic[i] == '/' {
			nextSlash = i - pos
			break
		}
	}

	if nextSlash == -1 {
		// No subtopic
		chanID = topic[pos:]
	} else {
		chanID = topic[pos : pos+nextSlash]
		subtopic = topic[pos+nextSlash+1:]
	}

	// Validate channelID
	if len(chanID) == 0 {
		return "", "", "", ErrMalformedTopic
	}

	return domainID, chanID, subtopic, nil
}
