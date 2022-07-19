package api

import (
	"fmt"
	"sync"

	"github.com/mainflux/mainflux/pkg/messaging"
)

type channel struct {
	Messages chan messaging.Message
	// Closed   chan bool
	closed bool
	mutex  sync.Mutex
}

func newChannel() *channel {
	return &channel{
		Messages: make(chan messaging.Message),
		closed:   false,
		// Closed:   make(chan bool),
		mutex: sync.Mutex{},
	}
}

// Send method sends message over Messages channel. // Not actually being used
func (c *channel) Send(msg messaging.Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.Messages <- msg
	}
}

// Close channel and stop message transfer.
func (c *channel) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true
	// c.Closed <- true

	fmt.Println("Closed both channels from ch.Close()")

	close(c.Messages)
	// close(c.Closed)

}
