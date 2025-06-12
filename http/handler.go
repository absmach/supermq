// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

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

type ctxKey string

const (
	protocol                = "http"
	clientIDCtxKey   ctxKey = "client_id"
	clientTypeCtxKey ctxKey = "client_type"
)

// Log message formats.
const (
	logInfoConnected         = "connected with client_key %s"
	logInfoPublished         = "published with client_type %s client_id %s to the topic %s"
	logInfoFailedAuthNToken  = "failed to authenticate token for topic %s with error %s"
	logInfoFailedAuthNClient = "failed to authenticate client key %s for topic %s with error %s"
)

// Error wrappers for MQTT errors.
var (
	errClientNotInitialized     = errors.New("client is not initialized")
	errFailedPublish            = errors.New("failed to publish")
	errFailedPublishToMsgBroker = errors.New("failed to publish to supermq message broker")
	errMalformedTopic           = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("malformed topic"))
	errMissingTopicPub          = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to publish due to missing topic"))
	errFailedResolveDomain      = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to resolve domain route"))
	errFailedResolveChannel     = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to resolve channel route"))
)

// Event implements events.Event interface.
type handler struct {
	publisher messaging.Publisher
	clients   grpcClientsV1.ClientsServiceClient
	channels  grpcChannelsV1.ChannelsServiceClient
	domains   grpcDomainsV1.DomainsServiceClient
	authn     smqauthn.Authentication
	logger    *slog.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(publisher messaging.Publisher, authn smqauthn.Authentication, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, domains grpcDomainsV1.DomainsServiceClient, logger *slog.Logger) session.Handler {
	return &handler{
		publisher: publisher,
		authn:     authn,
		clients:   clients,
		channels:  channels,
		domains:   domains,
		logger:    logger,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the HTTP server.
func (h *handler) AuthConnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}

	var tok string
	switch {
	case string(s.Password) == "":
		return mgate.NewHTTPProxyError(http.StatusBadRequest, errors.Wrap(apiutil.ErrValidation, apiutil.ErrBearerKey))
	case strings.HasPrefix(string(s.Password), apiutil.ClientPrefix):
		tok = strings.TrimPrefix(string(s.Password), apiutil.ClientPrefix)
	default:
		tok = string(s.Password)
	}

	h.logger.Info(fmt.Sprintf(logInfoConnected, tok))
	return nil
}

// AuthPublish is not used in HTTP service.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	return nil
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
	if topic == nil {
		return errMissingTopicPub
	}
	topic = &strings.Split(*topic, "?")[0]
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(errFailedPublish, errClientNotInitialized)
	}

	domain, channel, subtopic, err := messaging.ParsePublishTopic(*topic)
	if err != nil {
		return errors.Wrap(errMalformedTopic, err)
	}
	domainID, err := h.resolveDomain(ctx, domain)
	if err != nil {
		return errors.Wrap(errFailedResolveDomain, err)
	}
	channelID, err := h.resolveChannel(ctx, channel, domainID)
	if err != nil {
		return errors.Wrap(errFailedResolveChannel, err)
	}

	var clientID, clientType string
	switch {
	case strings.HasPrefix(string(s.Password), "Client"):
		secret := strings.TrimPrefix(string(s.Password), apiutil.ClientPrefix)
		authnRes, err := h.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{ClientSecret: secret})
		if err != nil {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNClient, secret, *topic, err))
			return mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		if !authnRes.Authenticated {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNClient, secret, *topic, svcerr.ErrAuthentication))
			return mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		clientType = policies.ClientType
		clientID = authnRes.GetId()
	case strings.HasPrefix(string(s.Password), apiutil.BearerPrefix):
		token := strings.TrimPrefix(string(s.Password), apiutil.BearerPrefix)
		authnSession, err := h.authn.Authenticate(ctx, token)
		if err != nil {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNToken, *topic, err))
			return mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		clientType = policies.UserType
		clientID = authnSession.DomainUserID
	default:
		return mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
	}

	msg := messaging.Message{
		Protocol: protocol,
		Domain:   domainID,
		Channel:  channelID,
		Subtopic: subtopic,
		Payload:  *payload,
		Created:  time.Now().UnixNano(),
	}

	ar := &grpcChannelsV1.AuthzReq{
		DomainId:   domainID,
		ClientId:   clientID,
		ClientType: clientType,
		ChannelId:  msg.Channel,
		Type:       uint32(connections.Publish),
	}
	res, err := h.channels.Authorize(ctx, ar)
	if err != nil {
		return mgate.NewHTTPProxyError(http.StatusBadRequest, err)
	}
	if !res.GetAuthorized() {
		return mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthorization)
	}

	if clientType == policies.ClientType {
		msg.Publisher = clientID
	}

	if err := h.publisher.Publish(ctx, messaging.EncodeMessageTopic(&msg), &msg); err != nil {
		return errors.Wrap(errFailedPublishToMsgBroker, err)
	}

	h.logger.Info(fmt.Sprintf(logInfoPublished, clientType, clientID, *topic))

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
