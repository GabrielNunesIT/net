package callhome

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"time"
)

// TLSListener listens for Call Home connections and initiates TLS as client.
type TLSListener struct {
	listener net.Listener
	config   *tls.Config
	trace    *Trace
}

// NewTLSListener creates a new TLS Call Home listener.
// The config is a TLS client configuration since the manager initiates TLS
// even though it receives the TCP connection.
func NewTLSListener(ctx context.Context, address string, port int, config *tls.Config) (*TLSListener, error) {
	listenAddr := fmt.Sprintf("%s:%d", address, port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("callhome: failed to listen on %s: %w", listenAddr, err)
	}

	trace := ContextTrace(ctx)
	trace.ListenStart(listener.Addr())

	return &TLSListener{
		listener: listener,
		config:   config,
		trace:    trace,
	}, nil
}

// Accept waits for a server to connect, then initiates TLS as client.
// Returns a TLS connection ready for NETCONF session establishment.
func (l *TLSListener) Accept(ctx context.Context) (*tls.Conn, error) {
	conn, err := l.listener.Accept()
	l.trace.AcceptDone(conn, err)
	if err != nil {
		return nil, fmt.Errorf("callhome: accept failed: %w", err)
	}

	// Initiate TLS as client (per RFC 8071, client initiates TLS)
	tlsConn := tls.Client(conn, l.config)

	// Perform handshake with context deadline if available
	if deadline, ok := ctx.Deadline(); ok {
		_ = tlsConn.SetDeadline(deadline)
	}

	if err = tlsConn.Handshake(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("callhome: TLS handshake failed: %w", err)
	}

	// Clear deadline after handshake
	_ = tlsConn.SetDeadline(time.Time{})

	l.trace.TLSConnected(conn.RemoteAddr().String(), tlsConn)

	return tlsConn, nil
}

// Close closes the listener.
func (l *TLSListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address.
func (l *TLSListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Port returns the port the listener is bound to.
func (l *TLSListener) Port() int {
	return l.listener.Addr().(*net.TCPAddr).Port
}

// TLSConnDialer wraps an established Call Home TLS connection as a Dialer.
// This allows integration with existing session factory functions.
type TLSConnDialer struct {
	conn *tls.Conn
}

// NewTLSConnDialer creates a dialer from an established Call Home connection.
func NewTLSConnDialer(conn *tls.Conn) *TLSConnDialer {
	return &TLSConnDialer{conn: conn}
}

// Target returns the remote address.
func (d *TLSConnDialer) Target() string {
	return d.conn.RemoteAddr().String()
}

// Dial returns the pre-established connection.
func (d *TLSConnDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	return d.conn, nil
}

// Close is a no-op; the connection should be closed separately.
func (d *TLSConnDialer) Close(conn io.ReadWriteCloser) error {
	return nil
}
