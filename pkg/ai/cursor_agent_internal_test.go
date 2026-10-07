package ai

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCursorAgentParsesPlaintextTranscript(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	transcriptDir := filepath.Join(home, ".cursor", "projects", "project", "agent-transcripts")
	require.NoError(t, os.MkdirAll(transcriptDir, 0o755))
	transcriptPath := filepath.Join(transcriptDir, "agent-1.txt")
	require.NoError(t, os.WriteFile(
		transcriptPath,
		[]byte("user:\n<user_query>Implement this parser</user_query>\nA:\nDone\n"),
		0o600,
	))

	got, err := (CursorAgent{After: time.Now().Add(-time.Minute)}).Parse(t.Context())
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Cursor Agent agent-1", got[0].Entity)
	assert.Equal(t, len([]rune("Implement this parser")), got[0].AIPromptLength)
}

func TestCursorAgentIgnoresTerminalSnapshots(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", home)

	root := filepath.Join(home, ".cursor", "projects", "demo")
	for _, name := range []string{"terminals/4.txt", "cache.txt", "cache/events.jsonl"} {
		path := filepath.Join(root, filepath.FromSlash(name))
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
		require.NoError(t, os.WriteFile(path, []byte("pid: 1234\nIdle shell prompt\n"), 0o600))
	}

	for scan := range 2 {
		stamp := time.Now().Add(time.Duration(scan) * time.Minute)
		path := filepath.Join(root, "terminals", "4.txt")
		require.NoError(t, os.Chtimes(path, stamp, stamp))
		got, err := (CursorAgent{}).Parse(t.Context())
		require.NoError(t, err)
		assert.Empty(t, got)
	}

	path := filepath.Join(root, "agent-transcripts", "session", "session.jsonl")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))

	data := `{"timestamp":"2026-09-23T12:00:00Z","role":"user","content":"Fix this","session_id":"real"}`
	require.NoError(t, os.WriteFile(path, []byte(data+"\n"), 0o600))
	got, err := (CursorAgent{}).Parse(t.Context())
	require.NoError(t, err)
	require.NotEmpty(t, got)
}

func TestCursorAgentSkipsIrrelevantSQLiteTables(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("WAKATIME_HOME", t.TempDir())

	dbPath := filepath.Join(home, ".cursor", "ai-tracking", "ai-code-tracking.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o755))

	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })

	_, err = db.Exec(`CREATE TABLE tracked_file_content (content TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO tracked_file_content VALUES (?)`, strings.Repeat("x", 2048))
	require.NoError(t, err)

	_, err = db.Exec(`CREATE TABLE ai_code_hashes (
		hash TEXT PRIMARY KEY,
		fileName TEXT,
		conversationId TEXT,
		timestamp INTEGER,
		model TEXT
	)`)
	require.NoError(t, err)

	filePath := filepath.Join(home, "project", "main.go")
	_, err = db.Exec(`INSERT INTO ai_code_hashes VALUES (?, ?, ?, ?, ?)`,
		"hash", filePath, "conversation", time.Now().UnixMilli(), "grok-4.5")
	require.NoError(t, err)

	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, &aiSQLiteBudget{
		bytes: aiSQLiteByteLimit - 1024,
	})
	got, err := (CursorAgent{}).Parse(ctx)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "Cursor Agent conversation", got[0].Entity)
	assert.Equal(t, "grok/4.5 Cursor Agent", got[0].UserAgent)
	assert.Equal(t, filepath.ToSlash(filePath), got[1].Entity)
}
