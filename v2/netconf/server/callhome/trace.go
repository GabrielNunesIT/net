package callhome

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

// Trace defines hooks for tracing Call Home operations.
type Trace struct {
	// DialStart is called when a dial operation begins.
	DialStart func(target string)

	// DialDone is called when a dial operation completes.
	DialDone func(target string, err error, duration time.Duration)

	// SSHConnected is called when SSH connection is established.
	SSHConnected func(target string, conn *ssh.ServerConn)

	// SubsystemReady is called when the netconf subsystem is ready.
	SubsystemReady func(target string)

	// TLSConnected is called when TLS connection is established.
	TLSConnected func(target string, conn *tls.Conn)

	// AcceptStart is called when starting to accept connections.
	AcceptStart func(addr net.Addr)

	// AcceptDone is called when a connection is accepted.
	AcceptDone func(conn net.Conn, err error)
}

// noOpTrace is a trace that does nothing (default).
var noOpTrace = &Trace{
	DialStart:      func(string) {},
	DialDone:       func(string, error, time.Duration) {},
	SSHConnected:   func(string, *ssh.ServerConn) {},
	SubsystemReady: func(string) {},
	TLSConnected:   func(string, *tls.Conn) {},
	AcceptStart:    func(net.Addr) {},
	AcceptDone:     func(net.Conn, error) {},
}

// DefaultLoggingHooks provides trace hooks that log operations.
var DefaultLoggingHooks = &Trace{
	DialStart: func(target string) {
		log.Printf("callhome: dialing %s", target)
	},
	DialDone: func(target string, err error, duration time.Duration) {
		if err != nil {
			log.Printf("callhome: dial to %s failed: %v (took %v)", target, err, duration)
		} else {
			log.Printf("callhome: dial to %s succeeded (took %v)", target, duration)
		}
	},
	SSHConnected: func(target string, conn *ssh.ServerConn) {
		log.Printf("callhome: SSH connected to %s, user=%s", target, conn.User())
	},
	SubsystemReady: func(target string) {
		log.Printf("callhome: netconf subsystem ready for %s", target)
	},
	TLSConnected: func(target string, conn *tls.Conn) {
		state := conn.ConnectionState()
		log.Printf("callhome: TLS connected to %s, version=0x%x", target, state.Version)
	},
	AcceptStart: func(addr net.Addr) {
		log.Printf("callhome: listening on %s", addr)
	},
	AcceptDone: func(conn net.Conn, err error) {
		if err != nil {
			log.Printf("callhome: accept failed: %v", err)
		} else {
			log.Printf("callhome: accepted connection from %s", conn.RemoteAddr())
		}
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
