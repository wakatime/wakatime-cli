package legacy_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/legacy"
	"github.com/wakatime/wakatime-cli/pkg/version"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRunCmd_SendDiagnostics_Error(t *testing.T) {
	// this is exclusively run in subprocess
	if os.Getenv("TEST_RUN") == "1" {
		version.OS = "some os"
		version.Arch = "some architecture"
		version.Version = "some version"

		offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
		require.NoError(t, err)

		defer os.Remove(offlineQueueFile.Name())

		logFile, err := ioutil.TempFile(os.TempDir(), "")
		require.NoError(t, err)

		defer os.Remove(logFile.Name())

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

	// run command in another runner, to effectively test os.Exit()
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmd_SendDiagnostics_Error") // nolint:gosec
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_SERVER_URL=%s", testServerURL))

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	require.True(t, ok)

	assert.Equal(t, 42, e.ExitCode())
}

func TestRunCmd_SendDiagnostics_Panic(t *testing.T) {
	// this is exclusively run in subprocess
	if os.Getenv("TEST_RUN") == "1" {
		version.OS = "some os"
		version.Arch = "some architecture"
		version.Version = "some version"

		offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
		require.NoError(t, err)

		defer os.Remove(offlineQueueFile.Name())

		logFile, err := ioutil.TempFile(os.TempDir(), "")
		require.NoError(t, err)

		defer os.Remove(logFile.Name())

		v := viper.New()
		v.Set("api-url", os.Getenv("TEST_SERVER_URL"))
		v.Set("entity", "/path/to/file")
		v.Set("key", "00000000-0000-4000-8000-000000000000")
		v.Set("log-file", logFile.Name())
		v.Set("log-to-stdout", true)
		v.Set("offline-queue-file", offlineQueueFile.Name())
		v.Set("plugin", "vim")

		legacy.RunCmd(v, func(v *viper.Viper) (int, error) {
			panic("fail")
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

		expectedBodyTpl, err := ioutil.ReadFile("testdata/diagnostics_request_template_no_logs.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
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
			jsonEscape(t, diagnostics.Stack),
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)
	})

	// run command in another runner, to effectively test os.Exit()
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmd_SendDiagnostics_Panic") // nolint:gosec
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_SERVER_URL=%s", testServerURL))

	err := cmd.Run()

	e, ok := err.(*exec.ExitError)
	require.True(t, ok)

	assert.Equal(t, 1, e.ExitCode())
}

func TestRunCmdWithOfflineSync(t *testing.T) {
	// this is exclusively run in subprocess
	if os.Getenv("TEST_RUN") == "1" {
		version.OS = "some os"
		version.Arch = "some architecture"
		version.Version = "some version"

		logFile, err := ioutil.TempFile(os.TempDir(), "")
		require.NoError(t, err)

		defer os.Remove(logFile.Name())

		v := viper.New()
		v.Set("api-url", os.Getenv("TEST_SERVER_URL"))
		v.Set("entity", "/path/to/file")
		v.Set("key", "00000000-0000-4000-8000-000000000000")
		v.Set("log-file", logFile.Name())
		v.Set("log-to-stdout", true)
		v.Set("offline-queue-file", os.Getenv("OFFLINE_QUEUE_FILE"))
		v.SetDefault("sync-offline-activity", 24)
		v.Set("plugin", "vim")

		legacy.RunCmdWithOfflineSync(v, func(v *viper.Viper) (int, error) {
			return 0, nil
		})

		return
	}

	// setup test queue
	offlineQueueFile, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(offlineQueueFile.Name())

	db, err := bolt.Open(offlineQueueFile.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
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

	db.Close()

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
		expectedBody, err := ioutil.ReadFile("testdata/api_heartbeats_request.json")
		require.NoError(t, err)

		body, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, string(expectedBody), string(body))

		// send response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	// run command in another runner, to effectively test os.Exit()
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmdWithOfflineSync") // nolint:gosec
	cmd.Env = append(os.Environ(), "TEST_RUN=1")
	cmd.Env = append(cmd.Env, fmt.Sprintf("TEST_SERVER_URL=%s", testServerURL))
	cmd.Env = append(cmd.Env, fmt.Sprintf("OFFLINE_QUEUE_FILE=%s", offlineQueueFile.Name()))

	err = cmd.Run()
	require.NoError(t, err)

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

	db.Close()

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
			return fmt.Errorf("failed put hearbeat: %s", err)
		}

		return nil
	})
	require.NoError(t, err)
}
