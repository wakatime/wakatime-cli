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

type cursorTokenSource string

type cursorEditSource string

const (
	cursorTokenSourceNone                    cursorTokenSource = ""
	cursorTokenSourceTokenCountPerRequest    cursorTokenSource = "token_count_per_request" // #nosec G101 -- metric source
	cursorTokenSourceUsagePerRequest         cursorTokenSource = "usage_per_request"       // #nosec G101 -- metric source
	cursorTokenSourceCumulative              cursorTokenSource = "cumulative"
	cursorTokenSourceContextSnapshotEstimate cursorTokenSource = "context_snapshot_estimate"
)

const (
	cursorEditSourceNone        cursorEditSource = ""
	cursorEditSourceEmbedded    cursorEditSource = "embedded"
	cursorEditSourceSnapshot    cursorEditSource = "snapshot"
	cursorEditSourceUnavailable cursorEditSource = "unavailable"
)

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
		BubbleID              string
		CreatedAt             time.Time                  `json:"createdAt"`
		SchemaVersion         int                        `json:"_v"`
		Type                  int                        `json:"type"`
		Text                  string                     `json:"text"`
		TokenCount            *cursorTokenCount          `json:"tokenCount"`
		TokenCountUpUntilHere *cursorTokenCount          `json:"tokenCountUpUntilHere"`
		Usage                 *cursorUsage               `json:"usage"`
		ContextWindowStatus   *cursorContextWindowStatus `json:"contextWindowStatusAtCreation"`
		ToolFormerData        *cursorToolFormerData      `json:"toolFormerData"`
		CodeBlocks            []cursorCodeBlock          `json:"codeBlocks"`
	}

	cursorTokenObservation struct {
		Input     *int64
		Output    *int64
		Source    cursorTokenSource
		Estimated bool
	}

	cursorParseStats struct {
		ContentIDs              int
		ContentIDsResolved      int
		ContextEstimateRows     int
		CumulativeTokenRows     int
		EmbeddedEditRows        int
		MalformedRows           int
		OtherSchemaRows         int
		PendingRows             int
		RowsScanned             int
		SchemaV3Rows            int
		SnapshotEdits           int
		TokenCountRows          int
		UnavailableEditRows     int
		UnresolvedSnapshotEdits int
		UsageRows               int
	}

	cursorLogRow struct {
		BubbleID string
		Value    string
	}

	cursorPendingRow struct {
		CWD     string
		LogLine cursorLogLine
		Model   string
		Tokens  heartbeat.AITokens
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

	startedAt := time.Now()
	stats := cursorParseStats{}

	defer func() {
		logger.Debugf(
			"cursor parser stats: rows=%d malformed=%d schema_v3=%d schema_other=%d "+
				"pending=%d token_count=%d usage=%d cumulative=%d context_estimates=%d "+
				"embedded_edits=%d snapshot_edits=%d unavailable_edits=%d "+
				"content_ids=%d content_ids_resolved=%d "+
				"snapshot_edits_unresolved=%d duration=%s",
			stats.RowsScanned,
			stats.MalformedRows,
			stats.SchemaV3Rows,
			stats.OtherSchemaRows,
			stats.PendingRows,
			stats.TokenCountRows,
			stats.UsageRows,
			stats.CumulativeTokenRows,
			stats.ContextEstimateRows,
			stats.EmbeddedEditRows,
			stats.SnapshotEdits,
			stats.UnavailableEditRows,
			stats.ContentIDs,
			stats.ContentIDsResolved,
			stats.UnresolvedSnapshotEdits,
			time.Since(startedAt),
		)
	}()

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

	bubbleCWDs := make(map[string]string)
	bubbleModels := make(map[string]string)
	bubbleTokens := make(map[string]heartbeat.AITokens)
	contentIDs := make([]string, 0)
	seenContentIDs := make(map[string]bool)
	pendingRows := make([]cursorPendingRow, 0)

	err = g.queryRows(ctx, db, dbPath, func(row cursorLogRow) error {
		stats.RowsScanned++

		var logLine cursorLogLine
		if err := json.Unmarshal([]byte(row.Value), &logLine); err != nil {
			stats.MalformedRows++
			return nil
		}

		logLine.BubbleID = row.BubbleID
		if logLine.SchemaVersion == 3 {
			stats.SchemaV3Rows++
		} else {
			stats.OtherSchemaRows++
		}

		if model := g.modelName([]byte(row.Value)); model != "" && logLine.BubbleID != "" {
			bubbleModels[logLine.BubbleID] = model
		}

		if cwd := g.projectPath(logLine); cwd != "" && logLine.BubbleID != "" {
			bubbleCWDs[logLine.BubbleID] = cwd
		}

		observation := g.cursorTokenObservation(logLine)
		stats.recordTokenSource(observation.Source)
		tokens := g.applyTokenObservation(observation, bubbleTokens[logLine.BubbleID])

		if logLine.CreatedAt.IsZero() || logLine.CreatedAt.Before(g.After) {
			bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)
			return nil
		}

		// Probe without snapshot line counts so token state can be advanced
		// while rows are streamed. Snapshot-backed metrics are resolved after
		// the bounded row scan and the heartbeat is rebuilt below.
		if len(g.cursorHeartbeats(
			logLine,
			bubbleCWDs[logLine.BubbleID],
			bubbleModels[logLine.BubbleID],
			tokens,
			nil,
		)) == 0 {
			bubbleTokens[logLine.BubbleID] = tokens
			return nil
		}

		pendingRows = append(pendingRows, cursorPendingRow{
			CWD:     bubbleCWDs[logLine.BubbleID],
			LogLine: logLine,
			Model:   bubbleModels[logLine.BubbleID],
			Tokens:  tokens,
		})
		stats.PendingRows++

		editSource := g.editSource(logLine)
		stats.recordEditSource(editSource)

		if editSource == cursorEditSourceSnapshot {
			for _, id := range g.editContentIDs(logLine) {
				if !seenContentIDs[id] {
					seenContentIDs[id] = true
					contentIDs = append(contentIDs, id)
				}
			}
		}

		bubbleTokens[logLine.BubbleID] = g.advanceTokens(tokens)

		return nil
	})
	if err != nil {
		return nil, err
	}

	stats.ContentIDs = len(contentIDs)

	contentLineCounts, err := g.queryContentLineCounts(ctx, db, contentIDs)
	if err != nil {
		logger.Debugf("failed reading cursor edit content line counts from %q: %s", dbPath, err)

		contentLineCounts = nil
	}

	stats.ContentIDsResolved = len(contentLineCounts)

	for _, row := range pendingRows {
		if g.editSource(row.LogLine) != cursorEditSourceSnapshot {
			continue
		}

		ids := g.editContentIDs(row.LogLine)

		resolved := len(ids) == 2
		for _, id := range ids {
			if _, ok := contentLineCounts[id]; !ok {
				resolved = false
			}
		}

		if !resolved {
			stats.UnresolvedSnapshotEdits++
		}
	}

	var heartbeats Heartbeats

	for _, row := range pendingRows {
		heartbeats = append(heartbeats, g.cursorHeartbeats(
			row.LogLine,
			row.CWD,
			row.Model,
			row.Tokens,
			contentLineCounts,
		)...)
	}

	return cursorApplySubscriptionPlan(heartbeats, subscriptionPlan), nil
}

func (s *cursorParseStats) recordTokenSource(source cursorTokenSource) {
	switch source {
	case cursorTokenSourceNone:
	case cursorTokenSourceTokenCountPerRequest:
		s.TokenCountRows++
	case cursorTokenSourceUsagePerRequest:
		s.UsageRows++
	case cursorTokenSourceCumulative:
		s.CumulativeTokenRows++
	case cursorTokenSourceContextSnapshotEstimate:
		s.ContextEstimateRows++
	}
}

func (s *cursorParseStats) recordEditSource(source cursorEditSource) {
	switch source {
	case cursorEditSourceNone:
	case cursorEditSourceEmbedded:
		s.EmbeddedEditRows++
	case cursorEditSourceSnapshot:
		s.SnapshotEdits++
	case cursorEditSourceUnavailable:
		s.UnavailableEditRows++
	}
}

func (g Cursor) cursorTokenCounts(line cursorLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
	return g.applyTokenObservation(g.cursorTokenObservation(line), previous)
}

func (Cursor) cursorTokenObservation(line cursorLogLine) cursorTokenObservation {
	if line.TokenCount != nil {
		input, output := cursorTokenPair(
			cursorFirstInt(line.TokenCount.InputTokens, line.TokenCount.InputTokensSnake),
			cursorFirstInt(line.TokenCount.OutputTokens, line.TokenCount.OutputTokensSnake),
			cursorFirstInt(line.TokenCount.TotalTokens, line.TokenCount.TotalTokensSnake),
		)
		if cursorHasPositiveToken(input, output) {
			return cursorTokenObservation{
				Input:  input,
				Output: output,
				Source: cursorTokenSourceTokenCountPerRequest,
			}
		}
	}

	if line.Usage != nil {
		input, output := cursorTokenPair(
			cursorFirstInt(line.Usage.InputTokens, line.Usage.InputTokensCamel),
			cursorFirstInt(line.Usage.OutputTokens, line.Usage.OutputTokensCamel),
			cursorFirstInt(line.Usage.TotalTokens, line.Usage.TotalTokensCamel),
		)
		if cursorHasPositiveToken(input, output) {
			return cursorTokenObservation{
				Input:  input,
				Output: output,
				Source: cursorTokenSourceUsagePerRequest,
			}
		}
	}

	if line.TokenCountUpUntilHere != nil {
		input, output := cursorTokenPair(
			cursorFirstInt(
				line.TokenCountUpUntilHere.InputTokens,
				line.TokenCountUpUntilHere.InputTokensSnake,
			),
			cursorFirstInt(
				line.TokenCountUpUntilHere.OutputTokens,
				line.TokenCountUpUntilHere.OutputTokensSnake,
			),
			cursorFirstInt(
				line.TokenCountUpUntilHere.TotalTokens,
				line.TokenCountUpUntilHere.TotalTokensSnake,
			),
		)
		if cursorHasPositiveToken(input, output) {
			return cursorTokenObservation{
				Input:  input,
				Output: output,
				Source: cursorTokenSourceCumulative,
			}
		}
	}

	// Cursor's context-window value is a point-in-time context snapshot. It
	// can shrink after compaction, lags the latest turn, and has no matching
	// output count. Preserve it as an estimate for diagnostics, but do not
	// publish it as exact ai_input_tokens.
	if line.ContextWindowStatus != nil &&
		line.ContextWindowStatus.TokensUsed != nil &&
		*line.ContextWindowStatus.TokensUsed > 0 {
		return cursorTokenObservation{
			Input:     cursorInt64Pointer(int64(*line.ContextWindowStatus.TokensUsed)),
			Source:    cursorTokenSourceContextSnapshotEstimate,
			Estimated: true,
		}
	}

	return cursorTokenObservation{Source: cursorTokenSourceNone}
}

func (Cursor) applyTokenObservation(
	observation cursorTokenObservation,
	previous heartbeat.AITokens,
) heartbeat.AITokens {
	current := previous

	switch observation.Source {
	case cursorTokenSourceNone, cursorTokenSourceContextSnapshotEstimate:
	case cursorTokenSourceTokenCountPerRequest, cursorTokenSourceUsagePerRequest:
		if observation.Input != nil && *observation.Input > 0 {
			current.CurrentInput = max(current.CurrentInput, current.LastInput) + *observation.Input
		}

		if observation.Output != nil && *observation.Output > 0 {
			current.CurrentOutput = max(current.CurrentOutput, current.LastOutput) + *observation.Output
		}
	case cursorTokenSourceCumulative:
		if observation.Input != nil && *observation.Input > 0 {
			current.CurrentInput = cumulativeTokenCount64(
				current.LastInput,
				current.CurrentInput,
				*observation.Input,
			)
		}

		if observation.Output != nil && *observation.Output > 0 {
			current.CurrentOutput = cumulativeTokenCount64(
				current.LastOutput,
				current.CurrentOutput,
				*observation.Output,
			)
		}
	}

	return current
}

func cumulativeTokenCount(last int64, current int64, value int) int64 {
	return cumulativeTokenCount64(last, current, int64(value))
}

func cumulativeTokenCount64(last int64, current int64, next int64) int64 {
	if next < last || next < current {
		return current
	}

	return next
}

func cursorTokenPair(input *int, output *int, total *int) (*int64, *int64) {
	var inputValue *int64
	if input != nil {
		inputValue = cursorInt64Pointer(int64(*input))
	}

	var outputValue *int64

	switch {
	case output != nil:
		outputValue = cursorInt64Pointer(int64(*output))
	case total != nil && input != nil && *total >= *input:
		outputValue = cursorInt64Pointer(int64(*total - *input))
	}

	return inputValue, outputValue
}

func cursorInt64Pointer(value int64) *int64 {
	return &value
}

func cursorHasPositiveToken(values ...*int64) bool {
	for _, value := range values {
		if value != nil && *value > 0 {
			return true
		}
	}

	return false
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

func (Cursor) queryRows(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
	visit func(cursorLogRow) error,
) error {
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
ORDER BY json_extract(value, '$.createdAt') ASC, key ASC;
`, cursorRecentBubbleRowLimit)
	if err != nil {
		return fmt.Errorf("failed querying cursor sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	for rows.Next() {
		var (
			key string
			row string
		)
		if err := rows.Scan(&key, &row); err != nil {
			return fmt.Errorf("failed scanning cursor sqlite row: %s", err)
		}

		row = strings.TrimSpace(row)
		if row != "" {
			bubbleID := strings.TrimPrefix(key, "bubbleId:")
			if idx := strings.Index(bubbleID, ":"); idx != -1 {
				bubbleID = bubbleID[:idx]
			}

			if err := visit(cursorLogRow{
				BubbleID: bubbleID,
				Value:    row,
			}); err != nil {
				return err
			}
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed reading cursor sqlite rows: %s", err)
	}

	return nil
}

func (g Cursor) editSource(logLine cursorLogLine) cursorEditSource {
	if logLine.ToolFormerData == nil || logLine.ToolFormerData.Status != "completed" {
		return cursorEditSourceNone
	}

	switch logLine.ToolFormerData.Name {
	case "edit_file_v2":
		var params cursorEditParams
		if err := json.Unmarshal([]byte(logLine.ToolFormerData.Params), &params); err != nil {
			return cursorEditSourceUnavailable
		}

		if strings.TrimSpace(params.StreamingContent) != "" {
			return cursorEditSourceEmbedded
		}
	case "edit_file":
		var rawArgs cursorEditRawArgs

		_ = json.Unmarshal([]byte(logLine.ToolFormerData.RawArgs), &rawArgs)

		content := rawArgs.CodeEdit
		if content == "" && len(logLine.CodeBlocks) > 0 {
			content = logLine.CodeBlocks[0].Content
		}

		if strings.TrimSpace(content) != "" {
			return cursorEditSourceEmbedded
		}
	default:
		return cursorEditSourceNone
	}

	if len(g.editContentIDs(logLine)) == 2 {
		return cursorEditSourceSnapshot
	}

	return cursorEditSourceUnavailable
}

func (Cursor) editContentIDs(logLine cursorLogLine) []string {
	if logLine.ToolFormerData == nil || logLine.ToolFormerData.Result == "" {
		return nil
	}

	var result cursorEditResult
	if err := json.Unmarshal([]byte(logLine.ToolFormerData.Result), &result); err != nil {
		return nil
	}

	ids := make([]string, 0, 2)
	if result.BeforeContentID != "" {
		ids = append(ids, result.BeforeContentID)
	}

	if result.AfterContentID != "" {
		ids = append(ids, result.AfterContentID)
	}

	return ids
}

func (Cursor) queryContentLineCounts(
	ctx context.Context,
	db *sql.DB,
	ids []string,
) (map[string]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	lineCounts := make(map[string]int)

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
SELECT
	key,
	CASE
		WHEN length(CAST(value AS TEXT)) = 0 THEN 0
		ELSE length(CAST(value AS TEXT))
			- length(replace(CAST(value AS TEXT), char(10), ''))
			+ 1
	END
FROM cursorDiskKV
WHERE key IN (%s);
`, placeholders), params...)
		if err != nil {
			return nil, fmt.Errorf("failed querying cursor content line counts: %s", err)
		}

		for rows.Next() {
			var (
				key       string
				lineCount int
			)
			if err := rows.Scan(&key, &lineCount); err != nil {
				_ = rows.Close()

				return nil, fmt.Errorf("failed scanning cursor content line count row: %s", err)
			}

			lineCounts[key] = lineCount
		}

		if err := rows.Err(); err != nil {
			_ = rows.Close()

			return nil, fmt.Errorf("failed reading cursor content line count rows: %s", err)
		}

		_ = rows.Close()
	}

	return lineCounts, nil
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
	contentLineCounts map[string]int,
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
		contentLineCounts,
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
	contentLineCounts map[string]int,
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

		var lineChanges *int
		if strings.TrimSpace(params.StreamingContent) == "" {
			lineChanges = g.lineChangesFromEditResult(logLine.ToolFormerData.Result, contentLineCounts)
		} else {
			lineChanges = heartbeat.PointerTo(g.lineChanges(params.StreamingContent))
		}

		h := g.newHeartbeat(
			lineChanges,
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

		var lineChanges *int
		if strings.TrimSpace(content) == "" {
			lineChanges = g.lineChangesFromEditResult(logLine.ToolFormerData.Result, contentLineCounts)
		} else {
			lineChanges = heartbeat.PointerTo(g.lineChanges(content))
		}

		h := g.newHeartbeat(
			lineChanges,
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

func (Cursor) lineChangesFromEditResult(result string, lineCounts map[string]int) *int {
	if result == "" || len(lineCounts) == 0 {
		return nil
	}

	var editResult cursorEditResult
	if err := json.Unmarshal([]byte(result), &editResult); err != nil {
		return nil
	}

	beforeLines, beforeOK := lineCounts[editResult.BeforeContentID]

	afterLines, afterOK := lineCounts[editResult.AfterContentID]
	if !beforeOK || !afterOK {
		return nil
	}

	return heartbeat.PointerTo(afterLines - beforeLines)
}

// lineChangesFromSnapshots returns the signed net file line-count delta. This
// matches ai_line_changes, where positive values are additions and negative
// values are deletions.
func (Cursor) lineChangesFromSnapshots(before, after string) int {
	return cursorSnapshotLineCount(after) - cursorSnapshotLineCount(before)
}

func cursorSnapshotLineCount(content string) int {
	if content == "" {
		return 0
	}

	return strings.Count(content, "\n") + 1
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
