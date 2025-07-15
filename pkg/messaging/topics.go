// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcCommonV1 "github.com/absmach/supermq/api/grpc/common/v1"
	grpcDomainsV1 "github.com/absmach/supermq/api/grpc/domains/v1"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/dgraph-io/ristretto/v2"
	"github.com/gofrs/uuid/v5"
)

const (
	MsgTopicPrefix     = 'm'
	ChannelTopicPrefix = 'c'
)

var (
	mqWildcards          = "+#"
	wildcards            = "*>"
	subtopicInvalidChars = " #+"
	wildcardsReplacer    = strings.NewReplacer("+", "*", "#", ">")
	pathReplacer         = strings.NewReplacer("/", ".")

	DefaultCacheConfig = CacheConfig{
		NumCounters: 2e5,     // 200k
		MaxCost:     1 << 20, // 1MB
		BufferItems: 64,
	}

	ErrMalformedTopic       = errors.New("malformed topic")
	ErrMalformedSubtopic    = errors.New("malformed subtopic")
	ErrEmptyRouteID         = errors.New("empty route or id")
	ErrFailedResolveDomain  = errors.New("failed to resolve domain route")
	ErrFailedResolveChannel = errors.New("failed to resolve channel route")
	ErrCreateCache          = errors.New("failed to create cache")
)

type CacheConfig struct {
	NumCounters int64 `env:"NUM_COUNTERS" envDefault:"200000"`  // number of keys to track frequency of.
	MaxCost     int64 `env:"MAX_COST"     envDefault:"1048576"` // maximum cost of cache.
	BufferItems int64 `env:"BUFFER_ITEMS" envDefault:"64"`      // number of keys per Get buffer.
}

type parsedTopic struct {
	domainID  string
	channelID string
	subtopic  string
	err       error
}

// TopicParser defines methods for parsing publish and subscribe topics.
// It uses a cache to store parsed topics for quick retrieval.
// It also resolves domain and channel IDs if requested.
type TopicParser interface {
	ParsePublishTopic(ctx context.Context, topic string, resolve bool) (domainID, channelID, subtopic string, err error)
	ParseSubscribeTopic(ctx context.Context, topic string, resolve bool) (domainID, channelID, subtopic string, err error)
}

type parser struct {
	resolver TopicResolver
	cache    *ristretto.Cache[string, *parsedTopic]
}

// NewTopicParser creates a new instance of TopicParser.
func NewTopicParser(cfg CacheConfig, channels grpcChannelsV1.ChannelsServiceClient, domains grpcDomainsV1.DomainsServiceClient) (TopicParser, error) {
	cache, err := ristretto.NewCache(&ristretto.Config[string, *parsedTopic]{
		NumCounters: cfg.NumCounters,
		MaxCost:     cfg.MaxCost,
		BufferItems: cfg.BufferItems,
		Cost:        costFunc,
	})
	if err != nil {
		return nil, errors.Wrap(ErrCreateCache, err)
	}
	return &parser{
		cache:    cache,
		resolver: NewTopicResolver(channels, domains),
	}, nil
}

func (p *parser) ParsePublishTopic(ctx context.Context, topic string, resolve bool) (string, string, string, error) {
	val, ok := p.cache.Get(topic)
	if ok {
		return val.domainID, val.channelID, val.subtopic, val.err
	}
	domainID, channelID, subtopic, err := ParsePublishTopic(topic)
	if err != nil {
		p.saveToCache(topic, "", "", "", err)
		return "", "", "", err
	}
	var isRoute bool
	if resolve {
		domainID, channelID, isRoute, err = p.resolver.Resolve(ctx, domainID, channelID)
		if err != nil {
			return "", "", "", err
		}
	}
	if !isRoute {
		p.saveToCache(topic, domainID, channelID, subtopic, nil)
	}

	return domainID, channelID, subtopic, nil
}

func (p *parser) ParseSubscribeTopic(ctx context.Context, topic string, resolve bool) (string, string, string, error) {
	domainID, channelID, subtopic, err := ParseSubscribeTopic(topic)
	if err != nil {
		return "", "", "", err
	}
	if resolve {
		domainID, channelID, _, err = p.resolver.Resolve(ctx, domainID, channelID)
		if err != nil {
			return "", "", "", err
		}
	}

	return domainID, channelID, subtopic, nil
}

func (p *parser) saveToCache(topic string, domainID, channelID, subtopic string, err error) {
	p.cache.Set(topic, &parsedTopic{
		domainID:  domainID,
		channelID: channelID,
		subtopic:  subtopic,
		err:       err,
	}, 0)
}

func costFunc(val *parsedTopic) int64 {
	errLen := 0
	if val.err != nil {
		errLen = len(val.err.Error())
	}
	cost := int64(len(val.domainID) + len(val.channelID) + len(val.subtopic) + errLen)

	return cost
}

// TopicResolver contains definitions for resolving domain and channel IDs
// from their respective routes from the message topic.
type TopicResolver interface {
	Resolve(ctx context.Context, domain, channel string) (domainID string, channelID string, isRoute bool, err error)
	ResolveTopic(ctx context.Context, topic string) (rtopic string, err error)
}

type resolver struct {
	channels grpcChannelsV1.ChannelsServiceClient
	domains  grpcDomainsV1.DomainsServiceClient
}

// NewTopicResolver creates a new instance of TopicResolver.
func NewTopicResolver(channelsClient grpcChannelsV1.ChannelsServiceClient, domainsClient grpcDomainsV1.DomainsServiceClient) TopicResolver {
	return &resolver{
		channels: channelsClient,
		domains:  domainsClient,
	}
}

func (r *resolver) Resolve(ctx context.Context, domain, channel string) (string, string, bool, error) {
	if domain == "" || channel == "" {
		return "", "", false, ErrEmptyRouteID
	}

	domainID, isdomainRoute, err := r.resolveDomain(ctx, domain)
	if err != nil {
		return "", "", false, errors.Wrap(ErrFailedResolveDomain, err)
	}
	channelID, isChannelRoute, err := r.resolveChannel(ctx, channel, domainID)
	if err != nil {
		return "", "", false, errors.Wrap(ErrFailedResolveChannel, err)
	}
	isRoute := isdomainRoute || isChannelRoute

	return domainID, channelID, isRoute, nil
}

func (r *resolver) ResolveTopic(ctx context.Context, topic string) (string, error) {
	domain, channel, subtopic, err := ParseTopic(topic)
	if err != nil {
		return "", errors.Wrap(ErrMalformedTopic, err)
	}

	domainID, channelID, _, err := r.Resolve(ctx, domain, channel)
	if err != nil {
		return "", err
	}
	rtopic := EncodeAdapterTopic(domainID, channelID, subtopic)

	return rtopic, nil
}

func (r *resolver) resolveDomain(ctx context.Context, domain string) (string, bool, error) {
	if validateUUID(domain) == nil {
		return domain, false, nil
	}
	d, err := r.domains.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route: domain,
	})
	if err != nil {
		return "", false, err
	}

	return d.Entity.Id, true, nil
}

func (r *resolver) resolveChannel(ctx context.Context, channel, domainID string) (string, bool, error) {
	if validateUUID(channel) == nil {
		return channel, false, nil
	}
	c, err := r.channels.RetrieveByRoute(ctx, &grpcCommonV1.RetrieveByRouteReq{
		Route:    channel,
		DomainId: domainID,
	})
	if err != nil {
		return "", false, err
	}

	return c.Entity.Id, true, nil
}

func validateUUID(extID string) (err error) {
	id, err := uuid.FromString(extID)
	if id.String() != extID || err != nil {
		return err
	}

	return nil
}

func ParsePublishTopic(topic string) (domainID, chanID, subtopic string, err error) {
	domainID, chanID, subtopic, err = ParseTopic(topic)
	if err != nil {
		return "", "", "", err
	}
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
	domainID, chanID, subtopic, err = ParseTopic(topic)
	if err != nil {
		return "", "", "", err
	}
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
	return fmt.Sprintf("%s.%s", string(MsgTopicPrefix), EncodeTopicSuffix(domainID, channelID, subtopic))
}

func EncodeTopicSuffix(domainID string, channelID string, subtopic string) string {
	subject := fmt.Sprintf("%s.%s.%s", domainID, string(ChannelTopicPrefix), channelID)
	if subtopic != "" {
		subject = fmt.Sprintf("%s.%s", subject, subtopic)
	}
	return subject
}

func EncodeMessageTopic(m *Message) string {
	return EncodeTopicSuffix(m.GetDomain(), m.GetChannel(), m.GetSubtopic())
}

func EncodeMessageMQTTTopic(m *Message) string {
	topic := fmt.Sprintf("%s/%s/%s/%s", string(MsgTopicPrefix), m.GetDomain(), string(ChannelTopicPrefix), m.GetChannel())
	if m.GetSubtopic() != "" {
		topic = topic + "/" + strings.ReplaceAll(m.GetSubtopic(), ".", "/")
	}
	return topic
}

func EncodeAdapterTopic(domain, channel, subtopic string) string {
	topic := fmt.Sprintf("%s/%s/%s/%s", string(MsgTopicPrefix), domain, string(ChannelTopicPrefix), channel)
	if subtopic != "" {
		topic = topic + "/" + subtopic
	}
	return topic
}

// ParseTopic parses a messaging topic string and returns the domain ID, channel ID, and subtopic.
// This is an optimized version with no regex and minimal allocations.
func ParseTopic(topic string) (domainID, chanID, subtopic string, err error) {
	// location of string "m"
	start := 0
	// Handle both formats: "/m/domain/c/channel/subtopic" and "m/domain/c/channel/subtopic".
	// If topic start with m/ then start is 0 , If topic start with /m/ then start is 1.
	n := len(topic)
	if n > 0 && topic[0] == '/' {
		start = 1
	}

	// length check - minimum: "m/<domain_id>/c/" = 5 characters if ignore <domain_id> and in this case start will be 0
	// length check - minimum: "/m/<domain_id>/c/" = 6 characters if ignore <domain_id> and in this case start will be 1
	if n < start+5 {
		return "", "", "", ErrMalformedTopic
	}
	if topic[start] != MsgTopicPrefix || topic[start+1] != '/' {
		return "", "", "", ErrMalformedTopic
	}
	pos := start + 2

	// Find "/c/" to locate domain ID
	cPos := -1
	for i := pos; i <= n-3; i++ {
		if topic[i] == '/' && topic[i+1] == ChannelTopicPrefix && topic[i+2] == '/' {
			cPos = i - pos
			break
		}
	}
	if cPos == -1 || cPos == 0 {
		return "", "", "", ErrMalformedTopic
	}
	domainID = topic[pos : pos+cPos]
	// skip "/c/"
	pos = pos + cPos + 3

	// Ensure channel exists
	if pos >= n {
		return "", "", "", ErrMalformedTopic
	}

	// Find '/' after channelID
	nextSlash := -1
	for i := pos; i < n; i++ {
		if topic[i] == '/' {
			nextSlash = i - pos
			break
		}
	}

	if nextSlash == -1 {
		// No subtopic
		chanID = topic[pos:]
	} else {
		chanID = topic[pos : pos+nextSlash]
		subtopic = topic[pos+nextSlash+1:]
	}

	// Validate channelID
	if len(chanID) == 0 {
		return "", "", "", ErrMalformedTopic
	}

	return domainID, chanID, subtopic, nil
}
