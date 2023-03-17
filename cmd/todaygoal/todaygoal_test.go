package todaygoal_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/todaygoal"
	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoal(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++

			// check request
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
			assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
			assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
			assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
				"%q should have suffix %q",
				req.Header["User-Agent"][0],
				plugin,
			))

			// send response
			w.WriteHeader(http.StatusOK)

			f, err := os.Open("testdata/api_goals_id_response.json")
			require.NoError(t, err)
			defer f.Close()

			_, err = io.Copy(w, f)
			require.NoError(t, err)
		})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("plugin", plugin)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	output, err := todaygoal.Goal(v)
	require.NoError(t, err)

	assert.Equal(t, "3 hrs 23 mins", output)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestGoal_ErrApi(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusInternalServerError)
		})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	expectedMsg := fmt.Sprintf(
		`failed fetching todays goal from api: `+
			`invalid response status from "%s/users/current/goals/00000000-0000-4000-8000-000000000000". `+
			`got: 500, want: 200. body: ""`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestGoal_ErrAuth(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusUnauthorized)
		})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	expectedMsg := fmt.Sprintf(
		`failed fetching todays goal from api: `+
			`authentication failed at "%s/users/current/goals/00000000-0000-4000-8000-000000000000". body: ""`,
		testServerURL,
	)
	assert.EqualError(t, err, expectedMsg)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestGoal_ErrBadRequest(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc(
		"/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusBadRequest)
		})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errbadRequest api.ErrBadRequest

	assert.True(t, errors.As(err, &errbadRequest))

	expectedMsg := fmt.Sprintf(
		`failed fetching todays goal from api: `+
			`bad request at "%s/users/current/goals/00000000-0000-4000-8000-000000000000"`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestGoal_ErrAuth_UnsetAPIKey(t *testing.T) {
	v := viper.New()
	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.Equal(
		t,
		"failed to load command parameters: failed to load API parameters: api key not found or empty",
		err.Error(),
	)
}

func TestLoadParams_GoalID(t *testing.T) {
	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000001")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "00000000-0000-4000-8000-000000000001", params.GoalID)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
