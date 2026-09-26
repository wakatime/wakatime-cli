//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

const (
	aiSQLiteTimeout   = 5 * time.Second
	aiSQLiteByteLimit = 64 * 1024 * 1024
)

type aiSQLiteBudgetKey struct{}

type aiSQLiteBudget struct {
	bytes  int64
	spent  time.Duration
	active bool
}

func openAISQLiteDB(ctx context.Context, path string) (*sql.DB, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// URI escaping keeps filename characters such as ? and # out of query options.
	// mode=ro still reads committed WAL data; immutable would ignore live writes.
	uriPath := filepath.ToSlash(absolute)
	if !strings.HasPrefix(uriPath, "/") {
		uriPath = "/" + uriPath
	}

	source := url.URL{Scheme: "file", Path: uriPath, RawQuery: "mode=ro"}

	db, err := sql.Open("sqlite", source.String())
	if err != nil {
		return nil, err
	}

	// Limits belong to a connection. Reuse the configured connection for every
	// query, including JSON functions executed inside SQLite before rows.Scan.
	db.SetMaxOpenConns(1)

	conn, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	_, err = sqlite.Limit(conn, sqlite3.SQLITE_LIMIT_LENGTH, maxTranscriptLineSize)

	_ = conn.Close()
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func aiSQLiteBudgetContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, aiSQLiteBudgetKey{}, &aiSQLiteBudget{})
}

// Share the work budget across SQLite parsers, without cancelling unrelated
// parsers or the editor heartbeat when the SQLite time allowance runs out.
func aiSQLiteContext(ctx context.Context) (context.Context, context.CancelFunc) {
	budget, ok := ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget)
	if !ok {
		ctx = aiSQLiteBudgetContext(ctx)
		budget = ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget)
	}

	if budget.active {
		return context.WithCancel(ctx) // nolint:gosec // The caller owns and defers the returned cancellation function.
	}

	budget.active = true
	started := time.Now()
	// nolint:gosec // The returned closure cancels this context and accounts for elapsed parsing time.
	ctx, cancel := context.WithTimeout(ctx, aiSQLiteTimeout-budget.spent)
	done := false

	return ctx, func() {
		cancel()

		if !done {
			budget.spent += time.Since(started)
			budget.active = false
			done = true
		}
	}
}

func aiSQLiteRead(ctx context.Context, values ...string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var size int64
	for _, value := range values {
		size += int64(len(value))
	}

	if size > maxTranscriptLineSize {
		return fmt.Errorf("sqlite row exceeds transcript size limit (%d bytes)", maxTranscriptLineSize)
	}

	if budget, ok := ctx.Value(aiSQLiteBudgetKey{}).(*aiSQLiteBudget); ok {
		budget.bytes += size
		if budget.bytes > aiSQLiteByteLimit {
			return fmt.Errorf("sqlite parsing exceeded byte budget (%d bytes)", aiSQLiteByteLimit)
		}
	}

	return nil
}

func aiSQLiteModifiedAfter(path string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	// Writers can commit to the WAL without changing the database's mtime.
	for _, filename := range []string{path, path + "-wal"} {
		if info, err := os.Stat(filename); err == nil && timestampAtOrAfterCutoff(info.ModTime(), after) {
			return true
		}
	}

	return false
}
