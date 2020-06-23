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

func TestQueue_PopMany(t *testing.T) {
	_, err := conn.Exec("CREATE TABLE heartbeat_2 (id TEXT, heartbeat TEXT)")
	require.NoError(t, err)

	defer func() {
		_, err := conn.Exec(`DROP TABLE heartbeat_2;`)
		if err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadFile("testdata/heartbeat_one.json")
	require.NoError(t, err)

	insertHearbeatRecords(t, []heartbeatRecord{
		{
			ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(data),
		},
	})

	data, err = ioutil.ReadFile("testdata/heartbeat_two.json")
	require.NoError(t, err)

	insertHearbeatRecords(t, []heartbeatRecord{
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(data),
		},
	})

	q := offline.NewQueue(conn)
	hh, err := q.PopMany(99)
	require.NoError(t, err)

	assert.Len(t, hh, 2)
	assert.Contains(t, hh, heartbeat.Heartbeat{
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
	assert.Contains(t, hh, heartbeat.Heartbeat{
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

	rows, err := conn.Query("SELECT id, heartbeat FROM heartbeat_2;")
	require.NoError(t, err)

	assert.False(t, rows.Next())
}

func TestQueue_PopMany_Limit(t *testing.T) {
	_, err := conn.Exec("CREATE TABLE heartbeat_2 (id TEXT, heartbeat TEXT)")
	require.NoError(t, err)

	defer func() {
		_, err := conn.Exec(`DROP TABLE heartbeat_2;`)
		if err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadFile("testdata/heartbeat_one.json")
	require.NoError(t, err)

	insertHearbeatRecords(t, []heartbeatRecord{
		{
			ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(data),
		},
	})

	data, err = ioutil.ReadFile("testdata/heartbeat_two.json")
	require.NoError(t, err)

	insertHearbeatRecords(t, []heartbeatRecord{
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(data),
		},
	})

	q := offline.NewQueue(conn)
	hh, err := q.PopMany(1)
	require.NoError(t, err)

	assert.Len(t, hh, 1)
	assert.Contains(t, testHeartbeats(), hh[0])

	rows, err := conn.Query("SELECT id, heartbeat FROM heartbeat_2;")
	require.NoError(t, err)

	var (
		ids        []string
		heartbeats []heartbeat.Heartbeat
	)

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

		ids = append(ids, id)

		var h heartbeat.Heartbeat
		err = json.Unmarshal([]byte(data), &h)
		require.NoError(t, err)

		assert.Equal(t, h.ID(), id)

		heartbeats = append(heartbeats, h)
	}
	require.NoError(t, rows.Err())

	assert.Equal(t, []string{"1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false"}, ids)
	assert.Equal(t, []heartbeat.Heartbeat{
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
	}, heartbeats)
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

type heartbeatRecord struct {
	ID        string
	Heartbeat string
}

func insertHearbeatRecords(t *testing.T, hh []heartbeatRecord) {
	for _, h := range hh {
		insertHearbeatRecord(t, h)
	}
}

func insertHearbeatRecord(t *testing.T, h heartbeatRecord) {
	t.Helper()

	_, err := conn.Exec(
		"INSERT INTO heartbeat_2 VALUES ($1, $2)",
		h.ID,
		h.Heartbeat,
	)
	require.NoError(t, err)
}
