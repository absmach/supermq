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
	msgTopicRegExp              = regexp.MustCompile(`^\/?` + MsgTopicPrefix + `\/([\w\-]+)\/` + ChannelTopicPrefix + `\/([\w\-]+)(\/[^?]*)?(\?.*)?$`)
	subTopicNotAllowedWildCards = []rune{'*', '>'}
	subTopicNotAllowedChars     = []rune{' ', '#', '+'}
)

func ParsePublishTopic(topic string) (domainID, chanID, subtopic string, err error) {
	msgParts := msgTopicRegExp.FindStringSubmatch(topic)
	if len(msgParts) < numGroups {
		return "", "", "", ErrMalformedTopic
	}

	domainID = msgParts[domainGroup]
	chanID = msgParts[channelGroup]
	subtopic = msgParts[subtopicGroup]

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

	subtopic, err = formatSubTopic(subtopic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedSubtopic, err)
	}

	for _, snaChars := range append(subTopicNotAllowedChars, subTopicNotAllowedWildCards...) {
		if strings.ContainsRune(subtopic, snaChars) {
			return "", ErrMalformedSubtopic
		}
	}

	return subtopic, nil
}

func ParseSubscribeTopic(topic string) (domainID string, chanID string, subtopic string, err error) {
	msgParts := msgTopicRegExp.FindStringSubmatch(topic)
	if len(msgParts) < numGroups {
		return "", "", "", ErrMalformedTopic
	}

	domainID = msgParts[domainGroup]
	chanID = msgParts[channelGroup]
	subtopic = msgParts[subtopicGroup]
	subtopic, err = ParseSubscribeSubtopic(subtopic)
	if err != nil {
		return "", "", "", errors.Wrap(ErrMalformedTopic, err)
	}

	return domainID, chanID, subtopic, nil
}

func ParseSubscribeSubtopic(subtopic string) (parseSubTopic string, err error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic = strings.ReplaceAll(subtopic, "+", "*")
	subtopic = strings.ReplaceAll(subtopic, "#", ">")

	subtopic, err = formatSubTopic(subtopic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedSubtopic, err)
	}

	elems := strings.Split(subtopic, ".")
	filteredElems := []string{}
	for _, elem := range elems {
		switch len(elem) {
		case 0:
			continue
		case 1:
			for _, snaChars := range subTopicNotAllowedChars {
				if strings.ContainsRune(elem, snaChars) {
					return "", ErrMalformedSubtopic
				}
			}
		default:
			for _, snaChars := range append(subTopicNotAllowedChars, subTopicNotAllowedWildCards...) {
				if strings.ContainsRune(elem, snaChars) {
					return "", ErrMalformedSubtopic
				}
			}
		}
		filteredElems = append(filteredElems, elem)
	}

	subtopic = strings.Join(filteredElems, ".")

	return subtopic, nil
}

func formatSubTopic(subtopic string) (string, error) {
	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", err
	}
	subtopic = strings.TrimPrefix(subtopic, "/")
	subtopic = strings.TrimSuffix(subtopic, "/")
	subtopic = strings.TrimSpace(subtopic)
	subtopic = strings.ReplaceAll(subtopic, "/", ".")
	return subtopic, nil
}

func EncodeTopic(domainID string, channelID string, subtopic string) string {
	return fmt.Sprintf("%s.%s", MsgTopicPrefix, EncodeTopicSuffix(domainID, channelID, subtopic))
}

func EncodeTopicSuffix(domainID string, channelID string, subtopic string) string {
	subject := fmt.Sprintf("%s.%s.%s", domainID, ChannelTopicPrefix, channelID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func (m *Message) EncodeTopicSuffix() string {
	return EncodeTopicSuffix(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func (m *Message) EncodeToMQTTTopic() string {
	topic := fmt.Sprintf("%s/%s/%s/%s", MsgTopicPrefix, m.GetDomain(), ChannelTopicPrefix, m.GetChannel())
	if m.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(m.GetSubtopic(), ".", "/")
	}
	return topic
}
