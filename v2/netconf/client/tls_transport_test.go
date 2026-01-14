package client

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"testing"
	"time"

	"github.com/damianoneill/net/v2/netconf/testserver"
	assert "github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func TestSuccessfulTLSConnection(t *testing.T) {
	ts := testserver.NewTLSServer(t)
	defer ts.Close()

	tlsConfig := createTestTLSClientConfig(t, ts.CertPEM)

	tr, err := newTLSTransport(dftContext, ts.Port(), tlsConfig)
	assert.NoError(t, err, "Not expecting new TLS transport to fail")
	defer tr.Close()
}

func TestFailingTLSConnection(t *testing.T) {
	ts := testserver.NewTLSServer(t)
	defer ts.Close()

	// Use a config that doesn't trust the server's certificate
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	tr, err := newTLSTransport(dftContext, ts.Port(), tlsConfig)
	assert.Error(t, err, "Not expecting new TLS transport to succeed")
	assert.Nil(t, tr, "Transport should not be defined")
}

func TestTLSWriteRead(t *testing.T) {
	ts := testserver.NewTLSServer(t)
	defer ts.Close()

	tlsConfig := createTestTLSClientConfig(t, ts.CertPEM)

	tr, err := newTLSTransport(dftContext, ts.Port(), tlsConfig)
	assert.NoError(t, err, "Not expecting new TLS transport to fail")
	defer tr.Close()

	rdr := bufio.NewReader(tr)
	_, _ = tr.Write([]byte("Message\n"))
	response, _ := rdr.ReadString('\n')
	assert.Equal(t, "GOT:Message\n", response, "Failed to get expected response")
}

func TestTLSTrace(t *testing.T) {
	ts := testserver.NewTLSServer(t)
	defer ts.Close()

	tlsConfig := createTestTLSClientConfig(t, ts.CertPEM)

	var traces []string
	trace := &ClientTrace{
		ConnectStart: func(target string) {
			traces = append(traces, fmt.Sprintf("ConnectStart %s", target))
		},
		ConnectDone: func(target string, err error, d time.Duration) {
			traces = append(traces, fmt.Sprintf("ConnectDone %s error:%v", target, err))
			assert.True(t, d > 0, "Duration should be defined")
		},
		DialStart: func(clientConfig *ssh.ClientConfig, target string) {
			traces = append(traces, fmt.Sprintf("DialStart %s", target))
		},
		DialDone: func(clientConfig *ssh.ClientConfig, target string, err error, d time.Duration) {
			traces = append(traces, fmt.Sprintf("DialDone %s error:%v", target, err))
			assert.True(t, d > 0, "Duration should be defined")
		},
		ConnectionClosed: func(target string, err error) {
			traces = append(traces, fmt.Sprintf("ConnectionClosed target:%s error:%v", target, err))
		},
	}

	ctx := WithClientTrace(context.Background(), trace)
	tr, _ := newTLSTransport(ctx, ts.Port(), tlsConfig)

	_, _ = tr.Write([]byte("Message\n"))
	_, _ = bufio.NewReader(tr).ReadString('\n')

	tr.Close()

	assert.Equal(t, fmt.Sprintf("ConnectStart localhost:%d", ts.Port()), traces[0])
	assert.Equal(t, fmt.Sprintf("DialStart localhost:%d", ts.Port()), traces[1])
	assert.Equal(t, fmt.Sprintf("DialDone localhost:%d error:<nil>", ts.Port()), traces[2])
	assert.Equal(t, fmt.Sprintf("ConnectDone localhost:%d error:<nil>", ts.Port()), traces[3])
	assert.Contains(t, traces[4], "ConnectionClosed target:localhost:")
}

func createTestTLSClientConfig(t *testing.T, caCertPEM []byte) *tls.Config {
	caCertPool := x509.NewCertPool()
	ok := caCertPool.AppendCertsFromPEM(caCertPEM)
	assert.True(t, ok, "Failed to parse CA certificate")

	return &tls.Config{
		RootCAs:    caCertPool,
		ServerName: "localhost",
		MinVersion: tls.VersionTLS12,
	}
}

func newTLSTransport(ctx context.Context, port int, cfg *tls.Config) (Transport, error) {
	target := fmt.Sprintf("localhost:%d", port)
	return NewTransport(ctx, NewTLSDialer(target, cfg))
}

