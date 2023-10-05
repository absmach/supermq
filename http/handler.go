// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/things/policies"
	"github.com/mainflux/mproxy/pkg/session"
)

var _ session.Handler = (*handler)(nil)

const protocol = "http"

// Log message formats.
const (
	LogInfoConnected = "connected with client_id %s"
	// ThingPrefix represents the key prefix for Thing authentication scheme.
	ThingPrefix      = "Thing "
	LogInfoPublished = "published with client_id %s to the topic %s"
)

// Error wrappers for MQTT errors.
var (
	ErrMalformedSubtopic         = errors.New("malformed subtopic")
	ErrClientNotInitialized      = errors.New("client is not initialized")
	ErrMalformedTopic            = errors.New("malformed topic")
	ErrMissingTopicPub           = errors.New("failed to publish due to missing topic")
	ErrMissingTopicSub           = errors.New("failed to subscribe due to missing topic")
	ErrFailedConnect             = errors.New("failed to connect")
	ErrFailedPublish             = errors.New("failed to publish")
	ErrFailedParseSubtopic       = errors.New("failed to parse subtopic")
	ErrFailedPublishConnectEvent = errors.New("failed to publish connect event")
	ErrFailedPublishToMsgBroker  = errors.New("failed to publish to mainflux message broker")
)

var channelRegExp = regexp.MustCompile(`^\/?channels\/([\w\-]+)\/messages(\/[^?]*)?(\?.*)?$`)

// Event implements events.Event interface.
type handler struct {
	publisher messaging.Publisher
	auth      policies.AuthServiceClient
	logger    logger.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(publisher messaging.Publisher, logger logger.Logger, auth policies.AuthServiceClient) session.Handler {
	return &handler{
		logger:    logger,
		publisher: publisher,
		auth:      auth,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the HTTP server.
func (h *handler) AuthConnect(ctx context.Context) error {
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the HTTP broker.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return ErrMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}

	tok := string(s.Password)
	if strings.HasPrefix(tok, "Thing") {
		tok = extractThingKey(tok)
	}

	return h.authAccess(ctx, tok, *topic, policies.WriteAction)
}

// AuthSubscribe is not used in HTTP service.
func (h *handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Connect - after client successfully connected.
func (h *handler) Connect(ctx context.Context) error {
	return nil
}

// Publish - after client successfully published.
func (h *handler) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	topic = &strings.Split(*topic, "?")[0]
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedPublish, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoPublished, s.ID, *topic))
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>

	channelParts := channelRegExp.FindStringSubmatch(*topic)
	if len(channelParts) < 2 {
		return errors.Wrap(ErrFailedPublish, ErrMalformedTopic)
	}

	chanID := channelParts[1]
	subtopic := channelParts[2]

	subtopic, err := parseSubtopic(subtopic)
	if err != nil {
		return errors.Wrap(ErrFailedParseSubtopic, err)
	}

	msg := messaging.Message{
		Protocol: protocol,
		Channel:  chanID,
		Subtopic: subtopic,
		Payload:  *payload,
		Created:  time.Now().UnixNano(),
	}
	tok := string(s.Password)
	if strings.HasPrefix(tok, "Thing") {
		tok = extractThingKey(tok)
	}
	ar := &policies.AuthorizeReq{
		Subject:    tok,
		Object:     msg.Channel,
		Action:     policies.WriteAction,
		EntityType: policies.ThingEntityType,
	}
	res, err := h.auth.Authorize(ctx, ar)
	if err != nil {
		return err
	}
	if !res.GetAuthorized() {
		return errors.ErrAuthorization
	}
	msg.Publisher = res.GetThingID()

	if err := h.publisher.Publish(ctx, msg.Channel, &msg); err != nil {
		return errors.Wrap(ErrFailedPublishToMsgBroker, err)
	}

	return nil
}

// Subscribe - not used for HTTP.
func (h *handler) Subscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Unsubscribe - not used for HTTP.
func (h *handler) Unsubscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Disconnect - not used for HTTP.
func (h *handler) Disconnect(ctx context.Context) error {
	return nil
}

func (h *handler) authAccess(ctx context.Context, password, topic, action string) error {
	// Topics are in the format:
	// channels/<channel_id>/messages/<subtopic>/.../ct/<content_type>
	if !channelRegExp.Match([]byte(topic)) {
		return ErrMalformedTopic
	}

	channelParts := channelRegExp.FindStringSubmatch(topic)
	if len(channelParts) < 1 {
		return ErrMalformedTopic
	}

	chanID := channelParts[1]

	ar := &policies.AuthorizeReq{
		Subject:    password,
		Object:     chanID,
		Action:     action,
		EntityType: policies.ThingEntityType,
	}
	res, err := h.auth.Authorize(ctx, ar)
	if err != nil {
		return err
	}
	if !res.GetAuthorized() {
		return errors.ErrAuthorization
	}

	return nil
}

func parseSubtopic(subtopic string) (string, error) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", ErrMalformedSubtopic
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

// extractThingKey returns value of the thing key. If there is no thing key - an empty value is returned.
func extractThingKey(topic string) string {
	if !strings.HasPrefix(topic, ThingPrefix) {
		return ""
	}

	return strings.TrimPrefix(topic, ThingPrefix)
}
