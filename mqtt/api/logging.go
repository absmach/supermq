package api

import (
	"fmt"
	"time"

	"github.com/mainflux/mainflux"
	log "github.com/mainflux/mainflux/logger"
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

func (lm *loggingMiddleware) Publish(msg mainflux.RawMessage, cfHandler mainflux.ConnFailHandler) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method publish to channel %s took %s to complete", msg.Channel, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Publish(msg, cfHandler)
}

func (lm *loggingMiddleware) Subscribe(sub mainflux.Subscription, cfHandler mainflux.ConnFailHandler) (_ mainflux.Unsubscribe, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method subscribe to channel %s took %s to complete", sub.ChanID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.Subscribe(sub, cfHandler)
}
