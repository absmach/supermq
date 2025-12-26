// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/absmach/mgate/pkg/session"
	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	"github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

var _ session.Handler = (*handler)(nil)

const protocol = "mqtt"

// Log message formats.
const (
	LogInfoSubscribed   = "subscribed with client_id %s to topics %s"
	LogInfoUnsubscribed = "unsubscribed client_id %s from topics %s"
	LogInfoConnected    = "connected with client_id %s"
	LogInfoDisconnected = "disconnected client_id %s and username %s"
	LogInfoPublished    = "published with client_id %s to the topic %s"
)

// Error wrappers for MQTT errors.
var (
	ErrClientNotInitialized         = errors.New("client is not initialized")
	ErrMissingClientID              = errors.New("client_id not found")
	ErrMissingTopicPub              = errors.New("failed to publish due to missing topic")
	ErrMissingTopicSub              = errors.New("failed to subscribe due to missing topic")
	ErrFailedConnect                = errors.New("failed to connect")
	ErrFailedSubscribe              = errors.New("failed to subscribe")
	ErrFailedUnsubscribe            = errors.New("failed to unsubscribe")
	ErrFailedPublish                = errors.New("failed to publish")
	ErrFailedDisconnect             = errors.New("failed to disconnect")
	ErrFailedPublishDisconnectEvent = errors.New("failed to publish disconnect event")
	ErrFailedPublishConnectEvent    = errors.New("failed to publish connect event")
	ErrFailedSubscribeEvent         = errors.New("failed to publish subscribe event")
	ErrFailedPublishToMsgBroker     = errors.New("failed to publish to supermq message broker")

	errInvalidUserId = errors.New("invalid user id")
)

// Event implements events.Event interface.
type handler struct {
	publisher messaging.Publisher
	clients   grpcClientsV1.ClientsServiceClient
	channels  grpcChannelsV1.ChannelsServiceClient
	parser    messaging.TopicParser
	logger    *slog.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(publisher messaging.Publisher, logger *slog.Logger, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, parser messaging.TopicParser) session.Handler {
	return &handler{
		logger:    logger,
		publisher: publisher,
		clients:   clients,
		channels:  channels,
		parser:    parser,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker.
func (h *handler) AuthConnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}

	if s.ID == "" {
		return ErrMissingClientID
	}

	pwd := string(s.Password)

	res, err := h.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{Token: authn.AuthPack(authn.BasicAuth, s.Username, pwd)})
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthentication, err)
	}
	if !res.GetAuthenticated() {
		return svcerr.ErrAuthentication
	}

	if s.Username != "" && res.GetId() != s.Username {
		return errInvalidUserId
	}

	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return ErrMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}

	domainID, chanID, _, topicType, err := h.parser.ParsePublishTopic(ctx, *topic, false)
	if err != nil {
		return err
	}

	return h.authAccess(ctx, string(s.Username), domainID, chanID, connections.Publish, topicType)
}

// AuthSubscribe is called on device subscribe,
// prior forwarding to the MQTT broker.
func (h *handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return ErrClientNotInitialized
	}
	if topics == nil || *topics == nil {
		return ErrMissingTopicSub
	}

	for _, topic := range *topics {
		domainID, chanID, _, topicType, err := h.parser.ParseSubscribeTopic(ctx, topic, false)
		if err != nil {
			return err
		}

		if err := h.authAccess(ctx, string(s.Username), domainID, chanID, connections.Subscribe, topicType); err != nil {
			return err
		}
	}

	return nil
}

// Connect - after client successfully connected.
func (h *handler) Connect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedConnect, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoConnected, s.ID))
	return nil
}

// Publish - after client successfully published.
func (h *handler) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	return nil
}

// Subscribe - after client successfully subscribed.
func (h *handler) Subscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Unsubscribe - after client unsubscribed.
func (h *handler) Unsubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedUnsubscribe, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoUnsubscribed, s.ID, strings.Join(*topics, ",")))

	return nil
}

// Disconnect - connection with broker or client lost.
func (h *handler) Disconnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(ErrFailedDisconnect, ErrClientNotInitialized)
	}
	h.logger.Info(fmt.Sprintf(LogInfoDisconnected, s.ID, s.Username))

	return nil
}

func (h *handler) authAccess(ctx context.Context, clientID, domainID, chanID string, msgType connections.ConnType, topicType messaging.TopicType) error {
	switch topicType {
	case messaging.HealthType:
		return nil
	default:
		ar := &grpcChannelsV1.AuthzReq{
			Type:       uint32(msgType),
			ClientId:   clientID,
			ClientType: policies.ClientType,
			ChannelId:  chanID,
			DomainId:   domainID,
		}
		res, err := h.channels.Authorize(ctx, ar)
		if err != nil {
			return err
		}
		if !res.GetAuthorized() {
			return svcerr.ErrAuthorization
		}

		return nil
	}
}
