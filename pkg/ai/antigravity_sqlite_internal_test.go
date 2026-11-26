//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"encoding/binary"
	"errors"
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

func testProtoTimestamp(t time.Time) []byte {
	ts := updateProtoInt(1, uint64(t.Unix()))
	if nanos := t.Nanosecond(); nanos > 0 {
		ts = append(ts, updateProtoInt(2, uint64(nanos))...)
	}

	return ts
}

func testProtoUvarints(values ...uint64) []byte {
	var out []byte
	for _, v := range values {
		out = binary.AppendUvarint(out, v)
	}

	return out
}

func testAntigravityGenRow(id string, input, output uint64, stepIndices ...uint64) []byte {
	usage := append(updateProtoInt(2, input), updateProtoInt(3, output)...)
	usage = append(usage, updateProtoBytes(11, []byte(id))...)

	chat := append(updateProtoBytes(4, usage), updateProtoBytes(19, []byte("gemini-3"))...)
	// Antigravity CLI 1.5.9 omits metadata field 4 (timestamp).
	chat = append(chat, updateProtoBytes(9, updateProtoInt(2, 1))...)

	return append(updateProtoBytes(1, chat), updateProtoBytes(2, testProtoUvarints(stepIndices...))...)
}

func TestAntigravitySQLiteStepTimestampsAndCLITools(t *testing.T) {
	home := updateTestHome(t)
	path := filepath.Join(home, ".gemini", "antigravity-cli", "conversations", "session-cli.db")
	db := updateDB(t, path)
	updateSQL(t, db, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB); CREATE TABLE steps(idx INTEGER, metadata BLOB);`)

	t1 := time.Date(2026, 11, 25, 12, 0, 0, 0, time.UTC)
	t1Tool := t1.Add(2 * time.Second)
	t2 := time.Date(2026, 11, 26, 1, 30, 0, 0, time.UTC)
	t2Write := t2.Add(3 * time.Second)
	t2Edit := t2.Add(5 * time.Second)

	step1 := updateProtoBytes(1, testProtoTimestamp(t1))

	readTool := append(
		updateProtoBytes(2, []byte("read_file")),
		updateProtoBytes(3, []byte(`{"Path":"/project/readme.md"}`))...,
	)
	step2 := append(updateProtoBytes(1, testProtoTimestamp(t1Tool)), updateProtoBytes(4, readTool)...)

	// Step 3 uses field 7 (completion timestamp) without field 1.
	step3 := updateProtoBytes(7, testProtoTimestamp(t2))

	writeTool := append(
		updateProtoBytes(2, []byte("write_file")),
		updateProtoBytes(3, []byte(`{"Path":"/project/new.go","Content":"package main\n\nfunc main() {}"}`))...,
	)
	step4 := append(updateProtoBytes(1, testProtoTimestamp(t2Write)), updateProtoBytes(4, writeTool)...)

	editTool := append(
		updateProtoBytes(2, []byte("edit_file")),
		updateProtoBytes(3, []byte(`{"Path":"/project/app.go","Target":"a\nb","Replacement":"a\nb\nc\nd"}`))...,
	)
	step5 := append(updateProtoBytes(1, testProtoTimestamp(t2Edit)), updateProtoBytes(4, editTool)...)

	updateSQL(t, db, `INSERT INTO steps VALUES (1,?),(2,?),(3,?),(4,?),(5,?)`, step1, step2, step3, step4, step5)
	updateSQL(t, db, `INSERT INTO gen_metadata VALUES (0,?),(1,?)`,
		testAntigravityGenRow("gen-1", 100, 20, 1, 2),
		testAntigravityGenRow("gen-2", 200, 40, 3, 4, 5),
	)
	require.NoError(t, db.Close())

	hh, err := (Gemini{After: time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)}).antigravitySQLite(t.Context(), home)
	require.NoError(t, err)

	var (
		turnHeartbeats []heartbeat.Heartbeat
		fileHeartbeats []heartbeat.Heartbeat
	)

	for _, h := range hh {
		switch {
		case h.EntityType == heartbeat.FileType:
			fileHeartbeats = append(fileHeartbeats, h)
		case h.AIInputTokens > 0:
			turnHeartbeats = append(turnHeartbeats, h)
		}
	}

	require.Len(t, turnHeartbeats, 2)
	assert.Equal(t, "Antigravity CLI", turnHeartbeats[0].Entity)
	assert.InDelta(t, float64(t1.Unix()), turnHeartbeats[0].Time, 0.001)
	assert.EqualValues(t, 100, turnHeartbeats[0].AIInputTokens)
	assert.Equal(t, "Antigravity CLI", turnHeartbeats[1].Entity)
	assert.InDelta(t, float64(t2.Unix()), turnHeartbeats[1].Time, 0.001)
	assert.EqualValues(t, 200, turnHeartbeats[1].AIInputTokens)

	require.Len(t, fileHeartbeats, 3)
	assert.Equal(t, "/project/readme.md", fileHeartbeats[0].Entity)
	assert.InDelta(t, float64(t1Tool.Unix()), fileHeartbeats[0].Time, 0.001)
	require.NotNil(t, fileHeartbeats[0].IsWrite)
	assert.False(t, *fileHeartbeats[0].IsWrite)
	assert.Nil(t, fileHeartbeats[0].AILineChanges)

	assert.Equal(t, "/project/new.go", fileHeartbeats[1].Entity)
	assert.InDelta(t, float64(t2Write.Unix()), fileHeartbeats[1].Time, 0.001)
	require.NotNil(t, fileHeartbeats[1].IsWrite)
	assert.True(t, *fileHeartbeats[1].IsWrite)
	require.NotNil(t, fileHeartbeats[1].AILineChanges)
	assert.Equal(t, 3, *fileHeartbeats[1].AILineChanges)

	assert.Equal(t, "/project/app.go", fileHeartbeats[2].Entity)
	assert.InDelta(t, float64(t2Edit.Unix()), fileHeartbeats[2].Time, 0.001)
	require.NotNil(t, fileHeartbeats[2].IsWrite)
	assert.True(t, *fileHeartbeats[2].IsWrite)
	require.NotNil(t, fileHeartbeats[2].AILineChanges)
	assert.Equal(t, 2, *fileHeartbeats[2].AILineChanges)

	// Cutoff on day 2 must skip turn 1 while still emitting turn 2 and its edits.
	hhDay2, err := (Gemini{After: time.Date(2026, 11, 26, 0, 0, 0, 0, time.UTC)}).antigravitySQLite(t.Context(), home)
	require.NoError(t, err)

	var (
		day2Turns []heartbeat.Heartbeat
		day2Files []heartbeat.Heartbeat
	)

	for _, h := range hhDay2 {
		switch {
		case h.EntityType == heartbeat.FileType:
			day2Files = append(day2Files, h)
		case h.AIInputTokens > 0:
			day2Turns = append(day2Turns, h)
		}
	}

	require.Len(t, day2Turns, 1)
	assert.InDelta(t, float64(t2.Unix()), day2Turns[0].Time, 0.001)
	require.Len(t, day2Files, 2)
	assert.Equal(t, "/project/new.go", day2Files[0].Entity)
	assert.Equal(t, "/project/app.go", day2Files[1].Entity)

	// Small non-timestamp varints must not be misread as 1970 timestamps.
	assert.True(t, antigravityStepTime(updateProtoBytes(1, updateProtoInt(1, 1))).IsZero())
	assert.True(t, antigravityStepTime(nil).IsZero())
}

func TestAntigravitySQLiteSkipsLockedDatabase(t *testing.T) {
	home := updateTestHome(t)
	convDir := filepath.Join(home, ".gemini", "antigravity-cli", "conversations")

	lockedPath := filepath.Join(convDir, "a-locked.db")
	lockedDB := updateDB(t, lockedPath)
	updateSQL(t, lockedDB, `CREATE TABLE gen_metadata(idx INTEGER, data BLOB)`)
	updateSQL(t, lockedDB, `BEGIN EXCLUSIVE`)
	t.Cleanup(func() {
		_, _ = lockedDB.Exec(`ROLLBACK`)
	})

	validPath := filepath.Join(convDir, "b-valid.db")
	validDB := updateDB(t, validPath)
	updateSQL(t, validDB,
		`CREATE TABLE gen_metadata(idx INTEGER, data BLOB); CREATE TABLE steps(idx INTEGER, metadata BLOB);`)

	ts := time.Date(2026, 11, 26, 2, 0, 0, 0, time.UTC)
	updateSQL(t, validDB, `INSERT INTO steps VALUES (1, ?)`, updateProtoBytes(1, testProtoTimestamp(ts)))
	updateSQL(t, validDB, `INSERT INTO gen_metadata VALUES (0, ?)`, testAntigravityGenRow("gen-valid", 50, 10, 1))
	require.NoError(t, validDB.Close())

	hh, err := (Gemini{}).antigravitySQLite(t.Context(), home)
	require.NoError(t, err)
	require.Len(t, hh, 1)
	assert.Equal(t, "b-valid", hh[0].AISession)
	assert.EqualValues(t, 50, hh[0].AIInputTokens)
	assert.False(t, isAntigravitySkippableSQLiteError(errors.New("non-sqlite error")))
}
