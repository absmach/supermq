// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws

import (
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/pkg/messaging"
)

type Connclient struct {
	conn  *websocket.Conn
	pubID string
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
	if msg.GetPublisher() == c.pubID {
		return nil
	}

	return c.conn.WriteMessage(websocket.TextMessage, msg.Payload)
}
