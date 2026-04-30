//go:build !freebsd && !openbsd && !netbsd && !dragonfly

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

func TestQoderParse(t *testing.T) {
	ctx := context.Background()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(
		home,
		"Library",
		"Application Support",
		"Qoder",
		"SharedClientCache",
		"cache",
		"db",
	)
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "local.db")
	createQoderDB(t, dbPath)
	createQoderConversationHistory(t, home)

	got, err := ai.Qoder{
		After:             time.UnixMilli(1777301083000),
		FallbackUserAgent: "plugin/0.0.1",
	}.Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 5)

	authorsPath := "/Users/user/git/wakatime-cli/AUTHORS"

	assert.Equal(t, "Qoder 99947f30-f6f8-4323-a2c1-4970f5329d9c", got[0].Entity)
	assert.Equal(t, heartbeat.AppType, got[0].EntityType)
	assert.Equal(t, "99947f30-f6f8-4323-a2c1-4970f5329d9c", got[0].AISession)
	assert.EqualValues(t, 20159, got[0].AIInputTokens)
	assert.EqualValues(t, 129, got[0].AIOutputTokens)
	assert.Equal(t, 45, got[0].AIPromptLength)
	assert.Equal(t, "/Users/user/git/wakatime-cli", got[0].ProjectPathOverride)
	assert.Contains(t, got[0].UserAgent, "Qoder")

	assert.Equal(t, "Qoder 99947f30-f6f8-4323-a2c1-4970f5329d9c", got[1].Entity)
	assert.EqualValues(t, 19812, got[1].AIInputTokens)
	assert.EqualValues(t, 80, got[1].AIOutputTokens)
	assert.Equal(t, 52, got[1].AIPromptLength)

	assert.Equal(t, authorsPath, got[2].Entity)
	assert.Equal(t, heartbeat.FileType, got[2].EntityType)
	assert.Equal(t, heartbeat.PointerTo(false), got[2].IsWrite)
	assert.Nil(t, got[2].AILineChanges)
	assert.Zero(t, got[2].AIPromptLength)

	assert.EqualValues(t, 20088, got[3].AIInputTokens)
	assert.EqualValues(t, 125, got[3].AIOutputTokens)
	assert.Zero(t, got[3].AIPromptLength)

	assert.Equal(t, authorsPath, got[4].Entity)
	assert.Equal(t, heartbeat.PointerTo(true), got[4].IsWrite)
	require.NotNil(t, got[4].AILineChanges)
	assert.Equal(t, -1, *got[4].AILineChanges)
	assert.Zero(t, got[4].AIPromptLength)
}

func TestQoderParse_NoDB(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got, err := ai.Qoder{After: time.Now()}.Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestQoderParse_SkipsStaleLocalDB(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	dbDir := filepath.Join(
		home,
		"Library",
		"Application Support",
		"Qoder",
		"SharedClientCache",
		"cache",
		"db",
	)
	require.NoError(t, os.MkdirAll(dbDir, 0o755))

	dbPath := filepath.Join(dbDir, "local.db")
	require.NoError(t, os.WriteFile(dbPath, []byte("not sqlite"), 0o600))

	staleTime := time.UnixMilli(1777301083000)
	require.NoError(t, os.Chtimes(dbPath, staleTime, staleTime))

	got, err := ai.Qoder{
		After: staleTime.Add(time.Millisecond),
	}.Parse(context.Background())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func createQoderDB(t *testing.T, dbPath string) {
	t.Helper()

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	defer db.Close() // nolint:errcheck

	_, err = db.Exec(`
CREATE TABLE chat_session (
	session_id varchar(64) primary key,
	project_uri varchar(512)
);
CREATE TABLE chat_message (
	id varchar(64) primary key,
	session_id VARCHAR(64),
	request_id VARCHAR(64),
	role VARCHAR(64),
	tool_result text,
	token_info text,
	gmt_create INTEGER
);
`)
	require.NoError(t, err)

	sessionID := "99947f30-f6f8-4323-a2c1-4970f5329d9c"
	firstRequestID := "c50ecfc2-0546-4712-a1a5-e5fe5fd4e94b"
	secondRequestID := "8ab982b7-15ca-43bc-9fac-1a7d607b2830"
	projectPath := "/Users/user/git/wakatime-cli"
	authorsPath := projectPath + "/AUTHORS"

	_, err = db.Exec(
		"INSERT INTO chat_session (session_id, project_uri) VALUES (?, ?)",
		sessionID,
		projectPath,
	)
	require.NoError(t, err)

	insertQoderMessage(t, db, "user-1", sessionID, firstRequestID, "user", "", "", 1777301083679)
	insertQoderMessage(
		t,
		db,
		"old",
		sessionID,
		"old-request",
		"assistant",
		"",
		qoderTokenInfoJSON(t, 1, 2, 0),
		1777301000000,
	)
	insertQoderMessage(
		t,
		db,
		"assistant-0",
		sessionID,
		firstRequestID,
		"assistant",
		"",
		qoderTokenInfoJSON(t, 20159, 129, 17130),
		1777301097768,
	)
	insertQoderMessage(t, db, "user-2", sessionID, secondRequestID, "user", "", "", 1777301165086)
	insertQoderMessage(
		t,
		db,
		"assistant-1",
		sessionID,
		secondRequestID,
		"assistant",
		"",
		qoderTokenInfoJSON(t, 19812, 80, 16491),
		1777301173625,
	)
	insertQoderMessage(
		t,
		db,
		"read-1",
		sessionID,
		secondRequestID,
		"tool",
		qoderToolResultJSON(t, sessionID, secondRequestID, projectPath, "read_file", authorsPath, nil),
		"",
		1777301173644,
	)
	insertQoderMessage(
		t,
		db,
		"assistant-2",
		sessionID,
		secondRequestID,
		"assistant",
		"",
		qoderTokenInfoJSON(t, 20088, 125, 19806),
		1777301180520,
	)
	insertQoderMessage(
		t,
		db,
		"write-1",
		sessionID,
		secondRequestID,
		"tool",
		qoderToolResultJSON(t, sessionID, secondRequestID, projectPath, "search_replace", authorsPath, map[string]int{
			"add":    1,
			"delete": 2,
		}),
		"",
		1777301184268,
	)
}

func createQoderConversationHistory(t *testing.T, home string) {
	t.Helper()

	historyDir := filepath.Join(
		home,
		".qoder",
		"cache",
		"projects",
		"wakatime-cli-4faf562b",
		"conversation-history",
		"99947f30",
	)
	require.NoError(t, os.MkdirAll(historyDir, 0o755))

	history := strings.Join([]string{
		qoderConversationLineJSON(t, "user", qoderPromptText("Where do you store your chat transcript logs?")),
		qoderConversationLineJSON(t, "assistant", "I don't have access to chat transcript logs."),
		qoderConversationLineJSON(t, "user", qoderPromptText("Yes, add your name to the AUTHORS file in this repo.")),
		"",
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(historyDir, "99947f30.jsonl"), []byte(history), 0o600))
}

func insertQoderMessage(
	t *testing.T,
	db *sql.DB,
	id string,
	sessionID string,
	requestID string,
	role string,
	toolResult string,
	tokenInfo string,
	createdAt int64,
) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO chat_message
		 (id, session_id, request_id, role, tool_result, token_info, gmt_create)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id,
		sessionID,
		requestID,
		role,
		toolResult,
		tokenInfo,
		createdAt,
	)
	require.NoError(t, err)
}

func qoderTokenInfoJSON(t *testing.T, promptTokens int, completionTokens int, cachedTokens int) string {
	t.Helper()

	raw, err := json.Marshal(map[string]int{
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"cached_tokens":     cachedTokens,
		"max_input_tokens":  180000,
	})
	require.NoError(t, err)

	return string(raw)
}

func qoderToolResultJSON(
	t *testing.T,
	sessionID string,
	requestID string,
	projectPath string,
	toolCallName string,
	filePath string,
	diffInfo map[string]int,
) string {
	t.Helper()

	result := map[string]any{
		"sessionId":    sessionID,
		"requestId":    requestID,
		"projectPath":  projectPath,
		"toolCallName": toolCallName,
		"parameters": map[string]any{
			"file_path": filePath,
		},
		"results": []map[string]any{
			{
				"path": filePath,
			},
		},
	}

	if diffInfo != nil {
		result["results"] = []map[string]any{
			{
				"path":     filePath,
				"diffInfo": diffInfo,
			},
		}
	}

	raw, err := json.Marshal(result)
	require.NoError(t, err)

	return string(raw)
}

func qoderConversationLineJSON(t *testing.T, role string, text string) string {
	t.Helper()

	raw, err := json.Marshal(map[string]any{
		"role": role,
		"message": map[string]any{
			"content": []map[string]string{
				{
					"type": "text",
					"text": text,
				},
			},
		},
	})
	require.NoError(t, err)

	return string(raw)
}

func qoderPromptText(query string) string {
	return strings.Join([]string{
		"<system-reminder>",
		"[IMPORTANT] You must always respond in en-us.",
		"</system-reminder>",
		"",
		"",
		"",
		"<user_query>",
		query,
		"</user_query>",
	}, "\n")
}
