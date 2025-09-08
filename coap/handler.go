// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
	mgate "github.com/felixgateru/mgate/pkg/http"
	"github.com/felixgateru/mgate/pkg/session"
)

var _ session.Handler = (*handler)(nil)

// Log message formats.
const (
	LogInfoSubscribed   = "subscribed with client_id %s to topics %s"
	LogInfoConnected    = "connected with client_id %s"
	LogInfoDisconnected = "disconnected client_id %s and username %s"
	LogInfoPublished    = "published with client_id %s to the topic %s"
)

// Error wrappers for MQTT errors.
var (
	errClientNotInitialized = errors.New("client is not initialized")
	errMissingTopicPub      = errors.New("failed to publish due to missing topic")
	errMissingTopicSub      = errors.New("failed to subscribe due to missing topic")
	errFailedPublish        = errors.New("failed to publish")
)

// Event implements events.Event interface.
type handler struct {
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	logger   *slog.Logger
	parser   messaging.TopicParser
}

// NewHandler creates new Handler entity.
func NewHandler(logger *slog.Logger, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, parser messaging.TopicParser) session.Handler {
	return &handler{
		logger:   logger,
		clients:  clients,
		channels: channels,
		parser:   parser,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the coap server.
func (h *handler) AuthConnect(ctx context.Context) error {
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the coap server.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return errMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}

	domainID, channelID, _, err := h.parser.ParsePublishTopic(ctx, *topic, true)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedPublish, err))
	}

	clientID, err := h.authAccess(ctx, string(s.Password), domainID, channelID, connections.Publish)
	if err != nil {
		return err
	}
	s.Username = clientID

	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker.
func (h *handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}
	if topics == nil || *topics == nil {
		return errMissingTopicSub
	}

	for _, topic := range *topics {
		domainID, channelID, _, err := h.parser.ParseSubscribeTopic(ctx, topic, true)
		if err != nil {
			return err
		}
		if _, err := h.authAccess(ctx, string(s.Password), domainID, channelID, connections.Subscribe); err != nil {
			return err
		}
	}
	return nil
}

// Connect - after client successfully connected.
func (h *handler) Connect(ctx context.Context) error {
	return nil
}

// Publish - after client successfully published.
func (h *handler) Publish(ctx context.Context, topic *string, payload *[]byte) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}

	if len(*payload) == 0 {
		h.logger.Warn("Empty payload, not publishing to broker", slog.String("client_id", s.Username))
		return nil
	}

	h.logger.Info(fmt.Sprintf(LogInfoPublished, s.Username, *topic))

	return nil
}

// Subscribe - after client successfully subscribed.
func (h *handler) Subscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}
	h.logger.Info(fmt.Sprintf(LogInfoSubscribed, s.Username, strings.Join(*topics, ",")))
	return nil
}

// Unsubscribe - after client unsubscribed.
func (h *handler) Unsubscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Disconnect - connection with broker or client lost.
func (h *handler) Disconnect(ctx context.Context) error {
	return nil
}

func (h *handler) authAccess(ctx context.Context, secret, domainID, chanID string, msgType connections.ConnType) (string, error) {
	authnRes, err := h.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{Token: smqauthn.AuthPack(smqauthn.DomainAuth, domainID, secret)})
	if err != nil {
		return "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}
	if !authnRes.Authenticated {
		return "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}

	ar := &grpcChannelsV1.AuthzReq{
		Type:       uint32(msgType),
		ClientId:   authnRes.GetId(),
		ClientType: policies.ClientType,
		ChannelId:  chanID,
		DomainId:   domainID,
	}
	res, err := h.channels.Authorize(ctx, ar)
	if err != nil {
		return "", mgate.NewHTTPProxyError(http.StatusUnauthorized, errors.Wrap(svcerr.ErrAuthentication, err))
	}
	if !res.GetAuthorized() {
		return "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}

	return authnRes.GetId(), nil
}
