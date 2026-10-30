package ai

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Pi writes `message.content` as a plain string for non-conversational roles
// (system, custom) and as an array of content blocks otherwise. The parser
// declared it as a slice only, so every string-content line failed to unmarshal
// and was skipped with a warning.
//
// Reproduces: json: cannot unmarshal string into Go struct field
// piMessage.message.content of type []ai.piContentBlock.
func TestPiLogLineUnmarshalStringContent(t *testing.T) {
	line := []byte(`{"type":"message","id":"52d6b381","timestamp":"2026-10-29T15:39:01.525Z",` +
		`"message":{"role":"system","content":"","sections":{"preamble":"You are an expert"}}}`)

	var logLine piLogLine

	require.NoError(t, json.Unmarshal(line, &logLine))
	require.NotNil(t, logLine.Message)
	assert.Equal(t, "system", logLine.Message.Role)
	assert.Empty(t, logLine.Message.Content)
}

// A non-empty string must be preserved, not dropped: user/custom messages can
// carry prose directly in the string form.
func TestPiLogLineUnmarshalStringContentPreservesText(t *testing.T) {
	line := []byte(`{"type":"message","timestamp":"2026-10-29T15:39:01.525Z",` +
		`"message":{"role":"custom","content":"alpha\nbeta\n"}}`)

	var logLine piLogLine

	require.NoError(t, json.Unmarshal(line, &logLine))
	require.NotNil(t, logLine.Message)
	require.Len(t, logLine.Message.Content, 1)
	assert.Equal(t, "text", logLine.Message.Content[0].Type)
	assert.Equal(t, "alpha\nbeta\n", logLine.Message.Content[0].Text)
}

// The array form is what most lines use and must keep working unchanged.
func TestPiLogLineUnmarshalArrayContent(t *testing.T) {
	line := []byte(`{"type":"message","timestamp":"2026-10-29T15:39:01.525Z",` +
		`"message":{"role":"assistant","content":[{"type":"text","text":"hi"},` +
		`{"type":"toolCall","id":"tool-1","name":"read","arguments":{"path":"a.go"}}]}}`)

	var logLine piLogLine

	require.NoError(t, json.Unmarshal(line, &logLine))
	require.NotNil(t, logLine.Message)
	require.Len(t, logLine.Message.Content, 2)
	assert.Equal(t, "hi", logLine.Message.Content[0].Text)
	assert.Equal(t, "tool-1", logLine.Message.Content[1].ID)
	assert.Equal(t, "read", logLine.Message.Content[1].Name)
}

// null content appears on bashExecution lines; it must stay a no-op.
func TestPiLogLineUnmarshalNullContent(t *testing.T) {
	line := []byte(`{"type":"message","timestamp":"2026-10-29T15:39:01.525Z",` +
		`"message":{"role":"bashExecution","content":null}}`)

	var logLine piLogLine

	require.NoError(t, json.Unmarshal(line, &logLine))
	require.NotNil(t, logLine.Message)
	assert.Empty(t, logLine.Message.Content)
}

// A reused value must not keep blocks from a previous unmarshal when the new
// content is null or an empty string.
func TestPiMessageContentClearsStaleBlocks(t *testing.T) {
	c := piMessageContent{{Type: "text", Text: "stale"}}

	require.NoError(t, json.Unmarshal([]byte(`null`), &c))
	assert.Empty(t, c)

	c = piMessageContent{{Type: "text", Text: "stale"}}

	require.NoError(t, json.Unmarshal([]byte(`""`), &c))
	assert.Empty(t, c)
}

// A malformed array element must surface the field that actually failed, not a
// misleading "cannot unmarshal array into string".
func TestPiMessageContentReportsInnerError(t *testing.T) {
	line := []byte(`{"type":"message","timestamp":"2026-10-29T15:39:01.525Z",` +
		`"message":{"role":"assistant","content":[{"type":"toolCall","id":"t1",` +
		`"name":"read","arguments":"not-a-map"}]}}`)

	var logLine piLogLine

	err := json.Unmarshal(line, &logLine)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "arguments")
	assert.NotContains(t, err.Error(), "of type string")
}

// Non-string, non-array content must still be rejected.
func TestPiMessageContentRejectsOtherTypes(t *testing.T) {
	var c piMessageContent

	require.Error(t, json.Unmarshal([]byte(`123`), &c))
	require.Error(t, json.Unmarshal([]byte(`{}`), &c))
}
