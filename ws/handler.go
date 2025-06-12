// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	mgate "github.com/absmach/mgate/pkg/http"
	"github.com/absmach/mgate/pkg/session"
	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	api "github.com/absmach/supermq/api/http"
	apiutil "github.com/absmach/supermq/api/http/util"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/policies"
)

var _ session.Handler = (*handler)(nil)

const protocol = "websocket"

// Log message formats.
const (
	LogInfoSubscribed   = "subscribed with client_id %s to topics %s"
	LogInfoConnected    = "connected with client_id %s"
	LogInfoDisconnected = "disconnected client_id %s and username %s"
	LogInfoPublished    = "published with client_id %s to the topic %s"
)

// Error wrappers for MQTT errors.
var (
	errClientNotInitialized     = errors.New("client is not initialized")
	errMissingTopicPub          = errors.New("failed to publish due to missing topic")
	errMissingTopicSub          = errors.New("failed to subscribe due to missing topic")
	errFailedPublish            = errors.New("failed to publish")
	errFailedPublishToMsgBroker = errors.New("failed to publish to supermq message broker")
)

// Event implements events.Event interface.
type handler struct {
	pubsub   messaging.PubSub
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
	domains  grpcDomainsV1.DomainsServiceClient
	authn    smqauthn.Authentication
	logger   *slog.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(pubsub messaging.PubSub, logger *slog.Logger, authn smqauthn.Authentication, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, domains grpcDomainsV1.DomainsServiceClient) session.Handler {
	return &handler{
		logger:   logger,
		pubsub:   pubsub,
		authn:    authn,
		clients:  clients,
		channels: channels,
		domains:  domains,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the ws server.
func (h *handler) AuthConnect(ctx context.Context) error {
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the ws server.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return errMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}

	var token string
	switch {
	case strings.HasPrefix(string(s.Password), "Client"):
		token = strings.ReplaceAll(string(s.Password), "Client ", "")
	default:
		token = string(s.Password)
	}

	domain, channel, _, err := messaging.ParsePublishTopic(*topic)
	if err != nil {
		return err
	}
	domainID, err := h.resolveDomain(ctx, domain)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveDomain, err))
	}
	chanID, err := h.resolveChannel(ctx, channel, domainID)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveChannel, err))
	}

	clientID, clientType, err := h.authAccess(ctx, token, domainID, chanID, connections.Publish)
	if err != nil {
		return err
	}

	if s.Username == "" && clientType == policies.ClientType {
		s.Username = clientID
	}

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
		domain, channel, _, err := messaging.ParseSubscribeTopic(topic)
		if err != nil {
			return err
		}
		domainID, err := h.resolveDomain(ctx, domain)
		if err != nil {
			return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveDomain, err))
		}
		chanID, err := h.resolveChannel(ctx, channel, domainID)
		if err != nil {
			return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveChannel, err))
		}
		if _, _, err := h.authAccess(ctx, string(s.Password), domainID, chanID, connections.Subscribe); err != nil {
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

	domain, channel, subtopic, err := messaging.ParsePublishTopic(*topic)
	if err != nil {
		return errors.Wrap(errFailedPublish, err)
	}
	domainID, err := h.resolveDomain(ctx, domain)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveDomain, err))
	}
	chanID, err := h.resolveChannel(ctx, channel, domainID)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(errFailedResolveChannel, err))
	}

	msg := messaging.Message{
		Protocol:  protocol,
		Domain:    domainID,
		Channel:   chanID,
		Subtopic:  subtopic,
		Payload:   *payload,
		Publisher: s.Username,
		Created:   time.Now().UnixNano(),
	}

	if err := h.pubsub.Publish(ctx, messaging.EncodeMessageTopic(&msg), &msg); err != nil {
		return mgate.NewHTTPProxyError(http.StatusInternalServerError, errors.Wrap(errFailedPublishToMsgBroker, err))
	}

	h.logger.Info(fmt.Sprintf(LogInfoPublished, s.ID, *topic))

	return nil
}

// Subscribe - after client successfully subscribed.
func (h *handler) Subscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}
	h.logger.Info(fmt.Sprintf(LogInfoSubscribed, s.ID, strings.Join(*topics, ",")))
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

func (h *handler) authAccess(ctx context.Context, token, domainID, chanID string, msgType connections.ConnType) (string, string, error) {
	authnReq := &grpcClientsV1.AuthnReq{
		ClientSecret: token,
	}
	if strings.HasPrefix(token, "Client") {
		authnReq.ClientSecret = extractClientSecret(token)
	}

	authnRes, err := h.clients.Authenticate(ctx, authnReq)
	if err != nil {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, errors.Wrap(svcerr.ErrAuthentication, err))
	}
	if !authnRes.GetAuthenticated() {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}
	clientType := policies.ClientType
	clientID := authnRes.GetId()

	ar := &grpcChannelsV1.AuthzReq{
		Type:       uint32(msgType),
		ClientId:   clientID,
		ClientType: clientType,
		ChannelId:  chanID,
		DomainId:   domainID,
	}
	res, err := h.channels.Authorize(ctx, ar)
	if err != nil {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, errors.Wrap(svcerr.ErrAuthentication, err))
	}
	if !res.GetAuthorized() {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}

	return clientID, clientType, nil
}

func (h *handler) resolveDomain(ctx context.Context, domain string) (string, error) {
	if api.ValidateUUID(domain) == nil {
		return domain, nil
	}
	d, err := h.domains.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route: domain,
	})
	if err != nil {
		return "", err
	}

	return d.Entity.Id, nil
}

func (h *handler) resolveChannel(ctx context.Context, channel, domainID string) (string, error) {
	if api.ValidateUUID(channel) == nil {
		return channel, nil
	}
	c, err := h.channels.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route:    channel,
		DomainId: domainID,
	})
	if err != nil {
		return "", err
	}

	return c.Entity.Id, nil
}

// extractClientSecret returns value of the client secret. If there is no client key - an empty value is returned.
func extractClientSecret(token string) string {
	if !strings.HasPrefix(token, apiutil.ClientPrefix) {
		return ""
	}

	return strings.TrimPrefix(token, apiutil.ClientPrefix)
}
