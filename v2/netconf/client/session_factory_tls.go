package client

import (
	"context"
	"crypto/tls"
)

// TLS-specific session factory methods for NETCONF over TLS (RFC 7589).

// NewRPCSessionTLS connects to the target using TLS and establishes
// a netconf session with default configuration.
func NewRPCSessionTLS(ctx context.Context, tlsConfig *tls.Config, target string) (Session, error) {
	return NewRPCSessionTLSWithConfig(ctx, tlsConfig, target, DefaultConfig)
}

// NewRPCSessionTLSWithConfig connects to the target using TLS and establishes
// a netconf session with the client configuration.
func NewRPCSessionTLSWithConfig(ctx context.Context, tlsConfig *tls.Config, target string, cfg *Config) (Session, error) {
	dialer := NewTLSDialer(target, tlsConfig)
	return NewRPCSessionFromDialer(ctx, dialer, cfg)
}

// NewRPCSessionFromTLSConn establishes a netconf session over the given TLS connection
// with default configuration.
func NewRPCSessionFromTLSConn(ctx context.Context, conn *tls.Conn) (Session, error) {
	return NewRPCSessionFromTLSConnWithConfig(ctx, conn, DefaultConfig)
}

// NewRPCSessionFromTLSConnWithConfig establishes a netconf session over the given
// TLS connection with the client configuration.
func NewRPCSessionFromTLSConnWithConfig(ctx context.Context, conn *tls.Conn, cfg *Config) (Session, error) {
	dialer := NewTLSConnDialer(conn)
	return NewRPCSessionFromDialer(ctx, dialer, cfg)
}
