//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	// Register the pure-Go SQLite driver used to read Cursor state.vscdb files.
	_ "modernc.org/sqlite"
)

// Cursor contains params for detecting heartbeats from Cursor transcripts.
type Cursor struct {
	After             time.Time
	FallbackUserAgent string
	UserAgents        map[string]string
}

type (
	cursorToolFormerData struct {
		Status  string `json:"status"`
		Name    string `json:"name"`
		Params  string `json:"params"`
		RawArgs string `json:"rawArgs"`
	}

	cursorURI struct {
		FSPath string `json:"_fsPath"`
	}

	cursorCodeBlock struct {
		URI     *cursorURI `json:"uri"`
		Content string     `json:"content"`
	}

	cursorLogLine struct {
		BubbleID       string
		CreatedAt      time.Time             `json:"createdAt"`
		Type           int                   `json:"type"`
		Text           string                `json:"text"`
		ToolFormerData *cursorToolFormerData `json:"toolFormerData"`
		CodeBlocks     []cursorCodeBlock     `json:"codeBlocks"`
	}

	cursorLogRow struct {
		BubbleID string
		Value    string
	}

	cursorEditParams struct {
		RelativeWorkspacePath string `json:"relativeWorkspacePath"`
		StreamingContent      string `json:"streamingContent"`
	}

	cursorEditRawArgs struct {
		TargetFile string `json:"target_file"`
		CodeEdit   string `json:"code_edit"`
	}

	cursorReadParams struct {
		TargetFile            string `json:"targetFile"`
		EffectiveURI          string `json:"effectiveUri"`
		Path                  string `json:"path"`
		RelativeWorkspacePath string `json:"relativeWorkspacePath"`
	}

	cursorReadRawArgs struct {
		TargetFile string `json:"target_file"`
	}
)

// Parse parses the Cursor SQLite state db for ai heartbeats.
func (g Cursor) Parse(ctx context.Context) (Heartbeats, error) {
	dbPath, err := stateDBPath(ctx)
	if err != nil {
		return nil, err
	}

	if dbPath == "" {
		return Heartbeats{}, nil
	}

	if !cursorStateDBModifiedAfter(dbPath, g.After) {
		return Heartbeats{}, nil
	}

	rows, err := g.queryRows(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	bubbleCWDs := make(map[string]string)

	for _, row := range rows {
		var logLine cursorLogLine
		if err := json.Unmarshal([]byte(row.Value), &logLine); err != nil {
			continue
		}

		logLine.BubbleID = row.BubbleID

		if cwd := cursorProjectPath(logLine); cwd != "" && logLine.BubbleID != "" {
			bubbleCWDs[logLine.BubbleID] = cwd
		}

		if logLine.CreatedAt.IsZero() || logLine.CreatedAt.Before(g.After) {
			continue
		}

		parsed := g.cursorHeartbeats(ctx, logLine, bubbleCWDs[logLine.BubbleID])
		if len(parsed) == 0 {
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func cursorStateDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(after)
}

func stateDBPath(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Cursor", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Cursor", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Cursor", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "cursor", "User", "globalStorage", "state.vscdb"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", nil
}

func (g Cursor) queryRows(ctx context.Context, dbPath string) ([]cursorLogRow, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening cursor sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	rows, err := db.QueryContext(ctx, `
SELECT key, CAST(value AS TEXT)
FROM cursorDiskKV
WHERE key LIKE 'bubbleId:%'
  AND json_extract(CAST(value AS TEXT), '$.createdAt') >= ?
  AND (
    (
      json_extract(CAST(value AS TEXT), '$.toolFormerData.status') = 'completed'
      AND json_extract(
        CAST(value AS TEXT),
        '$.toolFormerData.name'
      ) IN ('edit_file', 'edit_file_v2', 'read_file', 'read_file_v2')
    )
    OR json_extract(CAST(value AS TEXT), '$.text') IS NOT NULL
  )
ORDER BY json_extract(CAST(value AS TEXT), '$.createdAt') ASC;
`, g.After.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return nil, fmt.Errorf("failed querying cursor sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	var results []cursorLogRow

	for rows.Next() {
		var (
			key string
			row string
		)
		if err := rows.Scan(&key, &row); err != nil {
			return nil, fmt.Errorf("failed scanning cursor sqlite row: %s", err)
		}

		row = strings.TrimSpace(row)
		if row != "" {
			bubbleID := strings.TrimPrefix(key, "bubbleId:")
			if idx := strings.Index(bubbleID, ":"); idx != -1 {
				bubbleID = bubbleID[:idx]
			}

			results = append(results, cursorLogRow{
				BubbleID: bubbleID,
				Value:    row,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading cursor sqlite rows: %s", err)
	}

	return results, nil
}

func (g Cursor) cursorHeartbeats(ctx context.Context, logLine cursorLogLine, cwd string) Heartbeats {
	var heartbeats Heartbeats

	if heartbeat := g.cursorAppHeartbeat(ctx, logLine, cwd); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
	}

	if logLine.Type != 2 || logLine.ToolFormerData == nil || logLine.ToolFormerData.Status != "completed" {
		return heartbeats
	}

	if heartbeat := g.cursorFileHeartbeat(ctx, logLine); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
	}

	return heartbeats
}

func (g Cursor) cursorAppHeartbeat(ctx context.Context, logLine cursorLogLine, cwd string) *heartbeat.Heartbeat {
	if strings.TrimSpace(logLine.Text) == "" {
		return nil
	}

	if logLine.Type != 1 && logLine.Type != 2 {
		return nil
	}

	entity := logLine.BubbleID
	if entity == "" {
		entity = "Cursor"
	}

	h := heartbeat.New(
		nil,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		heartbeat.AppType,
		nil,
		false,
		heartbeat.PointerTo(false),
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		cwd,
		float64(logLine.CreatedAt.Unix()),
		aiUserAgent(ctx, entity, g.UserAgents, g.FallbackUserAgent, cursorPlugin()),
	)

	return &h
}

func (g Cursor) cursorFileHeartbeat(ctx context.Context, logLine cursorLogLine) *heartbeat.Heartbeat {
	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params cursorEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err != nil {
			return nil
		}

		filePath := cursorFilePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		if filePath == "" {
			return nil
		}

		lineChanges := cursorLineChanges(params.StreamingContent)
		h := heartbeat.New(
			heartbeat.PointerTo(lineChanges),
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			filePath,
			heartbeat.FileType,
			nil,
			false,
			heartbeat.PointerTo(true),
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(ctx, filePath, g.UserAgents, g.FallbackUserAgent, cursorPlugin()),
		)

		return &h
	case "edit_file":
		var (
			params  cursorEditParams
			rawArgs cursorEditRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath := cursorFilePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		if filePath == "" {
			filePath = rawArgs.TargetFile
		}

		if filePath == "" {
			return nil
		}

		content := rawArgs.CodeEdit
		if content == "" && len(logLine.CodeBlocks) > 0 {
			content = logLine.CodeBlocks[0].Content
		}

		lineChanges := cursorLineChanges(content)
		h := heartbeat.New(
			heartbeat.PointerTo(lineChanges),
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			filePath,
			heartbeat.FileType,
			nil,
			false,
			heartbeat.PointerTo(true),
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(ctx, filePath, g.UserAgents, g.FallbackUserAgent, cursorPlugin()),
		)

		return &h
	case "read_file_v2", "read_file":
		var (
			params  cursorReadParams
			rawArgs cursorReadRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath := params.TargetFile
		if filePath == "" {
			filePath = params.EffectiveURI
		}

		if filePath == "" {
			filePath = params.Path
		}

		if filePath == "" {
			filePath = params.RelativeWorkspacePath
		}

		if filePath == "" {
			filePath = rawArgs.TargetFile
		}

		if filePath == "" {
			filePath = cursorFilePath("", logLine.CodeBlocks)
		}

		if filePath == "" {
			return nil
		}

		h := heartbeat.New(
			heartbeat.PointerTo(0),
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			filePath,
			heartbeat.FileType,
			nil,
			false,
			heartbeat.PointerTo(false),
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(ctx, filePath, g.UserAgents, g.FallbackUserAgent, cursorPlugin()),
		)

		return &h
	default:
		return nil
	}
}

func cursorFilePath(path string, codeBlocks []cursorCodeBlock) string {
	if path != "" {
		return path
	}

	if len(codeBlocks) == 0 || codeBlocks[0].URI == nil {
		return ""
	}

	return codeBlocks[0].URI.FSPath
}

func cursorProjectPath(logLine cursorLogLine) string {
	if logLine.ToolFormerData == nil {
		return ""
	}

	var filePath string

	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params cursorEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err == nil {
			filePath = cursorFilePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		}
	case "edit_file":
		var (
			params  cursorEditParams
			rawArgs cursorEditRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath = cursorFilePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		if filePath == "" {
			filePath = rawArgs.TargetFile
		}
	case "read_file_v2", "read_file":
		var (
			params  cursorReadParams
			rawArgs cursorReadRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath = params.TargetFile
		if filePath == "" {
			filePath = params.EffectiveURI
		}

		if filePath == "" {
			filePath = params.Path
		}

		if filePath == "" {
			filePath = params.RelativeWorkspacePath
		}

		if filePath == "" {
			filePath = rawArgs.TargetFile
		}

		if filePath == "" {
			filePath = cursorFilePath("", logLine.CodeBlocks)
		}
	}

	if filePath == "" {
		return ""
	}

	return filepath.Dir(filePath)
}

func cursorLineChanges(content string) int {
	if strings.TrimSpace(content) == "" {
		return 0
	}

	if cursorLooksLikeUnifiedDiff(content) {
		return cursorLineChangesFromDiff(content)
	}

	lineChanges := 0

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			lineChanges++
		}
	}

	return lineChanges
}

func cursorLooksLikeUnifiedDiff(content string) bool {
	trimmed := strings.TrimLeft(content, " \t\r\n")
	if strings.HasPrefix(trimmed, "--- ") || strings.HasPrefix(trimmed, "+++ ") || strings.HasPrefix(trimmed, "@@") {
		return true
	}

	return strings.Contains(content, "\n--- ") ||
		strings.Contains(content, "\n+++ ") ||
		strings.Contains(content, "\n@@")
}

func cursorLineChangesFromDiff(diff string) int {
	lineChanges := 0

	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		}

		if strings.HasPrefix(line, "+") {
			lineChanges++
		} else if strings.HasPrefix(line, "-") {
			lineChanges--
		}
	}

	return lineChanges
}

func cursorPlugin() string {
	return "Cursor"
}

// ID returns its id.
func (Cursor) ID() ParserID {
	return CursorParser
}
