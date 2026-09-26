//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncCheckpointSQLiteAtomicProgress(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-01T12:00:00Z","session_id":"s","input_tokens":10}`)
	require.NoError(t, err)

	after := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "internal.cfg"))

	parse := func() (*syncCheckpoints, Heartbeats) {
		state, err := loadSyncCheckpoints(t.Context(), v, after)
		require.NoError(t, err)

		p := state.parser("ZCode", after)
		provider := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: p.discoveryAfter(), checkpoint: p}}
		got, err := parseGenericAISQLiteDB(t.Context(), provider, path)
		require.NoError(t, err)

		return state, p.filter(got)
	}
	state, got := parse()
	require.Len(t, got, 1)
	assert.EqualValues(t, 10, got[0].AIInputTokens)
	// A failed checkpoint commit must not advance the SQLite row cursor.
	state.path = t.TempDir()
	require.Error(t, state.save())
	state, got = parse()
	require.Len(t, got, 1)
	require.NoError(t, state.save())

	_, got = parse()
	assert.Empty(t, got)
	// Profiles have independent row cursors as well as independent timestamps.
	v.Set("internal-config", filepath.Join(t.TempDir(), "other.cfg"))

	_, got = parse()
	require.Len(t, got, 1)
}

func TestSyncCheckpointSQLiteMigratesLegacyCursor(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO events VALUES (?)`,
		`{"timestamp":"2026-09-01T12:00:00Z","session_id":"s","input_tokens":10}`)
	require.NoError(t, err)

	legacyAfter := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	upgradeAfter := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)

	// A pre-checkpoint build (no checkpoint field set) scans and checkpoints
	// this row to the legacy on-disk cursor file.
	legacy := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: legacyAfter}}
	got, err := parseGenericAISQLiteDB(t.Context(), legacy, path)
	require.NoError(t, err)
	require.Len(t, got, 1)

	// The global watermark keeps advancing (other syncs, other parsers)
	// before the process is upgraded to checkpoint tracking. Upgrading must
	// migrate the legacy cursor's progress in rather than force a full
	// rescan (and resend) of the same, unchanged row.
	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "internal.cfg"))
	state, err := loadSyncCheckpoints(t.Context(), v, upgradeAfter)
	require.NoError(t, err)

	p := state.parser("ZCode", upgradeAfter)
	upgraded := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: p.discoveryAfter(), checkpoint: p}}
	got, err = parseGenericAISQLiteDB(t.Context(), upgraded, path)
	require.NoError(t, err)
	assert.Empty(t, p.filter(got), "upgrading must not resend a row already scanned by the legacy cursor file")

	// The legacy file is only removed once the checkpoint holding its state is
	// saved, so a lost checkpoint cannot resurrect stale progress.
	cursorPath, err := genericAISQLiteCursorPath(t.Context(), upgraded, path, "events")
	require.NoError(t, err)
	require.FileExists(t, cursorPath)
	require.NoError(t, state.save())
	assert.NoFileExists(t, cursorPath)
	assert.Contains(t, p.Cursors, cursorPath)
}

func TestSyncCheckpointSQLiteQuietSessionDoesNotResendRows(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)

	insert := func(payload string) {
		_, err := db.Exec(`INSERT INTO events VALUES (?)`, payload)
		require.NoError(t, err)
	}
	insert(`{"timestamp":"2026-09-01T12:00:00Z","session_id":"s","input_tokens":10}`)

	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "internal.cfg"))

	// Like production, each run starts from the newest activity already emitted.
	global := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	sync := func() int64 {
		state, err := loadSyncCheckpoints(t.Context(), v, global)
		require.NoError(t, err)

		p := state.parser("ZCode", global)
		provider := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: p.discoveryAfter(), checkpoint: p}}
		got, err := parseGenericAISQLiteDB(t.Context(), provider, path)
		require.NoError(t, err)

		var tokens int64

		for _, h := range p.filter(got) {
			tokens += h.AIInputTokens

			if stamp := heartbeatTime(h.Time); stamp.After(global) {
				global = stamp
			}
		}

		require.NoError(t, state.save())

		return tokens
	}

	assert.EqualValues(t, 10, sync())

	// A new session that never emits anything appears in the same store.
	insert(`{"timestamp":"2026-09-01T12:10:00Z","session_id":"quiet"}`)
	assert.EqualValues(t, 0, sync())

	insert(`{"timestamp":"2026-09-01T12:30:00Z","session_id":"s","input_tokens":5}`)
	assert.EqualValues(t, 5, sync(), "rows already emitted must not be sent again")
}

func TestSyncCheckpointSQLitePendingOlderRows(t *testing.T) {
	db, path := sqliteTestDB(t)
	_, err := db.Exec(`CREATE TABLE events (payload TEXT)`)
	require.NoError(t, err)

	latest := `{"timestamp":"2026-09-01T12:45:00Z","session_id":"s","input_tokens":10}`
	earlier := `{"timestamp":"2026-09-01T12:30:00Z","session_id":"s","input_tokens":20}`
	_, err = db.Exec(`INSERT INTO events VALUES (?), (?)`, latest, earlier)
	require.NoError(t, err)

	after := time.Date(2026, 9, 1, 11, 0, 0, 0, time.UTC)
	v := viper.New()
	v.Set("internal-config", filepath.Join(t.TempDir(), "internal.cfg"))
	state, err := loadSyncCheckpoints(t.Context(), v, after)
	require.NoError(t, err)

	p := state.parser("ZCode", after)
	provider := genericAIProvider{parser: ZCode{}, config: ParserConfig{After: p.discoveryAfter(), checkpoint: p}}
	ctx := context.WithValue(t.Context(), aiSQLiteBudgetKey{},
		&aiSQLiteBudget{bytes: aiSQLiteByteLimit - int64(len(latest))})
	got, err := parseGenericAISQLiteIncremental(ctx, provider, db, path, "events", "rowid")
	require.ErrorContains(t, err, "byte budget")
	require.Len(t, p.filter(got), 1)
	require.NoError(t, state.save())

	state, err = loadSyncCheckpoints(t.Context(), v, after.Add(3*time.Hour))
	require.NoError(t, err)

	p = state.parser("ZCode", after)
	provider.config = ParserConfig{After: p.discoveryAfter(), checkpoint: p}
	got, err = parseGenericAISQLiteIncremental(t.Context(), provider, db, path, "events", "rowid")
	require.NoError(t, err)

	got = p.filter(got)
	require.Len(t, got, 1)
	assert.EqualValues(t, 20, got[0].AIInputTokens)
	assert.Equal(t, time.Date(2026, 9, 1, 12, 45, 0, 0, time.UTC), p.Sessions["s"].Cutoff)
}
