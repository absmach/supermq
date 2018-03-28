package mainflux

// MessagePublisher specifies a message publishing API.
type MessagePublisher interface {
	// Publishes message to the stream. A non-nil error is returned to indicate
	// operation failure.
	Publish(RawMessage) error
}

// MessageSubscriber specifies a message subscription API.
type MessageSubscriber interface {
	// Subscirbes to channel. A non-nil error is returned to indicate
	// operation failure.
	Subscribe(string, func(RawMessage)) (Subscription, error)
}

// MessagePubSub specifies a message publishing and subscription API.
type MessagePubSub interface {
	MessagePublisher
	MessageSubscriber
}

// MessageBroker specifies a message broadcasting API.
type MessageBroker interface {
	MessagePubSub
	// Broadcasts raw message to channel.
	Broadcast(RawMessage, func(RawMessage) error) error
}

// Subscription specifies subscription interface.
type Subscription interface {
	Unsubscribe() error
}
