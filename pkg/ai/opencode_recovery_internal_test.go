//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenCodeContinuesAfterInvalidSQLRows(t *testing.T) {
	home := updateTestHome(t)
	db := updateDB(t, filepath.Join(home, ".local", "share", "opencode", "opencode.db"))
	updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))
	updateSQL(t, db, `INSERT INTO session VALUES (NULL,'/bad','2',NULL,NULL);
 INSERT INTO session_message VALUES
 ('null-data','migrated','assistant',2,1800000000001,NULL),
 (NULL,'migrated','assistant',3,1800000000002,'{}'),
 ('bad-json','migrated','assistant',4,1800000000003,'{'),
 ('changed-shape','migrated','assistant',5,1800000000004,'{"tokens":[]}'),
 ('future-kind','migrated','future-kind',6,1800000000005,'{}'),
 ('later','migrated','user',7,1800000000006,'{"text":"later"}');
 INSERT INTO message VALUES
 (NULL,'legacy',1800000000001,'{}'),
 ('null-session',NULL,1800000000002,'{}'),
 ('changed-shape','legacy',1800000000003,'{"tokens":[]}'),
 ('later-legacy','legacy',1800000000004,'{"role":"user"}');
 INSERT INTO part VALUES
 (NULL,'later-legacy','legacy',1800000000000,'{}'),
 ('changed-part','later-legacy','legacy',1800000000001,'{"type":"text","text":{}}'),
 ('later-part','later-legacy','legacy',1800000000002,'{"type":"text","text":"legacy"}');`)

	parser := OpenCode{}
	hh, err := parser.Parse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 11, updateTotals(hh).prompt)
	assert.EqualValues(t, 132, updateTotals(hh).input)

	// A subsequent parse still discovers new rows after malformed records.
	updateSQL(t, db, `INSERT INTO session_message VALUES
 ('next','migrated','user',8,1800000000007,'{"text":"next"}')`)

	hh, err = parser.Parse(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 15, updateTotals(hh).prompt)
}

func TestOpenCodeContinuesAfterInvalidSeedRows(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "opencode.db")
	db := updateDB(t, dbPath)
	updateSQL(t, db, `CREATE TABLE message(id TEXT,session_id TEXT,time_created INTEGER,data TEXT);
 INSERT INTO message VALUES
 (NULL,'s',1999,'{}'),
 ('bad','s',1998,'{"tokens":[]}'),
 ('seed','s',1000,'{"role":"assistant","tokens":{"input":7}}'),
 ('current','s',3000,'{"role":"assistant","tokens":{"input":3}}');`)
	messages, _, err := (OpenCode{After: time.UnixMilli(2000)}).querySQLiteMessages(context.Background(), db, dbPath)
	require.NoError(t, err)
	require.Len(t, messages["s"], 2)
	assert.Equal(t, "seed", messages["s"][1].info.ID)
}

func TestOpenCodeContinuesAfterSchemaChanges(t *testing.T) {
	for _, table := range []string{"session_message", "message", "part"} {
		t.Run(table, func(t *testing.T) {
			home := updateTestHome(t)
			db := updateDB(t, filepath.Join(home, ".local", "share", "opencode", "opencode.db"))
			updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))
			// Simulate a future generation with an incompatible column layout.
			switch table {
			case "session_message":
				updateSQL(t, db, `ALTER TABLE session_message RENAME COLUMN data TO future_data`)
			case "message":
				updateSQL(t, db, `ALTER TABLE message RENAME COLUMN data TO future_data`)
			case "part":
				updateSQL(t, db, `ALTER TABLE part RENAME COLUMN data TO future_data`)
			}

			hh, err := (OpenCode{}).Parse(context.Background())
			require.NoError(t, err)

			switch table {
			case "session_message":
				assert.EqualValues(t, 1006, updateTotals(hh).input)
			case "message":
				assert.EqualValues(t, 125, updateTotals(hh).input)
			case "part":
				assert.EqualValues(t, 132, updateTotals(hh).input)
			}
		})
	}
}

func TestOpenCodeContinuesAfterInvalidFilesAndDatabases(t *testing.T) {
	home := updateTestHome(t)
	root := filepath.Join(home, ".local", "share", "opencode")
	updateWrite(t, filepath.Join(root, "opencode-0.db"), "not a database")
	db := updateDB(t, filepath.Join(root, "opencode-1.db"))
	updateSQL(t, db, `CREATE TABLE session(id TEXT); CREATE TABLE session_message(future_data TEXT);`)
	db = updateDB(t, filepath.Join(root, "opencode-2.db"))
	updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))

	storage := filepath.Join(root, "storage")
	updateWrite(t, filepath.Join(storage, "session", "project", "0-bad.json"), `{"id":[]}`)
	updateWrite(t, filepath.Join(storage, "session", "project", "1-good.json"), `{"id":"files"}`)
	updateWrite(t, filepath.Join(storage, "message", "files", "0-bad.json"), `{"tokens":[]}`)
	updateWrite(t, filepath.Join(storage, "message", "files", "1-good.json"),
		`{"id":"good","role":"user","time":{"created":1800000000000}}`)
	updateWrite(t, filepath.Join(storage, "part", "good", "0-bad.json"), `{"text":[]}`)
	updateWrite(t, filepath.Join(storage, "part", "good", "1-good.json"), `{"type":"text","text":"hello"}`)

	hh, err := (OpenCode{}).Parse(context.Background())
	require.NoError(t, err)
	assert.EqualValues(t, 132, updateTotals(hh).input)
	assert.Equal(t, 5, updateTotals(hh).prompt)
}

func TestOpenCodeV2ContinuesAfterInvalidContent(t *testing.T) {
	raw := `{"tokens":{"input":7},"content":[{"type":"tool","state":[]},
 {"type":"tool","name":"write","state":{"status":"completed",
 "input":{"path":"good.go","content":"hello"},"structured":{"existed":false}}}]} `
	message, ok := openCodeV2Message("msg", "session", "assistant", 1000, raw)
	require.True(t, ok)

	hh := (OpenCode{}).sessionHeartbeats(openCodeSessionInfo{ID: "session", Directory: "/project"},
		[]openCodeMessageWithParts{message})
	require.Len(t, hh, 2)
	assert.EqualValues(t, 7, updateTotals(hh).input)
	assert.Equal(t, "/project/good.go", filepath.ToSlash(hh[1].Entity))
}

func TestOpenCodeRecoveryHonorsCancellation(t *testing.T) {
	updateTestHome(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := (OpenCode{}).Parse(ctx)
	require.ErrorIs(t, err, context.Canceled)
}

func TestOpenCodeContinuesAfterOversizedRows(t *testing.T) {
	for _, payload := range []struct {
		name string
		data string
	}{
		{"text", strings.Repeat("x", maxTranscriptLineSize+1)},
		{"unicode", strings.Repeat("界", maxTranscriptLineSize/3+1)},
		// The payload alone fits, but the complete SQLite record does not.
		{"record", strings.Repeat("x", maxTranscriptLineSize-10)},
	} {
		t.Run(payload.name, func(t *testing.T) {
			home := updateTestHome(t)
			dbPath := filepath.Join(home, ".local", "share", "opencode", "opencode.db")
			db := updateDB(t, dbPath)
			updateSQL(t, db, updateFixture(t, "opencodemixedgenerations-1.txt"))

			for _, query := range []string{
				`INSERT INTO session VALUES ('oversized',?,'2',NULL,NULL)`,
				`INSERT INTO message VALUES ('oversized','legacy',1800000000001,?)`,
				`INSERT INTO part VALUES ('oversized','legacy-msg','legacy',1800000000001,?)`,
				`INSERT INTO session_message VALUES ('oversized','migrated','assistant',3,1800000000001,?)`,
			} {
				_, err := db.ExecContext(context.Background(), query, payload.data)

				require.NoError(t, err)
			}

			updateSQL(t, db, `INSERT INTO session_message VALUES
 ('oversized-session','oversized','user',1,1800000000001,'{"text":"ok"}')`)

			parser := OpenCode{}
			hh, err := parser.parseSQLiteDB(context.Background(), dbPath)
			require.NoError(t, err)
			assert.EqualValues(t, 132, updateTotals(hh).input)

			// Valid rows after each bad row must contribute heartbeats, including on
			// subsequent parses with a cutoff that puts the bad message in the seed query.
			updateSQL(t, db, `INSERT INTO message VALUES
    ('later','legacy',1800000003000,'{"role":"user"}');
    INSERT INTO part VALUES
    ('later-part','later','legacy',1800000003000,'{"type":"text","text":"hello"}');
    INSERT INTO session_message VALUES
    ('later-v2','migrated','user',4,1800000003000,'{"text":"world"}');`)

			parser.After = time.UnixMilli(1800000002000)
			hh, err = parser.parseSQLiteDB(context.Background(), dbPath)
			require.NoError(t, err)
			assert.Equal(t, 10, updateTotals(hh).prompt)
		})
	}
}
