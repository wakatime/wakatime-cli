package ai_test

import (
	"context"
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

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/params"
)

func TestWithAISyncUpdatesLastParsedAtBeforeParsing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpInternal.Close()

	v := viper.New()
	v.Set("internal-config", tmpInternal.Name())

	handle := ai.WithAISync(ai.Config{
		V: v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	_, err = handle(t.Context(), []heartbeat.Heartbeat{})
	require.NoError(t, err)

	writer, err := ini.NewWriter(t.Context(), v, ini.InternalFilePath)
	require.NoError(t, err)

	err = writer.File.Reload()
	require.NoError(t, err)

	lastParsedAt, err := writer.File.Section("internal").Key("ai_logs_last_parsed_at").TimeFormat(ini.DateFormat)
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now(), lastParsedAt, 2*time.Second)
}

func TestSendHeartbeats_WithAIParsing(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	local, err := filepath.Abs("testdata/localfile.go")
	require.NoError(t, err)

	local = strings.ReplaceAll(local, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"" + local + "\",\"content\":\"first\\nsecond\\nthird\"}}",
		"{\"timestamp\":\"2026-03-18T13:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/empty.go\",\"structuredPatch\":[]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime-config")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

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

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity           string  `json:"entity"`
			Project          *string `json:"project"`
			Language         *string `json:"language"`
			ProjectRootCount *int    `json:"project_root_count"`
			UserAgent        string  `json:"user_agent"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.Equal(t, 2, len(entities))

		assert.Equal(t, entity, entities[0].Entity)
		assert.Equal(t, "wakatime-cli", *entities[0].Project)
		assert.Equal(t, "Golang", *entities[0].Language)
		assert.Greater(t, *entities[0].ProjectRootCount, 1)
		assert.Equal(
			t,
			heartbeat.UserAgent(t.Context(), "ClaudeCode/2.1.45"),
			entities[0].UserAgent,
		)
		assert.Contains(t, entities[0].UserAgent, "ClaudeCode/2.1.45")
		assert.NotContains(t, entities[0].UserAgent, plugin)
		assert.Equal(t, local, entities[1].Entity)

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
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", entity)
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("projectmap..*", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("plugin", plugin)
	v.Set("time", 1773835200.1)
	v.Set("timeout", 5)
	v.Set("write", true)
	v.Set("internal.ai_logs_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	params, heartbeats, err := testLoadParamsAndHeartbeats(t.Context(), v)
	require.NoError(t, err)

	err = cmdheartbeat.SendHeartbeats(t.Context(), v, params, offlineQueueFile.Name(), heartbeats)
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_WithAIParsingDisabled(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	local, err := filepath.Abs("testdata/localfile.go")
	require.NoError(t, err)

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"" + local + "\",\"content\":\"first\\nsecond\\nthird\"}}",
		"{\"timestamp\":\"2026-03-18T13:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"/tmp/empty.go\",\"structuredPatch\":[]}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var (
		plugin   = "plugin/0.0.1"
		numCalls int
	)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime-config")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

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

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity   string `json:"entity"`
			Category string `json:"category"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		assert.Equal(t, 1, len(entities))

		assert.Equal(t, entity, entities[0].Entity)
		assert.Equal(t, "debugging", entities[0].Category)

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
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", entity)
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("projectmap..*", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("plugin", plugin)
	v.Set("time", 1773835200.1)
	v.Set("timeout", 5)
	v.Set("write", true)
	v.Set("internal.ai_logs_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))
	v.Set("sync-ai-disabled", true)

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	params, heartbeats, err := testLoadParamsAndHeartbeats(t.Context(), v)
	require.NoError(t, err)

	err = cmdheartbeat.SendHeartbeats(t.Context(), v, params, offlineQueueFile.Name(), heartbeats)
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSendHeartbeats_WithAIParsingBatchAppliedOnce(t *testing.T) {
	resetSingleton(t)

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	local, err := filepath.Abs("testdata/localfile.go")
	require.NoError(t, err)

	local = strings.ReplaceAll(local, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5},{\"oldLines\":4,\"newLines\":1}]}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"" + local + "\",\"content\":\"first\\nsecond\\nthird\"}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	testServerURL, router, tearDown := setupTestServer()
	defer tearDown()

	var numCalls int

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime-config")
	require.NoError(t, err)

	defer tmpFile.Close()

	tmpInternalFile, err := os.CreateTemp(t.TempDir(), "wakatime-internal-config")
	require.NoError(t, err)

	defer tmpInternalFile.Close()

	router.HandleFunc("/users/current/heartbeats.bulk", func(w http.ResponseWriter, req *http.Request) {
		numCalls++

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)

		var entities []struct {
			Entity string `json:"entity"`
		}

		err = json.Unmarshal(body, &entities)
		require.NoError(t, err)

		require.Len(t, entities, 2)
		assert.Equal(t, entity, entities[0].Entity)
		assert.Equal(t, local, entities[1].Entity)

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
	v.Set("internal-config", tmpInternalFile.Name())
	v.Set("category", "debugging")
	v.Set("cursorpos", 42)
	v.Set("entity", entity)
	v.Set("entity-type", "file")
	v.Set("key", "00000000-0000-4000-8000-000000000000")
	v.Set("language", "Golang")
	v.Set("hide-branch-names", true)
	v.Set("projectmap..*", "wakatime-cli")
	v.Set("lineno", 13)
	v.Set("plugin", "plugin/0.0.1")
	v.Set("time", 1773835200.1)
	v.Set("timeout", 5)
	v.Set("write", true)
	v.Set("internal.ai_logs_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	params, heartbeats, err := testLoadParamsAndHeartbeats(t.Context(), v)
	require.NoError(t, err)

	heartbeats = append(heartbeats, heartbeat.Heartbeat{
		Entity:     local,
		EntityType: heartbeat.FileType,
		Time:       1773837000.1,
		UserAgent:  heartbeat.UserAgent(t.Context(), params.API.Plugin),
	})

	offlineQueueFile, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer offlineQueueFile.Close()

	err = cmdheartbeat.SendHeartbeats(t.Context(), v, params, offlineQueueFile.Name(), heartbeats)
	require.NoError(t, err)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestWithAISyncMarksHumanHeartbeatsAsAICoding(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	local, err := filepath.Abs("testdata/localfile.go")
	require.NoError(t, err)

	local = strings.ReplaceAll(local, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := strings.Join([]string{
		"{\"timestamp\":\"2026-03-18T12:00:00Z\",\"version\":\"2.1.45\"," +
			"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
			"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5}]}}",
		"{\"timestamp\":\"2026-03-18T12:30:00Z\",\"toolUseResult\":{" +
			"\"filePath\":\"" + local + "\",\"content\":\"first\\nsecond\\nthird\"}}",
	}, "\n") + "\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpInternal.Close()

	v := viper.New()
	v.Set("internal-config", tmpInternal.Name())
	v.Set("internal.ai_logs_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	handle := ai.WithAISync(ai.Config{
		Plugin: "plugin/0.0.1",
		V:      v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	humanWithinTwoMinutes := "/tmp/human-within-two-minutes.go"
	humanWithinThirtyMinutesNoChangesBeforeEdit := "/tmp/human-within-thirty-minutes-no-changes-before-edit.go"
	humanWithinThirtyMinutesWithChanges := "/tmp/human-within-thirty-minutes-with-changes.go"
	humanWithinThirtyMinutesNoChangesAfterEdit := "/tmp/human-within-thirty-minutes-no-changes-after-edit.go"
	humanOutsideThirtyMinutes := "/tmp/human-outside-thirty-minutes.go"
	humanNearbyDebugging := "/tmp/human-nearby-debugging.go"
	humanAIDuplicate := entity
	humanAfterAIDifferentWrite := entity

	got, err := handle(t.Context(), []heartbeat.Heartbeat{
		{
			Entity:           humanWithinTwoMinutes,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(3),
			Time:             1773835260.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanWithinThirtyMinutesNoChangesBeforeEdit,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773836999.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanWithinThirtyMinutesWithChanges,
			EntityType:       heartbeat.FileType,
			Category:         "debugging",
			HumanLineChanges: heartbeat.PointerTo(1),
			Time:             1773838500.2,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanWithinThirtyMinutesNoChangesAfterEdit,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773838500.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanOutsideThirtyMinutes,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773839100.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanNearbyDebugging,
			EntityType:       heartbeat.FileType,
			Category:         "debugging",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773835261.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanAIDuplicate,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773835201.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanAfterAIDifferentWrite,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(2),
			Time:             1773835203.1,
			UserAgent:        "editor/1.2.3",
		},
	})
	require.NoError(t, err)

	categoriesByEntity := make(map[string]string, len(got))

	countByEntity := make(map[string]int, len(got))
	for _, result := range got {
		categoriesByEntity[result.Heartbeat.Entity] = result.Heartbeat.Category
		countByEntity[result.Heartbeat.Entity]++
	}

	assert.Equal(t, "ai coding", categoriesByEntity[humanWithinTwoMinutes])
	assert.Equal(t, "ai coding", categoriesByEntity[humanWithinThirtyMinutesNoChangesBeforeEdit])
	assert.Equal(t, "debugging", categoriesByEntity[humanWithinThirtyMinutesWithChanges])
	assert.Equal(t, "ai coding", categoriesByEntity[humanWithinThirtyMinutesNoChangesAfterEdit])
	assert.Equal(t, "coding", categoriesByEntity[humanOutsideThirtyMinutes])
	assert.Equal(t, "ai coding", categoriesByEntity[humanNearbyDebugging])
	assert.Equal(t, "ai coding", categoriesByEntity[humanAfterAIDifferentWrite])
	assert.Equal(t, 1, countByEntity[humanAfterAIDifferentWrite])
}

func TestWithAISyncSkipsTwoMinuteAICodingForHumanEditsWithChanges(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	entity, err := filepath.Abs("testdata/main.go")
	require.NoError(t, err)

	entity = strings.ReplaceAll(entity, "\\", "/")

	transcriptDir := filepath.Join(home, ".claude", "projects", "sample-project")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcriptPath := filepath.Join(transcriptDir, "session.jsonl")
	transcript := "{\"timestamp\":\"2026-03-18T12:30:00Z\",\"version\":\"2.1.45\"," +
		"\"toolUseResult\":{\"filePath\":\"" + entity + "\"," +
		"\"structuredPatch\":[{\"oldLines\":3,\"newLines\":5}]}}\n"
	require.NoError(t, os.WriteFile(transcriptPath, []byte(transcript), 0o644))

	tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpInternal.Close()

	v := viper.New()
	v.Set("internal-config", tmpInternal.Name())
	v.Set("internal.ai_logs_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

	handle := ai.WithAISync(ai.Config{
		Plugin: "plugin/0.0.1",
		V:      v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	humanWithinTwoMinutesWithChanges := "/tmp/human-within-two-minutes-with-changes.go"
	humanWithinTwoMinutesNoChanges := "/tmp/human-within-two-minutes-no-changes.go"

	got, err := handle(t.Context(), []heartbeat.Heartbeat{
		{
			Entity:           humanWithinTwoMinutesWithChanges,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(3),
			Time:             1773837062.1,
			UserAgent:        "editor/1.2.3",
		},
		{
			Entity:           humanWithinTwoMinutesNoChanges,
			EntityType:       heartbeat.FileType,
			Category:         "coding",
			HumanLineChanges: heartbeat.PointerTo(0),
			Time:             1773837062.2,
			UserAgent:        "editor/1.2.3",
		},
	})
	require.NoError(t, err)

	categoriesByEntity := make(map[string]string, len(got))
	for _, result := range got {
		categoriesByEntity[result.Heartbeat.Entity] = result.Heartbeat.Category
	}

	assert.Equal(t, "coding", categoriesByEntity[humanWithinTwoMinutesWithChanges])
	assert.Equal(t, "ai coding", categoriesByEntity[humanWithinTwoMinutesNoChanges])
}

func TestWithAISync_ProducesExpectedHeartbeats(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	now := time.Now()
	transcriptDir := filepath.Join(home, ".codex", "sessions", now.Format("2026"), now.Format("04"), now.Format("15"))
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))

	transcript := "rollout-2026-04-15T18-46-36-019d9353-333c-7c41-a909-26f5c6221a5a.jsonl"
	transcriptPath := filepath.Join(transcriptDir, transcript)
	copyFile(t, filepath.Join("testdata", transcript), transcriptPath)

	after := time.Date(2026, 4, 16, 0, 0, 0, 0, time.UTC)

	tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpInternal.Close()

	v := viper.New()
	v.Set("internal-config", tmpInternal.Name())
	v.Set("internal.ai_logs_last_parsed_at", after.Format(ini.DateFormat))

	handle := ai.WithAISync(ai.Config{
		Plugin: "plugin/0.0.1",
		V:      v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	heartbeats, err := handle(t.Context(), []heartbeat.Heartbeat{})
	require.NoError(t, err)

	require.Len(t, heartbeats, 3)

	h := heartbeats[0].Heartbeat
	assert.Equal(t, float64(1776297745), h.Time)
	assert.Equal(t, heartbeat.FileType, h.EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/templates/index.html", h.Entity)
	assert.Equal(t, "ai coding", h.Category)
	assert.False(t, *h.IsWrite)
	assert.Equal(t, int64(105134), h.AIInputTokens)
	assert.Equal(t, int64(479), h.AIOutputTokens)
	assert.Equal(t, 29, h.AIPromptLength)
	assert.Nil(t, h.AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", h.ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", h.AISession)
	assert.Equal(t, "plus", h.AISubscriptionPlan)

	h = heartbeats[1].Heartbeat
	assert.Equal(t, float64(1776297784), h.Time)
	assert.Equal(t, heartbeat.FileType, h.EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/templates/index.html", h.Entity)
	assert.Equal(t, "ai coding", h.Category)
	assert.True(t, *h.IsWrite)
	assert.Equal(t, int64(326506), h.AIInputTokens)
	assert.Equal(t, int64(1442), h.AIOutputTokens)
	assert.Equal(t, 10, h.AIPromptLength)
	assert.Equal(t, -1, *h.AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", h.ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", h.AISession)
	assert.Equal(t, "plus", h.AISubscriptionPlan)

	h = heartbeats[2].Heartbeat
	assert.Equal(t, float64(1776297789), h.Time)
	assert.Equal(t, heartbeat.FileType, h.EntityType)
	assert.Equal(t, "/Users/user/git/wakatime-cli/static/css/index.less", h.Entity)
	assert.Equal(t, "ai coding", h.Category)
	assert.True(t, *h.IsWrite)
	assert.Equal(t, int64(110200), h.AIInputTokens)
	assert.Equal(t, int64(223), h.AIOutputTokens)
	assert.Zero(t, h.AIPromptLength)
	assert.Equal(t, 23, *h.AILineChanges)
	assert.Equal(t, "/Users/user/git/wakatime-cli", h.ProjectPathOverride)
	assert.Equal(t, "019d9353-333c-7c41-a909-26f5c6221a5a", h.AISession)
	assert.Equal(t, "plus", h.AISubscriptionPlan)
}

func TestWithAISync_PreservesCopilotTokensAfterMergingAppHeartbeat(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	workspaceDir := filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage", "workspace-1")
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatSessions"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(workspaceDir, "chatEditingSessions", "session-1"), 0o755))

	mainFile := filepath.Join(home, "project", "main.go")
	secondFile := filepath.Join(home, "project", "util.go")

	require.NoError(t, os.MkdirAll(filepath.Dir(mainFile), 0o755))
	require.NoError(t, os.WriteFile(mainFile, []byte("package main\n"), 0o644))
	require.NoError(t, os.WriteFile(secondFile, []byte("package main\nfunc util() {}\n"), 0o644))

	sessionPath := filepath.Join(workspaceDir, "chatSessions", "session-1.json")
	session := map[string]any{
		"version":         3,
		"creationDate":    int64(1770000000000),
		"lastMessageDate": int64(1770000100000),
		"sessionId":       "session-1",
		"requests": []any{
			map[string]any{
				"requestId": "request-1",
				"timestamp": int64(1770000001000),
				"agent": map[string]any{
					"extensionVersion": "0.42.3",
				},
				"message": map[string]any{
					"text": "Update the file and explain what changed",
				},
				"result": map[string]any{
					"metadata": map[string]any{
						"promptTokens": 9,
						"outputTokens": 4,
					},
				},
				"modelState": map[string]any{
					"value":       1,
					"completedAt": int64(1770000009000),
				},
				"variableData": map[string]any{
					"variables": []any{
						map[string]any{
							"kind": "file",
							"id":   "file://" + strings.ReplaceAll(mainFile, " ", "%20"),
							"value": map[string]any{
								"fsPath": mainFile,
							},
						},
					},
				},
				"response": []any{
					map[string]any{
						"kind": "toolInvocationSerialized",
						"invocationMessage": map[string]any{
							"value": "Reading files",
							"uris": map[string]any{
								"file://" + strings.ReplaceAll(secondFile, " ", "%20"): map[string]any{
									"fsPath": secondFile,
								},
							},
						},
					},
				},
			},
		},
	}
	data, err := json.Marshal(session)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(sessionPath, data, 0o644))

	statePath := filepath.Join(workspaceDir, "chatEditingSessions", "session-1", "state.json")
	state := map[string]any{
		"version": 2,
		"timeline": map[string]any{
			"fileBaselines": []any{
				[]any{
					"file://" + mainFile + "::request-1",
					map[string]any{
						"content": "package main\n",
					},
				},
			},
			"operations": []any{
				map[string]any{
					"type":      "textEdit",
					"requestId": "request-1",
					"uri": map[string]any{
						"fsPath": mainFile,
					},
					"epoch": 1,
					"edits": []any{
						map[string]any{
							"text": "package main\n\nfunc main() {}\n",
							"range": map[string]any{
								"startLineNumber": 1,
								"startColumn":     1,
								"endLineNumber":   2,
								"endColumn":       1,
							},
						},
					},
				},
			},
		},
	}
	data, err = json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, data, 0o644))

	after := time.Unix(1769999990, 0)
	tmpInternal, err := os.CreateTemp(t.TempDir(), "wakatime-internal")
	require.NoError(t, err)

	defer tmpInternal.Close()

	v := viper.New()
	v.Set("internal-config", tmpInternal.Name())
	v.Set("internal.ai_logs_last_parsed_at", after.Format(ini.DateFormat))

	handle := ai.WithAISync(ai.Config{
		Plugin: "editor/1.0.0",
		V:      v,
	})(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Heartbeat: hh[i]}
		}

		return results, nil
	})

	results, err := handle(t.Context(), []heartbeat.Heartbeat{})
	require.NoError(t, err)
	require.Len(t, results, 4)

	var (
		mergedRead heartbeat.Heartbeat
		writeEdit  heartbeat.Heartbeat
		foundRead  bool
		foundEdit  bool
	)

	for _, result := range results {
		h := result.Heartbeat
		if h.Entity == mainFile && h.IsWrite != nil && !*h.IsWrite && h.AIPromptLength > 0 {
			mergedRead = h
			foundRead = true
		}

		if h.Entity == mainFile && h.IsWrite != nil && *h.IsWrite {
			writeEdit = h
			foundEdit = true
		}
	}

	require.True(t, foundRead)
	assert.Equal(t, heartbeat.FileType, mergedRead.EntityType)
	assert.Equal(t, int64(9), mergedRead.AIInputTokens)
	assert.Equal(t, int64(4), mergedRead.AIOutputTokens)
	assert.Equal(t, len([]rune("Update the file and explain what changed")), mergedRead.AIPromptLength)
	assert.Equal(t, filepath.Dir(mainFile), mergedRead.ProjectPathOverride)
	assert.Equal(t, "session-1", mergedRead.AISession)

	require.True(t, foundEdit)
	require.NotNil(t, writeEdit.AILineChanges)
	assert.Equal(t, 2, *writeEdit.AILineChanges)
	assert.Zero(t, writeEdit.AIInputTokens)
	assert.Zero(t, writeEdit.AIOutputTokens)
	assert.Zero(t, writeEdit.AIPromptLength)
	assert.Equal(t, filepath.Dir(mainFile), writeEdit.ProjectPathOverride)
	assert.Equal(t, "session-1", writeEdit.AISession)
}

func resetSingleton(t *testing.T) {
	t.Helper()

	params.Once = sync.Once{}
}

func setupTestServer() (string, *http.ServeMux, func()) {
	router := http.NewServeMux()
	srv := httptest.NewServer(router)

	return srv.URL, router, func() { srv.Close() }
}

func testLoadParamsAndHeartbeats(
	ctx context.Context,
	v *viper.Viper,
) (params.Params, []heartbeat.Heartbeat, error) {
	apiParams, err := params.LoadAPIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return params.Params{}, nil, fmt.Errorf("failed to load API parameters: %w", err)
	}

	heartbeatParams, err := params.LoadHeartbeatParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return params.Params{}, nil, fmt.Errorf("failed to load heartbeat params: %w", err)
	}

	aiParams, err := params.LoadAIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return params.Params{}, nil, fmt.Errorf("failed to load ai params: %w", err)
	}

	loaded := params.Params{
		AI:        aiParams,
		API:       apiParams,
		Heartbeat: heartbeatParams,
		Offline:   params.LoadOfflineParams(ctx, v, params.FlagReadOrderFlagPrecedence),
	}

	return loaded, cmdheartbeat.BuildHeartbeats(ctx, apiParams.Plugin, heartbeatParams), nil
}
