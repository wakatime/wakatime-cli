//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"

	// Register the pure-Go SQLite driver used to read Goose session databases.
	_ "modernc.org/sqlite"
)

// Goose contains params for detecting heartbeats from Goose SQLite session logs.
type Goose ParserConfig

type gooseSessionRow struct {
	ID                string
	Name              string
	WorkingDir        string
	Provider          string
	UpdatedAt         time.Time
	Input             int64
	Output            int64
	UsesSummaryTokens bool
	ThreadID          string
	Prompt            string
}

type gooseMessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Parse parses the Goose SQLite session db for ai heartbeats.
func (g Goose) Parse(ctx context.Context) (Heartbeats, error) {
	dbPath, err := g.dbPath(ctx)
	if err != nil {
		return nil, err
	}

	if dbPath == "" {
		return Heartbeats{}, nil
	}

	if !g.dbModifiedAfter(dbPath, g.After) {
		return Heartbeats{}, nil
	}

	rows, err := g.queryRows(ctx, dbPath)
	if err != nil {
		return nil, err
	}

	heartbeats := make(Heartbeats, 0, len(rows))
	for _, row := range rows {
		entity := appHeartbeatEntity(g.Name(), row.ID)

		tokens := heartbeat.AITokens{}
		if !row.UsesSummaryTokens {
			tokens.CurrentInput = row.Input
			tokens.CurrentOutput = row.Output
		}

		h := heartbeat.NewWithAITokens(
			nil,
			row.ID,
			tokens,
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
			row.WorkingDir,
			float64(row.UpdatedAt.Unix()),
			aiUserAgent(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, row.Provider)),
		)

		prompt := row.Prompt
		if prompt == "" {
			prompt = row.Name
		}

		h.AIPromptLength = promptLength(prompt)

		heartbeats = append(heartbeats, h)
	}

	return heartbeats, nil
}

func (Goose) dbPath(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, ".local", "share", "goose", "sessions", "sessions.db"),
		filepath.Join(home, "AppData", "Roaming", "Block", "goose", "data", "sessions", "sessions.db"),
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", nil
}

func (Goose) dbModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(after)
}

func (g Goose) queryRows(ctx context.Context, dbPath string) ([]gooseSessionRow, error) {
	logger := log.Extract(ctx)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening goose sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	columns, err := g.sessionColumns(ctx, db)
	if err != nil {
		return nil, err
	}

	minimum := []string{"id", "updated_at"}
	for _, column := range minimum {
		if !slices.Contains(columns, column) {
			logger.Warnf("skipping goose sqlite db %q: missing sessions column %q", dbPath, column)

			return []gooseSessionRow{}, nil
		}
	}

	if !slices.Contains(columns, "working_dir") {
		logger.Warnf("goose sqlite db %q is missing sessions column %q; project path will be empty", dbPath, "working_dir")
	}

	tables, err := g.tables(ctx, db)
	if err != nil {
		logger.Warnf("%s", err)
	}

	query, selectColumns, err := gooseSessionsQuery(columns)
	if err != nil {
		logger.Warnf("skipping goose sqlite db %q: %s", dbPath, err)

		return []gooseSessionRow{}, nil
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying goose sqlite db %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	var result []gooseSessionRow

	for rows.Next() {
		raw := make([]sql.NullString, len(selectColumns))

		dest := make([]interface{}, len(selectColumns))
		for i := range raw {
			dest[i] = &raw[i]
		}

		if err := rows.Scan(dest...); err != nil {
			return nil, fmt.Errorf("failed scanning goose sqlite row: %s", err)
		}

		row, ok := g.rowFromValues(selectColumns, raw)
		if !ok || row.UpdatedAt.IsZero() || row.UpdatedAt.Before(g.After) {
			continue
		}

		row.Prompt = g.sessionPrompt(ctx, db, tables, row)

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite rows: %s", err)
	}

	return result, nil
}

func (Goose) tables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table'")
	if err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite tables: %s", err)
	}
	defer rows.Close() // nolint:errcheck

	var tables []string

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed scanning goose sqlite table row: %s", err)
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite table rows: %s", err)
	}

	return tables, nil
}

func (Goose) sessionColumns(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA table_info(sessions)")
	if err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite schema: %s", err)
	}
	defer rows.Close() // nolint:errcheck

	var columns []string

	for rows.Next() {
		var (
			cid      int
			name     string
			typ      string
			notnull  int
			defaultV sql.NullString
			primaryK int
		)
		if err := rows.Scan(&cid, &name, &typ, &notnull, &defaultV, &primaryK); err != nil {
			return nil, fmt.Errorf("failed scanning goose sqlite schema row: %s", err)
		}

		columns = append(columns, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite schema rows: %s", err)
	}

	return columns, nil
}

func (Goose) rowFromValues(columns []string, raw []sql.NullString) (gooseSessionRow, bool) {
	row := gooseSessionRow{}

	for i, column := range columns {
		if !raw[i].Valid {
			continue
		}

		value := raw[i].String

		switch column {
		case "id":
			row.ID = value
		case "name":
			if value != "" {
				row.Name = value
			}
		case "description":
			if row.Name == "" {
				row.Name = value
			}
		case "working_dir":
			row.WorkingDir = value
		case "provider_name":
			row.Provider = value
		case "updated_at":
			row.UpdatedAt = parseGooseTime(value)
		case "input_tokens":
			row.Input = parseGooseInt(value)
		case "output_tokens":
			row.Output = parseGooseInt(value)
		case "accumulated_input_tokens":
			row.Input = parseGooseInt(value)
			row.UsesSummaryTokens = true
		case "accumulated_output_tokens":
			row.Output = parseGooseInt(value)
			row.UsesSummaryTokens = true
		case "thread_id":
			row.ThreadID = value
		}
	}

	if row.ID == "" {
		return gooseSessionRow{}, false
	}

	return row, true
}

func (Goose) sessionPrompt(ctx context.Context, db *sql.DB, tables []string, row gooseSessionRow) string {
	if slices.Contains(tables, "messages") {
		prompt := gooseQueryLatestPrompt(
			ctx,
			db,
			`SELECT content_json
			 FROM messages
			 WHERE session_id = ? AND role = 'user'
			 ORDER BY created_timestamp DESC
			 LIMIT 20`,
			row.ID,
		)
		if prompt != "" {
			return prompt
		}
	}

	if row.ThreadID != "" && slices.Contains(tables, "thread_messages") {
		prompt := gooseQueryLatestPrompt(
			ctx,
			db,
			`SELECT content_json
			 FROM thread_messages
			 WHERE thread_id = ? AND role = 'user'
			 ORDER BY created_timestamp DESC
			 LIMIT 20`,
			row.ThreadID,
		)
		if prompt != "" {
			return prompt
		}
	}

	return ""
}

func gooseQueryLatestPrompt(ctx context.Context, db *sql.DB, query, id string) string {
	rows, err := db.QueryContext(ctx, query, id)
	if err != nil {
		return ""
	}
	defer rows.Close() // nolint:errcheck

	for rows.Next() {
		var raw sql.NullString
		if err := rows.Scan(&raw); err != nil || !raw.Valid {
			continue
		}

		if prompt := goosePromptFromContent(raw.String); prompt != "" {
			return prompt
		}
	}

	return ""
}

func goosePromptFromContent(raw string) string {
	if raw == "" {
		return ""
	}

	var text string
	if err := json.Unmarshal([]byte(raw), &text); err == nil {
		return text
	}

	var items []gooseMessageContent
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return ""
	}

	var parts []string

	for _, item := range items {
		if item.Type == "text" && item.Text != "" {
			parts = append(parts, item.Text)
		}
	}

	return strings.Join(parts, "\n")
}

func parseGooseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999-07:00",
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}

	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		if unix > 1_000_000_000_000 {
			return time.UnixMilli(unix)
		}

		return time.Unix(unix, 0)
	}

	return time.Time{}
}

func parseGooseInt(value string) int64 {
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}

	return parsed
}

func gooseSessionsQuery(columns []string) (string, []string, error) {
	hasSessionType := slices.Contains(columns, "session_type")

	if !slices.Contains(columns, "id") || !slices.Contains(columns, "updated_at") {
		return "", nil, fmt.Errorf("missing minimum sessions columns")
	}

	inputColumn := ""

	switch {
	case slices.Contains(columns, "input_tokens"):
		inputColumn = "input_tokens"
	case slices.Contains(columns, "accumulated_input_tokens"):
		inputColumn = "accumulated_input_tokens"
	}

	outputColumn := ""

	switch {
	case slices.Contains(columns, "output_tokens"):
		outputColumn = "output_tokens"
	case slices.Contains(columns, "accumulated_output_tokens"):
		outputColumn = "accumulated_output_tokens"
	}

	titleColumns := []string{}

	for _, column := range []string{"name", "description"} {
		if slices.Contains(columns, column) {
			titleColumns = append(titleColumns, column)
		}
	}

	if slices.Contains(columns, "thread_id") {
		titleColumns = append(titleColumns, "thread_id")
	}

	selectColumns := []string{"id"}
	selectColumns = append(selectColumns, titleColumns...)

	for _, column := range []string{"working_dir", "updated_at", "provider_name"} {
		if slices.Contains(columns, column) {
			selectColumns = append(selectColumns, column)
		}
	}

	if inputColumn != "" {
		selectColumns = append(selectColumns, inputColumn)
	}

	if outputColumn != "" {
		selectColumns = append(selectColumns, outputColumn)
	}

	query := "SELECT " + strings.Join(selectColumns, ", ") + " FROM sessions"
	if hasSessionType {
		query += " WHERE LOWER(session_type) != 'hidden'"
	}

	query += " ORDER BY updated_at ASC"

	return query, selectColumns, nil
}

// Name returns its name.
func (Goose) Name() string {
	return "Goose"
}
