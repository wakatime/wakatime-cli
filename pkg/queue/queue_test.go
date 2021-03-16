package queue_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/queue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestQueueFilepath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperValue string
		EnvVar     string
		Expected   string
	}{
		"default": {
			Expected: filepath.Join(home, ".wakatime.bdb"),
		},
		"env_trailling_slash": {
			EnvVar:   "~/path2/",
			Expected: filepath.Join(home, "path2", ".wakatime.bdb"),
		},
		"env_without_trailling_slash": {
			EnvVar:   "~/path2",
			Expected: filepath.Join(home, "path2", ".wakatime.bdb"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			queueFilepath, err := queue.QueueFilepath()
			require.NoError(t, err)

			assert.Equal(t, test.Expected, queueFilepath)
		})
	}
}

func TestQueue_Delete(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, []heartbeatRecord{
		{
			ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	tx, err := db.Begin(true)
	require.NoError(t, err)

	// run
	q := queue.NewQueue(tx)
	err = q.Delete([]string{
		"1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
		"1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
	})
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// check
	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("heartbeats")).Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			stored = append(stored, heartbeatRecord{
				ID:        string(key),
				Heartbeat: string(value),
			})
		}

		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeatRecord{
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	}, stored)
}

func TestQueue_LoadMany(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, []heartbeatRecord{
		{
			ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	tx, err := db.Begin(true)
	require.NoError(t, err)

	// run
	q := queue.NewQueue(tx)
	ids, hh, err := q.LoadMany(2)
	require.NoError(t, err)

	// check
	assert.Equal(t, []string{
		"1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
		"1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
	}, ids)

	assert.Len(t, hh, 2)
	assert.Contains(t, hh, testHeartbeats()[0])
	assert.Contains(t, hh, testHeartbeats()[1])
	// assert.Contains(t, hh, heartbeat.Heartbeat{
	// 	Branch:         heartbeat.String("heartbeat"),
	// 	Category:       heartbeat.CodingCategory,
	// 	CursorPosition: heartbeat.Int(12),
	// 	Dependencies:   []string{"dep1", "dep2"},
	// 	Entity:         "/tmp/main.go",
	// 	EntityType:     heartbeat.FileType,
	// 	IsWrite:        heartbeat.Bool(true),
	// 	Language:       heartbeat.LanguagePtr(heartbeat.LanguageGo),
	// 	LineNumber:     heartbeat.Int(42),
	// 	Lines:          heartbeat.Int(100),
	// 	Project:        heartbeat.String("wakatime-cli"),
	// 	Time:           1592868367.219124,
	// 	UserAgent:      "wakatime/13.0.6",
	// })
	// assert.Contains(t, hh, heartbeat.Heartbeat{
	// 	Branch:         heartbeat.String("summary"),
	// 	Category:       heartbeat.DebuggingCategory,
	// 	CursorPosition: heartbeat.Int(13),
	// 	Dependencies:   []string{"dep3", "dep4"},
	// 	Entity:         "/tmp/main.py",
	// 	EntityType:     heartbeat.FileType,
	// 	IsWrite:        heartbeat.Bool(false),
	// 	Language:       heartbeat.LanguagePtr(heartbeat.LanguagePython),
	// 	LineNumber:     heartbeat.Int(43),
	// 	Lines:          heartbeat.Int(101),
	// 	Project:        heartbeat.String("wakatime"),
	// 	Time:           1592868386.079084,
	// 	UserAgent:      "wakatime/13.0.7",
	// })

	err = tx.Commit()
	require.NoError(t, err)

	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("heartbeats")).Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			stored = append(stored, heartbeatRecord{
				ID:        string(key),
				Heartbeat: string(value),
			})
		}

		return nil
	})
	require.NoError(t, err)

	assert.Equal(t, []heartbeatRecord{
		{
			ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	}, stored)
}

func TestQueue_StoreMany(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	insertHeartbeatRecord(t, db, heartbeatRecord{
		ID:        "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
		Heartbeat: string(dataGo),
	})

	var heartbeatPy heartbeat.Heartbeat

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	err = json.Unmarshal(dataPy, &heartbeatPy)
	require.NoError(t, err)

	var heartbeatJs heartbeat.Heartbeat

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	err = json.Unmarshal(dataJs, &heartbeatJs)
	require.NoError(t, err)

	tx, err := db.Begin(true)
	require.NoError(t, err)

	// run
	q := queue.NewQueue(tx)
	err = q.StoreMany([]heartbeat.Heartbeat{heartbeatPy, heartbeatJs})
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// check
	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("heartbeats")).Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			stored = append(stored, heartbeatRecord{
				ID:        string(key),
				Heartbeat: string(value),
			})
		}

		return nil
	})
	require.NoError(t, err)

	assert.Len(t, stored, 3)

	assert.Equal(t, "1592868313.541149-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[2].ID)
	assert.JSONEq(t, string(dataJs), stored[2].Heartbeat)
}

func initDB(t *testing.T) (*bolt.DB, func()) {
	// create tmp file
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	// init db
	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	return db, func() {
		defer os.Remove(f.Name())
		defer db.Close()
	}
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
			Language:       heartbeat.LanguagePtr(heartbeat.LanguageGo),
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
			Language:       heartbeat.LanguagePtr(heartbeat.LanguagePython),
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

func insertHeartbeatRecords(t *testing.T, db *bolt.DB, hh []heartbeatRecord) {
	for _, h := range hh {
		insertHeartbeatRecord(t, db, h)
	}
}

func insertHeartbeatRecord(t *testing.T, db *bolt.DB, h heartbeatRecord) {
	t.Helper()

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("heartbeats"))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %s", err)
		}

		err = b.Put([]byte(h.ID), []byte(h.Heartbeat))
		if err != nil {
			return fmt.Errorf("failed put hearbeat: %s", err)
		}

		return nil
	})
	require.NoError(t, err)
}
