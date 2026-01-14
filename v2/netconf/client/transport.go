package client

import (
	"context"
	"io"
	"time"
)

// The Secure Transport layer provides a communication path between
// the client and server.  NETCONF can be layered over any
// transport protocol that provides a set of basic requirements.

// Transport interface defines what characteristics make up a NETCONF transport
// layer object.
type Transport interface {
	io.ReadWriteCloser
	// Target returns the connection target address for tracing purposes.
	Target() string
}

// transportImpl is the generic transport implementation that wraps any Dialer.
type transportImpl struct {
	conn   io.ReadWriteCloser
	trace  *ClientTrace
	target string
	dialer Dialer
}

// NewTransport creates a new transport using the provided Dialer.
// This is the generic transport factory that works with any Dialer implementation.
func NewTransport(ctx context.Context, dialer Dialer) (rt Transport, err error) {
	impl := &transportImpl{
		target: dialer.Target(),
		dialer: dialer,
		trace:  ContextClientTrace(ctx),
	}

	impl.trace.ConnectStart(impl.target)

	defer func(begin time.Time) {
		impl.trace.ConnectDone(impl.target, err, time.Since(begin))
	}(time.Now())

	defer func() {
		if err != nil && impl.conn != nil {
			_ = dialer.Close(impl.conn)
		}
	}()

	impl.conn, err = dialer.Dial(ctx)
	if err != nil {
		return nil, err
	}

	return impl, nil
}

// Target returns the connection target address.
func (t *transportImpl) Target() string {
	return t.target
}

func (t *transportImpl) Read(p []byte) (n int, err error) {
	t.trace.ReadStart(p)
	defer func(begin time.Time) {
		t.trace.ReadDone(p, n, err, time.Since(begin))
	}(time.Now())

	return t.conn.Read(p)
}

func (t *transportImpl) Write(p []byte) (n int, err error) {
	t.trace.WriteStart(p)
	defer func(begin time.Time) {
		t.trace.WriteDone(p, n, err, time.Since(begin))
	}(time.Now())

	return t.conn.Write(p)
}

// Close closes the transport connection.
func (t *transportImpl) Close() (err error) {
	defer t.trace.ConnectionClosed(t.target, err)

	if t.conn != nil {
		err = t.dialer.Close(t.conn)
	}
	return err
}
