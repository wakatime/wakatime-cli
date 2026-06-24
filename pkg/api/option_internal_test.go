package api

import (
	"bytes"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func doAndClose(t *testing.T, client *Client, req *http.Request) error {
	t.Helper()

	resp, err := client.Do(t.Context(), req)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}

	return err
}

func TestClientDoDNSFallback(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, BaseURL+"/users/current", nil)
	require.NoError(t, err)

	var calls int

	client := &Client{
		baseURL: BaseURL,
		doFunc: func(_ *Client, req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return nil, &net.DNSError{Err: "no such host", Name: "api.wakatime.com"}
			}

			assert.Contains(t, []string{BaseIPAddrv4, BaseIPAddrv6}, req.URL.Host)

			return &http.Response{
				StatusCode: http.StatusAccepted,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		},
	}

	resp, err := client.Do(t.Context(), req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, 2, calls)
}

func TestClientDoDNSFallbackRetryError(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, BaseURL+"/users/current", nil)
	require.NoError(t, err)

	var calls int

	client := &Client{
		baseURL: BaseURL,
		doFunc: func(_ *Client, _ *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				return nil, &net.DNSError{Err: "no such host", Name: "api.wakatime.com"}
			}

			return nil, errors.New("retry failed")
		},
	}

	err = doAndClose(t, client, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "retry request failed: retry failed")
	assert.Equal(t, 2, calls)
}

func TestClientDoDoesNotFallback(t *testing.T) {
	dnsErr := &net.DNSError{Err: "no such host", Name: "custom.example.com"}
	tests := map[string]struct {
		BaseURL string
		Err     error
	}{
		"custom base url": {
			BaseURL: "https://custom.example.com/api",
			Err:     dnsErr,
		},
		"non dns error": {
			BaseURL: BaseURL,
			Err:     errors.New("connection refused"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, test.BaseURL, nil)
			require.NoError(t, err)

			var calls int

			client := &Client{
				baseURL: test.BaseURL,
				doFunc: func(_ *Client, _ *http.Request) (*http.Response, error) {
					calls++

					return nil, test.Err
				},
			}

			err = doAndClose(t, client, req)

			require.Error(t, err)
			assert.ErrorIs(t, err, test.Err)
			assert.Equal(t, 1, calls)
		})
	}
}

func TestRequestCloneHelpers(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	cloned, err := cloneRequest(req)
	require.NoError(t, err)
	assert.NotSame(t, req, cloned)

	req, err = http.NewRequest(http.MethodPost, "https://example.com", io.NopCloser(strings.NewReader("body")))
	require.NoError(t, err)

	_, err = cloneRequest(req)
	require.EqualError(t, err, "request body is not replayable")

	req, err = http.NewRequest(http.MethodPost, "https://example.com", bytes.NewReader([]byte("body")))
	require.NoError(t, err)

	cloned, err = cloneRequest(req)
	require.NoError(t, err)

	defer cloned.Body.Close()

	contents, err := io.ReadAll(cloned.Body)
	require.NoError(t, err)
	assert.Equal(t, "body", string(contents))
}

func TestShouldRetryProxyWithHTTP(t *testing.T) {
	assert.True(t, shouldRetryProxyWithHTTP(errors.New("proxyconnect tcp: server gave HTTP response to HTTPS client")))
	assert.True(t, shouldRetryProxyWithHTTP(errors.New("server gave HTTP response to HTTPS client")))
	assert.False(t, shouldRetryProxyWithHTTP(errors.New("tls handshake timeout")))
}

func TestTLSOptions(t *testing.T) {
	client := NewClient("https://example.com", WithDisableSSLVerify())
	transport := client.client.Transport.(*http.Transport)
	require.NotNil(t, transport.TLSClientConfig)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)

	pool := CACerts(t.Context())
	client = NewClient("https://example.com", WithSSLCertPool(pool))
	transport = client.client.Transport.(*http.Transport)
	require.NotNil(t, transport.TLSClientConfig)
	assert.Same(t, pool, transport.TLSClientConfig.RootCAs)

	transport = NewTransportWithHostVerificationDisabled(t.Context())
	require.NotNil(t, transport.TLSClientConfig)
	assert.Equal(t, serverName, transport.TLSClientConfig.ServerName)
	assert.NotNil(t, transport.TLSClientConfig.RootCAs)
}
