package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

type panicParser struct{}

func (panicParser) Parse(context.Context) (Heartbeats, error) {
	panic("OOM")
}

func (panicParser) Name() string {
	return "panic"
}

type heartbeatsParser struct{}

func (heartbeatsParser) Parse(context.Context) (Heartbeats, error) {
	return make(Heartbeats, 2), nil
}

func (heartbeatsParser) Name() string {
	return "heartbeats"
}

type errorParser struct{}

func (errorParser) Parse(context.Context) (Heartbeats, error) {
	return nil, errors.New("failed")
}

func (errorParser) Name() string {
	return "error"
}

func TestParseHeartbeats_Panic(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), panicParser{}, nil)

	require.Error(t, err)
	assert.Nil(t, heartbeats)
	assert.Contains(t, err.Error(), "panicked: OOM")
}

func TestParseHeartbeats_PanicReportsDiagnostic(t *testing.T) {
	var report aiPanicReport

	heartbeats, err := parseHeartbeats(t.Context(), panicParser{}, func(r aiPanicReport) {
		report = r
	})

	require.Error(t, err)
	assert.Nil(t, heartbeats)
	assert.Equal(t, "panic", report.ParserName)
	assert.Equal(t, "OOM", report.Recovered)
	assert.Contains(t, report.Stack, "panicParser.Parse")
}

func TestParseHeartbeats_Success(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), heartbeatsParser{}, nil)

	require.NoError(t, err)
	assert.Len(t, heartbeats, 2)
}

func TestParseHeartbeats_Error(t *testing.T) {
	heartbeats, err := parseHeartbeats(t.Context(), errorParser{}, nil)

	require.EqualError(t, err, "failed")
	assert.Nil(t, heartbeats)
}

func TestCaptureAIParsingLogsVerbose(t *testing.T) {
	var output bytes.Buffer

	logger := log.New(&output, log.WithVerbose(true))
	ctx := log.ToContext(t.Context(), logger)

	captured, reset := captureAIParsingLogs(ctx)

	logger.Debugln("captured debug")
	reset()
	logger.Debugln("after reset")

	assert.Contains(t, output.String(), "captured debug")
	assert.Contains(t, output.String(), "after reset")
	assert.Contains(t, captured.String(), "captured debug")
	assert.NotContains(t, captured.String(), "after reset")
}

func TestSendAIPanicDiagnostics(t *testing.T) {
	t.Setenv("HTTP_PROXY", "")
	t.Setenv("HTTPS_PROXY", "")
	t.Setenv("NO_PROXY", "*")

	var called bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		called = true

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "/plugins/errors", req.URL.Path)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
		assert.NotEmpty(t, req.Header.Get("Authorization"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var payload struct {
			ErrorMessage string `json:"error_message"`
			IsPanic      bool   `json:"is_panic"`
			Logs         string `json:"logs"`
			Plugin       string `json:"plugin"`
			Stack        string `json:"stacktrace"`
		}

		require.NoError(t, json.Unmarshal(body, &payload))
		assert.Equal(t, `ai parser "panic" panicked: OOM`, payload.ErrorMessage)
		assert.True(t, payload.IsPanic)
		assert.Equal(t, "parser logs", payload.Logs)
		assert.Equal(t, "plugin/0.0.1", payload.Plugin)
		assert.Equal(t, "stack", payload.Stack)

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	v := newDiagnosticViper(server.URL)

	err := sendAIPanicDiagnostics(t.Context(), v, aiPanicReport{
		Logs:       "parser logs",
		ParserName: "panic",
		Recovered:  "OOM",
		Stack:      "stack",
	})

	require.NoError(t, err)
	assert.True(t, called)
}

func TestSendAIPanicDiagnosticsErrorBranches(t *testing.T) {
	err := sendAIPanicDiagnostics(t.Context(), nil, aiPanicReport{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing viper instance")

	err = sendAIPanicDiagnostics(t.Context(), viper.New(), aiPanicReport{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load API parameters")

	v := newDiagnosticViper("http://127.0.0.1")
	v.Set("ssl-certs-file", filepath.Join(t.TempDir(), "missing.pem"))

	err = sendAIPanicDiagnostics(t.Context(), v, aiPanicReport{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to set up ssl cert file option")
}

func newDiagnosticViper(apiURL string) *viper.Viper {
	v := viper.New()
	v.Set("api-url", apiURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/0.0.1")
	v.Set("timeout", 5)

	return v
}
