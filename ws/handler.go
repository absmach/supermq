// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
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
	channelRegExp = regexp.MustCompile(`^\/?m\/([\w\-]+)\/c\/([\w\-]+)(\/[^?]*)?(\?.*)?$`)

	errMalformedSubtopic        = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("malformed subtopic"))
	errClientNotInitialized     = mgate.NewHTTPProxyError(http.StatusInternalServerError, errors.New("client is not initialized"))
	errMalformedTopic           = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("malformed topic"))
	errMissingTopicPub          = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to publish due to missing topic"))
	errMissingTopicSub          = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to subscribe due to missing topic"))
	errFailedPublishToMsgBroker = errors.New("failed to publish to supermq message broker")
	errFailedResolveDomain      = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to resolve domain route"))
	errFailedResolveChannel     = mgate.NewHTTPProxyError(http.StatusBadRequest, errors.New("failed to resolve channel route"))
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

	_, _, err := h.authAccess(ctx, token, *topic, connections.Publish)

	return err
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
		if _, _, err := h.authAccess(ctx, string(s.Password), topic, connections.Subscribe); err != nil {
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

	// Topics are in the format:
	// m/<domain>/c/<channel>/<subtopic>/.../ct/<content_type>
	channelParts := channelRegExp.FindStringSubmatch(*topic)
	if len(channelParts) < 3 {
		return errMalformedTopic
	}

	domainID, err := h.resolveDomain(ctx, channelParts[1])
	if err != nil {
		return errFailedResolveDomain
	}
	chanID, err := h.resolveChannel(ctx, channelParts[2], domainID)
	if err != nil {
		return errFailedResolveChannel
	}
	subtopic := channelParts[3]

	subtopic, err = parseSubtopic(subtopic)
	if err != nil {
		return err
	}

	clientID, clientType, err := h.authAccess(ctx, string(s.Password), *topic, connections.Publish)
	if err != nil {
		return err
	}

	msg := messaging.Message{
		Protocol: protocol,
		Domain:   domainID,
		Channel:  chanID,
		Subtopic: subtopic,
		Payload:  *payload,
		Created:  time.Now().UnixNano(),
	}

	if clientType == policies.ClientType {
		msg.Publisher = clientID
	}

	if err := h.pubsub.Publish(ctx, msg.GetChannel(), &msg); err != nil {
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

func (h *handler) authAccess(ctx context.Context, token, topic string, msgType connections.ConnType) (string, string, mgate.HTTPProxyError) {
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

	// Topics are in the format:
	// m/<domain>/c/<channel>/<subtopic>/.../ct/<content_type>
	if !channelRegExp.MatchString(topic) {
		return "", "", mgate.NewHTTPProxyError(http.StatusBadRequest, errMalformedTopic)
	}

	channelParts := channelRegExp.FindStringSubmatch(topic)
	if len(channelParts) < 3 {
		return "", "", mgate.NewHTTPProxyError(http.StatusBadRequest, errMalformedTopic)
	}

	domainID, err := h.resolveDomain(ctx, channelParts[1])
	if err != nil {
		return "", "", errFailedResolveDomain
	}
	chanID, err := h.resolveChannel(ctx, channelParts[2], domainID)
	if err != nil {
		return "", "", errFailedResolveChannel
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

func parseSubtopic(subtopic string) (string, mgate.HTTPProxyError) {
	if subtopic == "" {
		return subtopic, nil
	}

	subtopic, err := url.QueryUnescape(subtopic)
	if err != nil {
		return "", errMalformedSubtopic
	}
	subtopic = strings.ReplaceAll(subtopic, "/", ".")

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

// extractClientSecret returns value of the client secret. If there is no client key - an empty value is returned.
func extractClientSecret(token string) string {
	if !strings.HasPrefix(token, apiutil.ClientPrefix) {
		return ""
	}

	return strings.TrimPrefix(token, apiutil.ClientPrefix)
}
