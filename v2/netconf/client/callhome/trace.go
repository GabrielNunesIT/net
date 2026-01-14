package callhome

import (
	"context"
	"crypto/tls"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// Trace defines hooks for tracing Call Home listener operations.
type Trace struct {
	// ListenStart is called when the listener starts.
	ListenStart func(addr net.Addr)

	// AcceptDone is called when a connection is accepted.
	AcceptDone func(conn net.Conn, err error)

	// SSHConnected is called when SSH connection is established.
	SSHConnected func(target string, client *ssh.Client)

	// SubsystemReady is called when the netconf subsystem is ready.
	SubsystemReady func(target string)

	// TLSConnected is called when TLS connection is established.
	TLSConnected func(target string, conn *tls.Conn)
}

// noOpTrace is a trace that does nothing (default).
var noOpTrace = &Trace{
	ListenStart:    func(net.Addr) {},
	AcceptDone:     func(net.Conn, error) {},
	SSHConnected:   func(string, *ssh.Client) {},
	SubsystemReady: func(string) {},
	TLSConnected:   func(string, *tls.Conn) {},
}

// DefaultLoggingHooks provides trace hooks that log operations.
var DefaultLoggingHooks = &Trace{
	ListenStart: func(addr net.Addr) {
		log.Printf("callhome: listening on %s", addr)
	},
	AcceptDone: func(conn net.Conn, err error) {
		if err != nil {
			log.Printf("callhome: accept failed: %v", err)
		} else {
			log.Printf("callhome: accepted connection from %s", conn.RemoteAddr())
		}
	},
	SSHConnected: func(target string, client *ssh.Client) {
		log.Printf("callhome: SSH client connected to %s", target)
	},
	SubsystemReady: func(target string) {
		log.Printf("callhome: netconf subsystem ready for %s", target)
	},
	TLSConnected: func(target string, conn *tls.Conn) {
		state := conn.ConnectionState()
		log.Printf("callhome: TLS connected to %s, version=0x%x", target, state.Version)
	},
}

type traceKey struct{}

// WithTrace returns a context with the given trace attached.
func WithTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceKey{}, trace)
}

// ContextTrace returns the trace from the context, or noOpTrace if none.
func ContextTrace(ctx context.Context) *Trace {
	if trace, ok := ctx.Value(traceKey{}).(*Trace); ok && trace != nil {
		return trace
	}
	return noOpTrace
}
