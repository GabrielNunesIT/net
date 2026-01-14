package client

import (
	"context"

	"github.com/imdario/mergo"
)

// Session factory shared logic for creating NETCONF sessions using the Dialer interface.

// NewRPCSessionFromDialer creates a NETCONF session using any Dialer implementation.
// This is the core session creation logic that SSH and TLS factory methods delegate to.
func NewRPCSessionFromDialer(ctx context.Context, dialer Dialer, cfg *Config) (s Session, err error) {
	// Use supplied config, but apply any defaults to unspecified values.
	resolvedConfig := *cfg
	_ = mergo.Merge(&resolvedConfig, DefaultConfig)

	var t Transport
	if t, err = NewTransport(ctx, dialer); err != nil {
		return
	}

	if s, err = NewSession(ctx, t, &resolvedConfig); err != nil {
		_ = t.Close()
	}
	return
}

// NewRPCSessionFromDialerWithDefaults creates a NETCONF session using any Dialer implementation
// with default configuration.
func NewRPCSessionFromDialerWithDefaults(ctx context.Context, dialer Dialer) (Session, error) {
	return NewRPCSessionFromDialer(ctx, dialer, DefaultConfig)
}
