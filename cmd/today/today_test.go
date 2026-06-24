package today_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/today"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)

		f, err := os.Open("testdata/api_statusbar_today_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	var (
		code int
		err  error
	)

	output := captureStdout(t, func() {
		code, err = today.Run(t.Context(), v)
	})

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)
	assert.Equal(t, "10 secs\n", output)
}

func TestRunErr(t *testing.T) {
	v := viper.New()

	code, err := today.Run(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrGeneric, code)
	assert.Contains(t, err.Error(), "today fetch failed")
}

func TestToday(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
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

		f, err := os.Open("testdata/api_statusbar_today_response.json")
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

	output, err := today.Today(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, "10 secs", output)
	assert.Equal(t, 1, numCalls)
}

func TestToday_ErrApi(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusInternalServerError)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	_, err := today.Today(t.Context(), v)
	require.Error(t, err)

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	expectedMsg := fmt.Sprintf(
		`failed fetching today from api: `+
			`invalid response status from "%s/users/current/statusbar/today". got: 500, want: 200. body: ""`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Equal(t, 1, numCalls)
}

func TestToday_ErrAuth(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusUnauthorized)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	_, err := today.Today(t.Context(), v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	expectedMsg := fmt.Sprintf(
		`failed fetching today from api: `+
			`authentication failed at "%s/users/current/statusbar/today". body: ""`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())

	assert.Equal(t, 1, numCalls)
}

func TestToday_ErrBadRequest(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusBadRequest)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	_, err := today.Today(t.Context(), v)
	require.Error(t, err)

	var errbadRequest api.ErrBadRequest

	assert.True(t, errors.As(err, &errbadRequest))

	expectedMsg := fmt.Sprintf(
		`failed fetching today from api: `+
			`bad request at "%s/users/current/statusbar/today"`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Equal(t, 1, numCalls)
}

func TestToday_ErrAuth_UnsetAPIKey(t *testing.T) {
	v := viper.New()
	_, err := today.Today(t.Context(), v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.Equal(t, "failed to load API parameters: api key not found or empty", err.Error())
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w

	defer func() { os.Stdout = stdout }()

	outC := make(chan string, 1)
	errC := make(chan error, 1)

	go func() {
		var buf bytes.Buffer

		_, err := io.Copy(&buf, r)
		errC <- err

		outC <- buf.String()
	}()

	fn()

	require.NoError(t, w.Close())

	output := <-outC

	require.NoError(t, <-errC)
	require.NoError(t, r.Close())

	return output
}
