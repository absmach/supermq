// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package ws

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/logger"
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

func (c *Client) Publish(svc Service, logger logger.Logger, thingKey, chanID, subtopic string, msgs <-chan []byte) {
	for msg := range msgs {
		m := messaging.Message{
			Channel:  chanID,
			Subtopic: subtopic,
			Protocol: "websocket",
			Payload:  msg,
			Created:  time.Now().UnixNano(),
		}
		svc.Publish(context.Background(), thingKey, m)
	}
	if err := svc.Unsubscribe(context.Background(), thingKey, chanID, subtopic); err != nil {
		logger.Warn(fmt.Sprintf("Failed to subscribe to broker: %s", err.Error()))
		c.conn.Close()
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
