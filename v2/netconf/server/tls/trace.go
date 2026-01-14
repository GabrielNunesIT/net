package tls

import (
	"context"
	"log"
	"net"
)

// Trace defines the hooks for tracing TLS server events.
type Trace struct {
	// Listened is called when the server starts listening.
	Listened func(address string, err error)
	// StartAccepting is called when the server starts accepting connections.
	StartAccepting func()
	// Accepted is called when a connection is accepted.
	Accepted func(conn net.Conn, err error)
	// TLSHandshake is called after the TLS handshake completes.
	TLSHandshake func(conn net.Conn, err error)
	// ConnectionClosed is called when a connection is closed.
	ConnectionClosed func(conn net.Conn, err error)
}

type traceContextKeyType int

const traceContextKey traceContextKeyType = 1

// ContextTLSTrace returns the Trace from the context, or a no-op trace if none is set.
func ContextTLSTrace(ctx context.Context) *Trace {
	if ctx == nil {
		return noOpTrace
	}
	trace, ok := ctx.Value(traceContextKey).(*Trace)
	if !ok || trace == nil {
		return noOpTrace
	}
	return trace
}

// WithTLSTrace returns a context with the given Trace attached.
func WithTLSTrace(ctx context.Context, trace *Trace) context.Context {
	return context.WithValue(ctx, traceContextKey, trace)
}

// noOpTrace provides default no-op implementations for all hooks.
var noOpTrace = &Trace{
	Listened:         func(address string, err error) {},
	StartAccepting:   func() {},
	Accepted:         func(conn net.Conn, err error) {},
	TLSHandshake:     func(conn net.Conn, err error) {},
	ConnectionClosed: func(conn net.Conn, err error) {},
}

// DefaultLoggingHooks provides default logging for all trace hooks.
var DefaultLoggingHooks = &Trace{
	Listened: func(address string, err error) {
		log.Printf("TLS Server listening on %s, error: %v", address, err)
	},
	StartAccepting: func() {
		log.Printf("TLS Server started accepting connections")
	},
	Accepted: func(conn net.Conn, err error) {
		if err == nil {
			log.Printf("TLS Server accepted connection from %s", conn.RemoteAddr())
		} else {
			log.Printf("TLS Server accept error: %v", err)
		}
	},
	TLSHandshake: func(conn net.Conn, err error) {
		if err == nil {
			log.Printf("TLS handshake completed for %s", conn.RemoteAddr())
		} else {
			log.Printf("TLS handshake failed for %s: %v", conn.RemoteAddr(), err)
		}
	},
	ConnectionClosed: func(conn net.Conn, err error) {
		log.Printf("TLS connection closed from %s, error: %v", conn.RemoteAddr(), err)
	},
}

// DiagnosticLoggingHooks provides more verbose logging for debugging.
var DiagnosticLoggingHooks = &Trace{
	Listened: func(address string, err error) {
		log.Printf("[DIAG] TLS Server listening on %s, error: %v", address, err)
	},
	StartAccepting: func() {
		log.Printf("[DIAG] TLS Server started accepting connections")
	},
	Accepted: func(conn net.Conn, err error) {
		if err == nil {
			log.Printf("[DIAG] TLS Server accepted connection from %s", conn.RemoteAddr())
		} else {
			log.Printf("[DIAG] TLS Server accept error: %v", err)
		}
	},
	TLSHandshake: func(conn net.Conn, err error) {
		if err == nil {
			log.Printf("[DIAG] TLS handshake completed for %s", conn.RemoteAddr())
		} else {
			log.Printf("[DIAG] TLS handshake failed for %s: %v", conn.RemoteAddr(), err)
		}
	},
	ConnectionClosed: func(conn net.Conn, err error) {
		log.Printf("[DIAG] TLS connection closed from %s, error: %v", conn.RemoteAddr(), err)
	},
}
