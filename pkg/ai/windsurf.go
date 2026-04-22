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
)

// Windsurf contains params for detecting heartbeats from Windsurf transcripts.
type Windsurf ParserConfig

type (
	windsurfTokenCount struct {
		InputTokens  *int `json:"inputTokens"`
		OutputTokens *int `json:"outputTokens"`
		TotalTokens  *int `json:"totalTokens"`
	}

	windsurfUsage struct {
		InputTokens  *int `json:"input_tokens"`
		OutputTokens *int `json:"output_tokens"`
		TotalTokens  *int `json:"total_tokens"`
	}

	windsurfToolFormerData struct {
		Status  string `json:"status"`
		Name    string `json:"name"`
		Params  string `json:"params"`
		RawArgs string `json:"rawArgs"`
	}

	windsurfURI struct {
		FSPath string `json:"_fsPath"`
	}

	windsurfCodeBlock struct {
		URI     *windsurfURI `json:"uri"`
		Content string       `json:"content"`
	}

	windsurfLogLine struct {
		BubbleID       string
		CreatedAt      time.Time               `json:"createdAt"`
		Type           int                     `json:"type"`
		Text           string                  `json:"text"`
		TokenCount     *windsurfTokenCount     `json:"tokenCount"`
		Usage          *windsurfUsage          `json:"usage"`
		ToolFormerData *windsurfToolFormerData `json:"toolFormerData"`
		CodeBlocks     []windsurfCodeBlock     `json:"codeBlocks"`
	}

	windsurfLogRow struct {
		BubbleID string
		Value    string
	}

	windsurfEditParams struct {
		RelativeWorkspacePath string `json:"relativeWorkspacePath"`
		StreamingContent      string `json:"streamingContent"`
	}

	windsurfEditRawArgs struct {
		TargetFile string `json:"target_file"`
		CodeEdit   string `json:"code_edit"`
	}

	windsurfReadParams struct {
		TargetFile            string `json:"targetFile"`
		EffectiveURI          string `json:"effectiveUri"`
		Path                  string `json:"path"`
		RelativeWorkspacePath string `json:"relativeWorkspacePath"`
	}

	windsurfReadRawArgs struct {
		TargetFile string `json:"target_file"`
	}
)

// Parse parses the Windsurf SQLite state db for ai heartbeats.
func (g Windsurf) Parse(ctx context.Context) (Heartbeats, error) {
	dbPath, err := g.stateDBPath(ctx)
	if err != nil {
		return nil, err
	}

	if dbPath == "" {
		return Heartbeats{}, nil
	}

	if !g.stateDBModifiedAfter(dbPath, g.After) {
		return Heartbeats{}, nil
	}

	rows, err := g.queryRows(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	bubbleCWDs := make(map[string]string)
	bubbleTokens := make(map[string]heartbeat.AITokens)

	for _, row := range rows {
		var logLine windsurfLogLine
		if err := json.Unmarshal([]byte(row.Value), &logLine); err != nil {
			continue
		}

		logLine.BubbleID = row.BubbleID

		if cwd := g.projectPath(logLine); cwd != "" && logLine.BubbleID != "" {
			bubbleCWDs[logLine.BubbleID] = cwd
		}

		tokens := g.windsurfTokenCounts(logLine, bubbleTokens[logLine.BubbleID])

		if logLine.CreatedAt.IsZero() || logLine.CreatedAt.Before(g.After) {
			bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)
			continue
		}

		parsed := g.windsurfHeartbeats(logLine, bubbleCWDs[logLine.BubbleID], tokens)
		if len(parsed) == 0 {
			bubbleTokens[logLine.BubbleID] = tokens
			continue
		}

		bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (Windsurf) windsurfTokenCounts(line windsurfLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
	if line.TokenCount != nil {
		current := previous
		if line.TokenCount.InputTokens != nil {
			current.CurrentInput = previous.LastInput + int64(*line.TokenCount.InputTokens)
		}

		switch {
		case line.TokenCount.OutputTokens != nil:
			current.CurrentOutput = previous.LastOutput + int64(*line.TokenCount.OutputTokens)
		case line.TokenCount.TotalTokens != nil:
			current.CurrentOutput = previous.LastOutput + int64(*line.TokenCount.TotalTokens)
		}

		return current
	}

	if line.Usage == nil {
		return previous
	}

	current := previous
	if line.Usage.InputTokens != nil {
		current.CurrentInput = int64(*line.Usage.InputTokens)
	}

	switch {
	case line.Usage.OutputTokens != nil:
		current.CurrentOutput = int64(*line.Usage.OutputTokens)
	case line.Usage.TotalTokens != nil:
		current.CurrentOutput = int64(*line.Usage.TotalTokens)
	}

	return current
}

func (Windsurf) stateDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(after)
}

func (Windsurf) stateDBPath(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Windsurf", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "Library", "Application Support", "Windsurf - Next", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Windsurf", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Windsurf - Next", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Windsurf", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "windsurf", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Windsurf - Next", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "windsurf-next", "User", "globalStorage", "state.vscdb"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", nil
}

func (Windsurf) queryRows(ctx context.Context, dbPath string) ([]windsurfLogRow, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening windsurf sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	rows, err := db.QueryContext(ctx, `
SELECT key, CAST(value AS TEXT)
FROM cursorDiskKV
WHERE key LIKE 'bubbleId:%'
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
`)
	if err != nil {
		return nil, fmt.Errorf("failed querying windsurf sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	var results []windsurfLogRow

	for rows.Next() {
		var (
			key string
			row string
		)
		if err := rows.Scan(&key, &row); err != nil {
			return nil, fmt.Errorf("failed scanning windsurf sqlite row: %s", err)
		}

		row = strings.TrimSpace(row)
		if row != "" {
			bubbleID := strings.TrimPrefix(key, "bubbleId:")
			if idx := strings.Index(bubbleID, ":"); idx != -1 {
				bubbleID = bubbleID[:idx]
			}

			results = append(results, windsurfLogRow{
				BubbleID: bubbleID,
				Value:    row,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading windsurf sqlite rows: %s", err)
	}

	return results, nil
}

func (g Windsurf) windsurfHeartbeats(logLine windsurfLogLine, cwd string, tokens heartbeat.AITokens) Heartbeats {
	var heartbeats Heartbeats

	assignTokens := g.hasTokenDelta(tokens)

	appTokens := g.tokensForFirstHeartbeat(assignTokens, tokens)
	if heartbeat := g.windsurfAppHeartbeat(
		logLine,
		cwd,
		logLine.BubbleID,
		appTokens,
	); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
		assignTokens = false
	}

	if logLine.Type != 2 || logLine.ToolFormerData == nil || logLine.ToolFormerData.Status != "completed" {
		return heartbeats
	}

	fileTokens := g.tokensForFirstHeartbeat(assignTokens, tokens)
	if heartbeat := g.windsurfFileHeartbeat(
		logLine,
		logLine.BubbleID,
		fileTokens,
	); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
	}

	return heartbeats
}

func (g Windsurf) windsurfAppHeartbeat(
	logLine windsurfLogLine,
	cwd string,
	sessionID string,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	text := strings.TrimSpace(logLine.Text)
	if text == "" {
		return nil
	}

	if logLine.Type != 1 && logLine.Type != 2 {
		return nil
	}

	entity := appHeartbeatEntity("Windsurf", logLine.BubbleID)

	h := g.newHeartbeat(
		nil,
		sessionID,
		tokens,
		entity,
		heartbeat.AppType,
		heartbeat.PointerTo(false),
		cwd,
		float64(logLine.CreatedAt.Unix()),
		aiUserAgent(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
	)
	if logLine.Type == 1 {
		h.AIPromptLength = promptLength(logLine.Text)
	}

	return &h
}

func (g Windsurf) windsurfFileHeartbeat(
	logLine windsurfLogLine,
	sessionID string,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params windsurfEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err != nil {
			return nil
		}

		filePath := g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		if filePath == "" {
			return nil
		}

		lineChanges := g.lineChanges(params.StreamingContent)
		h := g.newHeartbeat(
			heartbeat.PointerTo(lineChanges),
			sessionID,
			tokens,
			filePath,
			heartbeat.FileType,
			heartbeat.PointerTo(true),
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(filePath, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
		)

		return &h
	case "edit_file":
		var (
			params  windsurfEditParams
			rawArgs windsurfEditRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath := g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
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

		lineChanges := g.lineChanges(content)
		h := g.newHeartbeat(
			heartbeat.PointerTo(lineChanges),
			sessionID,
			tokens,
			filePath,
			heartbeat.FileType,
			heartbeat.PointerTo(true),
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(filePath, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
		)

		return &h
	case "read_file_v2", "read_file":
		var (
			params  windsurfReadParams
			rawArgs windsurfReadRawArgs
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
			filePath = g.filePath("", logLine.CodeBlocks)
		}

		if filePath == "" {
			return nil
		}

		h := g.newHeartbeat(
			heartbeat.PointerTo(0),
			sessionID,
			tokens,
			filePath,
			heartbeat.FileType,
			heartbeat.PointerTo(false),
			"",
			float64(logLine.CreatedAt.Unix()),
			aiUserAgent(filePath, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
		)

		return &h
	default:
		return nil
	}
}

func (Windsurf) tokenDelta(tokens heartbeat.AITokens) (int64, int64) {
	input := tokens.CurrentInput - tokens.LastInput
	if input < 0 {
		input = 0
	}

	output := tokens.CurrentOutput - tokens.LastOutput
	if output < 0 {
		output = 0
	}

	return input, output
}

func (g Windsurf) hasTokenDelta(tokens heartbeat.AITokens) bool {
	input, output := g.tokenDelta(tokens)
	return input > 0 || output > 0
}

func (Windsurf) advanceTokens(tokens heartbeat.AITokens) heartbeat.AITokens {
	tokens.LastInput = tokens.CurrentInput
	tokens.LastOutput = tokens.CurrentOutput

	return tokens
}

func (Windsurf) tokensForFirstHeartbeat(assign bool, tokens heartbeat.AITokens) *heartbeat.AITokens {
	if !assign {
		return nil
	}

	copy := tokens

	return &copy
}

func (Windsurf) newHeartbeat(
	aiLineChanges *int,
	aiSession string,
	aiTokens *heartbeat.AITokens,
	entity string,
	entityType heartbeat.EntityType,
	isWrite *bool,
	projectPathOverride string,
	timestamp float64,
	userAgent string,
) heartbeat.Heartbeat {
	if aiTokens != nil {
		return heartbeat.NewWithAITokens(
			aiLineChanges,
			aiSession,
			*aiTokens,
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			entity,
			entityType,
			nil,
			false,
			isWrite,
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			projectPathOverride,
			timestamp,
			userAgent,
		)
	}

	h := heartbeat.New(
		aiLineChanges,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		entityType,
		nil,
		false,
		isWrite,
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		projectPathOverride,
		timestamp,
		userAgent,
	)
	h.AISession = aiSession

	return h
}

func (Windsurf) filePath(path string, codeBlocks []windsurfCodeBlock) string {
	if path != "" {
		return path
	}

	if len(codeBlocks) == 0 || codeBlocks[0].URI == nil {
		return ""
	}

	return codeBlocks[0].URI.FSPath
}

func (g Windsurf) projectPath(logLine windsurfLogLine) string {
	if logLine.ToolFormerData == nil {
		return ""
	}

	var filePath string

	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params windsurfEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err == nil {
			filePath = g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		}
	case "edit_file":
		var (
			params  windsurfEditParams
			rawArgs windsurfEditRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath = g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		if filePath == "" {
			filePath = rawArgs.TargetFile
		}
	case "read_file_v2", "read_file":
		var (
			params  windsurfReadParams
			rawArgs windsurfReadRawArgs
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
			filePath = g.filePath("", logLine.CodeBlocks)
		}
	}

	if filePath == "" {
		return ""
	}

	return filepath.Dir(filePath)
}

func (g Windsurf) lineChanges(content string) int {
	if strings.TrimSpace(content) == "" {
		return 0
	}

	if g.looksLikeUnifiedDiff(content) {
		return g.lineChangesFromDiff(content)
	}

	lineChanges := 0

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			lineChanges++
		}
	}

	return lineChanges
}

func (Windsurf) looksLikeUnifiedDiff(content string) bool {
	trimmed := strings.TrimLeft(content, " \t\r\n")
	if strings.HasPrefix(trimmed, "--- ") || strings.HasPrefix(trimmed, "+++ ") || strings.HasPrefix(trimmed, "@@") {
		return true
	}

	return strings.Contains(content, "\n--- ") ||
		strings.Contains(content, "\n+++ ") ||
		strings.Contains(content, "\n@@")
}

func (Windsurf) lineChangesFromDiff(diff string) int {
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

// Name returns its name.
func (Windsurf) Name() string {
	return "Windsurf"
}
