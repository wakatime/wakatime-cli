//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

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
		InputTokens       *int `json:"inputTokens"`
		OutputTokens      *int `json:"outputTokens"`
		TotalTokens       *int `json:"totalTokens"`
		InputTokensSnake  *int `json:"input_tokens"`
		OutputTokensSnake *int `json:"output_tokens"`
		TotalTokensSnake  *int `json:"total_tokens"`
	}

	cursorUsage struct {
		InputTokens       *int `json:"input_tokens"`
		OutputTokens      *int `json:"output_tokens"`
		TotalTokens       *int `json:"total_tokens"`
		InputTokensCamel  *int `json:"inputTokens"`
		OutputTokensCamel *int `json:"outputTokens"`
		TotalTokensCamel  *int `json:"totalTokens"`
	}

	cursorToolFormerData struct {
		Status  string `json:"status"`
		Name    string `json:"name"`
		Params  string `json:"params"`
		RawArgs string `json:"rawArgs"`
		Result  string `json:"result"`
	}

	cursorContextWindowStatus struct {
		TokensUsed *int `json:"tokensUsed"`
	}

	cursorEditResult struct {
		BeforeContentID string `json:"beforeContentId"`
		AfterContentID  string `json:"afterContentId"`
	}

	cursorURI struct {
		FSPath string `json:"_fsPath"`
	}

	cursorCodeBlock struct {
		URI     *cursorURI `json:"uri"`
		Content string     `json:"content"`
	}

	cursorLogLine struct {
		BubbleID            string
		CreatedAt           time.Time                  `json:"createdAt"`
		Type                int                        `json:"type"`
		Text                string                     `json:"text"`
		TokenCount          *cursorTokenCount          `json:"tokenCount"`
		Usage               *cursorUsage               `json:"usage"`
		ContextWindowStatus *cursorContextWindowStatus `json:"contextWindowStatusAtCreation"`
		ToolFormerData      *cursorToolFormerData      `json:"toolFormerData"`
		CodeBlocks          []cursorCodeBlock          `json:"codeBlocks"`
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

	contents, err := g.queryContentBlobs(ctx, db, g.editContentIDs(rows))
	if err != nil {
		logger.Debugf("failed reading cursor edit content blobs from %q: %s", dbPath, err)

		contents = nil
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

		parsed := g.cursorHeartbeats(logLine, bubbleCWDs[logLine.BubbleID], bubbleModels[logLine.BubbleID], tokens, contents)
		if len(parsed) == 0 {
			bubbleTokens[logLine.BubbleID] = tokens
			continue
		}

		bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)

		heartbeats = append(heartbeats, parsed...)
	}

	return cursorApplySubscriptionPlan(heartbeats, subscriptionPlan), nil
}

func (g Cursor) cursorTokenCounts(line cursorLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
	current := g.cursorExplicitTokenCounts(line, previous)

	// Recent Cursor versions only write zero tokenCount placeholders. The
	// cumulative context window usage on user bubbles is the remaining
	// token signal, so treat it as a cumulative input token counter.
	if line.ContextWindowStatus != nil &&
		line.ContextWindowStatus.TokensUsed != nil &&
		*line.ContextWindowStatus.TokensUsed > 0 {
		current.CurrentInput = cumulativeTokenCount(
			current.LastInput,
			current.CurrentInput,
			*line.ContextWindowStatus.TokensUsed,
		)
	}

	return current
}

func (Cursor) cursorExplicitTokenCounts(line cursorLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
	if line.TokenCount != nil {
		if line.TokenCount.isZero() {
			return previous
		}

		current := previous

		if inputTokens := cursorFirstInt(line.TokenCount.InputTokens, line.TokenCount.InputTokensSnake); inputTokens != nil {
			if *inputTokens != 0 {
				current.CurrentInput = cumulativeTokenCount(current.LastInput, current.CurrentInput, *inputTokens)
			}
		}

		outputTokens := cursorFirstInt(line.TokenCount.OutputTokens, line.TokenCount.OutputTokensSnake)

		totalTokens := cursorFirstInt(line.TokenCount.TotalTokens, line.TokenCount.TotalTokensSnake)
		switch {
		case outputTokens != nil && *outputTokens != 0:
			current.CurrentOutput = cumulativeTokenCount(current.LastOutput, current.CurrentOutput, *outputTokens)
		case totalTokens != nil && *totalTokens != 0:
			current.CurrentOutput = cumulativeTokenCount(current.LastOutput, current.CurrentOutput, *totalTokens)
		}

		return current
	}

	if line.Usage == nil {
		return previous
	}

	if line.Usage.isZero() {
		return previous
	}

	current := previous

	if inputTokens := cursorFirstInt(line.Usage.InputTokens, line.Usage.InputTokensCamel); inputTokens != nil {
		if *inputTokens != 0 {
			current.CurrentInput = int64(*inputTokens)
		}
	}

	outputTokens := cursorFirstInt(line.Usage.OutputTokens, line.Usage.OutputTokensCamel)

	totalTokens := cursorFirstInt(line.Usage.TotalTokens, line.Usage.TotalTokensCamel)
	switch {
	case outputTokens != nil && *outputTokens != 0:
		current.CurrentOutput = int64(*outputTokens)
	case totalTokens != nil && *totalTokens != 0:
		current.CurrentOutput = int64(*totalTokens)
	}

	return current
}

func cumulativeTokenCount(last int64, current int64, value int) int64 {
	next := int64(value)
	if next < last || next < current {
		return current
	}

	return next
}

func (c cursorTokenCount) isZero() bool {
	return cursorAllIntsZero(
		c.InputTokens,
		c.OutputTokens,
		c.TotalTokens,
		c.InputTokensSnake,
		c.OutputTokensSnake,
		c.TotalTokensSnake,
	)
}

func (u cursorUsage) isZero() bool {
	return cursorAllIntsZero(
		u.InputTokens,
		u.OutputTokens,
		u.TotalTokens,
		u.InputTokensCamel,
		u.OutputTokensCamel,
		u.TotalTokensCamel,
	)
}

func cursorAllIntsZero(values ...*int) bool {
	hasValue := false

	for _, value := range values {
		if value == nil {
			continue
		}

		hasValue = true

		if *value != 0 {
			return false
		}
	}

	return hasValue
}

func cursorFirstInt(values ...*int) *int {
	for _, value := range values {
		if value != nil {
			return value
		}
	}

	return nil
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
    OR json_extract(value, '$.tokenCount') IS NOT NULL
    OR json_extract(value, '$.usage') IS NOT NULL
    OR json_extract(value, '$.modelInfo') IS NOT NULL
    OR json_extract(value, '$.contextWindowStatusAtCreation') IS NOT NULL
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

func (Cursor) editContentIDs(rows []cursorLogRow) []string {
	var ids []string

	seen := make(map[string]bool)

	for _, row := range rows {
		if !strings.Contains(row.Value, "ContentId") {
			continue
		}

		var logLine cursorLogLine
		if err := json.Unmarshal([]byte(row.Value), &logLine); err != nil {
			continue
		}

		if logLine.ToolFormerData == nil || logLine.ToolFormerData.Result == "" {
			continue
		}

		var result cursorEditResult
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Result), &result); err != nil {
			continue
		}

		for _, id := range []string{result.BeforeContentID, result.AfterContentID} {
			if id != "" && !seen[id] {
				seen[id] = true

				ids = append(ids, id)
			}
		}
	}

	return ids
}

func (Cursor) queryContentBlobs(ctx context.Context, db *sql.DB, ids []string) (map[string]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	contents := make(map[string]string)

	const chunkSize = 500

	for start := 0; start < len(ids); start += chunkSize {
		end := start + chunkSize
		if end > len(ids) {
			end = len(ids)
		}

		chunk := ids[start:end]

		placeholders := strings.Repeat("?,", len(chunk))
		placeholders = placeholders[:len(placeholders)-1]

		params := make([]any, len(chunk))
		for i, id := range chunk {
			params[i] = id
		}

		rows, err := db.QueryContext(ctx, fmt.Sprintf(`
SELECT key, CAST(value AS TEXT)
FROM cursorDiskKV
WHERE key IN (%s);
`, placeholders), params...)
		if err != nil {
			return nil, fmt.Errorf("failed querying cursor content blobs: %s", err)
		}

		for rows.Next() {
			var key, value string
			if err := rows.Scan(&key, &value); err != nil {
				_ = rows.Close()

				return nil, fmt.Errorf("failed scanning cursor content blob row: %s", err)
			}

			contents[key] = value
		}

		if err := rows.Err(); err != nil {
			_ = rows.Close()

			return nil, fmt.Errorf("failed reading cursor content blob rows: %s", err)
		}

		_ = rows.Close()
	}

	return contents, nil
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

func (g Cursor) cursorHeartbeats(
	logLine cursorLogLine,
	cwd string,
	model string,
	tokens heartbeat.AITokens,
	contents map[string]string,
) Heartbeats {
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
		contents,
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
	contents map[string]string,
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
		if strings.TrimSpace(params.StreamingContent) == "" {
			lineChanges = g.lineChangesFromEditResult(logLine.ToolFormerData.Result, contents)
		}

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
		if strings.TrimSpace(content) == "" {
			lineChanges = g.lineChangesFromEditResult(logLine.ToolFormerData.Result, contents)
		}

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
	return aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, model, "")
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

func (g Cursor) lineChangesFromEditResult(result string, contents map[string]string) int {
	if result == "" || len(contents) == 0 {
		return 0
	}

	var editResult cursorEditResult
	if err := json.Unmarshal([]byte(result), &editResult); err != nil {
		return 0
	}

	after, ok := contents[editResult.AfterContentID]
	if !ok {
		return 0
	}

	return g.lineChangesFromSnapshots(contents[editResult.BeforeContentID], after)
}

// lineChangesFromSnapshots counts the lines in the after snapshot that are
// not part of the longest common subsequence with the before snapshot, i.e.
// lines added or modified by the edit.
func (Cursor) lineChangesFromSnapshots(before, after string) int {
	if strings.TrimSpace(after) == "" {
		return 0
	}

	beforeLines := strings.Split(before, "\n")
	afterLines := strings.Split(after, "\n")

	const maxDiffLines = 3000

	if len(beforeLines) > maxDiffLines || len(afterLines) > maxDiffLines {
		if diff := len(afterLines) - len(beforeLines); diff > 0 {
			return diff
		}

		return 0
	}

	prev := make([]int, len(beforeLines)+1)
	curr := make([]int, len(beforeLines)+1)

	for i := 1; i <= len(afterLines); i++ {
		for j := 1; j <= len(beforeLines); j++ {
			switch {
			case afterLines[i-1] == beforeLines[j-1]:
				curr[j] = prev[j-1] + 1
			case prev[j] >= curr[j-1]:
				curr[j] = prev[j]
			default:
				curr[j] = curr[j-1]
			}
		}

		prev, curr = curr, prev
	}

	return len(afterLines) - prev[len(beforeLines)]
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
