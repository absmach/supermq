package server

import (
	"bufio"
	"net"
)

// Client holds info about connection
type Client struct {
	conn   net.Conn
	server *Server
}

// Read client data from channel
func (c *Client) listen() {
	reader := bufio.NewReader(c.conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			c.conn.Close()
			//c.server.onClientConnectionClosed(c, err)
			return
		}
		//c.server.onNewMessage(c, message)
		println(message)
	}
}

// Send text message to client
func (c *Client) Send(message string) error {
	_, err := c.conn.Write([]byte(message))
	return err
}

// Send bytes to client
func (c *Client) SendBytes(b []byte) error {
	_, err := c.conn.Write(b)
	return err
}

func (c *Client) Conn() net.Conn {
	return c.conn
}
