// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/pkg/messaging"
)

type Client struct {
	conn *websocket.Conn
	id   string
}

func NewClient(c *websocket.Conn, thingKey string) *Client {
	return &Client{
		conn: c,
		id:   thingKey,
	}
}

func (c *Client) GetID() string {
	return c.id
}

func (c *Client) Cancel() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) Handle(msg messaging.Message) error {
	if msg.GetPublisher() == c.id {
		return nil
	}

	return c.conn.WriteMessage(websocket.TextMessage, msg.Payload)
}
