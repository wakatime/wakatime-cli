//go:build integration

package main_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/version"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

// nolint:gochecknoinits
func init() {
	version.Version = "<local-build>"
}

func TestSendHeartbeats(t *testing.T) {
	projectFolder, err := filepath.Abs(".")
	require.NoError(t, err)

	testSendHeartbeats(t, projectFolder, "testdata/main.go", "wakatime-cli")
}

func TestSendHeartbeats_EntityFileInTempDir(t *testing.T) {
	tmpDir, err := filepath.Abs(t.TempDir())
	require.NoError(t, err)

	tmpDir, err = realpath.Realpath(tmpDir)
	require.NoError(t, err)

	runCmd(exec.Command("cp", "./testdata/main.go", tmpDir), &bytes.Buffer{})

	testSendHeartbeats(t, tmpDir, filepath.Join(tmpDir, "main.go"), "")
}

func testSendHeartbeats(t *testing.T, projectFolder, entity, p string) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		// check body
		expectedBodyTpl, err := os.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		entityPath, err := realpath.Realpath(entity)
		require.NoError(t, err)

		entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
		expectedBody := fmt.Sprintf(
			string(expectedBodyTpl),
			entityPath,
			p,
			subfolders,
			heartbeat.UserAgent(""),
		)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, expectedBody, string(body))

		// write response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", entity,
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--line-additions", "123",
		"--line-deletions", "456",
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--project", p,
		"--project-folder", projectFolder,
		"--write",
		"--verbose",
	)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_SecondaryApiKey(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	rootPath, _ := filepath.Abs(".")
	subfolders := project.CountSlashesInProjectFolder(rootPath)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAx"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		// check body
		expectedBodyTpl, err := os.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		entityPath, err := realpath.Realpath("testdata/main.go")
		require.NoError(t, err)

		entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
		expectedBody := fmt.Sprintf(
			string(expectedBodyTpl),
			entityPath,
			"wakatime-cli",
			subfolders,
			heartbeat.UserAgent(""),
		)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, expectedBody, string(body))

		// write response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", "testdata/wakatime.cfg",
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--line-additions", "123",
		"--line-deletions", "456",
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--project", "wakatime-cli",
		"--write",
		"--verbose",
	)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ExtraHeartbeats(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		var filename string

		switch numCalls {
		case 1:
			filename = "testdata/api_heartbeats_response_extra_heartbeats.json"
		case 2:
			filename = "testdata/api_heartbeats_response_extra_heartbeats_extra.json"
		}

		// write response
		f, err := os.Open(filename)
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	buffer := bytes.NewBuffer(data)

	runWakatimeCli(
		t,
		buffer,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--extra-heartbeats", "true",
		"--cursorpos", "12",
		"--sync-offline-activity", "2",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--write",
		"--verbose",
	)

	offlineCount, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)

	assert.Eventually(t, func() bool { return numCalls == 2 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ExtraHeartbeats_SyncLegacyOfflineActivity(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		var filename string

		switch numCalls {
		case 1:
			filename = "testdata/api_heartbeats_response_extra_heartbeats.json"
		case 2:
			filename = "testdata/api_heartbeats_response_extra_heartbeats_legacy_offline.json"
		case 3:
			filename = "testdata/api_heartbeats_response_extra_heartbeats_extra.json"
		case 4:
			filename = "testdata/api_heartbeats_response_extra_heartbeats_extra.json"
		}

		// write response
		f, err := os.Open(filename)
		require.NoError(t, err)

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	// create legacy offline queue file and add some heartbeats
	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "legacy-offline-file")
	require.NoError(t, err)

	// early close to avoid file locking in Windows
	offlineQueueFileLegacy.Close()

	db, err := bolt.Open(offlineQueueFileLegacy.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "new-offline-file")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	buffer := bytes.NewBuffer(data)

	assert.FileExists(t, offlineQueueFileLegacy.Name())

	runWakatimeCli(
		t,
		buffer,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--extra-heartbeats", "true",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--heartbeat-rate-limit-seconds", "0",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--write",
		"--verbose",
	)

	assert.NoFileExists(t, offlineQueueFileLegacy.Name())

	offlineCount, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Zero(t, offlineCount)

	assert.Eventually(t, func() bool { return numCalls == 4 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_Err(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	projectFolder, err := filepath.Abs(".")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check headers
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		// check body
		expectedBodyTpl, err := os.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		entityPath, err := realpath.Realpath("testdata/main.go")
		require.NoError(t, err)

		entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
		expectedBody := fmt.Sprintf(
			string(expectedBodyTpl),
			entityPath,
			"wakatime-cli",
			subfolders,
			heartbeat.UserAgent(""),
		)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		assert.JSONEq(t, expectedBody, string(body))

		// write response
		w.WriteHeader(http.StatusBadGateway)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrAPI,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--line-additions", "123",
		"--line-deletions", "456",
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--project", "wakatime-cli",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ErrAuth_InvalidAPIKEY(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(_ http.ResponseWriter, _ *http.Request) {
		numCalls++
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrAuth,
		"--api-url", apiURL,
		"--key", "invalid",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--project", "wakatime-cli",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	count, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, count)

	assert.Eventually(t, func() bool { return numCalls == 0 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_MalformedConfig(t *testing.T) {
	tmpDir := t.TempDir()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrConfigFileParse,
		"--entity", "testdata/main.go",
		"--config", "./testdata/malformed.cfg",
		"--internal-config", tmpInternalConfigFile.Name(),
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--verbose",
	)

	assert.Empty(t, out)

	count, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, count)
}

func TestSendHeartbeats_MalformedInternalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrConfigFileParse,
		"--entity", "testdata/main.go",
		"--config", tmpConfigFile.Name(),
		"--internal-config", "./testdata/internal-malformed.cfg",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--verbose",
	)

	assert.Empty(t, out)

	count, err := offline.CountHeartbeats(offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, count)
}

func TestFileExperts(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	projectFolder, err := filepath.Abs(".")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	var numCalls int

	router.HandleFunc("/users/current/file_experts",
		func(w http.ResponseWriter, req *http.Request) {
			numCalls++

			// check headers
			assert.Equal(t, http.MethodPost, req.Method)
			assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
			assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
			assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
			assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

			// check body
			expectedBodyTpl, err := os.ReadFile("testdata/api_file_experts_request_template.json")
			require.NoError(t, err)

			entityPath, err := realpath.Realpath("testdata/main.go")
			require.NoError(t, err)

			entityPath = strings.ReplaceAll(entityPath, `\`, `/`)
			expectedBody := fmt.Sprintf(
				string(expectedBodyTpl),
				entityPath,
				"wakatime-cli",
				subfolders,
			)

			body, err := io.ReadAll(req.Body)
			require.NoError(t, err)

			assert.JSONEq(t, expectedBody, string(body))

			// write response
			f, err := os.Open("testdata/api_file_experts_response.json")
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			_, err = io.Copy(w, f)
			require.NoError(t, err)
		})

	tmpDir := t.TempDir()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	out := runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--file-experts",
		"--verbose",
	)

	assert.Equal(t, "You: 4 hrs 15 mins | Steve: 22 mins\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodayGoal(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	tmpDir := t.TempDir()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	router.HandleFunc("/users/current/goals/11111111-1111-4111-8111-111111111111",
		func(w http.ResponseWriter, req *http.Request) {
			numCalls++

			// check request
			assert.Equal(t, http.MethodGet, req.Method)
			assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
			assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
			assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

			// write response
			f, err := os.Open("testdata/api_goals_id_response.json")
			require.NoError(t, err)

			w.WriteHeader(http.StatusOK)
			_, err = io.Copy(w, f)
			require.NoError(t, err)
		})

	out := runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--today-goal", "11111111-1111-4111-8111-111111111111",
		"--verbose",
	)

	assert.Equal(t, "3 hrs 23 mins\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestTodaySummary(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	var numCalls int

	tmpDir := t.TempDir()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	router.HandleFunc("/users/current/statusbar/today", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.Equal(t, []string{heartbeat.UserAgent("")}, req.Header["User-Agent"])

		// write response
		f, err := os.Open("testdata/api_statusbar_today_response.json")
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	out := runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--today",
		"--verbose",
	)

	assert.Equal(t, "20 secs\n", out)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestOfflineCount(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.Copy(w, strings.NewReader("500 error test"))
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrAPI,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	out = runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--offline-count",
		"--verbose",
	)

	assert.Equal(t, "1\n", out)
}

func TestOfflineCountEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	out := runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--key", "00000000-0000-4000-8000-000000000000",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--offline-count",
		"--verbose",
	)

	assert.Equal(t, "0\n", out)
}

func TestPrintOfflineHeartbeats(t *testing.T) {
	apiURL, router, close := setupTestServer()
	defer close()

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, err := io.Copy(w, strings.NewReader("500 error test"))
		require.NoError(t, err)
	})

	tmpDir := t.TempDir()

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	offlineQueueFileLegacy, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFileLegacy.Close()

	tmpConfigFile, err := os.CreateTemp(tmpDir, "wakatime.cfg")
	require.NoError(t, err)

	defer tmpConfigFile.Close()

	tmpInternalConfigFile, err := os.CreateTemp(tmpDir, "wakatime-internal.cfg")
	require.NoError(t, err)

	defer tmpInternalConfigFile.Close()

	out := runWakatimeCliExpectErr(
		t,
		exitcode.ErrAPI,
		"--api-url", apiURL,
		"--key", "00000000-0000-4000-8000-000000000000",
		"--config", tmpConfigFile.Name(),
		"--internal-config", tmpInternalConfigFile.Name(),
		"--entity", "testdata/main.go",
		"--cursorpos", "12",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--lineno", "42",
		"--lines-in-file", "100",
		"--time", "1585598059",
		"--hide-branch-names", ".*",
		"--project", "wakatime-cli",
		"--write",
		"--verbose",
	)

	assert.Empty(t, out)

	out = runWakatimeCli(
		t,
		&bytes.Buffer{},
		"--key", "00000000-0000-4000-8000-000000000000",
		"--offline-queue-file", offlineQueueFile.Name(),
		"--offline-queue-file-legacy", offlineQueueFileLegacy.Name(),
		"--print-offline-heartbeats", "10",
		"--verbose",
	)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		entity = windows.FormatFilePath(entity)
	}

	projectFolder, err := filepath.Abs(".")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	offlineHeartbeat, err := os.ReadFile("testdata/offline_heartbeat_template.json")
	require.NoError(t, err)

	offlineHeartbeatStr := fmt.Sprintf(
		string(offlineHeartbeat),
		entity, subfolders,
		heartbeat.UserAgent(""),
	)

	assert.Equal(t, offlineHeartbeatStr+"\n", out)
}

func TestUserAgent(t *testing.T) {
	out := runWakatimeCli(t, &bytes.Buffer{}, "--user-agent")
	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgent("")), out)
}

func TestUserAgentWithPlugin(t *testing.T) {
	out := runWakatimeCli(t, &bytes.Buffer{}, "--user-agent", "--plugin", "Wakatime/1.0.4")

	assert.Equal(t, fmt.Sprintf("%s\n", heartbeat.UserAgent("Wakatime/1.0.4")), out)
}

func TestVersion(t *testing.T) {
	out := runWakatimeCli(t, &bytes.Buffer{}, "--version")

	assert.Equal(t, "<local-build>\n", out)
}

func TestVersionVerbose(t *testing.T) {
	out := runWakatimeCli(t, &bytes.Buffer{}, "--version", "--verbose")

	assert.Regexp(t, regexp.MustCompile(fmt.Sprintf(
		"wakatime-cli\n  Version: <local-build>\n  Commit: [0-9a-f]{7}\n  Built: [0-9-:T]{19} UTC\n  OS/Arch: %s/%s\n",
		runtime.GOOS,
		runtime.GOARCH,
	)), out)
}

func TestMultipleRunners(t *testing.T) {
	var wg sync.WaitGroup

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime.cfg")
	require.NoError(t, err)

	defer tmpFile.Close()

	for i := 0; i < 20; i++ {
		wg.Add(1)

		go func(filepath string) {
			defer wg.Done()

			out := runWakatimeCli(
				t,
				&bytes.Buffer{},
				"--config", filepath,
				"--config-write", "debug=true",
			)

			assert.Empty(t, out)
		}(tmpFile.Name())
	}

	wg.Wait()
}

func binaryPath(t *testing.T) string {
	filename := fmt.Sprintf("./build/wakatime-cli-%s-%s", runtime.GOOS, runtime.GOARCH)

	switch runtime.GOOS {
	case "darwin", "linux", "freebsd", "netbsd", "openbsd":
		return filename
	case "windows":
		return filename + ".exe"
	default:
		t.Fatalf("OS %q not supported", runtime.GOOS)
		return ""
	}
}

func runWakatimeCli(t *testing.T, buffer *bytes.Buffer, args ...string) string {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer func() {
		f.Close()
		data, err := os.ReadFile(f.Name())
		require.NoError(t, err)

		fmt.Printf("logs: %s\n", string(data))

		os.Remove(f.Name())
	}()

	args = append([]string{"--log-file", f.Name()}, args...)

	return runCmd(exec.Command(binaryPath(t), args...), buffer) // #nosec G204
}

func runWakatimeCliExpectErr(t *testing.T, exitcode int, args ...string) string {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer func() {
		f.Close()
		data, err := os.ReadFile(f.Name())
		require.NoError(t, err)

		fmt.Printf("logs: %s\n", string(data))

		os.Remove(f.Name())
	}()

	args = append([]string{"--log-file", f.Name()}, args...)

	stdout, code := runCmdExpectErr(exec.Command(binaryPath(t), args...)) // #nosec G204

	assert.Equal(t, exitcode, code)

	return stdout
}

func runCmd(cmd *exec.Cmd, buffer *bytes.Buffer) string {
	fmt.Println(cmd.String())

	cmd.Stdin = buffer

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		fmt.Printf("failed to run command %s: %s\n", cmd, err)

		return ""
	}

	return stdout.String()
}

func runCmdExpectErr(cmd *exec.Cmd) (string, int) {
	fmt.Println(cmd.String())

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		fmt.Printf("ran command successfully, but was expecting error: %s\n", cmd)

		return "", -1
	}

	if exitcode, ok := err.(*exec.ExitError); ok {
		return stdout.String(), exitcode.ExitCode()
	}

	return stdout.String(), -1
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	router.HandleFunc("/plugins/errors", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

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
