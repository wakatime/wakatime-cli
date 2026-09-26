package ai

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

func TestPatchHeartbeats_DeletedFiles(t *testing.T) {
	patch := "*** Begin Patch\n*** Delete File: gone.go\n*** Add File: new.go\n+new\n" +
		"*** Update File: edit.go\n-old\n*** Delete File: last.go\n*** End Patch"
	stamp := time.Now()
	base := t.TempDir()

	for name, hh := range map[string]Heartbeats{
		"Codex":    (Codex{}).patchHeartbeats(stamp, "", "", "", base, nil, "", patch, "session", heartbeat.AITokens{}),
		"Amp":      (Amp{}).patchHeartbeats(stamp, base, patch, "", "session", ""),
		"OpenCode": (OpenCode{}).patchHeartbeats("", "", "session", base, patch, stamp),
	} {
		t.Run(name, func(t *testing.T) {
			require.Len(t, hh, 4)

			for i, deleted := range []bool{true, false, false, true} {
				assert.Equal(t, deleted, hh[i].IsUnsavedEntity, "heartbeat %d", i)
			}
		})
	}
}

func TestOpenCodeDeletedFileMetadata(t *testing.T) {
	hh := (OpenCode{}).applyPatchHeartbeats("", "", "session", t.TempDir(), time.Now(), openCodeToolState{
		Metadata: json.RawMessage(`{"files":[{"filePath":"gone.go","type":"delete","deletions":3},` +
			`{"filePath":"edit.go","type":"update","deletions":3}]}`),
	})
	require.Len(t, hh, 2)
	assert.True(t, hh[0].IsUnsavedEntity)
	assert.False(t, hh[1].IsUnsavedEntity)
}

func TestClineDeletedFiles(t *testing.T) {
	dir := t.TempDir()
	data := `[{"ts":1781956800000,"say":"tool","text":"{\"tool\":\"fileDeleted\",\"path\":\"gone.go\"}"},` +
		`{"ts":1781956801000,"say":"tool",` +
		`"text":"{\"tool\":\"editedExistingFile\",\"path\":\"edit.go\",\"diff\":\"-old\"}"}]`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui_messages.json"), []byte(data), 0o600))
	hh, err := (Cline{}).parseTaskDir(dir)
	require.NoError(t, err)
	require.Len(t, hh, 2)
	assert.True(t, hh[0].IsUnsavedEntity)
	assert.Zero(t, *hh[0].AILineChanges)
	assert.False(t, hh[1].IsUnsavedEntity)
}

func TestCopilotDeletedPaths(t *testing.T) {
	for _, success := range []bool{true, false} {
		state := copilotCLIParseState{fileHeartbeatIDs: make(map[string]struct{})}
		data := `{"toolName":"apply_patch","success":true,"toolTelemetry":{"restrictedProperties":` +
			`{"deletedPaths":["/project/gone.go"],"addedPaths":["/project/new.go"],` +
			`"filePaths":["/project/edit.go","/project/gone.go"]}}}`

		var payload map[string]any
		require.NoError(t, json.Unmarshal([]byte(data), &payload))
		payload["success"] = success
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		(Copilot{}).handleCLIToolExecutionComplete(copilotCLIEvent{Timestamp: time.Now(), Data: raw}, &state)

		if !success {
			assert.Empty(t, state.heartbeats)
			continue
		}

		require.Len(t, state.heartbeats, 3)

		for _, timed := range state.heartbeats {
			assert.Equal(t, timed.heartbeat.Entity == "/project/gone.go", timed.heartbeat.IsUnsavedEntity)
		}
	}
}

func TestGenericDeletedFileTools(t *testing.T) {
	for _, tool := range []string{
		"delete_file", "deleteFile", "remove_file", "edit_file", "write_file", "delete_lines", "bash",
	} {
		t.Run(tool, func(t *testing.T) {
			event := genericAIEventFromValue(map[string]any{
				"tool_name": tool, "file_path": "missing.go", "timestamp": "2026-06-20T12:00:00Z",
			})
			hh := genericAIEventHeartbeats(Codex{}, ParserConfig{}, event)
			require.Len(t, hh, 2)
			assert.Equal(t, tool == "delete_file" || tool == "deleteFile" || tool == "remove_file", hh[1].IsUnsavedEntity)
		})
	}
}

func TestReplaceAppHeartbeats_DeletionThenRecreation(t *testing.T) {
	hh := replaceAppHeartbeats(Heartbeats{
		{AISession: "session", Entity: "file.go", EntityType: heartbeat.FileType, IsUnsavedEntity: true, Time: 100},
		{AISession: "session", Entity: "app", EntityType: heartbeat.AppType, AIPromptLength: 10, Time: 110},
		{AISession: "session", Entity: "file.go", EntityType: heartbeat.FileType, Time: 120},
		{AISession: "session", Entity: "app", EntityType: heartbeat.AppType, AIPromptLength: 20, Time: 130},
	})
	require.Len(t, hh, 2)
	assert.True(t, hh[0].IsUnsavedEntity)
	assert.Equal(t, 10, hh[0].AIPromptLength)
	assert.False(t, hh[1].IsUnsavedEntity)
	assert.Equal(t, 20, hh[1].AIPromptLength)
}
