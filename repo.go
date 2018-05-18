package mainflux

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	log "github.com/mainflux/mainflux/logger"
	nats "github.com/nats-io/go-nats"
)

// MessageRepository specifies message reading API.
type MessageRepository interface {

	// Save method is used to save published message. A non-nil error is returned to indicate
	// operation failure.
	Save(Message) error
}

// HandleMsg is default message handler for messages received from NATS.
func HandleMsg(name string, logger log.Logger, repo MessageRepository) nats.MsgHandler {
	return func(m *nats.Msg) {
		msg := &Message{}

		if err := proto.Unmarshal(m.Data, msg); err != nil {
			logger.Warn(fmt.Sprintf("%s failed to unmarshal received message: %s", name, err))
			return
		}

		if err := repo.Save(*msg); err != nil {
			logger.Warn(fmt.Sprintf("%s failed to save message: %s", name, err))
			return
		}
	}
}
