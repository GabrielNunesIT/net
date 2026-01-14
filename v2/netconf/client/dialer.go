package client

import (
	"context"
	"io"
)

// Dialer is the interface that wraps transport connection creation.
// It provides a unified way to create connections for different transport types (SSH, TLS).
type Dialer interface {
	// Dial establishes a connection to the target and returns an io.ReadWriteCloser.
	Dial(ctx context.Context) (io.ReadWriteCloser, error)

	// Close closes the connection. This method allows the dialer to track
	// and properly close connections it created.
	Close(conn io.ReadWriteCloser) error

	// Target returns the connection target address.
	Target() string
}
