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
	LogInfoPublished         = "published with client_id %s to the topic %s"
	LogInfoSubscribed        = "subscribed with client_id %s to topics %s"
	logInfoFailedAuthNToken  = "failed to authenticate token with error %s"
	logInfoFailedAuthNClient = "failed to authenticate client key %s with error %s"
)

// Error wrappers for MQTT errors.
var (
	errClientNotInitialized     = errors.New("client is not initialized")
	errFailedPublish            = errors.New("failed to publish")
	errFailedPublishToMsgBroker = errors.New("failed to publish to supermq message broker")
	errMalformedTopic           = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("malformed topic"))
	errMissingTopicPub          = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to publish due to missing topic"))
	errMissingTopicSub          = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to subscribe due to missing topic"))
)

// Event implements events.Event interface.
type handler struct {
	publisher messaging.Publisher
	clients   grpcClientsV1.ClientsServiceClient
	channels  grpcChannelsV1.ChannelsServiceClient
	authn     smqauthn.Authentication
	logger    *slog.Logger
}

// NewHandler creates new Handler entity.
func NewHandler(publisher messaging.Publisher, authn smqauthn.Authentication, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient, logger *slog.Logger) session.Handler {
	return &handler{
		publisher: publisher,
		authn:     authn,
		clients:   clients,
		channels:  channels,
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

// AuthPublish is called on device publish,
// prior forwarding to the HTTP server.
func (h *handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	if topic == nil {
		return errMissingTopicPub
	}
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}

	domainID, chanID, _, err := messaging.ParsePublishTopic(*topic)
	if err != nil {
		return err
	}

	clientID, clientType, err := h.authAccess(ctx, string(s.Password), domainID, chanID, connections.Publish)
	if err != nil {
		return err
	}

	if s.Username == "" && clientType == policies.ClientType {
		s.Username = clientID
	}
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the HTTP server.
func (h *handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		return errClientNotInitialized
	}
	if topics == nil || *topics == nil {
		return errMissingTopicSub
	}

	for _, topic := range *topics {
		domainID, chanID, _, err := messaging.ParseSubscribeTopic(topic)
		if err != nil {
			return err
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
	if topic == nil {
		return errMissingTopicPub
	}
	topic = &strings.Split(*topic, "?")[0]
	s, ok := session.FromContext(ctx)
	if !ok {
		return errors.Wrap(errFailedPublish, errClientNotInitialized)
	}
	if payload == nil || len(*payload) == 0 {
		h.logger.Warn("Empty payload, not publishing to broker", slog.String("client_id", s.Username))
		return nil
	}

	domainID, chanID, subtopic, err := messaging.ParsePublishTopic(*topic)
	if err != nil {
		return errors.Wrap(errMalformedTopic, err)
	}

	msg := messaging.Message{
		Protocol: protocol,
		Domain:   domainID,
		Channel:  chanID,
		Subtopic: subtopic,
		Payload:  *payload,
		Created:  time.Now().UnixNano(),
	}

	if err := h.publisher.Publish(ctx, messaging.EncodeMessageTopic(&msg), &msg); err != nil {
		return errors.Wrap(errFailedPublishToMsgBroker, err)
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

// Unsubscribe - not used for HTTP.
func (h *handler) Unsubscribe(ctx context.Context, topics *[]string) error {
	return nil
}

// Disconnect - not used for HTTP.
func (h *handler) Disconnect(ctx context.Context) error {
	return nil
}

func (h *handler) authAccess(ctx context.Context, token, domainID, chanID string, msgType connections.ConnType) (string, string, error) {
	var clientID, clientType, secret string
	switch {
	case strings.HasPrefix(string(token), apiutil.BearerPrefix):
		token := strings.TrimPrefix(string(token), apiutil.BearerPrefix)
		authnSession, err := h.authn.Authenticate(ctx, token)
		if err != nil {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNToken, err))
			return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		clientType = policies.UserType
		clientID = authnSession.DomainUserID
	default:
		if token == "" {
			return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		secret = token
		if strings.HasPrefix(string(token), "Client") {
			secret = strings.TrimPrefix(string(token), apiutil.ClientPrefix)
		}
		authnRes, err := h.clients.Authenticate(ctx, &grpcClientsV1.AuthnReq{ClientSecret: secret})
		if err != nil {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNClient, secret, err))
			return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		if !authnRes.Authenticated {
			h.logger.Info(fmt.Sprintf(logInfoFailedAuthNClient, secret, svcerr.ErrAuthentication))
			return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthentication)
		}
		clientType = policies.ClientType
		clientID = authnRes.GetId()
	}

	ar := &grpcChannelsV1.AuthzReq{
		Type:       uint32(msgType),
		ClientId:   clientID,
		ClientType: clientType,
		ChannelId:  chanID,
		DomainId:   domainID,
	}
	res, err := h.channels.Authorize(ctx, ar)
	if err != nil {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, errors.Wrap(svcerr.ErrAuthorization, err))
	}
	if !res.GetAuthorized() {
		return "", "", mgate.NewHTTPProxyError(http.StatusUnauthorized, svcerr.ErrAuthorization)
	}

	return clientID, clientType, nil
}
