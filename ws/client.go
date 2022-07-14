// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

// Client wraps WS client.
type Client interface {
	Token() string
	Handle(m messaging.Message) error
	Cancel() error
}

type Connclient struct {
	client *websocket.Conn
	token  string
	logger logger.Logger
}

// NewClient Instantiates a new Observer.
func NewClient(c *websocket.Conn, token string, l logger.Logger) Client {
	return &Connclient{
		client: c,
		token:  token,
		logger: l,
	}
}

type Channel struct {
	Messages chan messaging.Message
	Closed   chan bool
	closed   bool
	mutex    sync.Mutex
}

func NewChannel() *Channel {
	return &Channel{
		Messages: make(chan messaging.Message),
		closed:   false,
		Closed:   make(chan bool),
		mutex:    sync.Mutex{},
	}
}

func (c *Connclient) Cancel() error {
	m := messaging.Message{
		Protocol: "websocket",
		Created:  time.Now().UnixNano(),
	}
	if err := c.client.WriteMessage(1, m.Payload); err != nil {
		c.logger.Error(fmt.Sprintf("Error sending message: %s", err))
	}

	return c.client.Close()
}

func (c *Connclient) Token() string {
	return c.token
}

func (c *Connclient) Handle(msg messaging.Message) error {
	return c.client.WriteMessage(websocket.TextMessage, msg.Payload)
}

// Send method sends message over Messages channel.
func (c *Channel) Send(msg messaging.Message) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if !c.closed {
		c.Messages <- msg
	}
}

// Close channel and stop message transfer
func (c *Channel) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.closed = true
	c.Closed <- true

	close(c.Messages)
	close(c.Closed)
}
