// Package callhome provides NETCONF Call Home functionality as defined in RFC 8071.
// Call Home allows a NETCONF server (device) to initiate the TCP connection to a
// NETCONF client (manager), after which the client initiates the secure session.
package callhome

import (
	"context"
	"io"
	"net"
)

// Default ports as assigned by IANA for NETCONF Call Home.
const (
	// DefaultSSHPort is the IANA-assigned port for NETCONF Call Home over SSH.
	DefaultSSHPort = 4334

	// DefaultTLSPort is the IANA-assigned port for NETCONF Call Home over TLS.
	DefaultTLSPort = 4335
)

// NetDialer is an interface for creating TCP connections.
// This allows using custom dialers (e.g., with proxies, custom resolvers, rate limiting).
// The standard library's net.Dialer implements this interface.
type NetDialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// DefaultNetDialer is the default TCP dialer used when no custom dialer is provided.
var DefaultNetDialer NetDialer = &net.Dialer{}

// Dialer is the interface for initiating Call Home connections.
// The server uses this to connect to a client listener.
type Dialer interface {
	// Dial initiates a connection to the client and performs the transport handshake.
	// The returned connection is ready for NETCONF session establishment.
	Dial(ctx context.Context) (io.ReadWriteCloser, error)

	// Target returns the address of the client being dialed.
	Target() string

	// Close closes any resources held by the dialer.
	Close() error
}

