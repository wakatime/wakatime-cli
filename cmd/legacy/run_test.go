package legacy_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd_SendDiagnostics(t *testing.T) {
	// this is exclusively run in subprocess
	if os.Getenv("TEST_RUN") == "1" {
		version.OS = "some os"
		version.Arch = "some architecture"

		offlineQueueFile, err := ioutil.TempFile(t.TempDir(), "")
		require.NoError(t, err)

		logFile, err := ioutil.TempFile(t.TempDir(), "")
		require.NoError(t, err)

		v := viper.New()
		v.Set("api-url", os.Getenv("TEST_SERVER_URL"))
		v.Set("entity", "/path/to/file")
		v.Set("key", "00000000-0000-4000-8000-000000000000")
		v.Set("log-file", logFile.Name())
		v.Set("log-to-stdout", true)
		v.Set("offline-queue-file", offlineQueueFile.Name())
		v.Set("plugin", "vim")

		legacy.RunCmd(v, func(v *viper.Viper) (int, error) {
			return 42, errors.New("fail")
		})

		return
	}

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Nil(t, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := ioutil.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Platform     string `json:"platform"`
			Architecture string `json:"architecture"`
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

	// run command in another runner, to effectively test os.Exit()
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmd_SendDiagnostics") // nolint:gosec
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_SERVER_URL=%s", testServerURL))

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	require.True(t, ok)

	assert.Equal(t, 42, e.ExitCode())
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
