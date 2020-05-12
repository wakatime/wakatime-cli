package api_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Summaries(t *testing.T) {
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
			url, router, close := setupTestServer()
			defer close()

			var numCalls int

			router.HandleFunc("/api/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
				numCalls++

				// check method
				assert.Equal(t, http.MethodGet, req.Method)

				// check headers
				assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
				assert.Equal(t, []string{test.AuthHeaderValue}, req.Header["Authorization"])
				assert.Equal(t, []string{"wakatime/13.0.8"}, req.Header["User-Agent"])

				// write response
				f, err := os.Open("testdata/api_summaries_response.json")
				require.NoError(t, err)

				w.WriteHeader(http.StatusOK)
				_, err = io.Copy(w, f)
				require.NoError(t, err)
			})

			c := api.NewClient(url, http.DefaultClient)
			summaries, err := c.Summaries(
				time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
				api.SummaryConfig{
					Auth: api.BasicAuth{
						User:   test.User,
						Secret: "secret",
					},
					UserAgent: "wakatime/13.0.8",
				},
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

			assert.Equal(t, 1, numCalls)
		})
	}
}

func TestClient_SummariesByCategory(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/api/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		f, err := os.Open("testdata/api_summaries_by_category_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(url, http.DefaultClient)
	summaries, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		api.SummaryConfig{
			Auth: api.BasicAuth{
				Secret: "secret",
			},
			UserAgent: "wakatime/13.0.8",
		},
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

	assert.Equal(t, 1, numCalls)
}

func TestClient_Summaries_MissingAuthSecret(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/api/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
	})

	c := api.NewClient(url, http.DefaultClient)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		api.SummaryConfig{
			UserAgent: "wakatime/13.0.8",
		},
	)
	assert.Error(t, err)

	assert.Equal(t, 0, numCalls)
}

func TestClient_Summaries_Err(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/api/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(url, http.DefaultClient)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		api.SummaryConfig{
			Auth: api.BasicAuth{
				Secret: "secret",
			},
			UserAgent: "wakatime/13.0.8",
		},
	)
	assert.True(t, errors.Is(err, api.Err))
	assert.Equal(t, 1, numCalls)
}

func TestClient_Summaries_ErrAuth(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/api/v1/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(url, http.DefaultClient)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		api.SummaryConfig{
			Auth: api.BasicAuth{
				Secret: "secret",
			},
			UserAgent: "wakatime/13.0.8",
		},
	)
	assert.True(t, errors.Is(err, api.ErrAuth))
	assert.Equal(t, 1, numCalls)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
