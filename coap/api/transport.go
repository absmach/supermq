// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/absmach/supermq"
	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	api "github.com/absmach/supermq/api/http"
	"github.com/absmach/supermq/coap"
	"github.com/absmach/supermq/pkg/errors"
	svcerr "github.com/absmach/supermq/pkg/errors/service"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/go-chi/chi/v5"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	protocol     = "coap"
	authQuery    = "auth"
	startObserve = 0 // observe option value that indicates start of observation
)

var channelPartRegExp = regexp.MustCompile(`^/m/([\w\-]+)/c/([\w\-]+)(/[^?]*)?(\?.*)?$`)

const (
	numGroups     = 4 // entire expression+ domain group + channel group + subtopic group
	domainGroup   = 1 // domain group is first in channel regexp
	channelGroup  = 2 // channel group is third in channel regexp
	subtopicGroup = 3 // subtopic group is fourth in channel regexp
)

var (
	errMalformedSubtopic = errors.New("malformed subtopic")
	errBadOptions        = errors.New("bad options")
	errMethodNotAllowed  = errors.New("method not allowed")
)

var (
	logger   *slog.Logger
	service  coap.Service
	channels grpcChannelsV1.ChannelsServiceClient
	domains  grpcDomainsV1.DomainsServiceClient
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(instanceID string) http.Handler {
	b := chi.NewRouter()
	b.Get("/health", supermq.Health(protocol, instanceID))
	b.Handle("/metrics", promhttp.Handler())

	return b
}

// MakeCoAPHandler creates handler for CoAP messages.
func MakeCoAPHandler(svc coap.Service, channelsClient grpcChannelsV1.ChannelsServiceClient, domainsClient grpcDomainsV1.DomainsServiceClient, l *slog.Logger) mux.HandlerFunc {
	logger = l
	service = svc
	channels = channelsClient
	domains = domainsClient

	return handler
}

func sendResp(w mux.ResponseWriter, resp *pool.Message) {
	if err := w.Conn().WriteMessage(resp); err != nil {
		logger.Warn(fmt.Sprintf("Can't set response: %s", err))
	}
}

func handler(w mux.ResponseWriter, m *mux.Message) {
	resp := pool.NewMessage(w.Conn().Context())
	resp.SetToken(m.Token())
	for _, opt := range m.Options() {
		resp.AddOptionBytes(opt.ID, opt.Value)
	}
	defer sendResp(w, resp)

	msg, err := decodeMessage(m)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error decoding message: %s", err))
		resp.SetCode(codes.BadRequest)
		return
	}
	key, err := parseKey(m)
	if err != nil {
		logger.Warn(fmt.Sprintf("Error parsing auth: %s", err))
		resp.SetCode(codes.Unauthorized)
		return
	}

	switch m.Code() {
	case codes.GET:
		resp.SetCode(codes.Content)
		err = handleGet(m, w, msg, key)
	case codes.POST:
		resp.SetCode(codes.Created)
		err = service.Publish(m.Context(), key, msg)
	default:
		err = errMethodNotAllowed
	}

	if err != nil {
		switch {
		case err == errBadOptions:
			resp.SetCode(codes.BadOption)
		case err == errMethodNotAllowed:
			resp.SetCode(codes.MethodNotAllowed)
		case errors.Contains(err, svcerr.ErrAuthorization):
			resp.SetCode(codes.Forbidden)
		case errors.Contains(err, svcerr.ErrAuthentication):
			resp.SetCode(codes.Unauthorized)
		default:
			resp.SetCode(codes.InternalServerError)
		}
	}
}

func handleGet(m *mux.Message, w mux.ResponseWriter, msg *messaging.Message, key string) error {
	var obs uint32
	obs, err := m.Options().Observe()
	if err != nil {
		logger.Warn(fmt.Sprintf("Error reading observe option: %s", err))
		return errBadOptions
	}
	if obs == startObserve {
		c := coap.NewClient(w.Conn(), m.Token(), logger)
		w.Conn().AddOnClose(func() {
			_ = service.DisconnectHandler(context.Background(), msg.GetChannel(), msg.GetSubtopic(), c.Token())
		})
		return service.Subscribe(w.Conn().Context(), key, msg.GetDomain(), msg.GetChannel(), msg.GetSubtopic(), c)
	}
	return service.Unsubscribe(w.Conn().Context(), key, msg.GetDomain(), msg.GetChannel(), msg.GetSubtopic(), m.Token().String())
}

func decodeMessage(msg *mux.Message) (*messaging.Message, error) {
	if msg.Options() == nil {
		return &messaging.Message{}, errBadOptions
	}
	path, err := msg.Path()
	if err != nil {
		return &messaging.Message{}, err
	}
	channelParts := channelPartRegExp.FindStringSubmatch(path)
	if len(channelParts) < numGroups {
		return &messaging.Message{}, errMalformedSubtopic
	}

	domainID, err := resolveDomain(channelParts[domainGroup])
	if err != nil {
		return &messaging.Message{}, err
	}
	channelID, err := resolveChannel(channelParts[channelGroup], domainID)
	if err != nil {
		return &messaging.Message{}, err
	}

	st, err := parseSubtopic(channelParts[subtopicGroup])
	if err != nil {
		return &messaging.Message{}, err
	}
	ret := &messaging.Message{
		Protocol: protocol,
		Domain:   domainID,
		Channel:  channelID,
		Subtopic: st,
		Payload:  []byte{},
		Created:  time.Now().UnixNano(),
	}

	if msg.Body() != nil {
		buff, err := io.ReadAll(msg.Body())
		if err != nil {
			return ret, err
		}
		ret.Payload = buff
	}
	return ret, nil
}

func parseKey(msg *mux.Message) (string, error) {
	authKey, err := msg.Options().GetString(message.URIQuery)
	if err != nil {
		return "", err
	}
	vars := strings.Split(authKey, "=")
	if len(vars) != 2 || vars[0] != authQuery {
		return "", svcerr.ErrAuthorization
	}
	return vars[1], nil
}

func parseSubtopic(subtopic string) (string, error) {
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

func resolveDomain(domain string) (string, error) {
	if api.ValidateUUID(domain) == nil {
		return domain, nil
	}

	d, err := domains.RetrieveByRoute(context.Background(), &grpcCommonV1.RetrieveByRouteReq{
		Route: domain,
	})
	if err != nil {
		return "", err
	}

	return d.Entity.Id, nil
}

func resolveChannel(channel, domainID string) (string, error) {
	if api.ValidateUUID(channel) == nil {
		return channel, nil
	}

	c, err := channels.RetrieveByRoute(context.Background(), &grpcCommonV1.RetrieveByRouteReq{
		Route:    channel,
		DomainId: domainID,
	})
	if err != nil {
		return "", err
	}

	return c.Entity.Id, nil
}
