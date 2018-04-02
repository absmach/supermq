package mainflux

// WriteMessage is used to write message to channel.
type WriteMessage func(RawMessage) error

// ReadMessage is used to read message from channel.
type ReadMessage func() ([]byte, error)

// Unsubscribe tries to unsubscribe user from channel.
type Unsubscribe func() error

// ConnFailHandler is called when connection error happens.
type ConnFailHandler func()

// MessagePublisher specifies a message publishing API.
type MessagePublisher interface {
	// Publishes message to the stream. A non-nil error is returned to indicate
	// operation failure.
	Publish(RawMessage, ConnFailHandler) error
}

// MessageSubscriber specifies a message subscription API.
type MessageSubscriber interface {
	// Subscribes to channel and returns unsubscribe function.
	// A non-nil error is returned to indicate subscription failure.
	Subscribe(Subscription, ConnFailHandler) (Unsubscribe, error)
}

// MessagePubSub specifies a message publishing and subscription API.
type MessagePubSub interface {
	MessagePublisher
	MessageSubscriber
}

// Subscription contains publisher and channel id.
type Subscription struct {
	PubID  string
	ChanID string
	Write  WriteMessage
	Read   ReadMessage
}
