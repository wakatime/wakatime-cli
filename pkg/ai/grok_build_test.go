package ai_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

type grokBuildTestSession struct {
	grokHome   string
	dir        string
	id         string
	projectDir string
}

func TestGrokBuildParse(t *testing.T) {
	ctx := context.Background()
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	firstPrompt := 0
	secondPrompt := 1

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("fix ", "grok-3", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(2*time.Second),
			grokBuildTestUserUpdate("main.go", "grok-3", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(4*time.Second),
			grokBuildTestTurnCompleted(100, 40)),
		grokBuildTestRawUpdate(t, session.id, base.Add(5*time.Second),
			grokBuildTestUserUpdate("commit", "grok-4.5", &secondPrompt)),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(6*time.Second),
			grokBuildTestTurnCompleted(20, 5)),
	})

	editFile := filepath.Join(session.projectDir, "main.go")
	skipFile := filepath.Join(session.grokHome, "sessions", "plan.md")
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "agent-hunk", editFile, base.Add(10*time.Second),
			"added", "agent", &firstPrompt, 3, 1),
		grokBuildTestHunk(t, session.id, "mixed-hunk", editFile, base.Add(11*time.Second),
			"added", "human", nil, 5, 0),
		grokBuildTestHunk(t, session.id, "mixed-hunk", editFile, base.Add(12*time.Second),
			"updated", "agent", &secondPrompt, 1, 0),
		grokBuildTestHunk(t, session.id, "mixed-hunk", editFile, base.Add(13*time.Second),
			"removed", "", nil, -6, 0),
		grokBuildTestHunk(t, session.id, "agent-hunk", editFile, base.Add(14*time.Second),
			"removed", "", nil, -3, -1),
		grokBuildTestHunk(t, session.id, "skip-hunk", skipFile, base.Add(15*time.Second),
			"added", "agent", &secondPrompt, 10, 0),
	})

	parser := ai.GrokBuild{
		After:             base,
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			editFile: heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 6)

	assert.Equal(t, "Grok Build "+session.id, got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, len([]rune("fix main.go")), got[0].AIPromptLength)
	assert.EqualValues(t, 100, got[0].AIInputTokens)
	assert.EqualValues(t, 40, got[0].AIOutputTokens)
	assert.Equal(t, session.projectDir, got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "grok/3")
	assert.Contains(t, got[0].UserAgent, "grok-build/0.2.111")

	assert.Equal(t, len([]rune("commit")), got[1].AIPromptLength)
	assert.EqualValues(t, 20, got[1].AIInputTokens)
	assert.EqualValues(t, 5, got[1].AIOutputTokens)
	assert.Contains(t, got[1].UserAgent, "grok/4.5")

	assertGrokBuildFileHeartbeat(t, got[2], editFile, 2, "grok/3")
	assertGrokBuildFileHeartbeat(t, got[3], editFile, 1, "grok/4.5")
	assertGrokBuildFileHeartbeat(t, got[4], editFile, -1, "grok/4.5")
	assertGrokBuildFileHeartbeat(t, got[5], editFile, -2, "grok/3")
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")
}

func TestGrokBuildParse_PromptShapesAndRewind(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	firstPrompt := 0
	secondPrompt := 1

	bashUpdate := grokBuildTestUserUpdate("ignored bash", "grok-4.5", &firstPrompt)
	bashUpdate["content"].(map[string]any)["_meta"] = map[string]any{"bash_command": ""}

	hostUpdate := grokBuildTestUserUpdate("ignored host", "grok-4.5", &firstPrompt)
	hostUpdate["_meta"].(map[string]any)["hostTurn"] = true

	session.writeUpdates(t, []string{
		grokBuildTestRawUpdate(t, session.id, base.Add(time.Second),
			grokBuildTestUserUpdate("old ", "grok-4.5", nil)),
		grokBuildTestRawUpdate(t, session.id, base.Add(2*time.Second),
			grokBuildTestUserUpdate("format", "grok-4.5", nil)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(4*time.Second),
			grokBuildTestUserUpdate("first ", "grok-4.5", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(5*time.Second), bashUpdate),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(6*time.Second),
			grokBuildTestUserUpdate("second", "grok-4.5", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(7*time.Second), hostUpdate),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(8*time.Second),
			grokBuildTestUserUpdate("left", "grok-4.5", &firstPrompt)),
		`{"malformed":`,
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(9*time.Second),
			grokBuildTestUserUpdate("right", "grok-4.5", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(10*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(11*time.Second),
			grokBuildTestUserUpdate("phantom", "grok-4.5", nil)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(12*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(13*time.Second),
			grokBuildTestUserUpdate("discarded", "grok-4.5", &secondPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(14*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(15*time.Second),
			map[string]any{"sessionUpdate": "rewind_marker", "target_prompt_index": 5}),
		grokBuildTestEnvelope(t, session.id, "unexpected/update", base.Add(16*time.Second),
			grokBuildTestUserUpdate("after rewind", "grok-4.5", &secondPrompt)),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(17*time.Second),
			grokBuildTestTurnCompleted(7, 2)),
	})
	session.writeHunks(t, nil)

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 7)

	lengths := make([]int, 0, len(got))
	for _, h := range got {
		lengths = append(lengths, h.AIPromptLength)
	}

	assert.Equal(t, []int{10, 5, 6, 4, 5, 9, 12}, lengths)
	assert.EqualValues(t, 7, got[6].AIInputTokens)
	assert.EqualValues(t, 2, got[6].AIOutputTokens)
}

func TestGrokBuildParse_RemovalReversesPriorModelContributions(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	firstPrompt := 0
	secondPrompt := 1

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("first", "grok-3", &firstPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(2*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3*time.Second),
			grokBuildTestUserUpdate("second", "grok-4.5", &secondPrompt)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(4*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
	})

	editFile := filepath.Join(session.projectDir, "mixed.go")
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "mixed", editFile, base.Add(5*time.Second),
			"added", "human", nil, 5, 0),
		grokBuildTestHunk(t, session.id, "mixed", editFile, base.Add(6*time.Second),
			"updated", "agent", &firstPrompt, 1, 0),
		grokBuildTestHunk(t, session.id, "mixed", editFile, base.Add(7*time.Second),
			"updated", "agent", &secondPrompt, 2, 0),
		grokBuildTestHunk(t, session.id, "mixed", editFile, base.Add(20*time.Second),
			"removed", "", nil, -8, 0),
	})

	got, err := (ai.GrokBuild{After: base.Add(10 * time.Second)}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assertGrokBuildFileHeartbeat(t, got[0], editFile, -1, "grok/3")
	assertGrokBuildFileHeartbeat(t, got[1], editFile, -2, "grok/4.5")
}

func TestGrokBuildParse_FileHeartbeatUsesSessionProjectForOutOfTreeFile(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0

	session.writeUpdates(t, nil)

	outOfTreeFile := filepath.Join(t.TempDir(), "plan.md")
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "out-of-tree", outOfTreeFile, base.Add(time.Second),
			"added", "agent", &promptIndex, 1, 0),
	})

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assertGrokBuildFileHeartbeat(t, got[0], outOfTreeFile, 1, "grok/4.5")
	assert.Equal(t, session.projectDir, got[0].ProjectPathOverride)
}

func TestGrokBuildParse_ReusedPromptIndexUsesModelAtHunkTime(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("old branch", "grok-3", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(1200*time.Millisecond),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(2*time.Second),
			map[string]any{"sessionUpdate": "rewind_marker", "target_prompt_index": 0}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3*time.Second),
			grokBuildTestUserUpdate("new branch", "grok-4.5", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3200*time.Millisecond),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
	})

	oldFile := filepath.Join(session.projectDir, "old.go")
	newFile := filepath.Join(session.projectDir, "new.go")
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "old", oldFile, base.Add(1500*time.Millisecond),
			"added", "agent", &promptIndex, 1, 0),
		grokBuildTestHunk(t, session.id, "new", newFile, base.Add(4*time.Second),
			"added", "agent", &promptIndex, 1, 0),
	})

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)

	var fileHeartbeats []heartbeat.Heartbeat

	for _, h := range got {
		if h.EntityType == heartbeat.FileType {
			fileHeartbeats = append(fileHeartbeats, h)
		}
	}

	require.Len(t, fileHeartbeats, 2)
	assertGrokBuildFileHeartbeat(t, fileHeartbeats[0], oldFile, 1, "grok/3")
	assertGrokBuildFileHeartbeat(t, fileHeartbeats[1], newFile, 1, "grok/4.5")
}

func TestGrokBuildParse_RewindKeepsHistoricalPromptsAndTokens(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("old branch", "grok-3", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(2*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(3*time.Second),
			grokBuildTestTurnCompleted(10, 2)),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(4*time.Second),
			map[string]any{"sessionUpdate": "rewind_marker", "target_prompt_index": 0}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(5*time.Second),
			grokBuildTestUserUpdate("new branch", "grok-4.5", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(6*time.Second),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(7*time.Second),
			grokBuildTestTurnCompleted(4, 1)),
	})
	session.writeHunks(t, nil)

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, len([]rune("old branch")), got[0].AIPromptLength)
	assert.EqualValues(t, 10, got[0].AIInputTokens)
	assert.EqualValues(t, 2, got[0].AIOutputTokens)
	assert.Equal(t, len([]rune("new branch")), got[1].AIPromptLength)
	assert.EqualValues(t, 4, got[1].AIInputTokens)
	assert.EqualValues(t, 1, got[1].AIOutputTokens)
}

func TestGrokBuildParse_ReusedPromptIndexUsesCurrentGenerationForUpdatedHunk(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("old branch", "grok-3", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(1200*time.Millisecond),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(2*time.Second),
			map[string]any{"sessionUpdate": "rewind_marker", "target_prompt_index": 0}),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3*time.Second),
			grokBuildTestUserUpdate("new branch", "grok-4.5", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(3200*time.Millisecond),
			map[string]any{"sessionUpdate": "agent_message_chunk"}),
	})

	editFile := filepath.Join(session.projectDir, "main.go")
	createdAt := base.Add(1500 * time.Millisecond)
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "reused", editFile, createdAt,
			"added", "agent", &promptIndex, 1, 0),
		// Grok persists hunk.created_at for updated records, so this timestamp
		// intentionally remains older than the rewound prompt generation.
		grokBuildTestHunk(t, session.id, "reused", editFile, createdAt,
			"updated", "agent", &promptIndex, 2, 0),
	})

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)

	var fileHeartbeats []heartbeat.Heartbeat

	for _, h := range got {
		if h.EntityType == heartbeat.FileType {
			fileHeartbeats = append(fileHeartbeats, h)
		}
	}

	require.Len(t, fileHeartbeats, 2)
	assertGrokBuildFileHeartbeat(t, fileHeartbeats[0], editFile, 1, "grok/3")
	assertGrokBuildFileHeartbeat(t, fileHeartbeats[1], editFile, 2, "grok/4.5")
}

func TestGrokBuildParse_OversizedRecordsDoNotBlockFollowingLines(t *testing.T) {
	session := setupGrokBuildTestSession(t, false)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0
	oversized := strings.Repeat("x", 10*1024*1024+1)

	session.writeUpdates(t, []string{
		`{"padding":"` + oversized + `"}`,
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("still parsed", "grok-4.5", &promptIndex)),
		grokBuildTestEnvelope(t, session.id, "_x.ai/session/update", base.Add(2*time.Second),
			grokBuildTestTurnCompleted(3, 1)),
	})

	editFile := filepath.Join(session.projectDir, "main.go")
	session.writeHunks(t, []string{
		`{"padding":"` + oversized + `"}`,
		grokBuildTestHunk(t, session.id, "valid", editFile, base.Add(3*time.Second),
			"added", "agent", &promptIndex, 1, 0),
	})

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, len([]rune("still parsed")), got[0].AIPromptLength)
	assert.EqualValues(t, 3, got[0].AIInputTokens)
	assertGrokBuildFileHeartbeat(t, got[1], editFile, 1, "grok/4.5")
}

func TestGrokBuildParse_CustomHomeAndNativePaths(t *testing.T) {
	session := setupGrokBuildTestSession(t, true)
	base := time.Date(2026, 7, 23, 22, 0, 0, 0, time.UTC)
	promptIndex := 0

	session.writeUpdates(t, []string{
		grokBuildTestEnvelope(t, session.id, "session/update", base.Add(time.Second),
			grokBuildTestUserUpdate("custom home", "grok-4.5", &promptIndex)),
	})

	editFile := filepath.Join(session.projectDir, "main.go")
	session.writeHunks(t, []string{
		grokBuildTestHunk(t, session.id, "valid", editFile, base.Add(2*time.Second),
			"added", "agent", &promptIndex, 1, 0),
	})

	badSessionDir := filepath.Join(session.grokHome, "sessions", "bad", "bad-session")
	require.NoError(t, os.MkdirAll(badSessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(badSessionDir, "summary.json"),
		[]byte(`{"invalid":`),
		0o644,
	))

	got, err := (ai.GrokBuild{After: base}).Parse(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, session.projectDir, got[0].ProjectPathOverride)
	assert.Equal(t, editFile, got[1].Entity)
	assert.Contains(t, got[0].UserAgent, "grok-build/0.2.111")
}

func assertGrokBuildFileHeartbeat(
	t *testing.T,
	h heartbeat.Heartbeat,
	entity string,
	lineChanges int,
	model string,
) {
	t.Helper()
	assert.Equal(t, heartbeat.FileType, h.EntityType)
	assert.Equal(t, entity, h.Entity)
	require.NotNil(t, h.AILineChanges)
	assert.Equal(t, lineChanges, *h.AILineChanges)
	require.NotNil(t, h.IsWrite)
	assert.True(t, *h.IsWrite)
	assert.Contains(t, h.UserAgent, model)
}

func setupGrokBuildTestSession(t *testing.T, customHome bool) grokBuildTestSession {
	t.Helper()

	home := t.TempDir()

	grokHome := filepath.Join(home, ".grok")
	if customHome {
		grokHome = t.TempDir()
	}

	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	if customHome {
		t.Setenv("GROK_HOME", grokHome)
	} else {
		t.Setenv("GROK_HOME", "")
	}

	projectDir := filepath.Join(home, "project")
	require.NoError(t, os.MkdirAll(projectDir, 0o755))

	sessionID := "019f90c0-6137-79c2-b26e-ef83b2cecfd1"
	sessionDir := filepath.Join(grokHome, "sessions", "project", sessionID)
	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(grokHome, "version.json"),
		[]byte(`{"version":"0.2.111"}`),
		0o644,
	))
	require.NoError(t, os.WriteFile(
		filepath.Join(sessionDir, "summary.json"),
		[]byte(`{
  "info": {"id": "`+sessionID+`", "cwd": "`+filepath.ToSlash(projectDir)+`"},
  "current_model_id": "grok-4.5",
  "git_root_dir": "`+filepath.ToSlash(projectDir)+`/"
}`),
		0o644,
	))

	return grokBuildTestSession{
		grokHome:   grokHome,
		dir:        sessionDir,
		id:         sessionID,
		projectDir: projectDir,
	}
}

func (s grokBuildTestSession) writeUpdates(t *testing.T, lines []string) {
	t.Helper()
	writeGrokBuildJSONL(t, filepath.Join(s.dir, "updates.jsonl"), lines)
}

func (s grokBuildTestSession) writeHunks(t *testing.T, lines []string) {
	t.Helper()
	writeGrokBuildJSONL(t, filepath.Join(s.dir, "hunk_records.jsonl"), lines)
}

func writeGrokBuildJSONL(t *testing.T, path string, lines []string) {
	t.Helper()

	contents := ""
	if len(lines) > 0 {
		contents = strings.Join(lines, "\n") + "\n"
	}

	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
}

func grokBuildTestEnvelope(
	t *testing.T,
	sessionID string,
	method string,
	timestamp time.Time,
	update map[string]any,
) string {
	t.Helper()

	return mustJSONLine(t, map[string]any{
		"timestamp": timestamp.Unix(),
		"method":    method,
		"params":    grokBuildTestParams(sessionID, timestamp, update),
	})
}

func grokBuildTestRawUpdate(
	t *testing.T,
	sessionID string,
	timestamp time.Time,
	update map[string]any,
) string {
	t.Helper()
	return mustJSONLine(t, grokBuildTestParams(sessionID, timestamp, update))
}

func grokBuildTestParams(sessionID string, timestamp time.Time, update map[string]any) map[string]any {
	return map[string]any{
		"sessionId": sessionID,
		"update":    update,
		"_meta":     map[string]any{"agentTimestampMs": timestamp.UnixMilli()},
	}
}

func grokBuildTestUserUpdate(text string, model string, promptIndex *int) map[string]any {
	meta := map[string]any{"modelId": model}
	if promptIndex != nil {
		meta["promptIndex"] = *promptIndex
	}

	return map[string]any{
		"sessionUpdate": "user_message_chunk",
		"content":       map[string]any{"type": "text", "text": text},
		"_meta":         meta,
	}
}

func grokBuildTestTurnCompleted(input int64, output int64) map[string]any {
	return map[string]any{
		"sessionUpdate": "turn_completed",
		"usage": map[string]any{
			"inputTokens":  input,
			"outputTokens": output,
		},
	}
}

func grokBuildTestHunk(
	t *testing.T,
	sessionID string,
	hunkID string,
	filePath string,
	timestamp time.Time,
	eventType string,
	authorType string,
	promptIndex *int,
	linesAdded int,
	linesRemoved int,
) string {
	t.Helper()

	hunk := map[string]any{
		"hunkId":       hunkID,
		"filePath":     filepath.ToSlash(filePath),
		"linesAdded":   linesAdded,
		"linesRemoved": linesRemoved,
		"eventType":    eventType,
		"sessionId":    sessionID,
		"timestamp":    timestamp,
	}
	if authorType != "" {
		hunk["authorType"] = authorType
	}

	if promptIndex != nil {
		hunk["promptIndex"] = *promptIndex
	}

	return mustJSONLine(t, hunk)
}
