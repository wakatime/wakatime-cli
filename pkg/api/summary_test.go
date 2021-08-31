package api_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/summary"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Summary(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])

		// write response
		f, err := os.Open("testdata/api_statusbar_today_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(u)
	s, err := c.Summary()

	require.NoError(t, err)

	assert.Equal(t, s, &summary.Summary{
		Total: "20 secs",
	})

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummaryByCategory(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		f, err := os.Open("testdata/api_statusbar_today_by_category_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	c := api.NewClient(u)
	s, err := c.Summary()
	require.NoError(t, err)

	assert.Equal(t, s, &summary.Summary{
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

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_SummaryWithTimeout(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	block := make(chan struct{})

	called := make(chan struct{})
	defer close(called)

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		<-block
		called <- struct{}{}
	})

	opts := []api.Option{api.WithTimeout(20 * time.Millisecond)}
	c := api.NewClient(u, opts...)
	_, err := c.Summary()
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

func TestClient_Summary_Err(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	c := api.NewClient(u)
	_, err := c.Summary()

	var apierr api.Err

	assert.True(t, errors.As(err, &apierr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Summary_ErrAuth(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	c := api.NewClient(u)
	_, err := c.Summary()

	var autherr api.ErrAuth

	assert.True(t, errors.As(err, &autherr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Summary_ErrRequest(t *testing.T) {
	c := api.NewClient("invalid-url")
	_, err := c.Summary()

	var reqerr api.ErrRequest

	assert.True(t, errors.As(err, &reqerr))
}

func TestParseSummaryResponse_DayTotal(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_statusbar_today_response.json")
	require.NoError(t, err)

	s, err := api.ParseSummaryResponse(data)
	require.NoError(t, err)

	assert.Equal(t, s, &summary.Summary{
		Total: "20 secs",
	})
}

func TestParseSummaryResponse_TotalsByCategory(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/api_statusbar_today_by_category_response.json")
	require.NoError(t, err)

	s, err := api.ParseSummaryResponse(data)
	require.NoError(t, err)

	assert.Equal(t, s, &summary.Summary{
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
}
