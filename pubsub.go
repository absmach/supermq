package mainflux

// WriteMessage is used to write message to channel.
type WriteMessage func(RawMessage) error

// ReadMessage is used to read message from channel.
type ReadMessage func() ([]byte, error)

// MessagePublisher specifies a message publishing API.
type MessagePublisher interface {
	// Publishes message to the stream. A non-nil error is returned to indicate
	// operation failure.
	Publish(RawMessage) error
}

// MessageSubscriber specifies a message subscription API.
type MessageSubscriber interface {
	// Subscribes to channel and returns unsubscribe function.
	// A non-nil error is returned to indicate subscription failure.
	Subscribe(Subscription, WriteMessage, ReadMessage) (func(), error)
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
}
