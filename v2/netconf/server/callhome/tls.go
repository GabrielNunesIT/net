package callhome

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"time"
)

// TLSDialer implements Call Home for TLS transport.
// The server connects to the client, then waits for the client to initiate TLS.
type TLSDialer struct {
	target    string
	config    *tls.Config
	netDialer NetDialer
	trace     *Trace
}

// NewTLSDialer creates a new TLS Call Home dialer.
// The config is a TLS server configuration since the device acts as TLS server
// even though it initiates the TCP connection.
func NewTLSDialer(target string, config *tls.Config) *TLSDialer {
	return &TLSDialer{
		target:    target,
		config:    config,
		netDialer: DefaultNetDialer,
		trace:     noOpTrace,
	}
}

// NewTLSDialerWithOptions creates a new TLS Call Home dialer with custom options.
// If netDialer is nil, DefaultNetDialer is used.
// If trace is nil, a no-op trace is used.
func NewTLSDialerWithOptions(target string, config *tls.Config, netDialer NetDialer, trace *Trace) *TLSDialer {
	d := NewTLSDialer(target, config)
	if netDialer != nil {
		d.netDialer = netDialer
	}
	if trace != nil {
		d.trace = trace
	}
	return d
}

// Target returns the client address being dialed.
func (d *TLSDialer) Target() string {
	return d.target
}

// Dial connects to the client and performs TLS handshake as server.
// Per RFC 8071, the server initiates TCP but acts as TLS server.
func (d *TLSDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	d.trace.DialStart(d.target)

	var err error
	defer func(begin time.Time) {
		d.trace.DialDone(d.target, err, time.Since(begin))
	}(time.Now())

	// Dial the client using the configured NetDialer
	conn, err := d.netDialer.DialContext(ctx, "tcp", d.target)
	if err != nil {
		return nil, fmt.Errorf("callhome: failed to connect to client: %w", err)
	}

	// Wrap with TLS as server (device acts as TLS server per RFC 8071)
	tlsConn := tls.Server(conn, d.config)

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

	d.trace.TLSConnected(d.target, tlsConn)

	return tlsConn, nil
}

// Close closes the dialer (no-op for TLS dialer).
func (d *TLSDialer) Close() error {
	return nil
}
