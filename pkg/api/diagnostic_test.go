package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/diagnostic"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_SendDiagnostics(t *testing.T) {
	url, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check method and headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Nil(t, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		// check body
		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			Logs          string `json:"logs"`
			OriginalError string `json:"error_message"`
			Platform      string `json:"platform"`
			Plugin        string `json:"plugin"`
			Stack         string `json:"stacktrace"`
		}

		err = json.Unmarshal(body, &diagnostics)
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBodyTpl),
			jsonEscape(t, diagnostics.OriginalError),
			jsonEscape(t, diagnostics.Logs),
			jsonEscape(t, diagnostics.Stack),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// write response
		w.WriteHeader(http.StatusCreated)
	})

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	diagnostics := []diagnostic.Diagnostic{
		diagnostic.Error("some error"),
		diagnostic.Logs("some logs"),
		diagnostic.Stack("some stack"),
	}

	c := api.NewClient(url)
	err := c.SendDiagnostics("vim", false, diagnostics...)
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}
