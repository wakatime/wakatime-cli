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

	lastParsedAt, err := writer.File.Section("internal").Key("ai_heartbeats_last_parsed_at").TimeFormat(ini.DateFormat)
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
			heartbeat.UserAgent(t.Context(), "ClaudeCode/2.1.45 "+plugin),
			entities[0].UserAgent,
		)
		assert.Contains(t, entities[0].UserAgent, "ClaudeCode/2.1.45")
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
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

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
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))
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
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

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
	v.Set("internal.ai_heartbeats_last_parsed_at", time.Date(2026, 3, 18, 11, 0, 0, 0, time.UTC).Format(ini.DateFormat))

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
