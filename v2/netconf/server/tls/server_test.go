//nolint:dupl
package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"testing"

	assert "github.com/stretchr/testify/require"
)

type testHandler struct{}

func (h *testHandler) Handle(conn net.Conn) {
	buffer := make([]byte, 5)
	_, _ = conn.Read(buffer)
	_, _ = conn.Write([]byte(">" + string(buffer) + "<"))
}

func handlerFactory() HandlerFactory {
	return func(conn *tls.Conn) Handler {
		return &testHandler{}
	}
}

func TestTLSServer(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSignedCert()
	assert.NoError(t, err)

	tlsConfig, err := ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err)

	ctx := WithTLSTrace(context.Background(), DefaultLoggingHooks)
	server, err := NewServer(ctx, "localhost", 0, tlsConfig, handlerFactory())
	assert.NotNil(t, server)
	assert.NoError(t, err)
	defer server.Close()

	//----------------------------

	clientConfig := createClientConfig(t, certPEM)
	conn, err := tls.Dial("tcp", fmt.Sprintf("localhost:%d", server.Port()), clientConfig)
	assert.NoError(t, err, "Not expecting TLS dial to fail")
	defer conn.Close()

	_, _ = conn.Write([]byte("hello"))
	buffer := make([]byte, 7)
	_, _ = conn.Read(buffer)
	assert.Equal(t, ">hello<", string(buffer))
}

func TestTLSServerListenFailure(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSignedCert()
	assert.NoError(t, err)

	tlsConfig, err := ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err)

	ctx := WithTLSTrace(context.Background(), DefaultLoggingHooks)
	server, err := NewServer(ctx, "9.9.9.9", 9999, tlsConfig, handlerFactory())
	assert.Nil(t, server)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "assign requested address")
}

func TestTLSServerNoOpTraceHooks(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSignedCert()
	assert.NoError(t, err)

	tlsConfig, err := ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err)

	ctx := context.Background()
	server, err := NewServer(ctx, "localhost", 0, tlsConfig, handlerFactory())
	assert.NotNil(t, server)
	assert.NoError(t, err)
	defer server.Close()

	//----------------------------

	clientConfig := createClientConfig(t, certPEM)
	conn, err := tls.Dial("tcp", fmt.Sprintf("localhost:%d", server.Port()), clientConfig)
	assert.NoError(t, err, "Not expecting TLS dial to fail")
	defer conn.Close()

	_, _ = conn.Write([]byte("hello"))
	buffer := make([]byte, 7)
	_, _ = conn.Read(buffer)
	assert.Equal(t, ">hello<", string(buffer))
}

func TestTLSServerDiagnosticTraceHooks(t *testing.T) {
	certPEM, keyPEM, err := GenerateSelfSignedCert()
	assert.NoError(t, err)

	tlsConfig, err := ServerConfig(certPEM, keyPEM)
	assert.NoError(t, err)

	ctx := WithTLSTrace(context.Background(), DiagnosticLoggingHooks)
	server, err := NewServer(ctx, "localhost", 0, tlsConfig, handlerFactory())
	assert.NotNil(t, server)
	assert.NoError(t, err)
	defer server.Close()

	//----------------------------

	clientConfig := createClientConfig(t, certPEM)
	conn, err := tls.Dial("tcp", fmt.Sprintf("localhost:%d", server.Port()), clientConfig)
	assert.NoError(t, err, "Not expecting TLS dial to fail")
	defer conn.Close()

	_, _ = conn.Write([]byte("hello"))
	buffer := make([]byte, 7)
	_, _ = conn.Read(buffer)
	assert.Equal(t, ">hello<", string(buffer))
}

func createClientConfig(t *testing.T, caCertPEM []byte) *tls.Config {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCertPEM)
	assert.True(t, ok, "Failed to parse CA certificate")

	return &tls.Config{
		RootCAs:    caCertPool,
		MinVersion: tls.VersionTLS12,
	}
}
