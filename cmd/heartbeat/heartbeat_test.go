package heartbeat_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	cmdparams "github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/version"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"github.com/matishsiao/goInfo"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestSendHeartbeats(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	projectFolder, err := filepath.Abs("../..")
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime-config")
	require.NoError(t, err)

	defer tmpFile.Close()

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &[]any{&entity})
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entity.Entity,
			subfolders,
			heartbeat.UserAgent(context.Background(), plugin),
		)

		assert.True(t, strings.HasSuffix(entity.Entity, "testdata/main.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("config", tmpFile.Name())
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(context.Background(), v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_RateLimited(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(_ http.ResponseWriter, _ *http.Request) {
		// Should not be called
		numCalls++
	})

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime-config")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFileInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpFileInternal.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "offline-queue-file")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)
	v.Set("heartbeat-rate-limit-seconds", 500)
	v.Set("config", tmpFile.Name())
	v.Set("internal-config", tmpFileInternal.Name())
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("internal.heartbeats_last_sent_at", time.Now().Add(-time.Minute).Format(time.RFC3339))

	err = cmdheartbeat.SendHeartbeats(context.Background(), v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Zero(t, numCalls)
}

func TestSendHeartbeats_WithFiltering_Exclude(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("entity", `\tmp\main.go`)
	v.Set("exclude", `/tmp/`)
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin")
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(context.Background(), v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 0, numCalls)
}

func TestSendHeartbeats_ExtraHeartbeats(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	projectFolder, err := filepath.Abs("../..")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_extra_heartbeats_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[2].Entity, "testdata/main.py"))

		for i := 3; i < 25; i++ {
			assert.True(t, strings.HasSuffix(entities[i].Entity, "testdata/main.go"))
		}

		userAgent := heartbeat.UserAgent(ctx, plugin)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, subfolders, userAgent,
			entities[1].Entity, subfolders, userAgent,
			entities[2].Entity, subfolders, userAgent,
			entities[3].Entity, subfolders, userAgent,
			entities[4].Entity, subfolders, userAgent,
			entities[5].Entity, subfolders, userAgent,
			entities[6].Entity, subfolders, userAgent,
			entities[7].Entity, subfolders, userAgent,
			entities[8].Entity, subfolders, userAgent,
			entities[9].Entity, subfolders, userAgent,
			entities[10].Entity, subfolders, userAgent,
			entities[11].Entity, subfolders, userAgent,
			entities[12].Entity, subfolders, userAgent,
			entities[13].Entity, subfolders, userAgent,
			entities[14].Entity, subfolders, userAgent,
			entities[15].Entity, subfolders, userAgent,
			entities[16].Entity, subfolders, userAgent,
			entities[17].Entity, subfolders, userAgent,
			entities[18].Entity, subfolders, userAgent,
			entities[19].Entity, subfolders, userAgent,
			entities[20].Entity, subfolders, userAgent,
			entities[21].Entity, subfolders, userAgent,
			entities[22].Entity, subfolders, userAgent,
			entities[23].Entity, subfolders, userAgent,
			entities[24].Entity, subfolders, userAgent,
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 0)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 1)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("lineno", 2)
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ExtraHeartbeats_Sanitize(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	r, w, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		r.Close()
		w.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = r

	data, err := os.ReadFile("testdata/extra_heartbeats.json")
	require.NoError(t, err)

	go func() {
		_, err := w.Write(data)
		require.NoError(t, err)

		w.Close()
	}()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 0)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("hide-branch-names", true)
	v.Set("hide-file-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("lineno", 13)
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	db, err := bolt.Open(offlineQueueFile.Name(), 0600, nil)
	require.NoError(t, err)

	defer func() {
		err = db.Close()
		require.NoError(t, err)
	}()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)

	hh, err := q.PopMany(1)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)
	assert.Len(t, hh, 1)

	info, err := goInfo.GetInfo()
	require.NoError(t, err)

	userAgent := fmt.Sprintf(
		"wakatime/%s (%s-%s-%s) %s %s",
		version.Version,
		runtime.GOOS,
		info.Core,
		info.Platform,
		runtime.Version(),
		plugin,
	)

	assert.Equal(t, []heartbeat.Heartbeat{
		{
			Branch:           nil,
			Category:         heartbeat.CodingCategory,
			CursorPosition:   nil,
			Dependencies:     nil,
			Entity:           "HIDDEN.go",
			EntityType:       heartbeat.FileType,
			IsWrite:          heartbeat.PointerTo(true),
			Language:         heartbeat.PointerTo("Go"),
			LineNumber:       nil,
			Lines:            nil,
			Project:          heartbeat.PointerTo("wakatime-cli"),
			ProjectRootCount: nil,
			Time:             1585598059,
			UserAgent:        userAgent,
		}}, hh)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_NonExistingEntity(t *testing.T) {
	resetSingleton(t)

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	ctx := context.Background()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", "https://example.org")
	v.Set("entity", "nonexisting")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := cmd.SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestSendHeartbeats_IsUnsavedEntity(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	projectFolder, err := filepath.Abs("../..")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_is_unsaved_entity_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "missing"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "missing-from-extra-heartbeats"))
		assert.True(t, strings.HasSuffix(entities[2].Entity, "main.go"))

		userAgent := heartbeat.UserAgent(ctx, plugin)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, userAgent,
			entities[1].Entity, userAgent,
			entities[2].Entity, subfolders, userAgent,
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_is_unsaved_entity.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	inr, inw, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		inr.Close()
		inw.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = inr

	data, err := os.ReadFile("testdata/extra_heartbeats_is_unsaved_entity.json")
	require.NoError(t, err)

	go func() {
		_, err := inw.Write(data)
		require.NoError(t, err)

		inw.Close()
	}()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("is-unsaved-entity", true)
	v.Set("category", "coding")
	v.Set("cursorpos", 41)
	v.Set("entity", "missing")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("project", "wakatime-cli")
	v.Set("hide-branch-names", true)
	v.Set("lineno", 11)
	v.Set("lines-in-file", 91)
	v.Set("plugin", plugin)
	v.Set("time", 1585598051)
	v.Set("timeout", 5)
	v.Set("extra-heartbeats", true)
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := cmd.SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestSendHeartbeats_NonExistingExtraHeartbeatsEntity(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	projectFolder, err := filepath.Abs("../..")
	require.NoError(t, err)

	subfolders := project.CountSlashesInProjectFolder(projectFolder)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_extra_heartbeats_filtered_template.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.True(t, strings.HasSuffix(entities[0].Entity, "testdata/main.go"))
		assert.True(t, strings.HasSuffix(entities[1].Entity, "testdata/main.py"))

		userAgent := heartbeat.UserAgent(ctx, plugin)

		expectedBodyStr := fmt.Sprintf(
			string(expectedBody),
			entities[0].Entity, subfolders, userAgent,
			entities[1].Entity, subfolders, userAgent,
		)

		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response_extra_heartbeats_filtered.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	inr, inw, err := os.Pipe()
	require.NoError(t, err)

	defer func() {
		inr.Close()
		inw.Close()
	}()

	origStdin := os.Stdin

	defer func() { os.Stdin = origStdin }()

	os.Stdin = inr

	data, err := os.ReadFile("testdata/extra_heartbeats_nonexisting_entity.json")
	require.NoError(t, err)

	go func() {
		_, err := inw.Write(data)
		require.NoError(t, err)

		inw.Close()
	}()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("hide-branch-names", true)
	v.Set("project", "wakatime-cli")
	v.Set("extra-heartbeats", true)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := cmd.SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	offlineQueueFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "skipping because of non-existing file")
}

func TestSendHeartbeats_ErrAuth_UnsetAPIKey(t *testing.T) {
	resetSingleton(t)

	_, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusCreated)
	})

	v := viper.New()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(context.Background(), v, offlineQueueFile.Name())
	require.Error(t, err)

	var errauth api.ErrAuth

	assert.ErrorAs(t, err, &errauth)

	assert.EqualError(
		t,
		err,
		"failed to load command parameters: failed to load API parameters: api key not found or empty",
	)

	assert.Eventually(t, func() bool { return numCalls == 0 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ErrBackoff(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusInternalServerError)
	})

	ctx := context.Background()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("internal.backoff_at", time.Now().Add(10*time.Minute).Format(ini.DateFormat))
	v.Set("internal.backoff_retries", "1")
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())

	logger, err := cmd.SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.ErrorAs(t, err, &api.ErrBackoff{})

	assert.Equal(t, 0, numCalls)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Empty(t, string(output))
}

func TestSendHeartbeats_ErrBackoff_Verbose(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, _ *http.Request) {
		numCalls++

		w.WriteHeader(http.StatusInternalServerError)
	})

	ctx := context.Background()

	tmpDir := t.TempDir()

	logFile, err := os.CreateTemp(tmpDir, "")
	require.NoError(t, err)

	defer logFile.Close()

	v := viper.New()
	v.Set("internal.backoff_at", time.Now().Add(10*time.Minute).Format(ini.DateFormat))
	v.Set("internal.backoff_retries", "1")
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("entity", "testdata/main.go")
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("log-file", logFile.Name())
	v.Set("verbose", true)

	logger, err := cmd.SetupLogging(ctx, v)
	require.NoError(t, err)

	defer logger.Flush()

	ctx = log.ToContext(ctx, logger)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.Error(t, err)
	assert.ErrorAs(t, err, &api.ErrBackoff{})

	assert.Equal(t, 0, numCalls)

	offlineCount, err := offline.CountHeartbeats(ctx, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, offlineCount)

	output, err := io.ReadAll(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(output), "will retry again after")
}

func TestSendHeartbeats_ObfuscateProject(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	fp := setupTestGitBasic(t)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_template_obfuscated_project.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &[]any{&entity})
		require.NoError(t, err)

		lines, err := project.ReadFile(ctx, filepath.Join(fp, "wakatime-cli", ".wakatime-project"), 1)
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(string(expectedBody), entity.Entity, lines[0], heartbeat.UserAgent(ctx, plugin))

		assert.True(t, strings.HasSuffix(entity.Entity, "src/pkg/file.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", filepath.Join(fp, "wakatime-cli/src/pkg/file.go"))
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-project-names", true)
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_ObfuscateProjectNotBranch(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	ctx := context.Background()

	fp := setupTestGitBasic(t)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		assert.True(t, strings.HasSuffix(req.Header["User-Agent"][0], plugin), fmt.Sprintf(
			"%q should have suffix %q",
			req.Header["User-Agent"][0],
			plugin,
		))

		expectedBody, err := os.ReadFile("testdata/api_heartbeats_request_template_obfuscated_project_not_branch.json")
		require.NoError(t, err)

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entity struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &[]any{&entity})
		require.NoError(t, err)

		lines, err := project.ReadFile(ctx, filepath.Join(fp, "wakatime-cli", ".wakatime-project"), 1)
		require.NoError(t, err)

		expectedBodyStr := fmt.Sprintf(string(expectedBody), entity.Entity, lines[0], heartbeat.UserAgent(ctx, plugin))

		assert.True(t, strings.HasSuffix(entity.Entity, "src/pkg/file.go"))
		assert.JSONEq(t, expectedBodyStr, string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)

		numCalls++
	})

	v := viper.New()
	v.SetDefault("sync-offline-activity", 1000)
	v.Set("api-url", testServerURL)
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", filepath.Join(fp, "wakatime-cli/src/pkg/file.go"))
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Go")
	v.Set("alternate-language", "Golang")
	v.Set("hide-project-names", true)
	v.Set("hide-branch-names", false)
	v.Set("lineno", 13)
	v.Set("local-file", "testdata/localfile.go")
	v.Set("plugin", plugin)
	v.Set("time", 1585598059.1)
	v.Set("timeout", 5)
	v.Set("write", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(ctx, v, offlineQueueFile.Name())
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestRateLimited(t *testing.T) {
	resetSingleton(t)

	p := cmdheartbeat.RateLimitParams{
		Timeout:    time.Duration(offline.RateLimitDefaultSeconds) * time.Second,
		LastSentAt: time.Now(),
	}

	assert.True(t, cmdheartbeat.RateLimited(p))
}

func TestRateLimited_NotLimited(t *testing.T) {
	resetSingleton(t)

	p := cmdheartbeat.RateLimitParams{
		LastSentAt: time.Now().Add(time.Duration(-offline.RateLimitDefaultSeconds*2) * time.Second),
		Timeout:    time.Duration(offline.RateLimitDefaultSeconds) * time.Second,
	}

	assert.False(t, cmdheartbeat.RateLimited(p))
}

func TestRateLimited_Disabled(t *testing.T) {
	resetSingleton(t)

	p := cmdheartbeat.RateLimitParams{
		Disabled: true,
	}

	assert.False(t, cmdheartbeat.RateLimited(p))
}

func TestRateLimited_TimeoutZero(t *testing.T) {
	resetSingleton(t)

	p := cmdheartbeat.RateLimitParams{
		Timeout: 0,
	}

	assert.False(t, cmdheartbeat.RateLimited(p))
}

func TestRateLimited_LastSentAtZero(t *testing.T) {
	resetSingleton(t)

	p := cmdheartbeat.RateLimitParams{
		LastSentAt: time.Time{},
	}

	assert.False(t, cmdheartbeat.RateLimited(p))
}

func TestResetRateLimit(t *testing.T) {
	resetSingleton(t)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpFileInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpFileInternal.Close()

	ctx := context.Background()

	v := viper.New()
	v.Set("config", tmpFileInternal.Name())
	v.Set("internal-config", tmpFileInternal.Name())

	writer, err := ini.NewWriter(ctx, v, func(_ context.Context, vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFileInternal.Name(), nil
	})
	require.NoError(t, err)

	err = cmdheartbeat.ResetRateLimit(ctx, v)
	require.NoError(t, err)

	err = writer.File.Reload()
	require.NoError(t, err)

	lastSentAt := writer.File.Section("internal").Key("heartbeats_last_sent_at").MustTimeFormat(ini.DateFormat)

	assert.WithinDuration(t, time.Now(), lastSentAt, 1*time.Second)
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}

func setupTestGitBasic(t *testing.T) (fp string) {
	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	if runtime.GOOS == "windows" {
		tmpDir = windows.FormatFilePath(tmpDir)
	}

	err = os.MkdirAll(filepath.Join(tmpDir, "wakatime-cli/src/pkg"), os.FileMode(int(0700)))
	require.NoError(t, err)

	tmpFile, err := os.Create(filepath.Join(tmpDir, "wakatime-cli/src/pkg/file.go"))
	require.NoError(t, err)

	defer tmpFile.Close()

	err = os.Mkdir(filepath.Join(tmpDir, "wakatime-cli/.git"), os.FileMode(int(0700)))
	require.NoError(t, err)

	copyFile(t, "testdata/git_basic/config", filepath.Join(tmpDir, "wakatime-cli/.git/config"))
	copyFile(t, "testdata/git_basic/HEAD", filepath.Join(tmpDir, "wakatime-cli/.git/HEAD"))

	return tmpDir
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}

func resetSingleton(t *testing.T) {
	t.Helper()

	cmdparams.Once = sync.Once{}
}
