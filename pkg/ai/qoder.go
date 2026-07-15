//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"bufio"
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

// Qoder contains params for detecting heartbeats from Qoder local activity.
type Qoder ParserConfig

const (
	qoderRecentMessageRowLimit = 5000
	qoderRecentPromptRowLimit  = 5000
)

type (
	qoderMessageRow struct {
		SessionID   string
		ProjectURI  string
		RequestID   string
		Role        string
		ToolResult  string
		TokenInfo   string
		CreatedAtMS int64
	}

	qoderPromptRow struct {
		SessionID   string
		ProjectURI  string
		CreatedAtMS int64
	}

	qoderPrompt struct {
		SessionID  string
		ProjectURI string
		Timestamp  time.Time
		Length     int
	}

	qoderConversationLine struct {
		Role    string              `json:"role"`
		Message qoderHistoryMessage `json:"message"`
	}

	qoderHistoryMessage struct {
		Content []qoderContentBlock `json:"content"`
	}

	qoderContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	qoderTokenInfo struct {
		PromptTokens     int64 `json:"prompt_tokens"`
		CompletionTokens int64 `json:"completion_tokens"`
	}

	qoderToolResult struct {
		SessionID    string              `json:"sessionId"`
		RequestID    string              `json:"requestId"`
		ProjectPath  string              `json:"projectPath"`
		ToolCallName string              `json:"toolCallName"`
		Parameters   qoderToolParameters `json:"parameters"`
		Results      []qoderToolFile     `json:"results"`
	}

	qoderToolParameters struct {
		FilePath string `json:"file_path"`
		Path     string `json:"path"`
	}

	qoderToolFile struct {
		Path         string        `json:"path"`
		DiffInfo     qoderDiffInfo `json:"diffInfo"`
		LastDiffInfo qoderDiffInfo `json:"lastDiffInfo"`
	}

	qoderDiffInfo struct {
		Add    int `json:"add"`
		Delete int `json:"delete"`
	}
)

// Parse parses Qoder's local SQLite cache for ai heartbeats.
func (g Qoder) Parse(ctx context.Context) (Heartbeats, error) {
	dbPath, err := g.localDBPath(ctx)
	if err != nil {
		return nil, err
	}

	if dbPath == "" {
		return Heartbeats{}, nil
	}

	if !g.localDBModifiedAfter(dbPath, g.After) {
		return Heartbeats{}, nil
	}

	db, err := g.openDB(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	rows, err := g.queryRows(ctx, db, dbPath)
	if err != nil {
		return nil, err
	}

	promptRows, err := g.queryPromptRows(ctx, db, dbPath)
	if err != nil {
		return nil, err
	}

	prompts, err := g.qoderPrompts(ctx, promptRows)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, row := range rows {
		timestamp := time.UnixMilli(row.CreatedAtMS)
		if timestamp.IsZero() || !timestampAtOrAfterCutoff(timestamp, g.After) {
			continue
		}

		switch row.Role {
		case "assistant":
			if heartbeat := g.qoderAppHeartbeat(row, timestamp); heartbeat != nil {
				heartbeats = append(heartbeats, *heartbeat)
			}
		case "tool":
			heartbeats = append(heartbeats, g.qoderToolHeartbeats(row, timestamp)...)
		}
	}

	heartbeats = g.withPromptLengths(heartbeats, prompts)

	return heartbeats, nil
}

func (Qoder) localDBPath(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Qoder", "SharedClientCache", "cache", "db", "local.db"),
		filepath.Join(home, "AppData", "Roaming", "Qoder", "SharedClientCache", "cache", "db", "local.db"),
		filepath.Join(home, ".config", "Qoder", "SharedClientCache", "cache", "db", "local.db"),
		filepath.Join(home, ".config", "qoder", "SharedClientCache", "cache", "db", "local.db"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", nil
}

func (Qoder) localDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return timestampAtOrAfterCutoff(info.ModTime(), after)
}

func (g Qoder) afterUnixMilli() int64 {
	if g.After.IsZero() {
		return 0
	}

	return g.After.UnixMilli()
}

func (Qoder) openDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening qoder sqlite db %q: %s", dbPath, err)
	}

	// limit to a single connection, so the pure-Go sqlite driver only
	// allocates memory for one connection
	db.SetMaxOpenConns(1)

	return db, nil
}

func (g Qoder) queryRows(ctx context.Context, db *sql.DB, dbPath string) ([]qoderMessageRow, error) {
	rows, err := db.QueryContext(ctx, `
WITH recentQoderMessages AS (
  SELECT rowid, session_id, request_id, role, tool_result, token_info, gmt_create
  FROM chat_message
  ORDER BY rowid DESC
  LIMIT ?
)
SELECT
  m.session_id,
  COALESCE(s.project_uri, ''),
  COALESCE(m.request_id, ''),
  m.role,
  COALESCE(m.tool_result, ''),
  COALESCE(m.token_info, ''),
  COALESCE(m.gmt_create, 0)
FROM recentQoderMessages m
LEFT JOIN chat_session s ON s.session_id = m.session_id
WHERE m.role IN ('assistant', 'tool')
  AND COALESCE(m.gmt_create, 0) >= ?
ORDER BY m.gmt_create ASC;
`, qoderRecentMessageRowLimit, g.afterUnixMilli())
	if err != nil {
		return nil, fmt.Errorf("failed querying qoder sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	var results []qoderMessageRow

	for rows.Next() {
		var row qoderMessageRow
		if err := rows.Scan(
			&row.SessionID,
			&row.ProjectURI,
			&row.RequestID,
			&row.Role,
			&row.ToolResult,
			&row.TokenInfo,
			&row.CreatedAtMS,
		); err != nil {
			return nil, fmt.Errorf("failed scanning qoder sqlite row: %s", err)
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading qoder sqlite rows: %s", err)
	}

	return results, nil
}

func (Qoder) queryPromptRows(ctx context.Context, db *sql.DB, dbPath string) ([]qoderPromptRow, error) {
	rows, err := db.QueryContext(ctx, `
WITH recentQoderPromptMessages AS (
  SELECT rowid, session_id, gmt_create, role
  FROM chat_message
  ORDER BY rowid DESC
  LIMIT ?
)
SELECT
  m.session_id,
  COALESCE(s.project_uri, ''),
  COALESCE(m.gmt_create, 0)
FROM recentQoderPromptMessages m
LEFT JOIN chat_session s ON s.session_id = m.session_id
WHERE m.role = 'user'
ORDER BY m.session_id ASC, m.gmt_create ASC;
`, qoderRecentPromptRowLimit)
	if err != nil {
		return nil, fmt.Errorf("failed querying qoder prompt rows from sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	var results []qoderPromptRow

	for rows.Next() {
		var row qoderPromptRow
		if err := rows.Scan(&row.SessionID, &row.ProjectURI, &row.CreatedAtMS); err != nil {
			return nil, fmt.Errorf("failed scanning qoder prompt sqlite row: %s", err)
		}

		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading qoder prompt sqlite rows: %s", err)
	}

	return results, nil
}

func (g Qoder) qoderAppHeartbeat(row qoderMessageRow, timestamp time.Time) *heartbeat.Heartbeat {
	var tokenInfo qoderTokenInfo
	if err := json.Unmarshal([]byte(row.TokenInfo), &tokenInfo); err != nil {
		return nil
	}

	if tokenInfo.PromptTokens == 0 && tokenInfo.CompletionTokens == 0 {
		return nil
	}

	entity := appHeartbeatEntity("Qoder", row.SessionID)

	h := heartbeat.NewWithAITokens(
		nil,
		row.SessionID,
		heartbeat.AITokens{
			CurrentInput:  tokenInfo.PromptTokens,
			CurrentOutput: tokenInfo.CompletionTokens,
		},
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
		row.ProjectURI,
		float64(timestamp.UnixMilli())/1000,
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, "", ""),
	)

	return &h
}

func (g Qoder) withPromptLengths(heartbeats Heartbeats, prompts []qoderPrompt) Heartbeats {
	for _, prompt := range prompts {
		if prompt.Length == 0 || !timestampAtOrAfterCutoff(prompt.Timestamp, g.After) {
			continue
		}

		if idx := nearestPromptHeartbeat(heartbeats, prompt); idx >= 0 {
			heartbeats[idx].AIPromptLength = prompt.Length
			continue
		}

		heartbeats = append(heartbeats, g.qoderPromptHeartbeat(prompt))
	}

	return heartbeats
}

func nearestPromptHeartbeat(heartbeats Heartbeats, prompt qoderPrompt) int {
	idx := -1

	var nearest time.Duration

	for i, h := range heartbeats {
		if h.AIPromptLength != 0 || h.AISession != prompt.SessionID {
			continue
		}

		diff := time.UnixMilli(int64(h.Time * 1000)).Sub(prompt.Timestamp)
		if diff < 0 {
			diff = -diff
		}

		if idx == -1 || diff < nearest {
			idx = i
			nearest = diff
		}
	}

	return idx
}

func (g Qoder) qoderPromptHeartbeat(prompt qoderPrompt) heartbeat.Heartbeat {
	entity := appHeartbeatEntity("Qoder", prompt.SessionID)

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
		prompt.ProjectURI,
		float64(prompt.Timestamp.UnixMilli())/1000,
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, "", ""),
	)
	h.AISession = prompt.SessionID
	h.AIPromptLength = prompt.Length

	return h
}

func (g Qoder) qoderToolHeartbeats(row qoderMessageRow, timestamp time.Time) Heartbeats {
	var result qoderToolResult
	if err := json.Unmarshal([]byte(row.ToolResult), &result); err != nil {
		return nil
	}

	files := make([]qoderToolFile, 0, len(result.Results))
	for _, file := range result.Results {
		if file.Path != "" {
			files = append(files, file)
		}
	}

	if len(files) == 0 {
		files = []qoderToolFile{{Path: result.parameterFilePath()}}
	}

	isWrite := !strings.HasPrefix(result.ToolCallName, "read")

	sessionID := result.SessionID
	if sessionID == "" {
		sessionID = row.SessionID
	}

	var heartbeats Heartbeats

	for _, file := range files {
		if file.Path == "" {
			continue
		}

		var lineChanges *int
		if isWrite {
			lineChanges = heartbeat.PointerTo(file.lineChanges())
		}

		h := heartbeat.New(
			lineChanges,
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			file.Path,
			heartbeat.FileType,
			nil,
			false,
			heartbeat.PointerTo(isWrite),
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			"",
			float64(timestamp.UnixMilli())/1000,
			aiUserAgentWithModel(file.Path, g.UserAgents, g.FallbackUserAgent, "", ""),
		)
		h.AISession = sessionID
		heartbeats = append(heartbeats, h)
	}

	return heartbeats
}

func (r qoderToolResult) parameterFilePath() string {
	if r.Parameters.FilePath != "" {
		return r.Parameters.FilePath
	}

	return r.Parameters.Path
}

func (f qoderToolFile) lineChanges() int {
	diff := f.DiffInfo
	if diff.Add == 0 && diff.Delete == 0 {
		diff = f.LastDiffInfo
	}

	return diff.Add - diff.Delete
}

func (Qoder) qoderPrompts(ctx context.Context, rows []qoderPromptRow) ([]qoderPrompt, error) {
	if len(rows) == 0 {
		return nil, nil
	}

	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	history := make(map[string][]int)
	historyRoot := filepath.Join(home, ".qoder", "cache", "projects")

	paths, err := filepath.Glob(filepath.Join(
		historyRoot,
		"*",
		"conversation-history",
		"*",
		"*.jsonl",
	))
	if err != nil {
		return nil, fmt.Errorf("failed globbing qoder conversation history: %s", err)
	}

	for _, path := range paths {
		lengths, err := qoderPromptLengths(historyRoot, path)
		if err != nil {
			continue
		}

		if len(lengths) == 0 {
			continue
		}

		sessionPrefix := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		history[sessionPrefix] = append(history[sessionPrefix], lengths...)
	}

	promptOffsets := make(map[string]int)

	var prompts []qoderPrompt

	for _, row := range rows {
		timestamp := time.UnixMilli(row.CreatedAtMS)
		if timestamp.IsZero() {
			continue
		}

		lengths := qoderPromptLengthsForSession(row.SessionID, history)

		offset := promptOffsets[row.SessionID]
		if offset >= len(lengths) {
			continue
		}

		promptOffsets[row.SessionID]++
		prompts = append(prompts, qoderPrompt{
			SessionID:  row.SessionID,
			ProjectURI: row.ProjectURI,
			Timestamp:  timestamp,
			Length:     lengths[offset],
		})
	}

	return prompts, nil
}

func qoderPromptLengths(root string, path string) ([]int, error) {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	cleanPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return nil, err
	}

	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return nil, fmt.Errorf("qoder conversation history path outside root: %s", path)
	}

	file, err := os.Open(cleanPath) // #nosec G304 -- path is globbed and verified under Qoder history root.
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint:errcheck

	var lengths []int

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		var line qoderConversationLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}

		if line.Role != "user" {
			continue
		}

		for _, block := range line.Message.Content {
			if block.Type != "text" {
				continue
			}

			length := promptLength(qoderUserQuery(block.Text))
			if length > 0 {
				lengths = append(lengths, length)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lengths, nil
}

func qoderPromptLengthsForSession(sessionID string, history map[string][]int) []int {
	var (
		bestPrefix  string
		bestLengths []int
	)

	for prefix, lengths := range history {
		if strings.HasPrefix(sessionID, prefix) && len(prefix) > len(bestPrefix) {
			bestPrefix = prefix
			bestLengths = lengths
		}
	}

	return bestLengths
}

func qoderUserQuery(text string) string {
	const (
		openTag  = "<user_query>"
		closeTag = "</user_query>"
	)

	start := strings.LastIndex(text, openTag)
	if start == -1 {
		return strings.TrimSpace(text)
	}

	query := text[start+len(openTag):]

	end := strings.Index(query, closeTag)
	if end != -1 {
		query = query[:end]
	}

	return strings.TrimSpace(query)
}

// Name returns its name.
func (Qoder) Name() string {
	return "Qoder"
}
