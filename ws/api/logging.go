package api

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/mainflux/mainflux"
)

var _ mainflux.MessagePubSub = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    mainflux.MessagePubSub
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(svc mainflux.MessagePubSub, logger log.Logger) mainflux.MessagePubSub {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Publish(msg mainflux.RawMessage) error {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "publish",
			"took", time.Since(begin),
		)
	}(time.Now())

	return lm.svc.Publish(msg)
}

func (lm *loggingMiddleware) Subscribe(sub mainflux.Subscription, write mainflux.WriteMessage, read mainflux.ReadMessage) (func(), error) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "subscribe",
			"took", time.Since(begin),
		)
	}(time.Now())

	return lm.svc.Subscribe(sub, write, read)
}
