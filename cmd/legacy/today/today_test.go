package today_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy/today"
	"github.com/wakatime/wakatime-cli/pkg/api"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummary(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		dateToday = time.Now().Format("2006-01-02")
		plugin    = "plugin/0.0.1"
		numCalls  int
	)

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		values, err := url.ParseQuery(req.URL.RawQuery)
		require.NoError(t, err)

		assert.Equal(t, url.Values(map[string][]string{
			"start": {dateToday},
			"end":   {dateToday},
		}), values)

		// write response
		data, err := ioutil.ReadFile("testdata/api_summaries_response_template.json")
		require.NoError(t, err)

		_, err = w.Write([]byte(fmt.Sprintf(string(data), dateToday)))
		require.NoError(t, err)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("plugin", plugin)

	output, err := today.Summary(v)
	require.NoError(t, err)

	assert.Equal(t, "10 secs", output)
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSummary_ErrApi(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusInternalServerError)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	_, err := today.Summary(v)
	require.Error(t, err)

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	expectedMsg := fmt.Sprintf(
		`failed fetching summaries from api: `+
			`invalid response status from "%s/users/current/summaries". got: 500, want: 200. body: ""`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSummary_ErrAuth(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/summaries", func(w http.ResponseWriter, req *http.Request) {
		numCalls++
		w.WriteHeader(http.StatusUnauthorized)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)

	_, err := today.Summary(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.True(t, errors.As(err, &errauth))

	expectedMsg := fmt.Sprintf(
		`failed fetching summaries from api: `+
			`authentication failed at "%s/users/current/summaries". body: ""`,
		testServerURL,
	)
	assert.Equal(t, expectedMsg, err.Error())
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSummary_ErrAuth_UnsetAPIKey(t *testing.T) {
	v := viper.New()
	_, err := today.Summary(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.True(t, errors.As(err, &errauth))
	assert.Equal(t, "failed to load command parameters: failed to load api params: failed to load api key", err.Error())
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
