package api_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/Azure/go-ntlmssp"
	"github.com/matishsiao/goInfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOption_WithAuth(t *testing.T) {
	tests := map[string]struct {
		User            string
		AuthHeaderValue string
	}{
		"auth with user": {
			User:            "john",
			AuthHeaderValue: "Basic am9objpzZWNyZXQ=",
		},
		"auth without user": {
			AuthHeaderValue: "Basic c2VjcmV0",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			url, router, tearDown := setupTestServer()
			defer tearDown()

			var numCalls int

			router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				assert.Equal(t, []string{test.AuthHeaderValue}, req.Header["Authorization"])
				numCalls++
			})

			withAuth, err := api.WithAuth(api.BasicAuth{
				User:   test.User,
				Secret: "secret",
			})
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			c := api.NewClient("", []api.Option{withAuth}...)
			resp, err := c.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestOption_WithHostname(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, []string{"my-computer"}, req.Header["X-Machine-Name"])

		numCalls++
	})

	opts := []api.Option{api.WithHostname("my-computer")}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestOption_WithNTLM(t *testing.T) {
	tests := map[string]string{
		"default":  `domain\\john:123456`,
		"useronly": `domain\\john`,
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			url, router, close := setupTestServer()
			defer close()

			var numCalls int

			router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
				authHeader, ok := req.Header["Authorization"]
				if !ok {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				if strings.HasPrefix(authHeader[0], "Basic ") {
					w.Header().Set("WWW-Authenticate", "NTLM xyxyxyx")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				msg, err := ntlmssp.NewNegotiateMessage("domain", "")
				require.NoError(t, err)

				numCalls++
				assert.Equal(t, []string{"NTLM " + base64.StdEncoding.EncodeToString(msg)}, authHeader)
			})

			withNTLM, err := api.WithNTLM(proxyURL)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			c := api.NewClient("", []api.Option{withNTLM}...)
			resp, err := c.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestOption_WithNTLMRequestRetry(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		// provoking a request error on first request to trigger ntlm retry
		if numCalls == 0 {
			numCalls++

			hj, ok := w.(http.Hijacker)
			require.True(t, ok)

			conn, _, err := hj.Hijack()
			require.NoError(t, err)
			conn.Close()

			return
		}

		authHeader, ok := req.Header["Authorization"]
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if strings.HasPrefix(authHeader[0], "Basic ") {
			w.Header().Set("WWW-Authenticate", "NTLM xyxyxyx")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		msg, err := ntlmssp.NewNegotiateMessage("domain", "")
		require.NoError(t, err)

		numCalls++
		assert.Equal(t, []string{"NTLM " + base64.StdEncoding.EncodeToString(msg)}, authHeader)
	})

	withNTLMRetry, err := api.WithNTLMRequestRetry(`domain\\john:secret`)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", []api.Option{withNTLMRetry}...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 2 }, time.Second, 50*time.Millisecond)
}

func TestOption_WithProxy(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
	})

	withProxy, err := api.WithProxy(url)
	require.NoError(t, err)

	opts := []api.Option{withProxy}

	req, err := http.NewRequest(http.MethodGet, "http://example.org", nil)
	require.NoError(t, err)

	c := api.NewClient("", opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestOption_WithUserAgent(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		info := goInfo.GetInfo()
		expected := fmt.Sprintf(
			"wakatime/%s (%s-%s-%s) %s testplugin",
			version.Version,
			runtime.GOOS,
			info.Core,
			info.Platform,
			runtime.Version(),
		)
		assert.Equal(t, []string{expected}, req.Header["User-Agent"])

		numCalls++
	})

	opts := []api.Option{api.WithUserAgent("testplugin")}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestOption_WithUserAgentUnknownPlugin(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		info := goInfo.GetInfo()
		expected := fmt.Sprintf(
			"wakatime/%s (%s-%s-%s) %s Unknown/0",
			version.Version,
			runtime.GOOS,
			info.Core,
			info.Platform,
			runtime.Version(),
		)
		assert.Equal(t, []string{expected}, req.Header["User-Agent"])

		numCalls++
	})

	opts := []api.Option{api.WithUserAgentUnknownPlugin()}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}
