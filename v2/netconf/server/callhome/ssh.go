package callhome

import (
	"context"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHDialer implements Call Home for SSH transport.
// The server connects to the client, then waits for the client to initiate SSH.
type SSHDialer struct {
	target    string
	config    *ssh.ServerConfig
	netDialer NetDialer
	trace     *Trace
}

// NewSSHDialer creates a new SSH Call Home dialer.
// The config is an SSH server configuration since the device acts as SSH server
// even though it initiates the TCP connection.
func NewSSHDialer(target string, config *ssh.ServerConfig) *SSHDialer {
	return &SSHDialer{
		target:    target,
		config:    config,
		netDialer: DefaultNetDialer,
		trace:     noOpTrace,
	}
}

// NewSSHDialerWithOptions creates a new SSH Call Home dialer with custom options.
// If netDialer is nil, DefaultNetDialer is used.
// If trace is nil, a no-op trace is used.
func NewSSHDialerWithOptions(target string, config *ssh.ServerConfig, netDialer NetDialer, trace *Trace) *SSHDialer {
	d := NewSSHDialer(target, config)
	if netDialer != nil {
		d.netDialer = netDialer
	}
	if trace != nil {
		d.trace = trace
	}
	return d
}

// Target returns the client address being dialed.
func (d *SSHDialer) Target() string {
	return d.target
}

// Dial connects to the client and performs SSH handshake as server.
// Per RFC 8071, the server initiates TCP but the client initiates SSH.
func (d *SSHDialer) Dial(ctx context.Context) (io.ReadWriteCloser, error) {
	d.trace.DialStart(d.target)

	var err error
	defer func(begin time.Time) {
		d.trace.DialDone(d.target, err, time.Since(begin))
	}(time.Now())

	// Dial the client using the configured NetDialer
	conn, err := d.netDialer.DialContext(ctx, "tcp", d.target)
	if err != nil {
		return nil, fmt.Errorf("callhome: failed to connect to client: %w", err)
	}

	// Now perform SSH handshake as server (client initiates SSH per RFC 8071)
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, d.config)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("callhome: SSH handshake failed: %w", err)
	}

	go ssh.DiscardRequests(reqs)

	d.trace.SSHConnected(d.target, sshConn)

	// Wait for the netconf subsystem channel
	return d.handleChannels(sshConn, chans)
}

// handleChannels waits for a session channel and netconf subsystem request.
func (d *SSHDialer) handleChannels(sshConn *ssh.ServerConn, chans <-chan ssh.NewChannel) (io.ReadWriteCloser, error) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			return nil, fmt.Errorf("callhome: failed to accept channel: %w", err)
		}

		// Handle incoming requests (looking for subsystem request)
		subsystemChan := make(chan bool, 1)
		go func() {
			for req := range requests {
				if req.Type == "subsystem" && string(req.Payload[4:]) == "netconf" {
					_ = req.Reply(true, nil)
					subsystemChan <- true
					// Continue handling other requests
					for r := range requests {
						_ = r.Reply(false, nil)
					}
					return
				}
				_ = req.Reply(false, nil)
			}
			subsystemChan <- false
		}()

		// Wait for subsystem request
		if ok := <-subsystemChan; !ok {
			_ = channel.Close()
			return nil, fmt.Errorf("callhome: netconf subsystem not requested")
		}

		d.trace.SubsystemReady(d.target)

		return &sshCallhomeConn{
			conn:    sshConn,
			channel: channel,
		}, nil
	}

	return nil, fmt.Errorf("callhome: no session channel received")
}

// Close closes the dialer (no-op for SSH dialer).
func (d *SSHDialer) Close() error {
	return nil
}

// sshCallhomeConn wraps an SSH connection for Call Home.
type sshCallhomeConn struct {
	conn    *ssh.ServerConn
	channel ssh.Channel
}

func (c *sshCallhomeConn) Read(p []byte) (n int, err error) {
	return c.channel.Read(p)
}

func (c *sshCallhomeConn) Write(p []byte) (n int, err error) {
	return c.channel.Write(p)
}

func (c *sshCallhomeConn) Close() error {
	if err := c.channel.Close(); err != nil {
		return err
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}
