package ws

import (
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux"
)

var _ Service = (*adapterService)(nil)

// Service contains publish and subscribe methods necessary for
// message transfer.
type Service interface {
	mainflux.MessagePublisher

	// BroadcastMessage broadcasts raw message to channel.
	BroadcastMessage(mainflux.RawMessage)

	// AddConnection adds new client ws connection for given client and channel.
	AddConnection(IDPair, *websocket.Conn)
}

type adapterService struct {
	pub    mainflux.MessagePublisher
	conns  map[string]map[string]socket
	logger log.Logger
}

// New instantiates the domain service implementation.
func New(pub mainflux.MessagePublisher, logger log.Logger) Service {
	return &adapterService{
		pub:    pub,
		conns:  make(map[string]map[string]socket),
		logger: logger,
	}
}

func (as *adapterService) Publish(msg mainflux.RawMessage) error {
	as.BroadcastMessage(msg)
	return as.pub.Publish(msg)
}

func (as *adapterService) BroadcastMessage(msg mainflux.RawMessage) {
	chanID := msg.Channel
	for pubID, conn := range as.conns[chanID] {
		pid, sock := pubID, conn
		go func() {
			if msg.Publisher == pid {
				return
			}
			if err := sock.write(msg); err != nil {
				as.logger.Log("error", "Failed to write message: %s", err)
			}
		}()
	}

	return
}

func (as *adapterService) AddConnection(pair IDPair, conn *websocket.Conn) {
	if _, ok := as.conns[pair.ChanID]; !ok {
		as.conns[pair.ChanID] = make(map[string]socket)
	}

	if oldConn, ok := as.conns[pair.ChanID][pair.PubID]; ok {
		oldConn.Close()
	}

	as.conns[pair.ChanID][pair.PubID] = socket{conn, &sync.Mutex{}}

	// On close delete connection from map of connections.
	conn.SetCloseHandler(func(code int, text string) error {
		delete(as.conns[pair.ChanID], pair.PubID)
		return conn.CloseHandler()(code, text)
	})
}

// IDPair contains publisher and channel id.
type IDPair struct {
	PubID  string
	ChanID string
}
