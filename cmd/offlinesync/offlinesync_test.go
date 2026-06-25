package offlinesync_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/cmd/offlinesync"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestRunWithRateLimiting(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

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

		assert.JSONEq(t, string(expectedBody), string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	// setup offline queue
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	db, err := bolt.Open(f.Name(), 0600, nil)
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

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", f.Name())
	v.Set("sync-offline-activity", 100)
	v.Set("plugin", plugin)

	code, err := offlinesync.RunWithRateLimiting(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, exitcode.Success, code)
	assert.Equal(t, 1, numCalls)
}

func TestRunWithoutRateLimiting(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

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

		assert.JSONEq(t, string(expectedBody), string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	// setup offline queue
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	db, err := bolt.Open(f.Name(), 0600, nil)
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

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", f.Name())
	v.Set("sync-offline-activity", 100)
	v.Set("plugin", plugin)

	code, err := offlinesync.RunWithoutRateLimiting(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, exitcode.Success, code)
	assert.Equal(t, 1, numCalls)
}

func TestRunWithoutRateLimiting_Disabled(t *testing.T) {
	v := viper.New()
	v.Set("disable-offline", true)

	code, err := offlinesync.RunWithoutRateLimiting(t.Context(), v)

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)
}

func TestRunWithRateLimiting_Disabled(t *testing.T) {
	v := viper.New()
	v.Set("disable-offline", true)

	code, err := offlinesync.RunWithRateLimiting(t.Context(), v)

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)
}

func TestRunWithRateLimiting_RateLimited(t *testing.T) {
	resetSingleton(t)

	v := viper.New()
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("heartbeat-rate-limit-seconds", 500)
	v.Set("internal.heartbeats_last_sent_at", time.Now().Add(-time.Minute).Format(time.RFC3339))

	code, err := offlinesync.RunWithRateLimiting(t.Context(), v)
	require.NoError(t, err)

	assert.Equal(t, exitcode.Success, code)
}

func TestRunWithoutRateLimiting_QueueFilepathError(t *testing.T) {
	v := viper.New()
	v.Set("offline-queue-file", "~missing-user/wakatime.bdb")

	code, err := offlinesync.RunWithoutRateLimiting(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrGeneric, code)
	assert.Contains(t, err.Error(), "offline sync failed: failed to load offline queue filepath")
}

func TestRunWithoutRateLimiting_APIParamsError(t *testing.T) {
	resetSingleton(t)
	t.Setenv("WAKATIME_API_KEY", "")

	queueFile, err := os.CreateTemp(t.TempDir(), "offline-queue")
	require.NoError(t, err)
	require.NoError(t, queueFile.Close())

	v := viper.New()
	v.Set("offline-queue-file", queueFile.Name())

	code, err := offlinesync.RunWithoutRateLimiting(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrAuth, code)
	assert.Contains(t, err.Error(), "offline sync failed: failed to load API parameters")
}

func TestRunWithoutRateLimiting_CorruptDBReturnsWakaExitCode(t *testing.T) {
	resetSingleton(t)

	queueFilepath := filepath.Join(t.TempDir(), "offline_heartbeats.bdb")
	require.NoError(t, os.WriteFile(queueFilepath, []byte("not a bolt db"), 0600))

	v := viper.New()
	v.Set("api-url", "http://127.0.0.1:1")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", queueFilepath)

	code, err := offlinesync.RunWithoutRateLimiting(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.Success, code)
	assert.Contains(t, err.Error(), "offline sync failed")
	assert.Contains(t, err.Error(), "moved corrupt db file")
	assert.NoFileExists(t, queueFilepath)
}

func TestSyncOfflineActivity_APIParamsError(t *testing.T) {
	resetSingleton(t)
	t.Setenv("WAKATIME_API_KEY", "")

	err := offlinesync.SyncOfflineActivity(t.Context(), viper.New(), filepath.Join(t.TempDir(), "queue.bdb"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load API parameters")
}

func TestSyncOfflineActivity(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

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

		assert.JSONEq(t, string(expectedBody), string(body))

		// send response
		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	// setup offline queue
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	db, err := bolt.Open(f.Name(), 0600, nil)
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

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("sync-offline-activity", 100)
	v.Set("plugin", plugin)

	err = offlinesync.SyncOfflineActivity(t.Context(), v, f.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, numCalls)
}

func TestSyncOfflineActivity_MultipleApiKey(t *testing.T) {
	resetSingleton(t)

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		// check auth header
		switch numCalls {
		case 1:
			assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])
		case 2:
			assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAx"}, req.Header["Authorization"])
		}

		// send response
		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		w.WriteHeader(http.StatusCreated)
		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	// setup offline queue
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	var hgo heartbeat.Heartbeat

	err = json.Unmarshal(dataGo, &hgo)
	require.NoError(t, err)

	hgo.APIKey = "00000000-0000-4000-8000-000000000000"

	dataGoChanged, err := json.Marshal(hgo)
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	var hpy heartbeat.Heartbeat

	err = json.Unmarshal(dataPy, &hpy)
	require.NoError(t, err)

	hpy.APIKey = "00000000-0000-4000-8000-000000000001"

	dataPyChanged, err := json.Marshal(hpy)
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGoChanged),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPyChanged),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("sync-offline-activity", 100)
	v.Set("plugin", plugin)

	err = offlinesync.SyncOfflineActivity(t.Context(), v, f.Name())
	require.NoError(t, err)

	assert.Equal(t, 1, numCalls)
}

func TestSyncOfflineActivity_MultipleAPIURLs(t *testing.T) {
	resetSingleton(t)

	// Setup default API server
	defaultAPIURL, defaultRouter, closeDefault := setupTestServer()
	defer closeDefault()

	// Setup custom API server for work projects
	customAPIURL, customRouter, closeCustom := setupTestServer()
	defer closeCustom()

	var (
		plugin             = "plugin/0.0.1"
		defaultServerCalls int
		customServerCalls  int
		mu                 sync.Mutex
	)

	// Handler for default server
	defaultRouter.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()

		defaultServerCalls++

		mu.Unlock()

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAw"}, req.Header["Authorization"])

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var heartbeats []heartbeat.Heartbeat

		err = json.Unmarshal(body, &heartbeats)
		require.NoError(t, err)

		// Generate dynamic response
		var responses [][]any
		for _, h := range heartbeats {
			responses = append(responses, []any{
				map[string]any{"data": map[string]string{"id": h.Entity}},
				201,
			})
		}

		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(map[string]any{"responses": responses})
		require.NoError(t, err)
	})

	// Handler for custom server
	customRouter.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		mu.Lock()

		customServerCalls++

		mu.Unlock()

		// check request
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"Basic MDAwMDAwMDAtMDAwMC00MDAwLTgwMDAtMDAwMDAwMDAwMDAx"}, req.Header["Authorization"])

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var heartbeats []heartbeat.Heartbeat

		err = json.Unmarshal(body, &heartbeats)
		require.NoError(t, err)

		// Generate dynamic response
		var responses [][]any
		for _, h := range heartbeats {
			responses = append(responses, []any{
				map[string]any{"data": map[string]string{"id": h.Entity}},
				201,
			})
		}

		w.WriteHeader(http.StatusCreated)

		err = json.NewEncoder(w).Encode(map[string]any{"responses": responses})
		require.NoError(t, err)
	})

	// setup offline queue with heartbeat that matches api_urls pattern
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	// Create heartbeat with path that will match the api_urls pattern
	workHeartbeat := heartbeat.Heartbeat{
		Branch:         heartbeat.PointerTo("heartbeat"),
		Category:       "coding",
		CursorPosition: heartbeat.PointerTo(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/work/projects/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.PointerTo(true),
		Language:       heartbeat.PointerTo("Go"),
		LineNumber:     heartbeat.PointerTo(42),
		Lines:          heartbeat.PointerTo(100),
		Project:        heartbeat.PointerTo("wakatime-cli"),
		Time:           1592868367.219124,
		UserAgent:      "wakatime/13.0.6",
	}

	workHeartbeatData, err := json.Marshal(workHeartbeat)
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/work/projects/main.go-true",
			Heartbeat: string(workHeartbeatData),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	v := viper.New()
	v.Set("api-url", defaultAPIURL)
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("sync-offline-activity", 100)
	v.Set("plugin", plugin)
	// Set api_urls pattern to match /work/ paths
	v.Set("api_urls./work/", customAPIURL+"|00000000-0000-4000-8000-000000000001")

	err = offlinesync.SyncOfflineActivity(t.Context(), v, f.Name())
	require.NoError(t, err)

	mu.Lock()
	defaultCalls := defaultServerCalls
	customCalls := customServerCalls
	mu.Unlock()

	assert.GreaterOrEqual(t, defaultCalls, 1)
	assert.GreaterOrEqual(t, customCalls, 1)
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

func resetSingleton(t *testing.T) {
	t.Helper()

	params.Once = sync.Once{}
}
