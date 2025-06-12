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
	TopicRegExp          = regexp.MustCompile(`^\/?` + MsgTopicPrefix + `\/([\w\-]+)\/` + ChannelTopicPrefix + `\/([\w\-]+)(\/[^?]*)?(\?.*)?$`)
	mqWildcards          = "+#"
	wildcards            = "*>"
	subtopicInvalidChars = " #+"
	wildcardsReplacer    = strings.NewReplacer("+", "*", "#", ">")
	pathReplacer         = strings.NewReplacer("/", ".")
)

func ParsePublishTopic(topic string) (domainID, chanID, subtopic string, err error) {
	msgParts := TopicRegExp.FindStringSubmatch(topic)
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
	msgParts := TopicRegExp.FindStringSubmatch(topic)
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
	return fmt.Sprintf("%s.%s", MsgTopicPrefix, EncodeTopicSuffix(domainID, channelID, subtopic))
}

func EncodeTopicSuffix(domainID string, channelID string, subtopic string) string {
	subject := fmt.Sprintf("%s.%s.%s", domainID, ChannelTopicPrefix, channelID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func EncodeMessageTopic(m *Message) string {
	return EncodeTopicSuffix(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func EncodeMessageMQTTTopic(m *Message) string {
	topic := fmt.Sprintf("%s/%s/%s/%s", MsgTopicPrefix, m.GetDomain(), ChannelTopicPrefix, m.GetChannel())
	if m.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(m.GetSubtopic(), ".", "/")
	}
	return topic
}
