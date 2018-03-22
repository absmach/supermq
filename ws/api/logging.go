package api

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
)

var _ ws.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    ws.Service
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(svc ws.Service, logger log.Logger) ws.Service {
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

func (lm *loggingMiddleware) Broadcast(socket ws.Socket, msg mainflux.RawMessage) error {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "broadcast",
			"took", time.Since(begin),
		)
	}(time.Now())

	return lm.svc.Broadcast(socket, msg)
}

func (lm *loggingMiddleware) Listen(socket ws.Socket, sub ws.Subscription, onClose func()) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "start_listening",
			"took", time.Since(begin),
		)
	}(time.Now())

	lm.svc.Listen(socket, sub, onClose)
}
