package ws

import "github.com/mainflux/mainflux"

// Channel is used for recieving and sending messages.
type Channel struct {
	Messages chan mainflux.RawMessage
	Closed   chan bool
}

// Close channel and stop message transfer.
func (channel Channel) Close() {
	close(channel.Messages)
	close(channel.Closed)
}
