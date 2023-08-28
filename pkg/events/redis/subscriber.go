package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/events"
)

const (
	exists = "BUSYGROUP Consumer Group name already exists"
)

type subEventStore struct {
	client   *redis.Client
	stream   string
	consumer string
	logger   mflog.Logger
}

func NewEventStoreSubscriber(url, stream, consumer string, logger mflog.Logger) (events.Subscriber, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	return &subEventStore{
		client:   redis.NewClient(opts),
		stream:   stream,
		consumer: consumer,
		logger:   logger,
	}, nil
}

func (es *subEventStore) Subscribe(ctx context.Context, group string, handler events.EventHandler) error {
	err := es.client.XGroupCreateMkStream(ctx, es.stream, group, "$").Err()
	if err != nil && err.Error() != exists {
		return err
	}

	go func(ctx context.Context, group string) {
		for {
			msgs, err := es.client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    group,
				Consumer: es.consumer,
				Streams:  []string{es.stream, ">"},
				Count:    100,
			}).Result()
			if err != nil {
				es.logger.Warn(fmt.Sprintf("failed to read from Redis stream: %s", err))
				continue
			}
			if len(msgs) == 0 {
				continue
			}

			es.handle(ctx, group, msgs[0].Messages, handler)
		}
	}(ctx, group)

	return nil
}

func (es *subEventStore) Close() error {
	return es.client.Close()
}

type redisEvent struct {
	Data map[string]interface{}
}

func (re redisEvent) Encode() (map[string]interface{}, error) {
	return re.Data, nil
}

func (es *subEventStore) handle(ctx context.Context, group string, msgs []redis.XMessage, h events.EventHandler) {
	for _, msg := range msgs {
		var event = redisEvent{msg.Values}

		if err := h.Handle(ctx, event); err != nil {
			es.logger.Warn(fmt.Sprintf("failed to handle redis event: %s", err))
			return
		}

		es.client.XAck(ctx, es.stream, group, msg.ID)
	}
}
