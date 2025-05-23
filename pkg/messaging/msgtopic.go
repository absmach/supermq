// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/absmach/supermq/pkg/errors"
)

const (
	MsgTopicPrefix     = "m"
	ChannelTopicPrefix = "c"

	numGroups     = 4 // entire expression + domain group + channel group + subtopic group
	domainGroup   = 1 // domain group is first in msg topic regexp
	channelGroup  = 2 // channel group is second in msg topic regexp
	subtopicGroup = 3 // subtopic group is third in msg topic regexp
)

var (
	ErrMalformedTopic    = errors.New("malformed topic")
	ErrMalformedSubtopic = errors.New("malformed subtopic")
	// Regex to group topic in format m.<domain_id>.c.<channel_id>.<sub_topic> `^\/?m\/([\w\-]+)\/c\/([\w\-]+)(\/[^?]*)?(\?.*)?$`.
	msgTopicRegExp = regexp.MustCompile(`^\/?` + MsgTopicPrefix + `\/([\w\-]+)\/` + ChannelTopicPrefix + `\/([\w\-]+)(\/[^?]*)?(\?.*)?$`)
)

func ParseTopic(topic string) (string, string, string, error) {
	return ParseTopicWithOption(topic, false)
}

func ParseTopicWithOption(topic string, skipParseSubtopic bool) (string, string, string, error) {
	msgParts := msgTopicRegExp.FindStringSubmatch(topic)
	if len(msgParts) < numGroups {
		return "", "", "", ErrMalformedTopic
	}

	domainID := msgParts[domainGroup]
	chanID := msgParts[channelGroup]
	subtopic := msgParts[subtopicGroup]

	if !skipParseSubtopic {
		var err error
		subtopic, err = ParseSubtopic(subtopic)
		if err != nil {
			return "", "", "", errors.Wrap(ErrMalformedTopic, err)
		}
	}

	return domainID, chanID, subtopic, nil
}

func ParseSubtopic(subtopic string) (string, error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedSubtopic, err)
	}
	subtopic = strings.ReplaceAll(subtopic, "/", ".")

	elems := strings.Split(subtopic, ".")
	filteredElems := []string{}
	for _, elem := range elems {
		if elem == "" {
			continue
		}

		if len(elem) > 1 && (strings.Contains(elem, "*") || strings.Contains(elem, ">")) {
			return "", ErrMalformedSubtopic
		}

		filteredElems = append(filteredElems, elem)
	}

	subtopic = strings.Join(filteredElems, ".")
	return subtopic, nil
}

func EncodeToInternalSubject(domainID string, channelID string, subtopic string) string {
	return fmt.Sprintf("%s.%s", MsgTopicPrefix, EncodeToInternalSubjectSuffix(domainID, ChannelTopicPrefix, channelID))
}

func EncodeToInternalSubjectSuffix(domainID string, channelID string, subtopic string) string {
	subject := fmt.Sprintf("%s.%s.%s", domainID, ChannelTopicPrefix, channelID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func (m *Message) EncodeToInternalSubject() string {
	return EncodeToInternalSubject(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func (m *Message) EncodeToInternalSubjectSuffix() string {
	return EncodeToInternalSubjectSuffix(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func (m *Message) EncodeToMQTTTopic() string {
	topic := fmt.Sprintf("%s/%s/%s/%s", MsgTopicPrefix, m.GetDomain(), ChannelTopicPrefix, m.GetChannel())
	if m.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(m.GetSubtopic(), ".", "/")
	}
	return topic
}
