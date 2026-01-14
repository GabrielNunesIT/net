// Package callhome provides NETCONF Call Home listener functionality for clients.
// Per RFC 8071, the NETCONF client (manager) listens for incoming TCP connections
// from NETCONF servers (devices), then initiates the secure session.
package callhome

import (
	"net"
)

// Default ports as assigned by IANA for NETCONF Call Home.
const (
	// DefaultSSHPort is the IANA-assigned port for NETCONF Call Home over SSH.
	DefaultSSHPort = 4334

	// DefaultTLSPort is the IANA-assigned port for NETCONF Call Home over TLS.
	DefaultTLSPort = 4335
)

// Listener represents a Call Home listener that accepts incoming connections.
type Listener interface {
	// Accept waits for and returns the next connection.
	Accept() (net.Conn, error)

	// Close closes the listener.
	Close() error

	// Addr returns the listener's network address.
	Addr() net.Addr
}
