package callhome

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"

	clientcallhome "github.com/damianoneill/net/v2/netconf/client/callhome"
	sshserver "github.com/damianoneill/net/v2/netconf/server/ssh"
)

func TestSSHCallhomeDialer(t *testing.T) {
	// Create SSH server config for the device (Call Home server)
	serverConfig, err := sshserver.PasswordConfig("user", "password")
	assert.NoError(t, err)

	// Create SSH client config for the client (manager)
	clientConfig := &ssh.ClientConfig{
		User:            "user",
		Auth:            []ssh.AuthMethod{ssh.Password("password")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Start client listener (manager waits for device to connect)
	ctx := clientcallhome.WithTrace(context.Background(), clientcallhome.DefaultLoggingHooks)
	clientListener, err := clientcallhome.NewSSHListener(ctx, "localhost", 0, clientConfig)
	assert.NoError(t, err)
	defer clientListener.Close()

	target := fmt.Sprintf("localhost:%d", clientListener.Port())

	// Create server dialer (device connects to manager)
	serverDialer := NewSSHDialerWithOptions(target, serverConfig, nil, DefaultLoggingHooks)

	var wg sync.WaitGroup
	var serverConn, clientConn interface{}
	var serverErr, clientErr error

	// Server (device) connects to client
	wg.Add(1)
	go func() {
		defer wg.Done()
		serverConn, serverErr = serverDialer.Dial(context.Background())
	}()

	// Client (manager) accepts connection
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		clientConn, clientErr = clientListener.Accept(ctx)
	}()

	wg.Wait()

	assert.NoError(t, serverErr, "Server dial should succeed")
	assert.NoError(t, clientErr, "Client accept should succeed")
	assert.NotNil(t, serverConn, "Server connection should not be nil")
	assert.NotNil(t, clientConn, "Client connection should not be nil")

	// Test communication
	testMsg := []byte("hello from client")
	var writeErr, readErr error
	var readBuf = make([]byte, len(testMsg))

	wg.Add(2)
	go func() {
		defer wg.Done()
		_, writeErr = clientConn.(interface{ Write([]byte) (int, error) }).Write(testMsg)
	}()
	go func() {
		defer wg.Done()
		_, readErr = serverConn.(interface{ Read([]byte) (int, error) }).Read(readBuf)
	}()

	wg.Wait()

	assert.NoError(t, writeErr, "Client write should succeed")
	assert.NoError(t, readErr, "Server read should succeed")
	assert.Equal(t, testMsg, readBuf, "Message should match")

	// Clean up
	_ = serverConn.(interface{ Close() error }).Close()
	_ = clientConn.(interface{ Close() error }).Close()
}

func TestSSHDialerTarget(t *testing.T) {
	target := "192.168.1.1:4334"
	dialer := NewSSHDialer(target, nil)
	assert.Equal(t, target, dialer.Target())
}

func TestSSHDialerClose(t *testing.T) {
	dialer := NewSSHDialer("localhost:4334", nil)
	assert.NoError(t, dialer.Close())
}
