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

func parseGenericAISQLite(
	ctx context.Context,
	provider genericAIProvider,
) (Heartbeats, error) {
	parseCtx, cancel := aiSQLiteContext(ctx)
	defer cancel()

	paths, err := genericAISQLitePaths(provider.parser, provider.config, provider.sqliteRoots)
	if err != nil {
		return nil, err
	}

	logger := log.Extract(ctx)

	var heartbeats Heartbeats

	for _, path := range paths {
		parsed, err := parseGenericAISQLiteDB(parseCtx, provider, path)
		heartbeats = append(heartbeats, parsed...)

		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			logger.Warnf("failed parsing %s sqlite transcript %q: %s", provider.parser.Name(), path, err)

			if parseCtx.Err() != nil {
				break
			}
		}
	}

	return heartbeats, nil
}

func genericAISQLitePaths(parser Parser, _ ParserConfig, roots []string) ([]string, error) {
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
			if genericAISQLiteFilename(root) {
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

			seen[path] = true
			paths = append(paths, path)

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
	provider genericAIProvider,
	path string,
) (Heartbeats, error) {
	db, err := openAISQLiteDB(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed opening db: %s", err)
	}
	defer db.Close() // nolint:errcheck

	tables, err := genericAISQLiteTables(ctx, db)
	if err != nil {
		return nil, err
	}

	if err := prepareGenericAISQLiteCursors(ctx, provider, path, tables); err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	logger := log.Extract(ctx)

	for _, table := range tables {
		parsed, err := parseGenericAISQLiteTable(ctx, provider, db, path, table)
		heartbeats = append(heartbeats, parsed...)

		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return heartbeats, ctxErr
			}

			logger.Warnf("failed parsing %s sqlite table %q from %q: %s", provider.parser.Name(), table, path, err)
		}
	}

	return heartbeats, nil
}

func genericAISQLiteContainerValues(values []any, key string) []any {
	var expanded []any

	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok {
			expanded = append(expanded, value)
			continue
		}

		expanded = append(expanded, genericAIContainerValues(object, key)...)
	}

	return expanded
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

func parseGenericAISQLiteTable(
	ctx context.Context,
	provider genericAIProvider,
	db *sql.DB,
	path string,
	table string,
) (Heartbeats, error) {
	rowID, err := genericAISQLiteRowID(ctx, db, table)
	if err != nil {
		return nil, err
	}

	return parseGenericAISQLiteIncremental(ctx, provider, db, path, table, rowID)
}

func genericAISQLiteRow(ctx context.Context, rows *sql.Rows, columns []string) (map[string]any, error) {
	row, err := genericAISQLiteRawRow(ctx, rows, columns)
	if err != nil {
		return nil, err
	}

	return genericAISQLiteDecodeRow(row), nil
}

func genericAISQLiteRawRow(ctx context.Context, rows *sql.Rows, columns []string) (map[string]any, error) {
	values := make([]any, len(columns))

	destinations := make([]any, len(columns))
	for i := range values {
		destinations[i] = &values[i]
	}

	if err := rows.Scan(destinations...); err != nil {
		return nil, err
	}

	// Check before decoding: a single row may contain an entire conversation.
	// A context deadline cannot interrupt encoding/json while it decodes a value.
	var textValues []string

	for _, value := range values {
		switch typed := value.(type) {
		case string:
			textValues = append(textValues, typed)
		case []byte:
			textValues = append(textValues, string(typed))
		}
	}

	if err := aiSQLiteRead(ctx, textValues...); err != nil {
		return nil, err
	}

	result := make(map[string]any, len(columns))
	for i, column := range columns {
		result[column] = values[i]
	}

	return result, nil
}

func genericAISQLiteDecodeRow(row map[string]any) map[string]any {
	for column, value := range row {
		if contents, ok := value.([]byte); ok {
			row[column] = genericAIDecodedValue(string(contents))
		} else {
			row[column] = genericAICacheJSONText(value)
		}
	}

	return row
}
