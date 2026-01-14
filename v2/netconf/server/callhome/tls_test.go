package callhome

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	clientcallhome "github.com/damianoneill/net/v2/netconf/client/callhome"
	tlsconfig "github.com/damianoneill/net/v2/netconf/server/tls"
)

func TestTLSCallhomeDialer(t *testing.T) {
	// Generate self-signed certificate for testing
	certPEM, keyPEM, err := tlsconfig.GenerateSelfSignedCert()
	assert.NoError(t, err)

	// Create TLS server config for the device (Call Home server)
	serverConfig, err := tlsconfig.ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err)

	// Create TLS client config for the client (manager)
	// Use InsecureSkipVerify because both sides use the same cert
	clientConfig := tlsconfig.InsecureClientConfig()

	// Start client listener (manager waits for device to connect)
	ctx := clientcallhome.WithTrace(context.Background(), clientcallhome.DefaultLoggingHooks)
	clientListener, err := clientcallhome.NewTLSListener(ctx, "localhost", 0, clientConfig)
	assert.NoError(t, err)
	defer clientListener.Close()

	target := fmt.Sprintf("localhost:%d", clientListener.Port())

	// Create server dialer (device connects to manager)
	serverDialer := NewTLSDialerWithOptions(target, serverConfig, nil, DefaultLoggingHooks)

	var wg sync.WaitGroup
	var serverConn *tls.Conn
	var clientConn *tls.Conn
	var serverErr, clientErr error

	// Server (device) connects to client
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := serverDialer.Dial(context.Background())
		if err != nil {
			serverErr = err
			return
		}
		serverConn = result.(*tls.Conn)
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
		_, writeErr = clientConn.Write(testMsg)
	}()
	go func() {
		defer wg.Done()
		_, readErr = serverConn.Read(readBuf)
	}()

	wg.Wait()

	assert.NoError(t, writeErr, "Client write should succeed")
	assert.NoError(t, readErr, "Server read should succeed")
	assert.Equal(t, testMsg, readBuf, "Message should match")

	// Clean up
	_ = serverConn.Close()
	_ = clientConn.Close()
}

func TestTLSDialerTarget(t *testing.T) {
	target := "192.168.1.1:4335"
	dialer := NewTLSDialer(target, nil)
	assert.Equal(t, target, dialer.Target())
}

func TestTLSDialerClose(t *testing.T) {
	dialer := NewTLSDialer("localhost:4335", nil)
	assert.NoError(t, dialer.Close())
}
