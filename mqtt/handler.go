// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/mqtt/redis"
	"github.com/mainflux/mainflux/pkg/auth"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mproxy/pkg/session"
)

var _ session.Handler = (*handler)(nil)

const protocol = "mqtt"

var (
	channelRegExp                   = regexp.MustCompile(`^\/?channels\/([\w\-]+)\/messages(\/[^?]*)?(\?.*)?$`)
	infoSubscribed                  = "subscribed with client_id: %s to topics: %s"
	infoUnsubscribed                = "unsubscribed client_id: %s from topics: %s"
	infoConnected                   = "connected with client_id: %s"
	infoDisconnected                = "disconnected client_id: %s and username: %s"
	infoPublished                   = "published with client_id: %s to the topic: %s"
	infoMalformedTopic              = "Malformed topic: "
	errFailedUnsubscribe            = errors.New("failed to unsubscribe:")
	errFailedDisconnect             = errors.New("failed to disconnect:")
	errFailedSubscribe              = errors.New("failed to subscribe:")
	errFailedConnect                = errors.New("failed to connect:")
	errClientNotInitialized         = errors.New("client is not initialized")
	errMalformedTopic               = errors.New("malformed topic")
	errMalformedSubtopic            = errors.New("malformed subtopic")
	errMissingTopicPub              = errors.New("failed to publish due to missing topic")
	errMissingTopicSub              = errors.New("failed to subscribe due to missing topic")
	errFailedPublishConnectEvent    = errors.New("failed to publish connect event:")
	errFailedPublishDisconnectEvent = errors.New("failed to publish disconnect event:")
	errFailedPublish                = errors.New("failed to publish:")
	errFailedParseSubtopic          = errors.New("failed to parse subtopic:")
	errFailedPublishToMsgBroker     = errors.New("failed to publish to mainflux message broker")
)

// Event implements events.Event interface
type handler struct {
	publishers []messaging.Publisher
	auth       auth.Client
	logger     logger.Logger
	es         redis.EventStore
}

// NewHandler creates new Handler entity
func NewHandler(publishers []messaging.Publisher, es redis.EventStore,
	logger logger.Logger, auth auth.Client) session.Handler {
	return &handler{
		es:         es,
		logger:     logger,
		publishers: publishers,
		auth:       auth,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker
func (h *handler) AuthConnect(c *session.Client) error {
	if c == nil {
		return errClientNotInitialized
	}

	thid, err := h.auth.Identify(context.Background(), string(c.Password))
	if err != nil {
		return err
	}

	if thid != c.Username {
		return errors.ErrAuthentication
	}

	if err := h.es.Connect(c.Username); err != nil {
		h.logger.Error(errors.Wrap(errFailedPublishConnectEvent, err).Error())
	}

	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (h *handler) AuthPublish(c *session.Client, topic *string, payload *[]byte) error {
	if c == nil {
		return errClientNotInitialized
	}
	if topic == nil {
		return errMissingTopicPub
	}

	return h.authAccess(c.Username, *topic)
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (h *handler) AuthSubscribe(c *session.Client, topics *[]string) error {
	if c == nil {
		return errClientNotInitialized
	}
	if topics == nil || *topics == nil {
		return errMissingTopicSub
	}

	for _, v := range *topics {
		if err := h.authAccess(c.Username, v); err != nil {
			return err
		}

	}

	return nil
}

// Connect - after client successfully connected
func (h *handler) Connect(c *session.Client) {
	if c == nil {
		h.logger.Error(errors.Wrap(errFailedConnect, errClientNotInitialized).Error())
		return
	}
	h.logger.Info(fmt.Sprintf(infoConnected, c.ID))
}

// Publish - after client successfully published
func (h *handler) Publish(c *session.Client, topic *string, payload *[]byte) {
	if c == nil {
		h.logger.Error(errClientNotInitialized.Error())
		return
	}
	h.logger.Info(fmt.Sprintf(infoPublished, c.ID, *topic))
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>

	channelParts := channelRegExp.FindStringSubmatch(*topic)
	if len(channelParts) < 2 {
		h.logger.Error(errors.Wrap(errFailedPublish, errMalformedTopic).Error())
		return
	}

	chanID := channelParts[1]
	subtopic := channelParts[2]

	subtopic, err := parseSubtopic(subtopic)
	if err != nil {
		h.logger.Error(errors.Wrap(errFailedParseSubtopic, errFailedParseSubtopic).Error())
		return
	}

	msg := messaging.Message{
		Protocol:  protocol,
		Channel:   chanID,
		Subtopic:  subtopic,
		Publisher: c.Username,
		Payload:   *payload,
		Created:   time.Now().UnixNano(),
	}

	for _, pub := range h.publishers {
		if err := pub.Publish(msg.Channel, msg); err != nil {
			h.logger.Error(errors.Wrap(errFailedPublishToMsgBroker, err).Error())
		}
	}
}

// Subscribe - after client successfully subscribed
func (h *handler) Subscribe(c *session.Client, topics *[]string) {
	if c == nil {
		h.logger.Error(errors.Wrap(errFailedSubscribe, errClientNotInitialized).Error())
		return
	}
	h.logger.Info(fmt.Sprintf(infoSubscribed, c.ID, strings.Join(*topics, ",")))
}

// Unsubscribe - after client unsubscribed
func (h *handler) Unsubscribe(c *session.Client, topics *[]string) {
	if c == nil {
		h.logger.Error(errors.Wrap(errFailedUnsubscribe, errClientNotInitialized).Error())
		return
	}
	h.logger.Info(fmt.Sprintf(infoUnsubscribed, c.ID, strings.Join(*topics, ",")))
}

// Disconnect - connection with broker or client lost
func (h *handler) Disconnect(c *session.Client) {
	if c == nil {
		h.logger.Error(errors.Wrap(errFailedDisconnect, errClientNotInitialized).Error())
		return
	}
	h.logger.Error(fmt.Sprintf(infoDisconnected, c.ID, c.Username))
	if err := h.es.Disconnect(c.Username); err != nil {
		h.logger.Error(errors.Wrap(errFailedPublishDisconnectEvent, err).Error())
	}
}

func (h *handler) authAccess(username string, topic string) error {
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>
	if !channelRegExp.Match([]byte(topic)) {
		h.logger.Info(infoMalformedTopic + topic)
		return errMalformedTopic
	}

	channelParts := channelRegExp.FindStringSubmatch(topic)
	if len(channelParts) < 1 {
		return errMalformedTopic
	}

	chanID := channelParts[1]
	return h.auth.Authorize(context.Background(), chanID, username)
}

func parseSubtopic(subtopic string) (string, error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", errMalformedSubtopic
	}
	subtopic = strings.Replace(subtopic, "/", ".", -1)

	elems := strings.Split(subtopic, ".")
	filteredElems := []string{}
	for _, elem := range elems {
		if elem == "" {
			continue
		}

		if len(elem) > 1 && (strings.Contains(elem, "*") || strings.Contains(elem, ">")) {
			return "", errMalformedSubtopic
		}

		filteredElems = append(filteredElems, elem)
	}

	subtopic = strings.Join(filteredElems, ".")
	return subtopic, nil
}
