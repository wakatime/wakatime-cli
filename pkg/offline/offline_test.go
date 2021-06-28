package offline_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/offline"

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

			queueFilepath, err := offline.QueueFilepath()
			require.NoError(t, err)

			assert.Equal(t, test.Expected, queueFilepath)
		})
	}
}

func TestWithQueue(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	db.Close()

	opt, err := offline.WithQueue(f.Name(), 10)
	require.NoError(t, err)

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 2)
		assert.Contains(t, hh, testHeartbeats()[0])
		assert.Contains(t, hh, testHeartbeats()[1])

		return []heartbeat.Result{
			{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[0],
			},
			{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[1],
			},
		}, nil
	})

	// run
	results, err := handle([]heartbeat.Heartbeat{
		testHeartbeats()[0],
		testHeartbeats()[1],
	})
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status:    http.StatusCreated,
			Heartbeat: testHeartbeats()[0],
		},
		{
			Status:    http.StatusCreated,
			Heartbeat: testHeartbeats()[1],
		},
	}, results)

	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 1)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[0].ID)
	assert.JSONEq(t, string(dataJs), stored[0].Heartbeat)
}

func TestWithQueue_ApiError(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	opt, err := offline.WithQueue(f.Name(), 10)
	require.NoError(t, err)

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, []heartbeat.Heartbeat{
			testHeartbeats()[0],
			testHeartbeats()[1],
		})

		return []heartbeat.Result{}, errors.New("error")
	})

	// run
	_, err = handle([]heartbeat.Heartbeat{
		testHeartbeats()[0],
		testHeartbeats()[1],
	})
	require.Error(t, err)

	// check
	db, err := bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)
}

func TestWithQueue_InvalidResults(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	opt, err := offline.WithQueue(f.Name(), 10)
	require.NoError(t, err)

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, testHeartbeats())

		return []heartbeat.Result{
			{
				Status:    201,
				Heartbeat: testHeartbeats()[0],
			},
			{
				Status:    500,
				Heartbeat: testHeartbeats()[1],
			},
			{
				Status: 429,
				Errors: []string{"Too many heartbeats"},
			},
		}, nil
	})

	// run
	results, err := handle(testHeartbeats())
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status:    201,
			Heartbeat: testHeartbeats()[0],
		},
		{
			Status:    500,
			Heartbeat: testHeartbeats()[1],
		},
		{
			Status: 429,
			Errors: []string{"Too many heartbeats"},
		},
	}, results)

	// check db
	db, err := bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	assert.Len(t, stored, 2)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[1].ID)
	assert.JSONEq(t, string(dataJs), stored[1].Heartbeat)
}

func TestWithQueue_HandleLeftovers(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	opt, err := offline.WithQueue(f.Name(), 10)
	require.NoError(t, err)

	handle := opt(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, testHeartbeats())

		return []heartbeat.Result{
			{
				Status:    201,
				Heartbeat: testHeartbeats()[0],
			},
		}, nil
	})

	// run
	results, err := handle(testHeartbeats())
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status:    201,
			Heartbeat: testHeartbeats()[0],
		},
	}, results)

	db, err := bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[1].ID)
	assert.JSONEq(t, string(dataJs), stored[1].Heartbeat)
}

func TestSync(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	db.Close()

	syncFn := offline.Sync(f.Name(), 1000)

	var numCalls int

	err = syncFn(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Equal(t, []heartbeat.Heartbeat{
			testHeartbeats()[0],
			testHeartbeats()[1],
		}, hh)

		return []heartbeat.Result{
			{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[0],
			},
			{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[1],
			},
		}, nil
	})
	require.NoError(t, err)

	// check db
	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 0)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSync_MultipleRequests(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	for i := 0; i < 25; i++ {
		insertHeartbeatRecord(t, db, "heartbeats", heartbeatRecord{
			ID:        strconv.Itoa(i) + "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		})
	}

	db.Close()

	syncFn := offline.Sync(f.Name(), 1000)

	var numCalls int

	// run
	err = syncFn(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		// first request
		if numCalls == 1 {
			assert.Len(t, hh, 24)

			result := heartbeat.Result{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[0],
			}

			return []heartbeat.Result{
				result, result, result, result, result,
				result, result, result, result, result,
				result, result, result, result, result,
				result, result, result, result, result,
				result, result, result, result, result,
			}, nil
		}

		// second request
		assert.Len(t, hh, 1)

		results := []heartbeat.Result{
			{
				Status:    http.StatusCreated,
				Heartbeat: testHeartbeats()[0],
			},
		}

		return results, nil
	})
	require.NoError(t, err)

	// check db
	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 0)

	assert.Eventually(t, func() bool { return numCalls == 2 }, time.Second, 50*time.Millisecond)
}

func TestSync_APIError(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	db.Close()

	syncFn := offline.Sync(f.Name(), 10)

	var numCalls int

	// run
	err = syncFn(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Equal(t, []heartbeat.Heartbeat{
			testHeartbeats()[0],
			testHeartbeats()[1],
		}, hh)

		return nil, errors.New("failed")
	})
	require.Error(t, err)

	// check db
	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestSync_InvalidResults(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
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

	db.Close()

	syncFn := offline.Sync(f.Name(), 1000)

	var numCalls int

	// run
	err = syncFn(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		// first request
		if numCalls == 1 {
			require.Len(t, hh, 3)
			assert.Equal(t, []heartbeat.Heartbeat{
				testHeartbeats()[0],
				testHeartbeats()[1],
				testHeartbeats()[2],
			}, hh)

			return []heartbeat.Result{
				{
					Status:    201,
					Heartbeat: testHeartbeats()[0],
				},
				// any non 201/202/400 status results will be retried.
				{
					Status:    429,
					Errors:    []string{"Too many heartbeats"},
					Heartbeat: testHeartbeats()[1],
				},
				// 400 status results will be discarded
				{
					Status:    400,
					Heartbeat: testHeartbeats()[2],
				},
			}, nil
		}

		// second request: assert retry of 429 result
		require.Len(t, hh, 1)
		assert.Equal(t, []heartbeat.Heartbeat{
			testHeartbeats()[1],
		}, hh)

		return []heartbeat.Result{
			{
				Status:    201,
				Heartbeat: testHeartbeats()[1],
			},
		}, nil
	})
	require.NoError(t, err)

	// check db
	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 0)

	assert.Eventually(t, func() bool { return numCalls == 2 }, time.Second, 50*time.Millisecond)
}

func TestSync_SyncLimit(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.RemoveAll(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	db.Close()

	syncFn := offline.Sync(f.Name(), 1)

	var numCalls int

	// run
	err = syncFn(func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Len(t, hh, 1)

		return []heartbeat.Result{
			{
				Status:    201,
				Heartbeat: testHeartbeats()[0],
			},
		}, nil
	})
	require.NoError(t, err)

	// check db
	db, err = bolt.Open(f.Name(), 0600, nil)
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

	db.Close()

	require.Len(t, stored, 1)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Eventually(t, func() bool { return numCalls == 1 }, time.Second, 50*time.Millisecond)
}

func TestQueue_PopMany(t *testing.T) {
	// setup
	f, err := ioutil.TempFile(os.TempDir(), "")
	require.NoError(t, err)

	defer os.Remove(f.Name())

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := ioutil.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := ioutil.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "test_bucket", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
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
	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	hh, err := q.PopMany(2)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// check
	assert.Len(t, hh, 2)
	assert.Contains(t, hh, testHeartbeats()[0])
	assert.Contains(t, hh, testHeartbeats()[1])

	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("test_bucket")).Cursor()

		for key, value := c.First(); key != nil; key, value = c.Next() {
			stored = append(stored, heartbeatRecord{
				ID:        string(key),
				Heartbeat: string(value),
			})
		}

		return nil
	})
	require.NoError(t, err)

	assert.Len(t, stored, 1)
	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[0].ID)
	assert.JSONEq(t, string(dataJs), stored[0].Heartbeat)
}

func TestQueue_PushMany(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	dataGo, err := ioutil.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	insertHeartbeatRecord(t, db, "test_bucket", heartbeatRecord{
		ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
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
	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	err = q.PushMany([]heartbeat.Heartbeat{heartbeatPy, heartbeatJs})
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// check
	var stored []heartbeatRecord

	err = db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte("test_bucket")).Cursor()

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

	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[2].ID)
	assert.JSONEq(t, string(dataJs), stored[2].Heartbeat)
}

func TestQueue_Count(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"

	count, err := q.Count()
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	err = tx.Rollback()
	require.NoError(t, err)

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

	tx, err = db.Begin(true)
	require.NoError(t, err)

	// run
	q = offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	err = q.PushMany([]heartbeat.Heartbeat{heartbeatPy, heartbeatJs})
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	tx, err = db.Begin(true)
	require.NoError(t, err)

	q = offline.NewQueue(tx)
	q.Bucket = "test_bucket"

	count, err = q.Count()
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	err = tx.Rollback()
	require.NoError(t, err)
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
			Language:       heartbeat.String("Go"),
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
			Language:       heartbeat.String("Python"),
			LineNumber:     heartbeat.Int(43),
			Lines:          heartbeat.Int(101),
			Project:        heartbeat.String("wakatime"),
			Time:           1592868386.079084,
			UserAgent:      "wakatime/13.0.7",
		},
		{
			Branch:         heartbeat.String("todaygoal"),
			Category:       heartbeat.BuildingCategory,
			CursorPosition: heartbeat.Int(14),
			Dependencies:   []string{"dep5", "dep6"},
			Entity:         "/tmp/main.js",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.Bool(false),
			Language:       heartbeat.String("JavaScript"),
			LineNumber:     heartbeat.Int(44),
			Lines:          heartbeat.Int(102),
			Project:        heartbeat.String("wakatime"),
			Time:           1592868394.084354,
			UserAgent:      "wakatime/13.0.8",
		},
	}
}

type heartbeatRecord struct {
	ID        string
	Heartbeat string
}

func insertHeartbeatRecords(t *testing.T, db *bolt.DB, bucket string, hh []heartbeatRecord) {
	for _, h := range hh {
		insertHeartbeatRecord(t, db, bucket, h)
	}
}

func insertHeartbeatRecord(t *testing.T, db *bolt.DB, bucket string, h heartbeatRecord) {
	t.Helper()

	err := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
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
