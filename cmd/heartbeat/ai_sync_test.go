package heartbeat_test

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gandarez/go-realpath"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAISyncActivity_SendsAIHeartbeatsWithoutEntity(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	copyFile(t, "testdata/main.go", filepath.Join(tmpDir, "main.go"))

	entity, err := filepath.Abs(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity       string  `json:"entity"`
			Category     string  `json:"category"`
			Language     *string `json:"language"`
			Project      *string `json:"project"`
			UserAgent    string  `json:"user_agent"`
			AILineChange *int    `json:"ai_line_changes"`
		}

		require.NoError(t, json.Unmarshal(body, &entities))
		require.Len(t, entities, 1)
		assert.Equal(t, entity, entities[0].Entity)
		assert.Equal(t, "ai coding", entities[0].Category)
		require.NotNil(t, entities[0].Language)
		assert.Equal(t, "Go", *entities[0].Language)
		require.NotNil(t, entities[0].Project)
		assert.Equal(t, "myproject", *entities[0].Project)
		require.NotNil(t, entities[0].AILineChange)
		assert.Equal(t, -1, *entities[0].AILineChange)
		assert.Contains(t, entities[0].UserAgent, "ClaudeCode/2.1.45")
		assert.Contains(t, entities[0].UserAgent, "plugin/0.0.1")

		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("alternate-project", "fallback-project")
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/0.0.1")
	v.Set("project", "myproject")
	v.Set("project-folder", "/path/to/project")
	v.Set("timeout", 5)
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	code, err := cmdheartbeat.RunAISyncActivity(t.Context(), v)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, 1, numCalls)
}

func TestRunAISyncActivity_UsesAlternateProject(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	copyFile(t, "testdata/main.go", filepath.Join(tmpDir, "main.go"))

	entity, err := filepath.Abs(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, []string{"application/json"}, req.Header["Accept"])
		assert.Equal(t, []string{"application/json"}, req.Header["Content-Type"])

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity       string  `json:"entity"`
			Category     string  `json:"category"`
			Language     *string `json:"language"`
			Project      *string `json:"project"`
			UserAgent    string  `json:"user_agent"`
			AILineChange *int    `json:"ai_line_changes"`
		}

		require.NoError(t, json.Unmarshal(body, &entities))
		require.Len(t, entities, 1)
		assert.Equal(t, entity, entities[0].Entity)
		assert.Equal(t, "ai coding", entities[0].Category)
		require.NotNil(t, entities[0].Project)
		assert.Equal(t, "fallback-project", *entities[0].Project)

		w.WriteHeader(http.StatusCreated)

		f, err := os.Open("testdata/api_heartbeats_response.json")
		require.NoError(t, err)

		defer f.Close()

		_, err = io.Copy(w, f)
		require.NoError(t, err)
	})

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

	v := viper.New()
	v.Set("api-url", testServerURL)
	v.Set("alternate-project", "fallback-project")
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("plugin", "plugin/0.0.1")
	v.Set("timeout", 5)
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	code, err := cmdheartbeat.RunAISyncActivity(t.Context(), v)
	require.NoError(t, err)
	assert.Equal(t, 0, code)
	assert.Equal(t, 1, numCalls)
}

func TestRunAISyncActivity_RateLimitedWithoutEntity_SavesOffline(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tmpDir := t.TempDir()

	tmpDir, err := realpath.Realpath(tmpDir)
	require.NoError(t, err)

	copyFile(t, "testdata/main.go", filepath.Join(tmpDir, "main.go"))

	entity, err := filepath.Abs(filepath.Join(tmpDir, "main.go"))
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "offline-queue-file")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	v := viper.New()
	v.Set("heartbeat-rate-limit-seconds", 120)
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("internal.heartbeats_last_sent_at", time.Now().Format(ini.DateFormat))
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("offline-queue-file", offlineQueueFile.Name())
	v.Set("plugin", "plugin/0.0.1")
	v.Set("project", "myproject")
	v.Set("project-folder", "/path/to/project")
	v.Set("timeout", 5)

	code, err := cmdheartbeat.RunAISyncActivity(t.Context(), v)
	require.NoError(t, err)
	assert.Equal(t, 0, code)

	offlineCount, err := offline.CountHeartbeats(t.Context(), offlineQueueFile.Name())
	require.NoError(t, err)
	assert.Equal(t, 1, offlineCount)
}
