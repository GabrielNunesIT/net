package client

import (
	"context"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSH-specific transport implementation for NETCONF over SSH.

// SSHDialer implements the Dialer interface for SSH connections.
type SSHDialer struct {
	target string
	config *ssh.ClientConfig
}

// NewSSHDialer creates a new SSH dialer with the given target and configuration.
func NewSSHDialer(target string, config *ssh.ClientConfig) *SSHDialer {
	return &SSHDialer{target: target, config: config}
}

// Target returns the connection target address.
func (d *SSHDialer) Target() string {
	return d.target
}

// Dial establishes an SSH connection and returns the NETCONF subsystem channel.
func (d *SSHDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	tracer := ContextClientTrace(ctx)

	tracer.DialStart(d.config, d.target)
	var err error
	defer func(begin time.Time) {
		tracer.DialDone(d.config, d.target, err, time.Since(begin))
	}(time.Now())

	client, err := ssh.Dial("tcp", d.target, d.config)
	if err != nil {
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	if err = session.RequestSubsystem("netconf"); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, err
	}

	return &sshConn{client: client, session: session}, nil
}

// Close closes the SSH connection.
func (d *SSHDialer) Close(conn io.ReadWriteCloser) error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// sshConn wraps SSH client and session to implement io.ReadWriteCloser.
type sshConn struct {
	client  *ssh.Client
	session *ssh.Session
	reader  io.Reader
	writer  io.WriteCloser
}

func (c *sshConn) Read(p []byte) (n int, err error) {
	if c.reader == nil {
		c.reader, err = c.session.StdoutPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.reader.Read(p)
}

func (c *sshConn) Write(p []byte) (n int, err error) {
	if c.writer == nil {
		c.writer, err = c.session.StdinPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.writer.Write(p)
}

func (c *sshConn) Close() error {
	var firstErr error

	if c.writer != nil {
		if err := c.writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if c.session != nil {
		if err := c.session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if c.client != nil {
		if err := c.client.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// SSHClientDialer implements Dialer for an existing SSH client.
// Use this when the SSH client is already established.
type SSHClientDialer struct {
	client *ssh.Client
}

// NewSSHClientDialer creates a dialer from an existing SSH client.
func NewSSHClientDialer(client *ssh.Client) *SSHClientDialer {
	return &SSHClientDialer{client: client}
}

// Target returns the remote address of the SSH client.
func (d *SSHClientDialer) Target() string {
	return d.client.RemoteAddr().String()
}

// Dial creates a new session on the existing SSH client.
func (d *SSHClientDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	session, err := d.client.NewSession()
	if err != nil {
		return nil, err
	}

	if err = session.RequestSubsystem("netconf"); err != nil {
		_ = session.Close()
		return nil, err
	}

	// Note: we don't close the client in sshConn since it's pre-existing
	return &sshSessionConn{session: session}, nil
}

// Close does nothing for pre-existing SSH clients.
func (d *SSHClientDialer) Close(conn io.ReadWriteCloser) error {
	if conn != nil {
		return conn.Close()
	}
	return nil
}

// sshSessionConn wraps only the SSH session (for pre-existing clients).
type sshSessionConn struct {
	session *ssh.Session
	reader  io.Reader
	writer  io.WriteCloser
}

func (c *sshSessionConn) Read(p []byte) (n int, err error) {
	if c.reader == nil {
		c.reader, err = c.session.StdoutPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.reader.Read(p)
}

func (c *sshSessionConn) Write(p []byte) (n int, err error) {
	if c.writer == nil {
		c.writer, err = c.session.StdinPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.writer.Write(p)
}

func (c *sshSessionConn) Close() error {
	var firstErr error

	if c.writer != nil {
		if err := c.writer.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if c.session != nil {
		if err := c.session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}