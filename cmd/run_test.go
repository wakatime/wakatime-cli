package cmd_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRunCmd_Err(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(_ http.ResponseWriter, _ *http.Request) {
		numCalls++
	})

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++
		return 42, errors.New("fail")
	}

	err = cmd.RunCmd(v, false, false, cmdFn)
	require.Error(t, err)

	var errexitcode exitcode.Err

	assert.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 0 }, time.Second, 50*time.Millisecond)
}

func TestRunCmd_Verbose_Err(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(_ http.ResponseWriter, _ *http.Request) {
		numCalls++
	})

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++
		return 42, errors.New("fail")
	}

	err = cmd.RunCmd(v, true, false, cmdFn)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, 42, err.(exitcode.Err).Code)
	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 0 }, time.Second, 50*time.Millisecond)
}

func TestRunCmd_SendDiagnostics_Err(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

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

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++
		return 42, errors.New("fail")
	}

	err = cmd.RunCmd(v, true, true, cmdFn)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, 42, err.(exitcode.Err).Code)
	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestRunCmd_SendDiagnostics_Panic(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_panic_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			IsPanic       bool   `json:"is_panic"`
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

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++

		panic("fail")
	}

	err = cmd.RunCmd(v, true, false, cmdFn)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)
	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestRunCmd_SendDiagnostics_NoLogs_Panic(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		expectedBodyTpl, err := os.ReadFile("testdata/diagnostics_request_panic_no_logs_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var diagnostics struct {
			Architecture  string `json:"architecture"`
			CliVersion    string `json:"cli_version"`
			IsPanic       bool   `json:"is_panic"`
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
			jsonEscape(t, diagnostics.Stack),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++

		panic("fail")
	}

	err = cmd.RunCmd(v, false, false, cmdFn)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, exitcode.ErrGeneric, err.(exitcode.Err).Code)
	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestRunCmd_SendDiagnostics_WakaError(t *testing.T) {
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

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

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++
		return 42, offline.ErrOpenDB{Err: errors.New("fail")}
	}

	err = cmd.RunCmd(v, false, false, cmdFn)

	var errexitcode exitcode.Err

	require.ErrorAs(t, err, &errexitcode)

	assert.Equal(t, 42, err.(exitcode.Err).Code)
	assert.Equal(t, 1, cmdNumCalls)
	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestRunCmdWithOfflineSync(t *testing.T) {
	// setup test server
	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		// check body
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedBody), string(body))

		// send response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	version.OS = "some os"
	version.Arch = "some architecture"
	version.Version = "some version"

	logFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer logFile.Close()

	// setup test queue
	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	db, err := bolt.Open(offlineQueueFile.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("entity", "/path/to/file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("log-to-stdout", true)
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.SetDefault("sync-offline-activity", 24)
	v.Set("plugin", "vim")

	var cmdNumCalls int

	cmdFn := func(_ *viper.Viper) (int, error) {
		cmdNumCalls++
		return exitcode.Success, nil
	}

	err = cmd.RunCmdWithOfflineSync(v, false, false, cmdFn)
	require.NoError(t, err)

	assert.Equal(t, 1, cmdNumCalls)

	// check db
	db, err = bolt.Open(offlineQueueFile.Name(), 0600, nil)
	require.NoError(t, err)

	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("heartbeats")).Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			stored = append(stored, heartbeatRecord{
				ID:        string(key),
				Heartbeat: string(value),
			})
		}

		return nil
	})
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 0)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
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

type heartbeatRecord struct {
	ID        string
	Heartbeat string
}

func insertHeartbeatRecords(t *testing.T, db *bolt.DB, bucket string, hh []heartbeatRecord) {
	for _, h := range hh {
		insertHeartbeatRecord(t, db, bucket, h)
	}
}

func insertHeartbeatRecord(t *testing.T, db *bolt.DB, bucket string, h heartbeatRecord) {
	t.Helper()

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}

		err = b.Put([]byte(h.ID), []byte(h.Heartbeat))
		if err != nil {
			return fmt.Errorf("failed put heartbeat: %s", err)
		}

		return nil
	})
	require.NoError(t, err)
}
