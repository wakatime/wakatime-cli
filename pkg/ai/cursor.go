//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"

	// Register the pure-Go SQLite driver used to read Cursor state.vscdb files.
	_ "modernc.org/sqlite"
)

// Cursor contains params for detecting heartbeats from Cursor transcripts.
type Cursor ParserConfig

const cursorRecentBubbleRowLimit = 5000

type (
	cursorTokenCount struct {
		InputTokens  *int `json:"inputTokens"`
		OutputTokens *int `json:"outputTokens"`
		TotalTokens  *int `json:"totalTokens"`
	}

	cursorUsage struct {
		InputTokens  *int `json:"input_tokens"`
		OutputTokens *int `json:"output_tokens"`
		TotalTokens  *int `json:"total_tokens"`
	}

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
		TokenCount     *cursorTokenCount     `json:"tokenCount"`
		Usage          *cursorUsage          `json:"usage"`
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
	logger := log.Extract(ctx)

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

	db, err := g.openDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	subscriptionPlan, err := g.querySubscriptionPlan(ctx, db, dbPath)
	if err != nil {
		logger.Debugf("failed reading cursor subscription plan from %q: %s", dbPath, err)

		subscriptionPlan = ""
	}

	rows, err := g.queryRows(ctx, db, dbPath)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	bubbleCWDs := make(map[string]string)
	bubbleModels := make(map[string]string)
	bubbleTokens := make(map[string]heartbeat.AITokens)

	for _, row := range rows {
		var logLine cursorLogLine
		if err := json.Unmarshal([]byte(row.Value), &logLine); err != nil {
			continue
		}

		logLine.BubbleID = row.BubbleID

		if model := g.modelName([]byte(row.Value)); model != "" && logLine.BubbleID != "" {
			bubbleModels[logLine.BubbleID] = model
		}

		if cwd := g.projectPath(logLine); cwd != "" && logLine.BubbleID != "" {
			bubbleCWDs[logLine.BubbleID] = cwd
		}

		tokens := g.cursorTokenCounts(logLine, bubbleTokens[logLine.BubbleID])

		if logLine.CreatedAt.IsZero() || logLine.CreatedAt.Before(g.After) {
			bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)
			continue
		}

		parsed := g.cursorHeartbeats(logLine, bubbleCWDs[logLine.BubbleID], bubbleModels[logLine.BubbleID], tokens)
		if len(parsed) == 0 {
			bubbleTokens[logLine.BubbleID] = tokens
			continue
		}

		bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)

		heartbeats = append(heartbeats, parsed...)
	}

	return cursorApplySubscriptionPlan(heartbeats, subscriptionPlan), nil
}

func (Cursor) cursorTokenCounts(line cursorLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
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

func (Cursor) stateDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(after)
}

func (Cursor) stateDBPath(ctx context.Context) (string, error) {
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

func (Cursor) openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening cursor sqlite db %q: %s", dbPath, err)
	}

	// limit to a single connection, so the pure-Go sqlite driver only
	// allocates memory for one connection
	db.SetMaxOpenConns(1)

	return db, nil
}

func (Cursor) queryRows(ctx context.Context, db *sql.DB, dbPath string) ([]cursorLogRow, error) {
	rows, err := db.QueryContext(ctx, `
WITH recentCursorRows AS (
	SELECT key, CAST(value AS TEXT) AS value
	FROM cursorDiskKV
	WHERE key LIKE 'bubbleId:%'
	ORDER BY rowid DESC
	LIMIT ?
)
SELECT key, value
FROM recentCursorRows
WHERE json_valid(value)
  AND (
    (
      json_extract(value, '$.toolFormerData.status') = 'completed'
      AND json_extract(
        value,
        '$.toolFormerData.name'
      ) IN ('edit_file', 'edit_file_v2', 'read_file', 'read_file_v2')
    )
    OR json_extract(value, '$.text') IS NOT NULL
    OR json_extract(value, '$.model') IS NOT NULL
    OR json_extract(value, '$.modelName') IS NOT NULL
    OR json_extract(value, '$.modelId') IS NOT NULL
    OR json_extract(value, '$.modelID') IS NOT NULL
    OR json_extract(value, '$.model_id') IS NOT NULL
    OR json_extract(value, '$.modelSlug') IS NOT NULL
    OR json_extract(value, '$.modelDetails') IS NOT NULL
    OR json_extract(value, '$.modelConfig') IS NOT NULL
    OR json_extract(value, '$.selectedModel') IS NOT NULL
    OR json_extract(value, '$.selectedChatModel') IS NOT NULL
  )
ORDER BY json_extract(value, '$.createdAt') ASC;
`, cursorRecentBubbleRowLimit)
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

func (Cursor) querySubscriptionPlan(ctx context.Context, db *sql.DB, dbPath string) (string, error) {
	var tableName string

	err := db.QueryRowContext(ctx, `
SELECT name
FROM sqlite_master
WHERE type = 'table' AND name = 'ItemTable';
`).Scan(&tableName)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed checking cursor sqlite metadata table %q: %s", dbPath, err)
	}

	var plan string

	err = db.QueryRowContext(ctx, `
SELECT CAST(value AS TEXT)
FROM ItemTable
WHERE key = 'cursorAuth/stripeMembershipType'
LIMIT 1;
`).Scan(&plan)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed querying cursor subscription plan from sqlite db %q: %s", dbPath, err)
	}

	plan = strings.TrimSpace(strings.Trim(plan, `"`))
	if plan == "" {
		return "", nil
	}

	return plan, nil
}

func cursorApplySubscriptionPlan(heartbeats Heartbeats, plan string) Heartbeats {
	if plan == "" {
		return heartbeats
	}

	for i := range heartbeats {
		heartbeats[i].AISubscriptionPlan = plan
	}

	return heartbeats
}

func (g Cursor) cursorHeartbeats(logLine cursorLogLine, cwd string, model string, tokens heartbeat.AITokens) Heartbeats {
	var heartbeats Heartbeats

	assignTokens := g.hasTokenDelta(tokens)

	appTokens := g.tokensForFirstHeartbeat(assignTokens, tokens)
	if heartbeat := g.cursorAppHeartbeat(
		logLine,
		cwd,
		model,
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
	if heartbeat := g.cursorFileHeartbeat(
		logLine,
		model,
		logLine.BubbleID,
		fileTokens,
	); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
	}

	return heartbeats
}

func (g Cursor) cursorAppHeartbeat(
	logLine cursorLogLine,
	cwd string,
	model string,
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

	entity := appHeartbeatEntity("Cursor", logLine.BubbleID)

	h := g.newHeartbeat(
		nil,
		sessionID,
		tokens,
		entity,
		heartbeat.AppType,
		heartbeat.PointerTo(false),
		cwd,
		float64(logLine.CreatedAt.Unix()),
		g.userAgent(entity, model),
	)
	if logLine.Type == 1 {
		h.AIPromptLength = promptLength(logLine.Text)
	}

	return &h
}

func (g Cursor) cursorFileHeartbeat(
	logLine cursorLogLine,
	model string,
	sessionID string,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params cursorEditParams
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
			g.userAgent(filePath, model),
		)

		return &h
	case "edit_file":
		var (
			params  cursorEditParams
			rawArgs cursorEditRawArgs
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
			g.userAgent(filePath, model),
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
			g.userAgent(filePath, model),
		)

		return &h
	default:
		return nil
	}
}

func (g Cursor) userAgent(entity string, model string) string {
	return aiUserAgentWithAgentPrefix(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, ""), model)
}

func cursorUserAgentWithModel(userAgent string, modelToken string) string {
	return userAgentWithPrependedAgent(userAgent, modelToken)
}

func cursorModelUserAgentToken(model string) string {
	return aiAgentUserAgentToken(model)
}

func (Cursor) modelName(raw []byte) string {
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return ""
	}

	return cursorModelNameFromValue(value)
}

func cursorModelNameFromValue(value any) string {
	obj, ok := value.(map[string]any)
	if !ok {
		return cursorCleanModelName(value)
	}

	for _, key := range []string{
		"modelName",
		"model_name",
		"modelId",
		"modelID",
		"model_id",
		"modelSlug",
		"model",
		"selectedModel",
		"selectedChatModel",
	} {
		if model := cursorCleanModelName(obj[key]); model != "" {
			return model
		}
	}

	for _, key := range []string{
		"modelDetails",
		"modelConfig",
		"modelConfiguration",
		"modelInfo",
		"selectedModel",
		"selectedChatModel",
	} {
		if model := cursorCleanModelName(obj[key]); model != "" {
			return model
		}

		if model := cursorModelNameFromValue(obj[key]); model != "" {
			return model
		}
	}

	for key, nested := range obj {
		if !strings.Contains(strings.ToLower(key), "model") {
			continue
		}

		if model := cursorCleanModelName(nested); model != "" {
			return model
		}

		if model := cursorModelNameFromValue(nested); model != "" {
			return model
		}
	}

	return ""
}

func cursorCleanModelName(value any) string {
	switch v := value.(type) {
	case string:
		model := strings.Join(strings.Fields(strings.TrimSpace(v)), "-")

		model = strings.Trim(model, "/")
		if len(model) > 128 {
			return ""
		}

		return model
	case map[string]any:
		for _, key := range []string{
			"name",
			"id",
			"modelName",
			"model_name",
			"modelId",
			"modelID",
			"model_id",
			"slug",
			"value",
		} {
			if model := cursorCleanModelName(v[key]); model != "" {
				return model
			}
		}
	}

	return ""
}

func (Cursor) tokenDelta(tokens heartbeat.AITokens) (int64, int64) {
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

func (g Cursor) hasTokenDelta(tokens heartbeat.AITokens) bool {
	input, output := g.tokenDelta(tokens)
	return input > 0 || output > 0
}

func (Cursor) advanceTokens(tokens heartbeat.AITokens) heartbeat.AITokens {
	tokens.LastInput = tokens.CurrentInput
	tokens.LastOutput = tokens.CurrentOutput

	return tokens
}

func (Cursor) tokensForFirstHeartbeat(assign bool, tokens heartbeat.AITokens) *heartbeat.AITokens {
	if !assign {
		return nil
	}

	copy := tokens

	return &copy
}

func (Cursor) newHeartbeat(
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

func (Cursor) filePath(path string, codeBlocks []cursorCodeBlock) string {
	if path != "" {
		return path
	}

	if len(codeBlocks) == 0 || codeBlocks[0].URI == nil {
		return ""
	}

	return codeBlocks[0].URI.FSPath
}

func (g Cursor) projectPath(logLine cursorLogLine) string {
	if logLine.ToolFormerData == nil {
		return ""
	}

	var filePath string

	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params cursorEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err == nil {
			filePath = g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
		}
	case "edit_file":
		var (
			params  cursorEditParams
			rawArgs cursorEditRawArgs
		)

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params)
		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		filePath = g.filePath(params.RelativeWorkspacePath, logLine.CodeBlocks)
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
			filePath = g.filePath("", logLine.CodeBlocks)
		}
	}

	if filePath == "" {
		return ""
	}

	return filepath.Dir(filePath)
}

func (g Cursor) lineChanges(content string) int {
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

func (Cursor) looksLikeUnifiedDiff(content string) bool {
	trimmed := strings.TrimLeft(content, " \t\r\n")
	if strings.HasPrefix(trimmed, "--- ") || strings.HasPrefix(trimmed, "+++ ") || strings.HasPrefix(trimmed, "@@") {
		return true
	}

	return strings.Contains(content, "\n--- ") ||
		strings.Contains(content, "\n+++ ") ||
		strings.Contains(content, "\n@@")
}

func (Cursor) lineChangesFromDiff(diff string) int {
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
func (Cursor) Name() string {
	return "Cursor"
}
