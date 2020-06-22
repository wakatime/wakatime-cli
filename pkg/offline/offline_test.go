package offline_test

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	_ "github.com/mattn/go-sqlite3" // not used directly
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nolint
var conn *sql.DB

func TestMain(m *testing.M) {
	f, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		panic(err)
	}

	defer os.Remove(f.Name())

	// connect to DB
	conn, err = sql.Open("sqlite3", f.Name())
	if err != nil {
		panic(err)
	}

	// check DB connection
	for i := 0; i < 10; i++ {
		err = conn.Ping()
		if err == nil {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	if err != nil {
		panic(err)
	}

	// run tests
	os.Exit(m.Run())
}

func TestQueue_PushMany(t *testing.T) {
	_, err := conn.Exec("CREATE TABLE heartbeat_2 (id TEXT, heartbeat TEXT)")
	require.NoError(t, err)

	defer func() {
		_, err := conn.Exec(`DROP TABLE heartbeat_2;`)
		if err != nil {
			panic(err)
		}
	}()

	q := offline.NewQueue(conn)
	err = q.PushMany(testHeartbeats())
	require.NoError(t, err)

	rows, err := conn.Query("SELECT id, heartbeat FROM heartbeat_2;")
	require.NoError(t, err)

	var heartbeats []heartbeat.Heartbeat

	for rows.Next() {
		var (
			id   string
			data string
		)

		err := rows.Scan(
			&id,
			&data,
		)
		require.NoError(t, err)

		var h heartbeat.Heartbeat
		err = json.Unmarshal([]byte(data), &h)
		require.NoError(t, err)

		assert.Equal(t, h.ID(), id)

		heartbeats = append(heartbeats, h)
	}
	require.NoError(t, rows.Err())

	assert.Len(t, heartbeats, 2)
	assert.Contains(t, heartbeats, heartbeat.Heartbeat{
		Branch:         heartbeat.String("heartbeat"),
		Category:       heartbeat.CodingCategory,
		CursorPosition: heartbeat.Int(12),
		Dependencies:   []string{"dep1", "dep2"},
		Entity:         "/tmp/main.go",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(true),
		Language:       heartbeat.String("golang"),
		LineNumber:     heartbeat.Int(42),
		Lines:          heartbeat.Int(100),
		Project:        heartbeat.String("wakatime-cli"),
		Time:           1592868367.219124,
		UserAgent:      "wakatime/13.0.6",
	})
	assert.Contains(t, heartbeats, heartbeat.Heartbeat{
		Branch:         heartbeat.String("summary"),
		Category:       heartbeat.DebuggingCategory,
		CursorPosition: heartbeat.Int(13),
		Dependencies:   []string{"dep3", "dep4"},
		Entity:         "/tmp/main.py",
		EntityType:     heartbeat.FileType,
		IsWrite:        heartbeat.Bool(false),
		Language:       heartbeat.String("python"),
		LineNumber:     heartbeat.Int(43),
		Lines:          heartbeat.Int(101),
		Project:        heartbeat.String("wakatime"),
		Time:           1592868386.079084,
		UserAgent:      "wakatime/13.0.7",
	})
}

func testHeartbeats() []heartbeat.Heartbeat {
	return []heartbeat.Heartbeat{
		{
			Branch:         heartbeat.String("heartbeat"),
			Category:       heartbeat.CodingCategory,
			CursorPosition: heartbeat.Int(12),
			Dependencies:   []string{"dep1", "dep2"},
			Entity:         "/tmp/main.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(true),
			Language:       heartbeat.String("golang"),
			LineNumber:     heartbeat.Int(42),
			Lines:          heartbeat.Int(100),
			Project:        heartbeat.String("wakatime-cli"),
			Time:           1592868367.219124,
			UserAgent:      "wakatime/13.0.6",
		},
		{
			Branch:         heartbeat.String("summary"),
			Category:       heartbeat.DebuggingCategory,
			CursorPosition: heartbeat.Int(13),
			Dependencies:   []string{"dep3", "dep4"},
			Entity:         "/tmp/main.py",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(false),
			Language:       heartbeat.String("python"),
			LineNumber:     heartbeat.Int(43),
			Lines:          heartbeat.Int(101),
			Project:        heartbeat.String("wakatime"),
			Time:           1592868386.079084,
			UserAgent:      "wakatime/13.0.7",
		},
	}
}
