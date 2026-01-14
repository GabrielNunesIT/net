package tls

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
)

// Server represents a TLS-based NETCONF Server.
type Server struct {
	listener net.Listener
	trace    *Trace
}

// Handler is the interface that is implemented to handle a TLS connection.
type Handler interface {
	// Handle is a function that handles i/o to/from a TLS connection.
	Handle(conn net.Conn)
}

// HandlerFactory is a function that will deliver a Handler.
type HandlerFactory func(conn *tls.Conn) Handler

// NewServer creates a new TLS server with a custom connection handler.
func NewServer(ctx context.Context, address string, port int, tlsConfig *tls.Config, factory HandlerFactory) (server *Server, err error) {
	server = &Server{trace: ContextTLSTrace(ctx)}

	listenAddress := fmt.Sprintf("%s:%d", address, port)
	server.listener, err = tls.Listen("tcp", listenAddress, tlsConfig)
	server.trace.Listened(address, err)
	if err != nil {
		return nil, err
	}

	go server.acceptConnections(factory)

	return server, nil
}

// Port delivers the tcp port number on which the server is listening.
func (s *Server) Port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

// Close closes any resources used by the server.
func (s *Server) Close() {
	_ = s.listener.Close()
}

func (s *Server) acceptConnections(factory HandlerFactory) {
	s.trace.StartAccepting()
	for {
		conn, err := s.listener.Accept()
		s.trace.Accepted(conn, err)
		if err != nil {
			return
		}

		tlsConn, ok := conn.(*tls.Conn)
		if !ok {
			s.trace.TLSHandshake(conn, fmt.Errorf("connection is not TLS"))
			_ = conn.Close()
			continue
		}

		// Perform TLS handshake
		err = tlsConn.Handshake()
		s.trace.TLSHandshake(tlsConn, err)
		if err != nil {
			_ = conn.Close()
			continue
		}

		go func(c *tls.Conn) {
			defer c.Close()
			factory(c).Handle(c)
		}(tlsConn)
	}
}
