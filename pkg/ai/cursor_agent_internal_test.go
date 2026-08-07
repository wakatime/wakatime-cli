package ai

import (
	"os"
	"path/filepath"
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
