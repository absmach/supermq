package server

import (
	"errors"
	"fmt"
	"net"
	"io"
	"time"

	log "github.com/mainflux/mainflux/logger"
	"github.com/surgemq/message"
)

var (
	ErrInvalidConnectionType  error = errors.New("service: Invalid connection type")
	ErrInvalidSubscriber      error = errors.New("service: Invalid subscriber")
	ErrBufferNotReady         error = errors.New("service: buffer is not ready")
	ErrBufferInsufficientData error = errors.New("service: buffer has insufficient data.")
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
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, e := l.Accept()
		if e != nil {
			//select {
			//case <-srv.getDoneChan():
			//	return ErrServerClosed
			//default:
			//}
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.logger.Error(fmt.Sprintf("http: Accept error: %v; retrying in %v", e, tempDelay))
				time.Sleep(tempDelay)
				continue
			}
			return e
		}

		go srv.handle(conn)
	}

	srv.logger.Info("Server Exiting..")

	return nil
}

func (srv *Server) handle(c io.Closer) error {
	if c == nil {
		return ErrInvalidConnectionType
	}

	conn, ok := c.(net.Conn)
	if !ok {
		return ErrInvalidConnectionType
	}

	// To establish a connection, we must
	// 1. Read and decode the message.ConnectMessage from the wire
	// 2. If no decoding errors, then authenticate using username and password.
	//    Otherwise, write out to the wire message.ConnackMessage with
	//    appropriate error.
	// 3. If authentication is successful, then either create a new session or
	//    retrieve existing session
	// 4. Write out to the wire a successful message.ConnackMessage message

	// Read the CONNECT message from the wire, if error, then check to see if it's
	// a CONNACK error. If it's CONNACK error, send the proper CONNACK error back
	// to client. Exit regardless of error type.

	//conn.SetReadDeadline(time.Now().Add(time.Second * time.Duration(srv.ConnectTimeout)))

	resp := message.NewConnackMessage()

	req, err := getConnectMessage(conn)
	if err != nil {
		if cerr, ok := err.(message.ConnackCode); ok {
			resp.SetReturnCode(cerr)
			resp.SetSessionPresent(false)
			writeMessage(conn, resp)
		}
		return err
	}

	if req.KeepAlive() == 0 {
		//req.SetKeepAlive(minKeepAlive)
		req.SetKeepAlive(30)
	}


	resp.SetReturnCode(message.ConnectionAccepted)

	if err = writeMessage(c, resp); err != nil {
		return err
	}

	return nil
}
