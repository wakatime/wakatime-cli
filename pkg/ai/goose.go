//go:build !freebsd && !openbsd && !netbsd && !dragonfly

package ai

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"

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
		h.AIPromptLength = promptLength(row.Name)

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

func (g Goose) queryRows(ctx context.Context, dbPath string) ([]gooseSessionRow, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening goose sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	columns, err := g.sessionColumns(ctx, db)
	if err != nil {
		return nil, err
	}

	required := []string{"id", "name", "working_dir", "updated_at"}
	for _, column := range required {
		if !slices.Contains(columns, column) {
			return nil, fmt.Errorf("missing goose sessions column %q in %q", column, dbPath)
		}
	}

	query, selectColumns, err := gooseSessionsQuery(columns)
	if err != nil {
		return nil, err
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

		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading goose sqlite rows: %s", err)
	}

	return result, nil
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
			row.Name = value
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
		}
	}

	if row.ID == "" {
		return gooseSessionRow{}, false
	}

	return row, true
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
	hasProvider := slices.Contains(columns, "provider_name")
	hasSessionType := slices.Contains(columns, "session_type")

	inputColumn := ""

	switch {
	case slices.Contains(columns, "accumulated_input_tokens"):
		inputColumn = "accumulated_input_tokens"
	case slices.Contains(columns, "input_tokens"):
		inputColumn = "input_tokens"
	}

	outputColumn := ""

	switch {
	case slices.Contains(columns, "accumulated_output_tokens"):
		outputColumn = "accumulated_output_tokens"
	case slices.Contains(columns, "output_tokens"):
		outputColumn = "output_tokens"
	}

	type queryOption struct {
		hasProvider    bool
		hasSessionType bool
		inputColumn    string
		outputColumn   string
		columns        []string
	}

	options := []queryOption{
		{
			hasProvider:    false,
			hasSessionType: false,
			inputColumn:    "",
			outputColumn:   "",
			columns:        []string{"id", "name", "working_dir", "updated_at"},
		},
		{
			hasProvider:    true,
			hasSessionType: false,
			inputColumn:    "",
			outputColumn:   "",
			columns:        []string{"id", "name", "working_dir", "updated_at", "provider_name"},
		},
		{
			hasProvider:    false,
			hasSessionType: true,
			inputColumn:    "",
			outputColumn:   "",
			columns:        []string{"id", "name", "working_dir", "updated_at", "session_type"},
		},
		{
			hasProvider:    true,
			hasSessionType: true,
			inputColumn:    "",
			outputColumn:   "",
			columns:        []string{"id", "name", "working_dir", "updated_at", "provider_name", "session_type"},
		},
	}

	baseOptions := append([]queryOption(nil), options...)
	for _, option := range baseOptions {
		for _, inColumn := range []string{"", inputColumn} {
			if inColumn == "" && inputColumn != "" {
				continue
			}

			for _, outColumn := range []string{"", outputColumn} {
				if outColumn == "" && outputColumn != "" {
					continue
				}

				if inColumn == "" && outColumn == "" {
					continue
				}

				cols := append([]string(nil), option.columns...)
				if inColumn != "" {
					cols = append(cols, inColumn)
				}

				if outColumn != "" {
					cols = append(cols, outColumn)
				}

				options = append(options, queryOption{
					hasProvider:    option.hasProvider,
					hasSessionType: option.hasSessionType,
					inputColumn:    inColumn,
					outputColumn:   outColumn,
					columns:        cols,
				})
			}
		}
	}

	for _, option := range options {
		if option.hasProvider == hasProvider &&
			option.hasSessionType == hasSessionType &&
			option.inputColumn == inputColumn &&
			option.outputColumn == outputColumn {
			query := "SELECT " + strings.Join(option.columns, ", ") + " FROM sessions"
			if option.hasSessionType {
				query += " WHERE session_type != 'Hidden'"
			}

			query += " ORDER BY updated_at ASC"

			return query, option.columns, nil
		}
	}

	return "", nil, fmt.Errorf("unsupported goose sessions schema")
}

// Name returns its name.
func (Goose) Name() string {
	return "Goose"
}
