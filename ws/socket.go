package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

// Socket represents threadsafe websocket connection.
type Socket struct {
	*websocket.Conn
	mu *sync.Mutex
}

// NewSocket returns new threadsafe websocket connection.
func NewSocket(conn *websocket.Conn) Socket {
	return Socket{conn, &sync.Mutex{}}
}

func (s Socket) write(rawMsg mainflux.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.WriteMessage(websocket.TextMessage, rawMsg.Payload)
}
