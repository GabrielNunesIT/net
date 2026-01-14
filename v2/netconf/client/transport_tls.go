package client

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"time"
)

// TLS-specific transport implementation for NETCONF over TLS (RFC 7589).

// TLSDialer implements the Dialer interface for TLS connections.
type TLSDialer struct {
	target string
	config *tls.Config
}

// NewTLSDialer creates a new TLS dialer with the given target and configuration.
func NewTLSDialer(target string, config *tls.Config) *TLSDialer {
	return &TLSDialer{target: target, config: config}
}

// Target returns the connection target address.
func (d *TLSDialer) Target() string {
	return d.target
}

// Dial establishes a TLS connection to the target.
func (d *TLSDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	tracer := ContextClientTrace(ctx)

	tracer.DialStart(nil, d.target) // nil for ssh config as this is TLS
	var err error
	defer func(begin time.Time) {
		tracer.DialDone(nil, d.target, err, time.Since(begin))
	}(time.Now())

	// Use a dialer with context support for timeout/cancellation
	dialer := &net.Dialer{}
	netConn, err := dialer.DialContext(ctx, "tcp", d.target)
	if err != nil {
		return nil, err
	}

	// Wrap with TLS
	conn := tls.Client(netConn, d.config)

	// Perform handshake with context deadline if available
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}

	if err = conn.Handshake(); err != nil {
		_ = netConn.Close()
		return nil, err
	}

	// Clear deadline after handshake
	_ = conn.SetDeadline(time.Time{})

	return conn, nil
}

// Close closes the TLS connection.
func (d *TLSDialer) Close(conn io.ReadWriteCloser) error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// TLSConnDialer implements Dialer for an existing TLS connection.
// Use this when the TLS connection is already established.
type TLSConnDialer struct {
	conn *tls.Conn
}

// NewTLSConnDialer creates a dialer from an existing TLS connection.
func NewTLSConnDialer(conn *tls.Conn) *TLSConnDialer {
	return &TLSConnDialer{conn: conn}
}

// Target returns the remote address of the TLS connection.
func (d *TLSConnDialer) Target() string {
	return d.conn.RemoteAddr().String()
}

// Dial returns the existing TLS connection.
func (d *TLSConnDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	return d.conn, nil
}

// Close does nothing for pre-existing TLS connections.
func (d *TLSConnDialer) Close(conn io.ReadWriteCloser) error {
	// Don't close a pre-existing connection
	return nil
}
