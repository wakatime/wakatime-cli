package api_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Summaries(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
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

	c := api.NewClient(u)
	summaries, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "10 secs",
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "20 secs",
	})

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummariesByCategory(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		f, err := os.Open("testdata/api_summaries_by_category_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(u)
	summaries, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "Coding",
				Total:    "30 secs",
			},
			{
				Category: "Debugging",
				Total:    "20 secs",
			},
		},
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "Coding",
				Total:    "50 secs",
			},
		},
	})

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummariesWithTimeout(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	block := make(chan struct{})

	called := make(chan struct{})
	defer close(called)

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		<-block
		called <- struct{}{}
	})

	opts := []api.Option{api.WithTimeout(20 * time.Millisecond)}
	c := api.NewClient(u, opts...)
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

func TestClient_Summaries_Err(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(u)
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

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(u)
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	var autherr api.ErrAuth

	assert.True(t, errors.As(err, &autherr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Summaries_ErrRequest(t *testing.T) {
	c := api.NewClient("invalid-url")
	_, err := c.Summaries(
		time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
	)

	var reqerr api.ErrRequest

	assert.True(t, errors.As(err, &reqerr))
}

func TestParseSummariesResponse_DayTotal(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_summaries_response.json")
	require.NoError(t, err)

	summaries, err := api.ParseSummariesResponse(data)
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "10 secs",
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "20 secs",
	})
}

func TestParseSummariesResponse_TotalsByCategory(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_summaries_by_category_response.json")
	require.NoError(t, err)

	summaries, err := api.ParseSummariesResponse(data)
	require.NoError(t, err)

	assert.Len(t, summaries, 2)
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 1, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "Coding",
				Total:    "30 secs",
			},
			{
				Category: "Debugging",
				Total:    "20 secs",
			},
		},
	})
	assert.Contains(t, summaries, summary.Summary{
		Date:  time.Date(2020, time.April, 2, 0, 0, 0, 0, time.UTC),
		Total: "50 secs",
		ByCategory: []summary.Category{
			{
				Category: "Coding",
				Total:    "50 secs",
			},
		},
	})
}
