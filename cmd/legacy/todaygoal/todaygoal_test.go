package todaygoal_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy/todaygoal"
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
		"/v1/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
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

			// write response
			data, err := ioutil.ReadFile("testdata/api_goals_id_response.json")
			require.NoError(t, err)

			_, err = w.Write([]byte(string(data)))
			require.NoError(t, err)
		})

	v := viper.New()
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
		"/v1/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusInternalServerError)
		})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errapi api.Err

	assert.True(t, errors.As(err, &errapi))

	expectedMsg := fmt.Sprintf(
		`failed fetching todays goal from api: `+
			`invalid response status from "%s/v1/users/current/goals/00000000-0000-4000-8000-000000000000". `+
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
		"/v1/users/current/goals/00000000-0000-4000-8000-000000000000", func(w http.ResponseWriter, req *http.Request) {
			numCalls++
			w.WriteHeader(http.StatusUnauthorized)
		})

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("api-url", testServerURL)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	_, err := todaygoal.Goal(v)
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.True(t, errors.As(err, &errauth))

	expectedMsg := fmt.Sprintf(
		`failed fetching todays goal from api: `+
			`authentication failed at "%s/v1/users/current/goals/00000000-0000-4000-8000-000000000000". body: ""`,
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

	assert.True(t, errors.As(err, &errauth))
	assert.Equal(t, "failed to load command parameters: failed to load api key", err.Error())
}

func TestLoadParams_APIKey(t *testing.T) {
	tests := map[string]struct {
		ViperAPIKey          string
		ViperAPIKeyConfig    string
		ViperAPIKeyConfigOld string
		Expected             todaygoal.Params
	}{
		"api key flag takes preceedence": {
			ViperAPIKey:          "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfig:    "10000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "20000000-0000-4000-8000-000000000000",
			Expected: todaygoal.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api from config takes preceedence": {
			ViperAPIKeyConfig:    "00000000-0000-4000-8000-000000000000",
			ViperAPIKeyConfigOld: "10000000-0000-4000-8000-000000000000",
			Expected: todaygoal.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api key from config deprecated": {
			ViperAPIKeyConfigOld: "00000000-0000-4000-8000-000000000000",
			Expected: todaygoal.Params{
				APIKey: "00000000-0000-4000-8000-000000000000",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", test.ViperAPIKey)
			v.Set("settings.api_key", test.ViperAPIKeyConfig)
			v.Set("settings.apikey", test.ViperAPIKeyConfigOld)
			v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

			params, err := todaygoal.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoadParams_APIUrl(t *testing.T) {
	tests := map[string]struct {
		ViperAPIUrl       string
		ViperAPIUrlConfig string
		ViperAPIUrlOld    string
		Expected          todaygoal.Params
	}{
		"api url flag takes preceedence": {
			ViperAPIUrl:       "http://localhost:8080",
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: todaygoal.Params{
				APIUrl: "http://localhost:8080",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api url deprecated flag takes preceedence": {
			ViperAPIUrlConfig: "http://localhost:8081",
			ViperAPIUrlOld:    "http://localhost:8082",
			Expected: todaygoal.Params{
				APIUrl: "http://localhost:8082",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
		"api url from config": {
			ViperAPIUrlConfig: "http://localhost:8081",
			Expected: todaygoal.Params{
				APIUrl: "http://localhost:8081",
				GoalID: "00000000-0000-4000-8000-000000000000",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("api-url", test.ViperAPIUrl)
			v.Set("apiurl", test.ViperAPIUrlOld)
			v.Set("settings.api_url", test.ViperAPIUrlConfig)
			v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

			params, err := todaygoal.LoadParams(v)
			require.NoError(t, err)

			test.Expected.APIKey = "00000000-0000-4000-8000-000000000000"
			assert.Equal(t, test.Expected, params)
		})
	}
}

func TestLoadParams_Plugin(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/10.0.0")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, todaygoal.Params{
		APIKey: "00000000-0000-4000-8000-000000000000",
		Plugin: "plugin/10.0.0",
		GoalID: "00000000-0000-4000-8000-000000000000",
	}, params)
}

func TestLoadParams_Timeout_FlagTakesPreceedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("timeout", 5)
	v.Set("settings.timeout", 10)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 5*time.Second, params.Timeout)
}

func TestLoadParams_Timeout_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("settings.timeout", 10)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, 10*time.Second, params.Timeout)
}

func TestLoadParamsErr_InvalidAPIKey(t *testing.T) {
	tests := map[string]string{
		"unset":            "",
		"invalid format 1": "not-uuid",
		"invalid format 2": "00000000-0000-0000-8000-000000000000",
		"invalid format 3": "00000000-0000-4000-0000-000000000000",
	}

	for name, value := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", value)

			_, err := todaygoal.LoadParams(v)
			require.Error(t, err)

			var errauth api.ErrAuth
			require.True(t, errors.As(err, &errauth))
		})
	}
}

func TestLoadParams_Network_DisableSSLVerify_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("no-ssl-verify", true)
	v.Set("settings.no_ssl_verify", false)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_DisableSSLVerify_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.no_ssl_verify", true)
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.True(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_DisableSSLVerify_Default(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.False(t, params.Network.DisableSSLVerify)
}

func TestLoadParams_Network_ProxyURL(t *testing.T) {
	tests := map[string]string{
		"https":  "https://john:secret@example.org:8888",
		"http":   "http://john:secret@example.org:8888",
		"socks5": "socks5://john:secret@example.org:8888",
		"ntlm":   `domain\\john:123456`,
	}

	for name, proxyURL := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("key", "00000000-0000-4000-8000-000000000000")
			v.Set("entity", "/path/to/file")
			v.Set("proxy", proxyURL)
			v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

			params, err := todaygoal.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, proxyURL, params.Network.ProxyURL)
		})
	}
}

func TestLoadParams_Network_ProxyURL_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", "https://john:secret@example.org:8888")
	v.Set("settings.proxy", "ignored")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.Network.ProxyURL)
}

func TestLoadParams_Network_ProxyURL_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.proxy", "https://john:secret@example.org:8888")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "https://john:secret@example.org:8888", params.Network.ProxyURL)
}

func TestLoadParams_Network_ProxyURL_InvalidFormat(t *testing.T) {
	proxyURL := "ftp://john:secret@example.org:8888"

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("proxy", proxyURL)

	_, err := todaygoal.LoadParams(v)
	require.Error(t, err)
}

func TestLoadParams_Network_SSLCertFilepath_FlagTakesPrecedence(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("ssl-certs-file", "/path/to/cert.pem")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.Network.SSLCertFilepath)
}

func TestLoadParams_Network_SSLCertFilepath_FromConfig(t *testing.T) {
	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("entity", "/path/to/file")
	v.Set("settings.ssl_certs_file", "/path/to/cert.pem")
	v.Set("today-goal", "00000000-0000-4000-8000-000000000000")

	params, err := todaygoal.LoadParams(v)
	require.NoError(t, err)

	assert.Equal(t, "/path/to/cert.pem", params.Network.SSLCertFilepath)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
