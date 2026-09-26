package ai

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"time"
)

// ZCode contains parameters for detecting heartbeats from ZCode logs.
type ZCode ParserConfig

// Parse parses ZCode logs for AI heartbeats.
func (g ZCode) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	dbPath := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")

	return parseGenericAIProvider(ctx, g.sqliteProvider(dbPath))
}

func (g ZCode) sqliteProvider(dbPath string) genericAIProvider {
	return genericAIProvider{
		parser: g, config: ParserConfig(g), sqliteRoots: []string{dbPath},
		// Messages own request usage; turn/model/tool telemetry overlaps it.
		// Parts own prompt text and file edits. Session titles are not prompts.
		sqliteTables:        []string{"message", "part"},
		sqliteQuery:         zCodeSQLiteQuery,
		sqliteEvents:        zCodeSQLiteEvents,
		sqliteParserVersion: 1,
	}
}

// Correlated primary-key lookups retain one result per source row and its rowid,
// allowing the shared scanner to checkpoint, resume, and detect in-place updates.
// In particular, parts are not filtered by their parent message's creation time:
// a tool may finish after that message has fallen outside the cutoff.
func zCodeSQLiteQuery(table, projection string, after time.Time) (string, []any) {
	if table == "part" {
		projection = strings.ReplaceAll(projection, "*", `id, session_id, time_created, time_updated,
 `+zCodeSQLitePartData+` AS data`)

		return "SELECT " + projection + `,
 (SELECT ` + zCodeSQLiteParentData + ` FROM message WHERE id = part.message_id) AS zcode_message,
 (SELECT directory FROM session WHERE id = part.session_id) AS zcode_directory
 FROM part
 WHERE (
   COALESCE(NULLIF(time_updated, 0), time_created) >= ?
   OR EXISTS (
     SELECT 1 FROM message WHERE id = part.message_id
     AND CASE WHEN json_valid(data) THEN json_extract(data, '$.role') = 'user' END
   )
 )`, []any{after.UnixMilli()}
	}

	projection = strings.ReplaceAll(projection, "*", `id, session_id, time_created, time_updated,
 `+zCodeSQLiteMessageData+` AS data`)

	return "SELECT " + projection + `,
 (SELECT directory FROM session WHERE id = message.session_id) AS zcode_directory
 FROM message WHERE 1=1`, nil
}

// Project only fields used for accounting. In particular, never transfer or
// hash assistant response text, reasoning parts, or tool output.
// Keep historical request counters and user prompts as deduplication baselines;
// old assistant parts can be rejected using scalar timestamps before SQLite
// reads their potentially very large data column.
const zCodeSQLiteMessageFields = `
 'role', json_extract(data, '$.role'),
 'modelID', COALESCE(json_extract(data, '$.modelID'), json_extract(data, '$.modelId')),
 'path', json_extract(data, '$.path')`

const zCodeSQLiteParentData = `CASE WHEN json_valid(data) THEN json_object(` + zCodeSQLiteMessageFields + `) END`

const zCodeSQLiteMessageData = `CASE WHEN json_valid(data) THEN json_object(` + zCodeSQLiteMessageFields + `,
 'tokens', json_extract(data, '$.tokens')
) END`

const zCodeSQLitePartData = `CASE WHEN json_valid(data) THEN
 CASE json_extract(data, '$.type')
 WHEN 'text' THEN CASE WHEN EXISTS (
   SELECT 1 FROM message WHERE id = part.message_id
   AND CASE WHEN json_valid(data) THEN json_extract(data, '$.role') = 'user' END
 ) THEN json_object(
   'type', 'text', 'text', json_extract(data, '$.text'),
   'ignored', data -> '$.ignored', 'synthetic', data -> '$.synthetic'
 ) END
 WHEN 'tool' THEN CASE WHEN json_extract(data, '$.state.status') = 'completed' THEN json_object(
   'type', 'tool', 'tool', json_extract(data, '$.tool'),
   'state', json_object(
     'status', 'completed', 'input', json_extract(data, '$.state.input'),
     'metadata', json_extract(data, '$.state.metadata')
   )
 ) END
 END
END`

type zCodeMessage struct {
	Role    string `json:"role"`
	ModelID string `json:"modelID"`
	Path    struct {
		Cwd  string `json:"cwd"`
		Root string `json:"root"`
	} `json:"path"`
	Tokens *struct {
		Input  int64 `json:"input"`
		Output int64 `json:"output"`
		Cache  struct {
			Read int64 `json:"read"`
		} `json:"cache"`
	} `json:"tokens"`
}

type zCodePart struct {
	Type      string `json:"type"`
	Text      string `json:"text"`
	Ignored   bool   `json:"ignored"`
	Synthetic bool   `json:"synthetic"`
	State     *struct {
		Status string `json:"status"`
	} `json:"state"`
}

func zCodeSQLiteEvents(table string, row map[string]any) []genericAIEvent {
	data := row["data"]
	if table == "part" {
		data = row["zcode_message"]
	}

	var message zCodeMessage
	if !zCodeDecodeJSON(data, &message) {
		return nil
	}

	event := genericAIEvent{
		timestamp: genericAITime(row["time_updated"]),
		sessionID: genericAIString(row["session_id"]),
		model:     message.ModelID,
		cwd:       firstNonEmptyString(message.Path.Cwd, message.Path.Root, genericAIString(row["zcode_directory"])),
	}
	if event.timestamp.IsZero() {
		event.timestamp = genericAITime(row["time_created"])
	}

	if table == "message" {
		if message.Role != "assistant" || message.Tokens == nil {
			return nil
		}

		// These are per-request totals, unlike Codex's session counters. Input
		// includes cache reads/writes, and output already includes reasoning,
		// just as in Codex. Do not add those subsets a second time.
		// https://github.com/xiufengsun/TokenTracker/issues/554
		event.recordID = "message:" + genericAIString(row["id"])
		event.cachedInput = max(message.Tokens.Cache.Read, 0)
		event.input = max(message.Tokens.Input-event.cachedInput, 0)
		event.output = max(message.Tokens.Output, 0)
		event.tokensFound = true

		return []genericAIEvent{event}
	}

	var part zCodePart
	if !zCodeDecodeJSON(row["data"], &part) {
		return nil
	}

	if message.Role == "user" && part.Type == "text" && !part.Ignored && !part.Synthetic {
		event.recordID = "part:" + genericAIString(row["id"])
		event.prompt = strings.TrimSpace(part.Text)

		return []genericAIEvent{event}
	}

	if message.Role != "assistant" || part.Type != "tool" || part.State == nil || part.State.Status != "completed" {
		return nil
	}

	// Keep the existing generic file-change extraction and gross line counts.
	tool := genericAIEventFromValue(row)
	event.toolName = tool.toolName
	event.filePath = tool.filePath
	event.isWrite = tool.isWrite
	event.lineChanges = tool.lineChanges

	return []genericAIEvent{event}
}

func zCodeDecodeJSON(value any, target any) bool {
	switch data := value.(type) {
	case string:
		return json.Unmarshal([]byte(data), target) == nil
	case []byte:
		return json.Unmarshal(data, target) == nil
	default:
		return false
	}
}

// Name returns the ZCode parser name.
func (ZCode) Name() string { return "ZCode" }
