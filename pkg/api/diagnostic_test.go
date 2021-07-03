package api_test

import (
	"io/ioutil"
	"net/http"
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
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Nil(t, req.Header["Authorization"])

		// check body
		expectedBody, err := ioutil.ReadFile("testdata/diagnostics_request.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedBody), string(body))

		// write response
		w.WriteHeader(http.StatusCreated)
	})

	version.OS = "linux"
	version.Arch = "amd64"
	version.Version = "<local-build>"

	diagnostics := []diagnostic.Diagnostic{
		diagnostic.Logs("some logs"),
		diagnostic.Stack("some stack"),
	}

	c := api.NewClient(url)
	err := c.SendDiagnostics("vim", diagnostics...)
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}
