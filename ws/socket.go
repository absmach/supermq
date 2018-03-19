package ws

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

type socket struct {
	*websocket.Conn
	mu *sync.Mutex
}

func (s socket) write(msg mainflux.RawMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.WriteJSON(msg)
}
