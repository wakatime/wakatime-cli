//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/ai"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	_ "modernc.org/sqlite"
)

func TestOpenCodeParse_LegacyStorage(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	baseDir := filepath.Join(home, ".local", "share", "opencode", "storage")
	sessionDir := filepath.Join(baseDir, "session", "proj_1")
	messageDir := filepath.Join(baseDir, "message", "ses_123")
	partDirUser := filepath.Join(baseDir, "part", "msg_user")
	partDirAssistant := filepath.Join(baseDir, "part", "msg_assistant")

	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.MkdirAll(messageDir, 0o755))
	require.NoError(t, os.MkdirAll(partDirUser, 0o755))
	require.NoError(t, os.MkdirAll(partDirAssistant, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "ses_123.json"), []byte(`{
  "id": "ses_123",
  "directory": "/workspace/project",
  "version": "1.4.4",
  "parentID": "ses_parent",
  "time": { "created": 1740000000000, "updated": 1740000004000 }
}`), 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(messageDir, "msg_user.json"), []byte(`{
  "id": "msg_user",
  "sessionID": "ses_123",
  "role": "user",
  "time": { "created": 1740000001000 }
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirUser, "part_user_text.json"), []byte(`{
  "id": "part_user_text",
  "messageID": "msg_user",
  "sessionID": "ses_123",
  "type": "text",
  "text": "Refactor this function"
}`), 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(messageDir, "msg_assistant.json"), []byte(`{
  "id": "msg_assistant",
  "sessionID": "ses_123",
  "role": "assistant",
  "modelID": "claude-3.5",
  "path": { "cwd": "/workspace/project", "root": "/workspace/project" },
  "tokens": { "input": 120, "output": 30 },
  "time": { "created": 1740000002000 }
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirAssistant, "part_assistant_text.json"), []byte(`{
  "id": "part_assistant_text",
  "messageID": "msg_assistant",
  "sessionID": "ses_123",
  "type": "text",
  "text": "I updated the file."
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirAssistant, "part_assistant_edit.json"), []byte(`{
  "id": "part_assistant_edit",
  "messageID": "msg_assistant",
  "sessionID": "ses_123",
  "type": "tool",
  "tool": "edit",
  "state": {
    "status": "completed",
    "input": {
      "filePath": "/workspace/project/main.go",
      "oldString": "old line",
      "newString": "new line\nextra line"
    },
    "metadata": {
      "filediff": {
        "filePath": "/workspace/project/main.go",
        "additions": 2,
        "deletions": 1
      }
    }
  }
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirAssistant, "part_assistant_write.json"), []byte(`{
  "id": "part_assistant_write",
  "messageID": "msg_assistant",
  "sessionID": "ses_123",
  "type": "tool",
  "tool": "write",
  "state": {
    "status": "completed",
    "input": {
      "filePath": "/workspace/project/existing.go",
      "content": "one\ntwo\nthree"
    },
    "metadata": {
      "exists": true
    }
  }
}`), 0o600))

	parser := ai.OpenCode{
		After:             time.Date(2025, 2, 19, 17, 20, 0, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/workspace/project/main.go":     heartbeat.UserAgent(ctx, "editor/1.2.3"),
			"/workspace/project/existing.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 4)

	assert.Equal(t, "OpenCode ses_123", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "ses_123", got[0].AISession)
	assert.Equal(t, "/workspace/project", got[0].ProjectPathOverride)
	assert.Equal(t, len([]rune("Refactor this function")), got[0].AIPromptLength)
	assert.Contains(t, got[0].UserAgent, "opencode-cli/1.4.4")
	assert.NotContains(t, got[0].UserAgent, "OpenCode/")

	assert.Equal(t, "OpenCode ses_123", got[1].Entity)
	assert.Equal(t, heartbeat.AppType, got[1].EntityType)
	assert.EqualValues(t, 120, got[1].AIInputTokens)
	assert.EqualValues(t, 30, got[1].AIOutputTokens)
	assert.Equal(t, "/workspace/project", got[1].ProjectPathOverride)
	assert.Contains(t, got[1].UserAgent, "claude/3.5 opencode-cli/1.4.4")
	assert.True(t, strings.Index(got[1].UserAgent, "claude/3.5") <
		strings.Index(got[1].UserAgent, "opencode-cli/1.4.4"))

	assert.Equal(t, "/workspace/project/main.go", got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	assert.Equal(t, "ses_123", got[2].AISession)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 1, *got[2].AILineChanges)
	require.NotNil(t, got[2].IsWrite)
	assert.True(t, *got[2].IsWrite)
	assert.Contains(t, got[2].UserAgent, "claude/3.5 opencode-cli/1.4.4")
	assert.True(t, strings.Index(got[2].UserAgent, "claude/3.5") <
		strings.Index(got[2].UserAgent, "opencode-cli/1.4.4"))
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")

	assert.Equal(t, "/workspace/project/existing.go", got[3].Entity)
	require.NotNil(t, got[3].AILineChanges)
	assert.Zero(t, *got[3].AILineChanges)
}

func TestOpenCodeParse_SQLiteFallback(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	baseDir := filepath.Join(home, ".local", "share", "opencode")
	storageDir := filepath.Join(baseDir, "storage")
	dbPath := filepath.Join(baseDir, "opencode.db")

	require.NoError(t, os.MkdirAll(filepath.Join(storageDir, "session", "proj_1"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(storageDir, "session", "proj_1", "ses_old.json"), []byte(`{
  "id": "ses_old",
  "directory": "/workspace/old",
  "version": "1.4.4",
  "time": { "created": 1740000000000, "updated": 1740000001000 }
}`), 0o600))

	createOpenCodeDB(t, dbPath, openCodeDBFixture{
		Sessions: []openCodeDBSession{
			{
				ID:        "ses_sqlite",
				Directory: "/workspace/project",
				Version:   "1.4.4",
			},
		},
		Messages: []openCodeDBMessage{
			{
				ID:        "msg_user",
				SessionID: "ses_sqlite",
				CreatedAt: 1740000003000,
				Data: map[string]any{
					"role": "user",
					"time": map[string]any{
						"created": int64(1740000003000),
					},
				},
			},
			{
				ID:        "msg_assistant",
				SessionID: "ses_sqlite",
				CreatedAt: 1740000004000,
				Data: map[string]any{
					"role": "assistant",
					"path": map[string]any{
						"cwd":  "/workspace/project",
						"root": "/workspace/project",
					},
					"tokens": map[string]any{
						"input":  200,
						"output": 50,
					},
					"time": map[string]any{
						"created": int64(1740000004000),
					},
				},
			},
		},
		Parts: []openCodeDBPart{
			{
				ID:        "part_user",
				MessageID: "msg_user",
				SessionID: "ses_sqlite",
				CreatedAt: 1740000003001,
				Data: map[string]any{
					"type": "text",
					"text": "Implement the patch",
				},
			},
			{
				ID:        "part_apply_patch",
				MessageID: "msg_assistant",
				SessionID: "ses_sqlite",
				CreatedAt: 1740000004001,
				Data: map[string]any{
					"type": "tool",
					"tool": "apply_patch",
					"state": map[string]any{
						"status": "completed",
						"metadata": map[string]any{
							"files": []map[string]any{
								{
									"filePath":  "/workspace/project/new.go",
									"additions": 3,
									"deletions": 1,
								},
							},
						},
					},
				},
			},
		},
	})

	parser := ai.OpenCode{
		After:             time.Date(2025, 2, 19, 17, 20, 2, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
		UserAgents: map[string]string{
			"/workspace/project/new.go": heartbeat.UserAgent(ctx, "editor/1.2.3"),
		},
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "OpenCode ses_sqlite", got[0].Entity)
	assert.Equal(t, "OpenCode ses_sqlite", got[1].Entity)
	assert.Equal(t, "/workspace/project/new.go", got[2].Entity)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 2, *got[2].AILineChanges)
	assert.Contains(t, got[2].UserAgent, "editor/1.2.3")
}

func TestOpenCodeParse_SQLiteFallback_UnicodeAndMalformedRows(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	baseDir := filepath.Join(home, ".local", "share", "opencode")
	dbPath := filepath.Join(baseDir, "opencode.db")
	editedPath := "/workspace/项目/文件.go"
	prompt := "请更新这个文件"

	require.NoError(t, os.MkdirAll(baseDir, 0o755))

	createOpenCodeDB(t, dbPath, openCodeDBFixture{
		Sessions: []openCodeDBSession{
			{
				ID:        "ses_unicode",
				Directory: "/workspace/项目",
				Version:   "1.15.3",
			},
		},
		Messages: []openCodeDBMessage{
			{
				ID:        "msg_user",
				SessionID: "ses_unicode",
				CreatedAt: 1740000003000,
				Data: map[string]any{
					"role": "user",
					"time": map[string]any{
						"created": int64(1740000003000),
					},
				},
			},
			{
				ID:        "msg_assistant",
				SessionID: "ses_unicode",
				CreatedAt: 1740000004000,
				Data: map[string]any{
					"role": "assistant",
					"path": map[string]any{
						"cwd":  "/workspace/项目",
						"root": "/workspace/项目",
					},
					"tokens": map[string]any{
						"input":  50,
						"output": 20,
					},
					"time": map[string]any{
						"created": int64(1740000004000),
					},
				},
			},
		},
		Parts: []openCodeDBPart{
			{
				ID:        "part_user",
				MessageID: "msg_user",
				SessionID: "ses_unicode",
				CreatedAt: 1740000003001,
				Data: map[string]any{
					"type": "text",
					"text": prompt,
				},
			},
			{
				ID:        "part_apply_patch",
				MessageID: "msg_assistant",
				SessionID: "ses_unicode",
				CreatedAt: 1740000004001,
				Data: map[string]any{
					"type": "tool",
					"tool": "apply_patch",
					"state": map[string]any{
						"status": "completed",
						"metadata": map[string]any{
							"files": []map[string]any{
								{
									"filePath":  editedPath,
									"additions": 2,
									"deletions": 1,
								},
							},
						},
					},
				},
			},
		},
	})

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO message(id, session_id, data, time_created) VALUES(?, ?, ?, ?)`,
		"msg_bad",
		"ses_unicode",
		"{",
		int64(1740000005000),
	)
	require.NoError(t, err)
	_, err = db.Exec(
		`INSERT INTO part(id, message_id, session_id, data, time_created) VALUES(?, ?, ?, ?, ?)`,
		"part_bad",
		"msg_assistant",
		"ses_unicode",
		"{",
		int64(1740000005001),
	)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	got, err := ai.OpenCode{
		After:             time.Date(2025, 2, 19, 17, 20, 2, 0, time.UTC),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, "OpenCode ses_unicode", got[0].Entity)
	assert.Equal(t, len([]rune(prompt)), got[0].AIPromptLength)
	assert.Equal(t, "/workspace/项目", got[0].ProjectPathOverride)

	assert.Equal(t, "OpenCode ses_unicode", got[1].Entity)
	assert.EqualValues(t, 50, got[1].AIInputTokens)
	assert.EqualValues(t, 20, got[1].AIOutputTokens)

	assert.Equal(t, editedPath, got[2].Entity)
	require.NotNil(t, got[2].AILineChanges)
	assert.Equal(t, 1, *got[2].AILineChanges)
}

func TestOpenCodeParse_LegacyStorage_AfterUsesSeedTokens(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	baseDir := filepath.Join(home, ".local", "share", "opencode", "storage")
	sessionDir := filepath.Join(baseDir, "session", "proj_1")
	messageDir := filepath.Join(baseDir, "message", "ses_456")
	partDirBefore := filepath.Join(baseDir, "part", "msg_before")
	partDirAfter := filepath.Join(baseDir, "part", "msg_after")

	require.NoError(t, os.MkdirAll(sessionDir, 0o755))
	require.NoError(t, os.MkdirAll(messageDir, 0o755))
	require.NoError(t, os.MkdirAll(partDirBefore, 0o755))
	require.NoError(t, os.MkdirAll(partDirAfter, 0o755))

	require.NoError(t, os.WriteFile(filepath.Join(sessionDir, "ses_456.json"), []byte(`{
  "id": "ses_456",
  "directory": "/workspace/project",
  "version": "1.4.4",
  "time": { "created": 1740000000000, "updated": 1740000005000 }
}`), 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(messageDir, "msg_before.json"), []byte(`{
  "id": "msg_before",
  "sessionID": "ses_456",
  "role": "assistant",
  "path": { "cwd": "/workspace/project", "root": "/workspace/project" },
  "tokens": { "input": 100, "output": 20 },
  "time": { "created": 1740000001000 }
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirBefore, "part_before_text.json"), []byte(`{
  "id": "part_before_text",
  "messageID": "msg_before",
  "sessionID": "ses_456",
  "type": "text",
  "text": "Earlier response"
}`), 0o600))

	require.NoError(t, os.WriteFile(filepath.Join(messageDir, "msg_after.json"), []byte(`{
  "id": "msg_after",
  "sessionID": "ses_456",
  "role": "assistant",
  "path": { "cwd": "/workspace/project", "root": "/workspace/project" },
  "tokens": { "input": 130, "output": 27 },
  "time": { "created": 1740000003000 }
}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(partDirAfter, "part_after_text.json"), []byte(`{
  "id": "part_after_text",
  "messageID": "msg_after",
  "sessionID": "ses_456",
  "type": "text",
  "text": "Later response"
}`), 0o600))

	parser := ai.OpenCode{
		After:             time.UnixMilli(1740000002000),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "OpenCode ses_456", got[0].Entity)
	assert.EqualValues(t, 30, got[0].AIInputTokens)
	assert.EqualValues(t, 7, got[0].AIOutputTokens)
}

func TestOpenCodeParse_SQLiteFallback_AfterUsesSeedTokens(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	baseDir := filepath.Join(home, ".local", "share", "opencode")
	dbPath := filepath.Join(baseDir, "opencode.db")

	require.NoError(t, os.MkdirAll(baseDir, 0o755))

	createOpenCodeDB(t, dbPath, openCodeDBFixture{
		Sessions: []openCodeDBSession{
			{
				ID:        "ses_seed",
				Directory: "/workspace/project",
				Version:   "1.4.4",
			},
		},
		Messages: []openCodeDBMessage{
			{
				ID:        "msg_before",
				SessionID: "ses_seed",
				CreatedAt: 1740000001000,
				Data: map[string]any{
					"role": "assistant",
					"path": map[string]any{
						"cwd":  "/workspace/project",
						"root": "/workspace/project",
					},
					"tokens": map[string]any{
						"input":  100,
						"output": 20,
					},
					"time": map[string]any{
						"created": int64(1740000001000),
					},
				},
			},
			{
				ID:        "msg_after",
				SessionID: "ses_seed",
				CreatedAt: 1740000003000,
				Data: map[string]any{
					"role": "assistant",
					"path": map[string]any{
						"cwd":  "/workspace/project",
						"root": "/workspace/project",
					},
					"tokens": map[string]any{
						"input":  130,
						"output": 27,
					},
					"time": map[string]any{
						"created": int64(1740000003000),
					},
				},
			},
		},
		Parts: []openCodeDBPart{
			{
				ID:        "part_after",
				MessageID: "msg_after",
				SessionID: "ses_seed",
				CreatedAt: 1740000003001,
				Data: map[string]any{
					"type": "text",
					"text": "Later response",
				},
			},
		},
	})

	parser := ai.OpenCode{
		After:             time.UnixMilli(1740000002000),
		FallbackUserAgent: "plugin/0.0.1",
	}

	got, err := parser.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	assert.Equal(t, "OpenCode ses_seed", got[0].Entity)
	assert.EqualValues(t, 30, got[0].AIInputTokens)
	assert.EqualValues(t, 7, got[0].AIOutputTokens)
}

func TestOpenCodeParse_NoStorageDir(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.OpenCode{After: time.Now()}.Parse(ctx)
	require.NoError(t, err)
	assert.Empty(t, got)
}

type openCodeDBFixture struct {
	Sessions []openCodeDBSession
	Messages []openCodeDBMessage
	Parts    []openCodeDBPart
}

type openCodeDBSession struct {
	ID        string
	Directory string
	Version   string
}

type openCodeDBMessage struct {
	ID        string
	SessionID string
	CreatedAt int64
	Data      map[string]any
}

type openCodeDBPart struct {
	ID        string
	MessageID string
	SessionID string
	CreatedAt int64
	Data      map[string]any
}

func createOpenCodeDB(t *testing.T, dbPath string, fixture openCodeDBFixture) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE session (
  id TEXT PRIMARY KEY,
  directory TEXT,
  version TEXT
);
CREATE TABLE message (
  id TEXT PRIMARY KEY,
  session_id TEXT,
  data TEXT,
  time_created INTEGER
);
CREATE TABLE part (
  id TEXT PRIMARY KEY,
  message_id TEXT,
  session_id TEXT,
  data TEXT,
  time_created INTEGER
);
`)
	require.NoError(t, err)

	for _, session := range fixture.Sessions {
		_, err := db.Exec(
			`INSERT INTO session(id, directory, version) VALUES(?, ?, ?)`,
			session.ID,
			session.Directory,
			session.Version,
		)
		require.NoError(t, err)
	}

	for _, message := range fixture.Messages {
		data, err := json.Marshal(message.Data)
		require.NoError(t, err)

		_, err = db.Exec(
			`INSERT INTO message(id, session_id, data, time_created) VALUES(?, ?, ?, ?)`,
			message.ID,
			message.SessionID,
			string(data),
			message.CreatedAt,
		)
		require.NoError(t, err)
	}

	for _, part := range fixture.Parts {
		data, err := json.Marshal(part.Data)
		require.NoError(t, err)

		_, err = db.Exec(
			`INSERT INTO part(id, message_id, session_id, data, time_created) VALUES(?, ?, ?, ?, ?)`,
			part.ID,
			part.MessageID,
			part.SessionID,
			string(data),
			part.CreatedAt,
		)
		require.NoError(t, err)
	}
}
