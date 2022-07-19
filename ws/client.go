// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

type Client interface {
	Handle(m messaging.Message) error
	Cancel() error
}

type Connclient struct {
	conn     *websocket.Conn
	thingKey string
	logger   logger.Logger
}

func NewClient(c *websocket.Conn, token string, l logger.Logger) *Connclient {
	return &Connclient{
		conn:     c,
		thingKey: token,
		logger:   l,
	}
}

func (c *Connclient) Cancel() error {
	// m := messaging.Message{
	// 	Protocol: "websocket",
	// 	Created:  time.Now().UnixNano(),
	// }
	// if err := c.client.WriteMessage(1, m.Payload); err != nil {
	// 	c.logger.Error(fmt.Sprintf("Error sending message: %s", err))
	// }

	return c.conn.Close()
}

func (c *Connclient) Handle(msg messaging.Message) error {
	fmt.Println("Using Handle() function")
	fmt.Println("msg.Pubslisher: ", msg.GetPublisher())
	fmt.Println("c.token: ", c.thingKey)
	fmt.Println("######")
	if msg.GetPublisher() == c.thingKey {
		return nil
	}

	return c.conn.WriteMessage(websocket.TextMessage, msg.Payload)
}
