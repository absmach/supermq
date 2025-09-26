package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/absmach/mgate/pkg/session"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"google.golang.org/protobuf/proto"
)

type beforeHandler struct {
	resolver messaging.TopicResolver
	parser   messaging.TopicParser
	logger   *slog.Logger
}

func NewBeforeHandler(resolver messaging.TopicResolver, parser messaging.TopicParser, logger *slog.Logger) session.Interceptor {
	return &beforeHandler{
		resolver: resolver,
		parser:   parser,
		logger:   logger,
	}
}

// This interceptor is used to replace domain and channel routes with relevant domain and channel IDs in the message topic.
func (bh beforeHandler) Intercept(ctx context.Context, pkt packets.ControlPacket, dir session.Direction) (packets.ControlPacket, error) {
	switch pt := pkt.(type) {
	case *packets.SubscribePacket:
		for i, topic := range pt.Topics {
			ft, err := bh.resolver.ResolveTopic(ctx, topic)
			if err != nil {
				return nil, err
			}
			pt.Topics[i] = ft
		}

		return pt, nil
	case *packets.UnsubscribePacket:
		for i, topic := range pt.Topics {
			ft, err := bh.resolver.ResolveTopic(ctx, topic)
			if err != nil {
				return nil, err
			}
			pt.Topics[i] = ft
		}
		return pt, nil
	case *packets.PublishPacket:
		ft, err := bh.resolver.ResolveTopic(ctx, pt.TopicName)
		if err != nil {
			return nil, err
		}
		pt.TopicName = ft

		s, ok := session.FromContext(ctx)
		if !ok {
			return pt, errors.Wrap(ErrFailedPublish, ErrClientNotInitialized)
		}
		bh.logger.Info(fmt.Sprintf(LogInfoPublished, s.ID, ft))

		switch dir {
		case session.Up:
			domainID, chanID, subTopic, _, err := bh.parser.ParsePublishTopic(ctx, ft, false)
			if err != nil {
				return pt, errors.Wrap(ErrFailedPublish, err)
			}

			msg := &messaging.Message{
				Protocol:  "mqtt",
				Domain:    domainID,
				Channel:   chanID,
				Subtopic:  subTopic,
				Publisher: s.Username,
				Payload:   pt.Payload,
				Created:   time.Now().UnixNano(),
			}

			data, err := proto.Marshal(msg)
			if err != nil {
				return pt, err
			}
			pt.Payload = data

		case session.Down:
			var msg messaging.Message

			if err := proto.Unmarshal(pt.Payload, &msg); err != nil {
				return pt, errors.Wrap(ErrFailedPublish, err)
			}
			pt.Payload = msg.GetPayload()
		}

		return pt, nil
	}

	return pkt, nil
}
