//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"
)

const aiSQLiteBatchSize = 128

// A scan is checkpointed after each complete row. Row hashes distinguish updates
// from unchanged records when a changed database is scanned again.
type genericAISQLiteCursor struct {
	Version       int
	ParserVersion int
	RowID         *int64
	Offset        int64
	Cutoff        time.Time
	Pending       bool
	Rescan        bool
	Fingerprint   string
	State         genericAIParseState
	Hashes        map[string]string
	ScanHashes    map[string]string
}

const genericAISQLiteCursorVersion = 2

func aiSQLiteQuote(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// Find an addressable rowid for keyset pagination, not as evidence that rows
// are immutable. Every changed database is checked for updates to existing rows.
func genericAISQLiteRowID(ctx context.Context, db *sql.DB, table string) (string, error) {
	rows, err := db.QueryContext(ctx, "SELECT * FROM "+aiSQLiteQuote(table)+" LIMIT 0") // nolint:gosec
	if err != nil {
		return "", err
	}

	columns, err := rows.Columns()
	_ = rows.Close()

	if err != nil {
		return "", err
	}

	for i, column := range columns {
		columns[i] = strings.ToLower(column)
	}

	for _, alias := range []string{"rowid", "_rowid_", "oid"} {
		if slices.Contains(columns, alias) {
			continue
		}

		// Do not quote the alias: SQLite can treat an unknown quoted identifier
		// as a string literal on WITHOUT ROWID tables.
		rows, err := db.QueryContext(ctx, "SELECT "+alias+" FROM "+aiSQLiteQuote(table)+" LIMIT 0") // nolint:gosec
		if err != nil {
			return "", ctx.Err()
		}

		_ = rows.Close()

		return alias, nil
	}

	return "", nil
}

func genericAISQLiteCursorPath(ctx context.Context, provider genericAIProvider, path, table string) (string, error) {
	dir, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return "", err
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	key := sha256.Sum256([]byte(provider.parser.Name() + "\x00" + absolute + "\x00" + table))

	return filepath.Join(dir, "ai-sqlite", fmt.Sprintf("%x.json", key)), nil
}

func prepareGenericAISQLiteCursors(
	ctx context.Context, provider genericAIProvider, path string, tables []string,
) error {
	// Save the cutoff for all tables before any table can exhaust the budget.
	var pending, completed []string

	for _, table := range tables {
		cursorPath, err := genericAISQLiteCursorPath(ctx, provider, path, table)
		if err != nil {
			return err
		}

		cursor, err := provider.readCheckpointCursor(cursorPath)
		if err != nil {
			return err
		}

		if cursor.Version != genericAISQLiteCursorVersion || cursor.ParserVersion != provider.sqliteParserVersion {
			cutoff := provider.config.After
			if (cursor.Pending || cursor.Rescan) && cursor.Cutoff.Before(cutoff) {
				cutoff = cursor.Cutoff
			}

			cursor = genericAISQLiteCursor{
				Version: genericAISQLiteCursorVersion, ParserVersion: provider.sqliteParserVersion,
				Cutoff: cutoff, Pending: true,
			}
			if err := provider.writeCheckpointCursor(cursorPath, cursor); err != nil {
				return err
			}
		}

		if cursor.Pending {
			pending = append(pending, table)
		} else {
			completed = append(completed, table)
		}
	}
	// A busy, already completed table must not repeatedly consume the budget
	// before later tables get to finish their first scan.
	copy(tables, append(pending, completed...))

	return nil
}

func readGenericAISQLiteCursor(path string) (genericAISQLiteCursor, error) {
	contents, err := os.ReadFile(filepath.Clean(path))
	if os.IsNotExist(err) {
		return genericAISQLiteCursor{}, nil
	}

	if err != nil {
		return genericAISQLiteCursor{}, err
	}

	var cursor genericAISQLiteCursor

	err = json.Unmarshal(contents, &cursor)

	return cursor, err
}

func writeGenericAISQLiteCursor(path string, cursor genericAISQLiteCursor) error {
	contents, err := json.Marshal(cursor)
	if err != nil {
		return err
	}

	return atomicWriteFile(filepath.Dir(path), ".cursor-*", path, contents)
}

// Fingerprint both files: WAL commits need not touch the database file. Include
// header and tail bytes as well as file times to handle reused WAL files and
// filesystems whose modification times have coarse resolution.
func genericAISQLiteFingerprint(path string) (string, error) {
	digest := sha256.New()

	for _, name := range []string{path, path + "-wal"} {
		file, err := os.Open(filepath.Clean(name))
		if os.IsNotExist(err) {
			continue
		}

		if err != nil {
			return "", err
		}

		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return "", err
		}

		_, _ = fmt.Fprintf(digest, "%s:%d:%d:", name, info.Size(), info.ModTime().UnixNano())

		_, err = io.Copy(digest, io.NewSectionReader(file, 0, 100))
		if err == nil && info.Size() > 100 {
			_, err = io.Copy(digest, io.NewSectionReader(file, max(100, info.Size()-64), 64))
		}

		_ = file.Close()

		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", digest.Sum(nil)), nil
}

func parseGenericAISQLiteIncremental(
	ctx context.Context, provider genericAIProvider, db *sql.DB, path, table, rowID string,
) (Heartbeats, error) {
	cursorPath, err := genericAISQLiteCursorPath(ctx, provider, path, table)
	if err != nil {
		return nil, err
	}

	cursor, err := provider.readCheckpointCursor(cursorPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading sqlite cursor: %w", err)
	}

	fingerprint, err := genericAISQLiteFingerprint(path)
	if err != nil {
		return nil, err
	}

	if cursor.Version != genericAISQLiteCursorVersion || cursor.ParserVersion != provider.sqliteParserVersion ||
		provider.config.After.Before(cursor.Cutoff) {
		cursor = genericAISQLiteCursor{
			Version: genericAISQLiteCursorVersion, ParserVersion: provider.sqliteParserVersion,
			Cutoff: provider.config.After,
		}
	}

	if !cursor.Pending && cursor.Fingerprint == fingerprint {
		return nil, nil
	}

	if !cursor.Pending {
		cutoff := provider.config.After
		if cursor.Rescan {
			cutoff = cursor.Cutoff
		}

		cursor = genericAISQLiteCursor{
			Version: genericAISQLiteCursorVersion, ParserVersion: provider.sqliteParserVersion,
			Cutoff: cutoff, Pending: true,
			Hashes: cursor.Hashes,
			State:  genericAIParseState{Records: cursor.State.Records},
		}
	}

	if cursor.Fingerprint == "" {
		cursor.Fingerprint = fingerprint
	}

	if cursor.ScanHashes == nil {
		cursor.ScanHashes = make(map[string]string)
	}

	provider.config.After = cursor.Cutoff
	provider.incremental = provider.config.checkpoint != nil

	order, keys, err := genericAISQLiteOrder(ctx, db, table, rowID)
	if err != nil {
		return nil, err
	}

	var singleRow bool

	query := "SELECT NOT EXISTS(SELECT 1 FROM " + aiSQLiteQuote(table) + " LIMIT 1 OFFSET 1)" // nolint:gosec
	if err := db.QueryRowContext(ctx, query).Scan(&singleRow); err != nil {
		return nil, err
	}

	var (
		heartbeats Heartbeats
		parseErr   error
	)

	for cursor.Pending {
		parsed, done, err := genericAISQLiteBatch(ctx, provider, db, path, table, order, rowID, keys, singleRow, &cursor)
		heartbeats = append(heartbeats, parsed...)

		if err != nil {
			parseErr = err
			break
		}

		if done {
			cursor.Pending = false
			cursor.Hashes = cursor.ScanHashes
			cursor.ScanHashes = nil

			current, err := genericAISQLiteFingerprint(path)
			if err != nil {
				parseErr = err
				break
			}
			// A writer can change an already scanned row during a multi-run scan.
			// Revisit it with this scan's cutoff, even if the global cutoff advanced.
			cursor.Rescan = current != cursor.Fingerprint
		}
	}

	if err := provider.writeCheckpointCursor(cursorPath, cursor); err != nil {
		return nil, fmt.Errorf("failed saving sqlite cursor: %w", err)
	}

	if provider.config.checkpoint != nil {
		provider.config.checkpoint.markIncremental(heartbeats)
	}

	return heartbeats, parseErr
}

// WITHOUT ROWID tables are stored in primary-key order. Explicitly order all
// other fallback scans as well so OFFSET resumes deterministically.
func genericAISQLiteOrder(ctx context.Context, db *sql.DB, table, rowID string) (string, []string, error) {
	if rowID != "" {
		return rowID, nil, nil
	}

	rows, err := db.QueryContext(ctx, "SELECT name FROM pragma_table_info(?) WHERE pk > 0 ORDER BY pk", table)
	if err != nil {
		return "", nil, err
	}

	var keys []string

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			_ = rows.Close()
			return "", nil, err
		}

		keys = append(keys, key)
	}

	err = rows.Err()
	_ = rows.Close()

	if err != nil {
		return "", nil, err
	}

	columns := slices.Clone(keys)
	if len(columns) == 0 {
		// All rowid aliases can be shadowed even without a declared primary key.
		rows, err := db.QueryContext(ctx, "SELECT * FROM "+aiSQLiteQuote(table)+" LIMIT 0") // nolint:gosec
		if err != nil {
			return "", nil, err
		}

		columns, err = rows.Columns()
		_ = rows.Close()

		if err != nil {
			return "", nil, err
		}
	}

	for i, column := range columns {
		columns[i] = aiSQLiteQuote(column)
	}

	return strings.Join(columns, ", "), keys, nil
}

func genericAISQLiteBatch(
	ctx context.Context, provider genericAIProvider, db *sql.DB, path, table, order, rowID string,
	keys []string, singleRow bool, cursor *genericAISQLiteCursor,
) (Heartbeats, bool, error) {
	projection := "*"
	if rowID != "" {
		projection = rowID + ", *"
	}

	query := "SELECT " + projection + " FROM " + aiSQLiteQuote(table) + " WHERE 1=1" // nolint:gosec

	var args []any

	if provider.sqliteQuery != nil {
		query, args = provider.sqliteQuery(table, projection, provider.config.After)
	}

	if rowID != "" && cursor.RowID != nil {
		query += " AND " + rowID + " > ?"

		args = append(args, *cursor.RowID)
	}

	query += " ORDER BY " + order + " LIMIT ?"

	args = append(args, aiSQLiteBatchSize)

	if rowID == "" {
		query += " OFFSET ?"

		args = append(args, cursor.Offset)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close() // nolint:errcheck

	columns, err := rows.Columns()
	if err != nil {
		return nil, false, err
	}

	if rowID != "" {
		columns[0] = "\x00rowid"
	}

	var heartbeats Heartbeats

	count := 0

	for rows.Next() {
		row, err := genericAISQLiteRawRow(ctx, rows, columns)
		if err != nil {
			return heartbeats, false, err
		}

		key := fmt.Sprint(cursor.Offset)

		var id int64

		if rowID != "" {
			var ok bool

			id, ok = row[columns[0]].(int64)
			if !ok {
				return heartbeats, false, fmt.Errorf("invalid sqlite rowid")
			}

			key = fmt.Sprint(id)

			delete(row, columns[0])
		}

		if len(keys) > 0 {
			values := make([]any, 0, len(keys))
			for _, column := range keys {
				values = append(values, row[column])
			}

			encoded, err := json.Marshal(values)
			if err != nil {
				return heartbeats, false, err
			}

			key = string(encoded)
		}

		digest := sha256.New()
		if err := json.NewEncoder(digest).Encode(row); err != nil {
			return heartbeats, false, err
		}

		hash := fmt.Sprintf("%x", digest.Sum(nil))

		parsed, err := genericAISQLiteRowHeartbeats(provider, path, table, row, singleRow, &cursor.State)
		if err != nil {
			return heartbeats, false, err
		}

		if cursor.Hashes[key] != hash && cursor.ScanHashes[key] != hash {
			heartbeats = append(heartbeats, parsed...)
		}

		cursor.ScanHashes[key] = hash
		// Only checkpoint fully parsed rows. A later budget/deadline error returns
		// these heartbeats together with the matching checkpoint.
		cursor.Offset++
		if rowID != "" {
			cursor.RowID = &id
		}

		count++
	}

	return heartbeats, count < aiSQLiteBatchSize, rows.Err()
}

func genericAISQLiteRowHeartbeats(
	provider genericAIProvider, path, table string, row map[string]any, singleRow bool, state *genericAIParseState,
) (Heartbeats, error) {
	if provider.sqliteEvents != nil {
		events := provider.sqliteEvents(table, row)
		return genericAIHeartbeatsFromEvents(context.Background(), provider, path, slices.Values(events), state)
	}

	values := []any{genericAISQLiteDecodeRow(row)}
	if provider.containerKey != "" {
		values = genericAISQLiteContainerValues(values, provider.containerKey)
	}

	events := make([]genericAIEvent, 0, len(values))
	for _, value := range values {
		events = append(events, genericAIEventFromValue(value))
	}

	if singleRow && len(events) == 1 && events[0].timestamp.IsZero() {
		if info, err := os.Stat(path); err == nil {
			events[0].timestamp = info.ModTime()
		}
	}
	// Finish this size-limited row atomically with its checkpoint. The scan checks
	// cancellation before reading the next row, including within each SQL batch.
	return genericAIHeartbeatsFromEvents(context.Background(), provider, path, slices.Values(events), state)
}

func (p genericAIProvider) readCheckpointCursor(path string) (genericAISQLiteCursor, error) {
	if p.config.checkpoint == nil {
		return readGenericAISQLiteCursor(path)
	}

	if raw, ok := p.config.checkpoint.Cursors[path]; ok {
		var cursor genericAISQLiteCursor
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &cursor); err != nil {
				return cursor, err
			}
		}

		return cursor, nil
	}

	// The checkpoint has never seen this cursor before: fall back to the
	// legacy on-disk cursor file so upgrading does not force a full rescan
	// (and resend) of rows already scanned by a pre-checkpoint build.
	cursor, err := readGenericAISQLiteCursor(path)
	if err != nil || cursor.Version == 0 {
		return cursor, err
	}

	// Adopt the legacy cursor into the checkpoint right away. It is otherwise
	// only written back when the database changed, which would leave an
	// unchanged store reading the legacy file forever.
	state := p.config.checkpoint
	if state.legacy == nil {
		state.legacy = make(map[string]struct{})
	}

	state.legacy[path] = struct{}{}

	return cursor, p.writeCheckpointCursor(path, cursor)
}

func (p genericAIProvider) writeCheckpointCursor(path string, cursor genericAISQLiteCursor) error {
	if p.config.checkpoint == nil {
		return writeGenericAISQLiteCursor(path, cursor)
	}

	raw, err := json.Marshal(cursor)
	if err != nil {
		return err
	}

	state := p.config.checkpoint
	if state.Cursors == nil {
		state.Cursors = make(map[string]json.RawMessage)
	}

	state.Cursors[path] = raw

	return nil
}
