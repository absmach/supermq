// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws

import (
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/pkg/messaging"
)

type Client interface {
	Handle(m messaging.Message) error
	Cancel() error
}

type Connclient struct {
	conn  *websocket.Conn
	pubID string
	// logger   logger.Logger
}

func NewClient(c *websocket.Conn, token string) *Connclient {
	return &Connclient{
		conn:  c,
		pubID: token,
	}
}

func (c *Connclient) Cancel() error {
	return c.conn.Close()
}

func (c *Connclient) Handle(msg messaging.Message) error {
	fmt.Println("Using Handle() function")
	fmt.Println("msg.Pubslisher: ", msg.GetPublisher())
	fmt.Println("c.token: ", c.pubID)
	fmt.Println("######")
	if msg.GetPublisher() == c.pubID {
		return nil
	}

	return c.conn.WriteMessage(websocket.TextMessage, msg.Payload)
}
