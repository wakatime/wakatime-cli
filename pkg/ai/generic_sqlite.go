//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"

	// Register the pure-Go SQLite driver used to read provider session databases.
	_ "modernc.org/sqlite"
)

const genericAISQLiteRowLimit = 5000

func parseGenericAISQLite(
	ctx context.Context,
	parser Parser,
	config ParserConfig,
	roots []string,
) (Heartbeats, error) {
	paths, err := genericAISQLitePaths(parser, config, roots)
	if err != nil {
		return nil, err
	}

	logger := log.Extract(ctx)

	var heartbeats Heartbeats

	for _, path := range paths {
		parsed, err := parseGenericAISQLiteDB(ctx, parser, config, path)
		if err != nil {
			logger.Warnf("failed parsing %s sqlite transcript %q: %s", parser.Name(), path, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func genericAISQLitePaths(parser Parser, config ParserConfig, roots []string) ([]string, error) {
	var paths []string

	seen := make(map[string]bool)

	for _, root := range roots {
		if root == "" {
			continue
		}

		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to stat %s sqlite path %q: %s", parser.Name(), root, err)
		}

		if !info.IsDir() {
			if genericAISQLiteFilename(root) && timestampAtOrAfterCutoff(info.ModTime(), config.After) {
				paths = append(paths, root)
			}

			continue
		}

		err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() || !genericAISQLiteFilename(path) || seen[path] {
				return nil
			}

			info, err := entry.Info()
			if err == nil && timestampAtOrAfterCutoff(info.ModTime(), config.After) {
				seen[path] = true
				paths = append(paths, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk %s sqlite path %q: %s", parser.Name(), root, err)
		}
	}

	return paths, nil
}

func genericAISQLiteFilename(path string) bool {
	extension := strings.ToLower(filepath.Ext(path))
	return extension == ".db" || extension == ".sqlite" || extension == ".sqlite3"
}

func parseGenericAISQLiteDB(
	ctx context.Context,
	parser Parser,
	config ParserConfig,
	path string,
) (Heartbeats, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed opening db: %s", err)
	}
	defer db.Close() // nolint:errcheck

	tables, err := genericAISQLiteTables(ctx, db)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	logger := log.Extract(ctx)
	provider := genericAIProvider{parser: parser, config: config}

	for _, table := range tables {
		values, err := genericAISQLiteRows(ctx, db, table)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}

			logger.Warnf("failed parsing %s sqlite table %q from %q: %s", parser.Name(), table, path, err)

			continue
		}

		parsed, err := genericAIHeartbeats(ctx, provider, path, values)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func genericAISQLiteTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%'")
	if err != nil {
		return nil, fmt.Errorf("failed listing tables: %s", err)
	}
	defer rows.Close() // nolint:errcheck

	var tables []string

	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed scanning table name: %s", err)
		}

		tables = append(tables, table)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading table names: %s", err)
	}

	return tables, nil
}

func genericAISQLiteRows(ctx context.Context, db *sql.DB, table string) ([]any, error) {
	quotedTable := `"` + strings.ReplaceAll(table, `"`, `""`) + `"`
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", quotedTable, genericAISQLiteRowLimit) // nolint:gosec

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed querying table %q: %s", table, err)
	}
	defer rows.Close() // nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed reading columns for table %q: %s", table, err)
	}

	var values []any

	for rows.Next() {
		row, err := genericAISQLiteRow(rows, columns)
		if err != nil {
			return nil, fmt.Errorf("failed scanning table %q: %s", table, err)
		}

		values = append(values, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading table %q: %s", table, err)
	}

	return values, nil
}

func genericAISQLiteRow(rows *sql.Rows, columns []string) (map[string]any, error) {
	values := make([]any, len(columns))

	destinations := make([]any, len(columns))
	for i := range values {
		destinations[i] = &values[i]
	}

	if err := rows.Scan(destinations...); err != nil {
		return nil, err
	}

	result := make(map[string]any, len(columns))
	for i, column := range columns {
		if contents, ok := values[i].([]byte); ok {
			result[column] = genericAIDecodedValue(string(contents))
			continue
		}

		result[column] = values[i]
	}

	return result, nil
}
