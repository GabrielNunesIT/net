package callhome

import (
	"context"
	"fmt"
	"io"
	"net"

	"golang.org/x/crypto/ssh"
)

// SSHListener listens for Call Home connections and initiates SSH as client.
type SSHListener struct {
	listener net.Listener
	config   *ssh.ClientConfig
	trace    *Trace
}

// NewSSHListener creates a new SSH Call Home listener.
// The config is an SSH client configuration since the manager initiates SSH
// even though it receives the TCP connection.
func NewSSHListener(ctx context.Context, address string, port int, config *ssh.ClientConfig) (*SSHListener, error) {
	listenAddr := fmt.Sprintf("%s:%d", address, port)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("callhome: failed to listen on %s: %w", listenAddr, err)
	}

	trace := ContextTrace(ctx)
	trace.ListenStart(listener.Addr())

	return &SSHListener{
		listener: listener,
		config:   config,
		trace:    trace,
	}, nil
}

// Accept waits for a server to connect, then initiates SSH as client.
// Returns a connection ready for NETCONF session establishment.
func (l *SSHListener) Accept(ctx context.Context) (io.ReadWriteCloser, error) {
	conn, err := l.listener.Accept()
	l.trace.AcceptDone(conn, err)
	if err != nil {
		return nil, fmt.Errorf("callhome: accept failed: %w", err)
	}

	// Initiate SSH as client (per RFC 8071, client initiates SSH)
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, conn.RemoteAddr().String(), l.config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("callhome: SSH client handshake failed: %w", err)
	}

	client := ssh.NewClient(sshConn, chans, reqs)
	l.trace.SSHConnected(conn.RemoteAddr().String(), client)

	// Create a session and request NETCONF subsystem
	session, err := client.NewSession()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("callhome: failed to create SSH session: %w", err)
	}

	if err = session.RequestSubsystem("netconf"); err != nil {
		_ = session.Close()
		_ = client.Close()
		return nil, fmt.Errorf("callhome: failed to request netconf subsystem: %w", err)
	}

	l.trace.SubsystemReady(conn.RemoteAddr().String())

	return &sshCallhomeConn{
		client:  client,
		session: session,
	}, nil
}

// Close closes the listener.
func (l *SSHListener) Close() error {
	return l.listener.Close()
}

// Addr returns the listener's network address.
func (l *SSHListener) Addr() net.Addr {
	return l.listener.Addr()
}

// Port returns the port the listener is bound to.
func (l *SSHListener) Port() int {
	return l.listener.Addr().(*net.TCPAddr).Port
}

// sshCallhomeConn wraps an SSH client connection for Call Home.
type sshCallhomeConn struct {
	client  *ssh.Client
	session *ssh.Session
	reader  io.Reader
	writer  io.WriteCloser
}

func (c *sshCallhomeConn) Read(p []byte) (n int, err error) {
	if c.reader == nil {
		c.reader, err = c.session.StdoutPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.reader.Read(p)
}

func (c *sshCallhomeConn) Write(p []byte) (n int, err error) {
	if c.writer == nil {
		c.writer, err = c.session.StdinPipe()
		if err != nil {
			return 0, err
		}
	}
	return c.writer.Write(p)
}

func (c *sshCallhomeConn) Close() error {
	if c.writer != nil {
		if err := c.writer.Close(); err != nil {
			return err
		}
	}
	if err := c.session.Close(); err != nil {
		return err
	}
	if err := c.client.Close(); err != nil {
		return err
	}
	return nil
}

// SSHConnDialer wraps an established Call Home SSH connection as a Dialer.
// This allows integration with existing session factory functions.
type SSHConnDialer struct {
	conn io.ReadWriteCloser
	addr string
}

// NewSSHConnDialer creates a dialer from an established Call Home connection.
func NewSSHConnDialer(conn io.ReadWriteCloser, remoteAddr string) *SSHConnDialer {
	return &SSHConnDialer{conn: conn, addr: remoteAddr}
}

// Target returns the remote address.
func (d *SSHConnDialer) Target() string {
	return d.addr
}

// Dial returns the pre-established connection.
func (d *SSHConnDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	return d.conn, nil
}

// Close is a no-op; the connection should be closed separately.
func (d *SSHConnDialer) Close(conn io.ReadWriteCloser) error {
	return nil
}
