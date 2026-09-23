//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type copilotUsageRow struct {
	id, session, model, cwd string
	timestamp               time.Time
	input, cached, output   int64
}

func (g Copilot) reconcileSQLiteUsage(ctx context.Context, timed []copilotTimedHeartbeat) ([]copilotTimedHeartbeat, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return timed, err
	}

	path := filepath.Join(home, ".copilot", "session-store.db")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return timed, nil
	}

	ctx, cancel := aiSQLiteContext(ctx)
	defer cancel()

	db, err := openAISQLiteDB(ctx, path)
	if err != nil {
		return timed, err
	}
	defer db.Close() // nolint:errcheck
	// Optional columns differ by CLI release; inspect once, never require billing metadata.
	columns, err := aiSQLiteColumns(ctx, db, "assistant_usage_events")
	if err != nil {
		return timed, err
	}

	output := "0"
	if columns["initiator"] && columns["output_tokens"] {
		output = `CASE WHEN e.initiator = 'compaction' THEN COALESCE(e.output_tokens,0) ELSE 0 END`
	}

	rows, err := db.QueryContext(ctx, `SELECT e.id, e.session_id, COALESCE(e.model,''),
 COALESCE(e.input_tokens,0), COALESCE(e.cache_read_tokens,0), `+output+`,
 COALESCE(e.created_at,s.created_at), COALESCE(s.cwd,'')
 FROM assistant_usage_events e LEFT JOIN sessions s ON s.id=e.session_id ORDER BY e.created_at, e.id`)
	if err != nil {
		return timed, err
	}
	defer rows.Close() // nolint:errcheck

	var usage []copilotUsageRow

	for rows.Next() {
		var (
			row       copilotUsageRow
			timestamp any
		)

		if err := rows.Scan(&row.id, &row.session, &row.model, &row.input, &row.cached, &row.output, &timestamp,
			&row.cwd); err != nil {
			return timed, err
		}

		if err := aiSQLiteRead(ctx, row.id, row.session, row.model, row.cwd); err != nil {
			return timed, err
		}

		row.timestamp = genericAITime(timestamp)
		row.input = max(row.input-row.cached, 0)
		row.cached, row.output = max(row.cached, 0), max(row.output, 0)
		usage = append(usage, row)
	}

	if err := rows.Err(); err != nil {
		return timed, err
	}
	// Reconcile using all rows in each shutdown leg, including those before After.
	// Only the emission is cutoff-filtered; otherwise a later shutdown replays old input.
	for i := range timed {
		item := &timed[i]
		if !item.shutdown {
			continue
		}

		for _, row := range usage {
			if row.session != item.heartbeat.AISession || row.timestamp.IsZero() ||
				row.timestamp.After(item.timestamp) || (!item.legStart.IsZero() && row.timestamp.Before(item.legStart)) {
				continue
			}

			h := &item.heartbeat
			h.AIInputTokens = max(h.AIInputTokens-row.input, 0)
			h.AICachedInputTokens = max(h.AICachedInputTokens-row.cached, 0)
			h.AIOutputTokens = max(h.AIOutputTokens-row.output, 0)
		}
	}

	for _, row := range usage {
		if row.timestamp.IsZero() || !timestampAtOrAfterCutoff(row.timestamp, g.After) {
			continue
		}

		event := genericAIEvent{sessionID: row.session, timestamp: row.timestamp, model: row.model, cwd: row.cwd,
			input: row.input, cachedInput: row.cached, output: row.output, tokensFound: true}
		for _, h := range genericAIEventHeartbeats(g, ParserConfig(g), event) {
			timed = append(timed, copilotTimedHeartbeat{timestamp: row.timestamp, heartbeat: h})
		}
	}

	return timed, nil
}

func aiSQLiteColumns(ctx context.Context, db *sql.DB, table string) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM "+aiSQLiteQuote(table)+" LIMIT 0") // nolint:gosec
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint:errcheck

	names, err := rows.Columns()

	columns := make(map[string]bool)
	for _, name := range names {
		columns[name] = true
	}

	return columns, err
}

func (g Copilot) otelHeartbeats(ctx context.Context, existing []copilotTimedHeartbeat) ([]copilotTimedHeartbeat, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, root := range vscodeExtensionTaskRoots(home, "github.copilot-chat") {
		paths = append(paths, filepath.Join(filepath.Dir(root), "agent-traces.db"))
	}

	var timed []copilotTimedHeartbeat

	seen := make(map[string]bool)

	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			continue
		}

		parsed, err := g.parseOTelDB(ctx, path, existing, seen)
		if err != nil {
			return timed, err
		}

		timed = append(timed, parsed...)
	}

	return timed, nil
}

func (g Copilot) parseOTelDB(ctx context.Context, path string, existing []copilotTimedHeartbeat,
	seen map[string]bool) ([]copilotTimedHeartbeat, error) {
	ctx, cancel := aiSQLiteContext(ctx)
	defer cancel()

	db, err := openAISQLiteDB(ctx, path)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	tables, err := genericAISQLiteTables(ctx, db)
	if err != nil {
		return nil, err
	}

	attrs := make(map[string]map[string]any)

	for _, table := range tables {
		if table != "span_attributes" {
			continue
		}

		rows, err := db.QueryContext(ctx, `SELECT span_id,key,value FROM span_attributes`)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var (
				id, key string
				value   any
			)

			if err := rows.Scan(&id, &key, &value); err != nil {
				_ = rows.Close()
				return nil, err
			}

			if attrs[id] == nil {
				attrs[id] = make(map[string]any)
			}

			attrs[id][key] = value
		}

		err = rows.Err()
		_ = rows.Close()

		if err != nil {
			return nil, err
		}
	}

	rows, err := db.QueryContext(ctx, `SELECT * FROM spans WHERE start_time_ms >= ? ORDER BY start_time_ms`,
		g.After.UnixMilli())
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []copilotTimedHeartbeat

	for rows.Next() {
		row, err := genericAISQLiteRow(ctx, rows, columns)
		if err != nil {
			return nil, err
		}

		id := genericAIString(row["span_id"])
		if seen[id] {
			continue
		}

		a := attrs[id]
		if a == nil {
			a = make(map[string]any)

			switch v := row["attributes"].(type) {
			case map[string]any:
				a = v
			case string:
				_ = json.Unmarshal([]byte(v), &a)
			}
		}

		operation := firstNonEmptyString(genericAIString(a["gen_ai.operation.name"]), genericAIString(row["operation_name"]))
		if operation != "chat" {
			continue
		}

		event := genericAIEvent{sessionID: genericAIString(a["gen_ai.conversation.id"]),
			timestamp: genericAITime(row["start_time_ms"]),
			model: firstNonEmptyString(genericAIString(a["gen_ai.response.model"]),
				genericAIString(a["gen_ai.request.model"]), genericAIString(row["response_model"])),
			input:       genericAIInt64(a["gen_ai.usage.input_tokens"]),
			output:      genericAIInt64(a["gen_ai.usage.output_tokens"]),
			cachedInput: genericAIInt64(a["gen_ai.usage.cache_read.input_tokens"]), tokensFound: true}

		event.input = max(event.input-event.cachedInput, 0)
		if event.sessionID == "" {
			event.sessionID = genericAIString(row["trace_id"])
		}
		// The same chat request may already exist in chatSessions JSONL.
		duplicate := false

		for _, item := range existing {
			h := item.heartbeat
			if h.AISession == event.sessionID && (h.AIInputTokens > 0 || h.AIOutputTokens > 0) &&
				absFloat64(h.Time-heartbeatTimestamp(event.timestamp)) < 1 && (event.model == "" ||
				strings.Contains(h.UserAgent, aiModelUserAgentToken(event.model, ""))) {
				duplicate = true
				break
			}
		}

		seen[id] = true

		if duplicate {
			continue
		}

		for _, h := range genericAIEventHeartbeats(g, ParserConfig(g), event) {
			result = append(result, copilotTimedHeartbeat{timestamp: event.timestamp, heartbeat: h})
		}
	}

	return result, rows.Err()
}
