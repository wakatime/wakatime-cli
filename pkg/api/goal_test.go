package api_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Goal(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++

			// check request
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, []string{"application/json"}, req.Header["Accept"])

			// write response
			f, err := os.Open("testdata/api_goals_id_response.json")
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			_, err = io.Copy(w, f)
			require.NoError(t, err)
		})

	c := api.NewClient(u)
	goal, err := c.Goal("00000000-0000-4000-8000-000000000000")

	require.NoError(t, err)

	assert.Equal(t, "3 hrs 23 mins", goal.Total)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_GoalWithTimeout(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	block := make(chan struct{})

	called := make(chan struct{})
	defer close(called)

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			<-block
			called <- struct{}{}
		})

	opts := []api.Option{api.WithTimeout(20 * time.Millisecond)}
	c := api.NewClient(u, opts...)
	_, err := c.Goal("00000000-0000-4000-8000-000000000000")
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

func TestClient_Goal_Err(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusInternalServerError)
		})

	c := api.NewClient(u)
	_, err := c.Goal("00000000-0000-4000-8000-000000000000")

	var apierr api.Err

	assert.True(t, errors.As(err, &apierr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Goal_ErrAuth(t *testing.T) {
	u, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusUnauthorized)
		})

	c := api.NewClient(u)
	_, err := c.Goal("00000000-0000-4000-8000-000000000000")

	var autherr api.ErrAuth

	assert.True(t, errors.As(err, &autherr))
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestClient_Goal_ErrRequest(t *testing.T) {
	c := api.NewClient("invalid-url")
	_, err := c.Goal("00000000-0000-4000-8000-000000000000")

	var reqerr api.ErrRequest

	assert.True(t, errors.As(err, &reqerr))
}
