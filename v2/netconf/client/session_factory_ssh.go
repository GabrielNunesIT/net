package client

import (
	"context"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSH-specific session factory methods for NETCONF over SSH.

// NewRPCSession connects to the target using SSH and establishes
// a netconf session with default configuration.
func NewRPCSession(ctx context.Context, sshcfg *ssh.ClientConfig, target string) (Session, error) {
	return NewRPCSessionWithConfig(ctx, sshcfg, target, DefaultConfig)
}

// NewRPCSessionWithConfig connects to the target using SSH and establishes
// a netconf session with the client configuration.
func NewRPCSessionWithConfig(ctx context.Context, sshcfg *ssh.ClientConfig, target string, cfg *Config) (Session, error) {
	dialer := NewSSHDialer(target, sshcfg)
	return NewRPCSessionFromDialer(ctx, dialer, cfg)
}

// NewRPCSessionFromSSHClient establishes a netconf session over the given SSH client
// with default configuration.
func NewRPCSessionFromSSHClient(ctx context.Context, client *ssh.Client) (Session, error) {
	return NewRPCSessionFromSSHClientWithConfig(ctx, client, DefaultConfig)
}

// NewRPCSessionFromSSHClientWithConfig establishes a netconf session over the given
// SSH client with the client configuration.
func NewRPCSessionFromSSHClientWithConfig(ctx context.Context, client *ssh.Client, cfg *Config) (Session, error) {
	dialer := NewSSHClientDialer(client)
	return NewRPCSessionFromDialer(ctx, dialer, cfg)
}

// NewDialer creates a new SSH dialer.
// DEPRECATED: Use NewSSHDialer instead.
func NewDialer(target string, clientConfig *ssh.ClientConfig) *RealDialer {
	return &RealDialer{target: target, config: clientConfig}
}

// RealDialer implements SSHClientFactory for backward compatibility.
// DEPRECATED: Use SSHDialer instead.
type RealDialer struct {
	target string
	config *ssh.ClientConfig
}

// Dial establishes an SSH client connection.
func (rd *RealDialer) Dial(ctx context.Context) (*ssh.Client, error) {
	tracer := ContextClientTrace(ctx)

	tracer.DialStart(rd.config, rd.target)
	var err error
	defer func(begin time.Time) {
		tracer.DialDone(rd.config, rd.target, err, time.Since(begin))
	}(time.Now())

	return ssh.Dial("tcp", rd.target, rd.config)
}

// Close closes the SSH client.
func (rd *RealDialer) Close(cli *ssh.Client) error {
	if cli != nil {
		return cli.Close()
	}
	return nil
}
