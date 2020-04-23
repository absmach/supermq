package mainflux

// Publisher specifies message publishing API.
type Publisher interface {
	// Publishes message to the stream.
	Publish(msg Message) error
}

// MessageHandler represents Message handler for Subscriber.
type MessageHandler func(Message) error

// Subscriber specifies message subscription API.
type Subscriber interface {
	// Subscribe subscribes to the message stream and consumes messages.
	Subscribe(MessageHandler) error

	// Unsubscribe unsubscribes from the message stream and
	// stops consuming messages.
	Unsubscribe() error
}

// PubSub represents aggregation interface for publisher and subscriber.
type PubSub interface {
	Publisher
	Subscriber
}
