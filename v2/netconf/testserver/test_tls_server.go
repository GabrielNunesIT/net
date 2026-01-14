package testserver

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"

	xtls "github.com/damianoneill/net/v2/netconf/server/tls"

	assert "github.com/stretchr/testify/require"
)

// TLSServer represents a test TLS Server.
type TLSServer struct {
	server  *xtls.Server
	CertPEM []byte
	KeyPEM  []byte
}

// TLSHandler is the interface that is implemented to handle a TLS connection.
type TLSHandler interface {
	// Handle is a function that handles i/o to/from a TLS connection.
	Handle(t assert.TestingT, conn net.Conn)
}

// TLSHandlerFactory is a test function that will deliver a TLSHandler.
type TLSHandlerFactory func(t assert.TestingT) TLSHandler

// NewTLSServer delivers a new test TLS Server, with a Handler that simply echoes lines received.
// The server uses a self-signed certificate for testing.
func NewTLSServer(t assert.TestingT) *TLSServer {
	return NewTLSServerHandler(t, func(t assert.TestingT) TLSHandler { return &tlsEchoer{} })
}

// NewTLSServerHandler delivers a new test TLS Server with a custom handler.
func NewTLSServerHandler(t assert.TestingT, factory TLSHandlerFactory) *TLSServer {
	// Generate self-signed certificate for testing
	certPEM, keyPEM, err := xtls.GenerateSelfSignedCert()
	assert.NoError(t, err, "Failed to generate self-signed certificate")

	tlsConfig, err := xtls.ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err, "Failed to create TLS server config")

	server, err := xtls.NewServer(context.Background(), "localhost", 0, tlsConfig, func(conn *tls.Conn) xtls.Handler {
		return &tlsTestHandler{t: t, factory: factory}
	})
	assert.NoError(t, err, "Failed to create TLS server")

	return &TLSServer{
		server:  server,
		CertPEM: certPEM,
		KeyPEM:  keyPEM,
	}
}

// Port delivers the tcp port number on which the server is listening.
func (ts *TLSServer) Port() int {
	return ts.server.Port()
}

// Close closes any resources used by the server.
func (ts *TLSServer) Close() {
	ts.server.Close()
}

// tlsTestHandler wraps a TLSHandler to implement tls.Handler.
type tlsTestHandler struct {
	t       assert.TestingT
	factory TLSHandlerFactory
}

func (h *tlsTestHandler) Handle(conn net.Conn) {
	h.factory(h.t).Handle(h.t, conn)
}

// tlsEchoer is a simple handler that echoes received lines.
type tlsEchoer struct{}

func (e *tlsEchoer) Handle(t assert.TestingT, conn net.Conn) {
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		_, err = writer.WriteString(fmt.Sprintf("GOT:%s", input))
		assert.NoError(t, err, "Write failed")
		err = writer.Flush()
		assert.NoError(t, err, "Flush failed")
	}
}
