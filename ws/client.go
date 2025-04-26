// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package ws

import (
	"context"
	"log/slog"
	"time"

	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/messaging"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

var (
	errHandlerBlockedMsgChan = errors.New("message handler msg chan blocked (full)")
	errHandlerClosedMsgChan  = errors.New("message handler closed msg chan")
	errFailedToWriteMsg      = errors.New("failed to write message to connection")
	errFailedToWritePing     = errors.New("failed to write ping to connection")
	errReadMsg               = errors.New("failed to read messages ")
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 30 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
)

// Client handles messaging and websocket connection.
type Client struct {
	logger *slog.Logger
	conn   *websocket.Conn
	id     string
	msg    chan *messaging.Message
}

// NewClient returns a new websocket client.
func NewClient(logger *slog.Logger, conn *websocket.Conn, sessionID string) *Client {
	c := &Client{
		logger: logger,
		conn:   conn,
		id:     sessionID,
		msg:    make(chan *messaging.Message, 1024),
	}
	return c
}

// Cancel handles the websocket connection after unsubscribing.
func (c *Client) Cancel() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Handle handles the sending and receiving of messages via the broker.
func (c *Client) Handle(msg *messaging.Message) error {
	select {
	case c.msg <- msg:
		return nil
	default:
		return errHandlerBlockedMsgChan
	}
}

// CloseHandler will work only if messages are read.
func (c *Client) readPump(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()
	c.conn.SetPongHandler(func(string) error {
		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			return err
		}
		return nil
	})
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("read_pump: received context Done")
			return nil
		default:
			msgType, msg, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Debug("read_pump: unexpected close error", slog.String("error", err.Error()))
					return nil
				}
				return errors.Wrap(errReadMsg, err)
			}
			c.logger.Debug("read_pump: received message ", slog.Int("message_type", msgType), slog.String("message", string(msg)))
		}
	}
}

func (c *Client) writePump(ctx context.Context, cancel context.CancelFunc) error {
	defer cancel()
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			c.logger.Debug("write_pump: received context Done ")
			return nil
		case msg, ok := <-c.msg:
			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					return errors.Wrap(errHandlerClosedMsgChan, err)
				}
				return errHandlerClosedMsgChan
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, msg.GetPayload()); err != nil {
				return errors.Wrap(errFailedToWriteMsg, err)
			}
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return errors.Wrap(errFailedToWritePing, err)
			}
		}
	}
}

// SetCloseHandler sets a close handler for the WebSocket connection.
func (c *Client) SetCloseHandler(handler func(code int, text string) error) {
	c.conn.SetCloseHandler(func(code int, text string) error {
		c.logger.Debug("WebSocket closed", slog.String("session_id", c.id), slog.Int("code", code), slog.String("text", text))
		if err := handler(code, text); err != nil {
			c.logger.Warn("Error in close handler", slog.String("error", err.Error()))
		}
		return nil
	})
}

func (c *Client) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return c.readPump(ctx, cancel)
	})

	g.Go(func() error {
		return c.writePump(ctx, cancel)
	})

	err := g.Wait()
	if err != nil {
		c.logger.Warn("websocket client error", slog.String("session_id", c.id), slog.String("error", err.Error()))
	}
}
