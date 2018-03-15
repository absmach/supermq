package api

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/ws"
	broker "github.com/nats-io/go-nats"
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

func (lm *loggingMiddleware) HandleMessage(msg *broker.Msg) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "handle_message",
			"took", time.Since(begin),
		)
	}(time.Now())

	lm.svc.HandleMessage(msg)
}

func (lm *loggingMiddleware) AddConnection(channelID, publisherID string, conn *websocket.Conn) {
	defer func(begin time.Time) {
		lm.logger.Log(
			"method", "add_connection",
			"took", time.Since(begin),
		)
	}(time.Now())

	lm.svc.AddConnection(channelID, publisherID, conn)
}
