package api_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/summary"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/matishsiao/goInfo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Summaries(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])

		values, err := url.ParseQuery(req.URL.RawQuery)
		require.NoError(t, err)

		assert.Equal(t, url.Values(map[string][]string{
			"start": {"2020-04-01"},
			"end":   {"2020-04-02"},
		}), values)

		// write response
		f, err := os.Open("testdata/api_summaries_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(u, http.DefaultClient)
	summaries, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		GrandTotal: "10 secs",
		Date:       time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
	})
	assert.Contains(t, summaries, summary.Summary{
		GrandTotal: "20 secs",
		Date:       time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	})

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummariesByCategory(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		f, err := os.Open("testdata/api_summaries_by_category_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(u, http.DefaultClient)
	summaries, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)

	assert.Len(t, summaries, 3)
	assert.Contains(t, summaries, summary.Summary{
		Category:   "coding",
		Date:       time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		GrandTotal: "30 secs",
	})
	assert.Contains(t, summaries, summary.Summary{
		Category:   "debugging",
		Date:       time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		GrandTotal: "40 secs",
	})
	assert.Contains(t, summaries, summary.Summary{
		Category:   "coding",
		Date:       time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		GrandTotal: "50 secs",
	})

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummariesWithAuth(t *testing.T) {
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
			u, router, tearDown := setupTestServer()
			defer tearDown()

			var numCalls int

			router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
				numCalls++
				assert.Equal(t, []string{test.AuthHeaderValue}, req.Header["Authorization"])
			})

			withAuth, err := api.WithAuth(api.BasicAuth{
				User:   test.User,
				Secret: "secret",
			})
			require.NoError(t, err)

			c := api.NewClient(u, http.DefaultClient, []api.Option{withAuth}...)
			_, _ = c.Summaries(
				time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
			)

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestClient_SummariesWithTimeout(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	block := make(chan struct{})

	called := make(chan struct{})
	defer close(called)

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		<-block
		called <- struct{}{}
	})

	opts := []api.Option{api.WithTimeout(20 * time.Millisecond)}
	c := api.NewClient(u, http.DefaultClient, opts...)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)
	require.Error(t, err)

	errMsg := fmt.Sprintf("error %q does not contain string 'Timeout'", err)
	assert.True(t, strings.Contains(err.Error(), "Timeout"), errMsg)

	close(block)
	select {
	case <-called:
		break
	case <-time.After(50 * time.Millisecond):
		t.Fatal("failed")
	}
}

func TestClient_SummariesWithUserAgent(t *testing.T) {
	info := goInfo.GetInfo()

	tests := map[string]struct {
		Plugin   string
		Expected string
	}{
		"with plugin": {
			Plugin: "testplugin",
			Expected: fmt.Sprintf(
				"wakatime/%s (%s-%s-%s) %s testplugin",
				version.Version,
				runtime.GOOS,
				info.Core,
				info.Platform,
				runtime.Version(),
			),
		},
		"without plugin": {
			Expected: fmt.Sprintf(
				"wakatime/%s (%s-%s-%s) %s Unknown/0",
				version.Version,
				runtime.GOOS,
				info.Core,
				info.Platform,
				runtime.Version(),
			),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			u, router, tearDown := setupTestServer()
			defer tearDown()

			var numCalls int

			router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
				numCalls++
				assert.Equal(t, []string{test.Expected}, req.Header["User-Agent"])
			})

			var opts []api.Option
			if test.Plugin != "" {
				opts = []api.Option{api.WithUserAgent(test.Plugin)}
			} else {
				opts = []api.Option{api.WithUserAgentUnknownPlugin()}
			}

			c := api.NewClient(u, http.DefaultClient, opts...)
			_, _ = c.Summaries(
				time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
			)

			assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
		})
	}
}

func TestClient_Summaries_Err(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(u, http.DefaultClient)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	var apierr api.Err

	assert.True(t, errors.As(err, &apierr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Summaries_ErrAuth(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(u, http.DefaultClient)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	var autherr api.ErrAuth

	assert.True(t, errors.As(err, &autherr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
