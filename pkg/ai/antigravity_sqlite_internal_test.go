//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"bytes"
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
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

func TestAntigravitySQLiteSkipsOversizedRows(t *testing.T) {
	for _, tc := range []struct {
		name string
		data any
	}{
		{"blob", bytes.Repeat([]byte("x"), maxTranscriptLineSize+1)},
		{"unicode", strings.Repeat("界", maxTranscriptLineSize/3+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := updateTestHome(t)
			dir := filepath.Join(home, ".gemini", "antigravity", "conversations")
			onePath := filepath.Join(dir, "one.db")

			db := updateDB(t, onePath)
			updateSQL(t, db, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB);
CREATE TABLE steps(idx INTEGER, metadata BLOB);`)

			rootBefore := antigravityTestRoot("response-before", 100, 20, "2026-09-20T10:00:00Z", 1)
			rootAfter := antigravityTestRoot("response-after", 50, 10, "2026-09-20T10:01:00Z", 2, 3)
			stepBefore := antigravityTestStep("/project/before.go")
			stepAfter := antigravityTestStep("/project/after.go")

			_, err := db.ExecContext(t.Context(), `INSERT INTO gen_metadata VALUES (1, ?), (2, ?), (3, ?)`,
				rootBefore, tc.data, rootAfter)
			require.NoError(t, err)

			_, err = db.ExecContext(t.Context(), `INSERT INTO steps VALUES (1, ?), (2, ?), (3, ?)`,
				stepBefore, tc.data, stepAfter)
			require.NoError(t, err)
			require.NoError(t, db.Close())

			twoPath := filepath.Join(dir, "two.db")
			dbTwo := updateDB(t, twoPath)
			updateSQL(t, dbTwo, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB);`)

			rootTwo := antigravityTestRoot("response-two", 30, 5, "2026-09-20T10:02:00Z")
			_, err = dbTwo.ExecContext(t.Context(), `INSERT INTO gen_metadata VALUES (1, ?)`, rootTwo)
			require.NoError(t, err)
			require.NoError(t, dbTwo.Close())

			hh, err := (Gemini{}).antigravitySQLite(t.Context(), home)
			require.NoError(t, err)

			totals := updateTotals(hh)
			assert.EqualValues(t, 180, totals.input)
			assert.EqualValues(t, 35, totals.output)

			var files []string

			for _, h := range hh {
				if h.EntityType == heartbeat.FileType {
					files = append(files, h.Entity)
				}
			}

			assert.Equal(t, []string{"/project/before.go", "/project/after.go"}, files)
		})
	}
}

func TestAntigravitySQLitePreservesMaxPayloadRow(t *testing.T) {
	home := updateTestHome(t)
	dir := filepath.Join(home, ".gemini", "antigravity", "conversations")
	onePath := filepath.Join(dir, "one.db")

	db := updateDB(t, onePath)
	updateSQL(t, db, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB);
CREATE TABLE steps(idx INTEGER, metadata BLOB);`)

	rootBefore := antigravityTestRoot("response-before", 100, 20, "2026-09-20T10:00:00Z", 1)
	// Payload fits within maxTranscriptLineSize, but its encoded SQLite record overflows
	// an unadjusted SQLITE_LIMIT_LENGTH on the ordered query.
	rootBoundary := antigravityTestPaddedRoot(
		t, "response-boundary", 70, 15, "2026-09-20T10:00:30Z", maxTranscriptLineSize-1,
	)
	rootAfter := antigravityTestRoot("response-after", 50, 10, "2026-09-20T10:01:00Z", 2)
	stepBefore := antigravityTestStep("/project/before.go")
	stepAfter := antigravityTestStep("/project/after.go")

	_, err := db.ExecContext(t.Context(), `INSERT INTO gen_metadata VALUES (1, ?), (2, ?), (3, ?)`,
		rootBefore, rootBoundary, rootAfter)
	require.NoError(t, err)

	_, err = db.ExecContext(t.Context(), `INSERT INTO steps VALUES (1, ?), (2, ?)`, stepBefore, stepAfter)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	// Confirm the original connection limit fails on this row with "too big".
	origDB, err := openAISQLiteDB(t.Context(), onePath)
	require.NoError(t, err)

	origRows, origErr := origDB.QueryContext(t.Context(), `SELECT idx, data FROM gen_metadata ORDER BY idx`)
	if origErr == nil {
		for origRows.Next() {
			var (
				idx int64
				raw []byte
			)

			if err := origRows.Scan(&idx, &raw); err != nil {
				origErr = err
				break
			}
		}

		if origErr == nil {
			origErr = origRows.Err()
		}

		_ = origRows.Close()
	}

	require.NoError(t, origDB.Close())
	require.ErrorContains(t, origErr, "too big")

	hh, err := (Gemini{}).antigravitySQLite(t.Context(), home)
	require.NoError(t, err)

	totals := updateTotals(hh)
	assert.EqualValues(t, 220, totals.input)
	assert.EqualValues(t, 45, totals.output)

	var files []string

	for _, h := range hh {
		if h.EntityType == heartbeat.FileType {
			files = append(files, h.Entity)
		}
	}

	assert.Equal(t, []string{"/project/before.go", "/project/after.go"}, files)
}

func antigravityTestPaddedRoot(t *testing.T, id string, input, output uint64, stamp string, totalSize int) []byte {
	t.Helper()

	base := antigravityTestRoot(id, input, output, stamp)
	tag := binary.AppendUvarint(nil, 99<<3|2)

	padLen := totalSize - len(base) - len(tag)
	for headerLen := len(binary.AppendUvarint(nil, uint64(padLen))); len(base)+len(tag)+headerLen+padLen != totalSize; {
		padLen = totalSize - len(base) - len(tag) - headerLen
		headerLen = len(binary.AppendUvarint(nil, uint64(padLen)))
	}

	base = append(base, updateProtoBytes(99, bytes.Repeat([]byte("p"), padLen))...)
	require.Len(t, base, totalSize)

	return base
}

func antigravityTestRoot(id string, input, output uint64, stamp string, steps ...uint64) []byte {
	usage := append(updateProtoInt(2, input), updateProtoInt(3, output)...)
	usage = append(usage, updateProtoBytes(11, []byte(id))...)

	chat := append(updateProtoBytes(4, usage), updateProtoBytes(19, []byte("gemini-3"))...)
	chat = append(chat, updateProtoBytes(9, updateProtoBytes(4, []byte(stamp)))...)

	root := updateProtoBytes(1, chat)
	for _, step := range steps {
		root = append(root, updateProtoInt(2, step)...)
	}

	return root
}

func antigravityTestStep(file string) []byte {
	args := []byte(`{"TargetFile":"` + file + `","CodeContent":"one\ntwo"}`)
	tool := append(updateProtoBytes(2, []byte("write_to_file")), updateProtoBytes(3, args)...)

	return updateProtoBytes(4, tool)
}
