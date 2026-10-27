//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAntigravitySQLiteModifiedAfter(t *testing.T) {
	cutoff := time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC)

	for _, product := range antigravityProductDirs() {
		for _, directory := range []string{"conversations", "implicit"} {
			t.Run(product.appDir+"/"+directory, func(t *testing.T) {
				for _, tc := range []struct {
					name     string
					after    time.Time
					modified time.Time
					wal      bool
					wantRead bool
				}{
					{"historical", cutoff, cutoff.Add(-time.Hour), false, false},
					{"recent", cutoff, cutoff.Add(time.Hour), false, true},
					{"at cutoff", cutoff, cutoff, false, true},
					{"no cutoff", time.Time{}, cutoff.Add(-time.Hour), false, true},
					{"recent WAL", cutoff, cutoff.Add(-time.Hour), true, true},
				} {
					t.Run(tc.name, func(t *testing.T) {
						home := updateTestHome(t)
						path := filepath.Join(home, ".gemini", product.appDir, directory, "one.db")
						db := updateDB(t, path)
						if tc.wal {
							updateSQL(t, db, `PRAGMA journal_mode = WAL`)
						}
						updateSQL(t, db, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB)`)
						updateSQL(t, db, `INSERT INTO gen_metadata VALUES(1, ?)`, strings.Repeat("x", 2048))
						if !tc.wal {
							require.NoError(t, db.Close())
						}
						require.NoError(t, os.Chtimes(path, tc.modified, tc.modified))
						if tc.wal {
							recent := cutoff.Add(time.Hour)
							require.NoError(t, os.Chtimes(path+"-wal", recent, recent))
						}

						// Leave only 1 KiB to prove historical payloads never consume the shared budget.
						budget := &aiSQLiteBudget{bytes: aiSQLiteByteLimit - 1024}
						ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{}, budget)
						_, err := (Gemini{After: tc.after}).antigravitySQLite(ctx, home)
						if tc.wantRead {
							require.ErrorContains(t, err, "sqlite parsing exceeded byte budget")
						} else {
							require.NoError(t, err)
							assert.EqualValues(t, aiSQLiteByteLimit-1024, budget.bytes)
						}
					})
				}
			})
		}
	}
}
