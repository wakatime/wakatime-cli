package api_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	router := http.NewServeMux()

	server := httptest.NewServer(router)
	defer server.Close()

	var calls int

	router.HandleFunc("/", func(_ http.ResponseWriter, req *http.Request) {
		calls++

		assert.Equal(t, "application/json", req.Header.Get("Accept"))
		assert.Equal(t, "Basic c2VjcmV0", req.Header.Get("Authorization"))
		assert.Equal(t, "workstation", req.Header.Get("X-Machine-Name"))
		assert.Contains(t, req.Header.Get("User-Agent"), "testplugin")
	})

	client, err := cmdapi.NewClient(t.Context(), params.API{
		Hostname:         " workstation ",
		Key:              "secret",
		Plugin:           "testplugin",
		Timeout:          time.Second,
		URL:              server.URL,
		DisableSSLVerify: true,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(t.Context(), req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, 1, calls)
}

func TestNewClientWithoutAuth(t *testing.T) {
	router := http.NewServeMux()

	server := httptest.NewServer(router)
	defer server.Close()

	router.HandleFunc("/", func(_ http.ResponseWriter, req *http.Request) {
		assert.Empty(t, req.Header.Get("Authorization"))
	})

	client, err := cmdapi.NewClientWithoutAuth(t.Context(), params.API{
		Timeout:          time.Second,
		URL:              server.URL,
		DisableSSLVerify: true,
	})
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(t.Context(), req)
	require.NoError(t, err)

	defer resp.Body.Close()
}

func TestNewClient_SSLCertFileError(t *testing.T) {
	client, err := cmdapi.NewClientWithoutAuth(t.Context(), params.API{
		SSLCertFilepath: "missing.pem",
		Timeout:         time.Second,
		URL:             "https://example.com",
	})

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to set up ssl cert file option on api client")
}

func TestNewClient_ProxyError(t *testing.T) {
	client, err := cmdapi.NewClientWithoutAuth(t.Context(), params.API{
		DisableSSLVerify: true,
		ProxyURL:         "%",
		Timeout:          time.Second,
		URL:              "https://example.com",
	})

	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "failed to set up proxy option on api client")
}

func TestNewClient_SSLCertFile(t *testing.T) {
	certFile := filepath.Join(t.TempDir(), "cert.pem")
	require.NoError(t, os.WriteFile(certFile, []byte("not a pem cert"), 0600))

	client, err := cmdapi.NewClientWithoutAuth(t.Context(), params.API{
		SSLCertFilepath: certFile,
		Timeout:         time.Second,
		URL:             "https://example.com",
	})

	require.NoError(t, err)
	assert.NotNil(t, client)
}
