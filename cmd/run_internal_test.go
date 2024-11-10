package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	v := viper.New()

	err := runCmd(context.Background(), v, false, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		return exitcode.Success, nil
	})

	assert.Nil(t, err)
}

func TestRunCmd_Err(t *testing.T) {
	v := viper.New()

	err := runCmd(context.Background(), v, false, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		return exitcode.ErrGeneric, errors.New("fail")
	})

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)
}

func TestRunCmd_Panic(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			Editor        string `json:"editor"`
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

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("log-file", logFile.Name())

	logger, err := SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	err = runCmd(ctx, v, false, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		panic("fail")
	})

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "panicked")
}

func TestRunCmd_Panic_Verbose(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			Editor        string `json:"editor"`
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

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("log-file", logFile.Name())

	logger, err := SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	err = runCmd(ctx, v, true, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		panic("fail")
	})

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "panicked")
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
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			Editor        string `json:"editor"`
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

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "vim")

	err := runCmd(context.Background(), v, true, false, func(_ context.Context, _ *viper.Viper) (int, error) {
		return exitcode.ErrGeneric, errors.New("fail")
	})

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)
}

func TestRunCmd_BackoffLoggedWithVerbose(t *testing.T) {
	ctx := context.Background()
	verbose := true

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusCreated)
	})

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	// close to avoid "The process cannot access the file because it is being used by another process" error on Windows
	offlineQueueFile.Close()

	entity, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer entity.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", entity.Name())
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("internal.backoff_at", time.Now().Add(10*time.Minute).Format(ini.DateFormat))
	v.Set("internal.backoff_retries", "1")
	v.Set("verbose", verbose)

	logger, err := SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	err = runCmd(ctx, v, verbose, false, cmdheartbeat.Run)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrBackoff, err.(exitcode.Err).Code)
	assert.Equal(t, 0, numCalls)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "failed to run command: sending heartbeat")
}

func TestRunCmd_BackoffNotLogged(t *testing.T) {
	ctx := context.Background()
	verbose := false

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusCreated)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	// close to avoid "The process cannot access the file because it is being used by another process" error on Windows
	offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	entity, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer entity.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", entity.Name())
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("internal.backoff_at", time.Now().Add(10*time.Minute).Format(ini.DateFormat))
	v.Set("internal.backoff_retries", "1")
	v.Set("verbose", verbose)

	logger, err := SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	err = runCmd(ctx, v, verbose, false, cmdheartbeat.Run)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)
	assert.Equal(t, exitcode.ErrBackoff, err.(exitcode.Err).Code)
	assert.Equal(t, 0, numCalls)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Empty(t, string(output))
}

func TestParseConfigFiles(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/.wakatime.cfg")
	v.Set("internal-config", "testdata/.wakatime-internal.cfg")

	err := parseConfigFiles(context.Background(), v)
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
