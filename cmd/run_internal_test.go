package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	v := viper.New()

	ret := runCmd(v, false, func(v *viper.Viper) (int, error) {
		return exitcode.Success, nil
	})

	assert.Equal(t, exitcode.Success, ret)
}

func TestRunCmd_Err(t *testing.T) {
	v := viper.New()

	ret := runCmd(v, false, func(v *viper.Viper) (int, error) {
		return exitcode.ErrGeneric, errors.New("fail")
	})

	assert.Equal(t, exitcode.ErrGeneric, ret)
}

func TestRunCmd_ErrOfflineEnqueue(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Nil(t, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Platform     string `json:"platform"`
			Architecture string `json:"architecture"`
			CliVersion   string `json:"cli_version"`
			Editor       string `json:"editor"`
			Logs         string `json:"logs"`
			Stack        string `json:"stacktrace"`
		}

		err = json.Unmarshal(body, &diagnostics)
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBodyTpl),
			jsonEscape(t, diagnostics.Logs),
			jsonEscape(t, diagnostics.Stack),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "vim")

	ret := runCmd(v, true, func(v *viper.Viper) (int, error) {
		return exitcode.ErrGeneric, errors.New("fail")
	})

	assert.Equal(t, exitcode.ErrGeneric, ret)
}

func TestParseConfigFiles(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("internal-config", "testdata/.wakatime-internal.cfg")

	err := parseConfigFiles(v)
	require.NoError(t, err)

	assert.Equal(t, "true", v.GetString("settings.debug"))
	assert.Equal(t, "testdata/.import.cfg", v.GetString("settings.import_cfg"))
	assert.Equal(t,
		"00000000-0000-4000-8000-000000000000",
		v.GetString("settings.api_key"))
	assert.Equal(t, "1", v.GetString("internal.backoff_retries"))
	assert.Equal(t,
		"2006-01-02T15:04:05Z07:00",
		v.GetString("internal.backoff_at"))
}

func jsonEscape(t *testing.T, i string) string {
	b, err := json.Marshal(i)
	require.NoError(t, err)

	s := string(b)

	return s[1 : len(s)-1]
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}
