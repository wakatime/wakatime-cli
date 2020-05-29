package api_test

import (
	"fmt"
	"net/http"
	"runtime"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/version"

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

			c := api.NewClient("", http.DefaultClient, []api.Option{withAuth}...)
			resp, err := c.Do(req)
			require.NoError(t, err)

			defer resp.Body.Close()

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestOption_WithHostName(t *testing.T) {
	url, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, []string{"my-computer"}, req.Header["X-Machine-Name"])

		numCalls++
	})

	opts := []api.Option{api.WithHostName("my-computer")}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	c := api.NewClient("", http.DefaultClient, opts...)
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

	c := api.NewClient("", http.DefaultClient, opts...)
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

	c := api.NewClient("", http.DefaultClient, opts...)
	resp, err := c.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}
