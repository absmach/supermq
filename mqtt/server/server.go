package server

import (
	"fmt"
	"net"
	"time"

	log "github.com/mainflux/mainflux/logger"
)

// Server is our main struct.
type Server struct {
	running      bool
	listener     net.Listener
	clients      map[uint64]*Client
	totalClients uint64
	start        time.Time
	addr         string
	logger       log.Logger
}

// Protected check on running state
func (srv *Server) isRunning() bool {
	return srv.running
}

// Start up the server, this will block.
func ListenAndServe(addr string, logger log.Logger) error {
	srv := &Server{
		addr:   addr,
		clients: make(map[uint64]*Client),
		start:  time.Now(),
		logger: logger,
		running:  true,
	}

	// Wait for clients.
	return srv.ListenAndServe()
}

func (srv *Server) ListenAndServe() error {
	l, err := net.Listen("tcp", srv.addr)
	if err != nil {
		srv.logger.Error(fmt.Sprintf("Failed to start: %s", err))
		return err
	}
	defer l.Close()

	// Setup state that can enable shutdown
	srv.listener = l

	return srv.Serve(l)
}

func (srv *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			srv.logger.Error(fmt.Sprintf("Accept error: %s", err))
			continue
		}

		srv.logger.Info("Accepted new client")

		client := &Client{
			conn:   conn,
			server: srv,
		}
		go client.listen()
	}

	srv.logger.Info("Server Exiting..")

	return nil
}
