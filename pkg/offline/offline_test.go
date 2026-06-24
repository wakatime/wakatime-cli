package offline_test

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
)

func TestQueueFilepath(t *testing.T) {
	ctx := t.Context()

	tests := map[string]struct {
		EnvVar string
	}{
		"default": {},
		"env_trailing_slash": {
			EnvVar: "~/path2/",
		},
		"env_without_trailing_slash": {
			EnvVar: "~/path2",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("WAKATIME_HOME", test.EnvVar)

			folder, err := ini.WakaResourcesDir(ctx)
			require.NoError(t, err)

			v := viper.New()
			queueFilepath, err := offline.QueueFilepath(ctx, v)
			require.NoError(t, err)

			expected := filepath.Join(folder, "offline_heartbeats.bdb")

			assert.Equal(t, expected, queueFilepath)
		})
	}
}

func TestWithQueue(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	opt := offline.WithQueue(f.Name())

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 2)
		assert.Contains(t, hh, testHeartbeats()[0])
		assert.Contains(t, hh, testHeartbeats()[1])

		return []heartbeat.Result{
			{
				Status: http.StatusCreated,
				ID:     "A1B2C3D4-E5F6-4789-A123-456789ABCDEF",
			},
			{
				Status: http.StatusCreated,
				ID:     "B2C3D4E5-F6A7-4890-B234-567890BCDEFG",
			},
		}, nil
	})

	// run
	results, err := handle(t.Context(), []heartbeat.Heartbeat{
		testHeartbeats()[0],
		testHeartbeats()[1],
	})
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status: http.StatusCreated,
			ID:     "A1B2C3D4-E5F6-4789-A123-456789ABCDEF",
		},
		{
			Status: http.StatusCreated,
			ID:     "B2C3D4E5-F6A7-4890-B234-567890BCDEFG",
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 1)

	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[0].ID)
	assert.JSONEq(t, string(dataJs), stored[0].Heartbeat)
}

func TestWithQueue_NoHeartbeats(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	opt := offline.WithQueue(f.Name())

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Len(t, hh, 0)

		return []heartbeat.Result{}, nil
	})

	// run
	results, err := handle(t.Context(), []heartbeat.Heartbeat{})
	require.NoError(t, err)

	// check
	assert.Nil(t, results)
}

func TestWithQueue_ApiError(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	opt := offline.WithQueue(f.Name())

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, []heartbeat.Heartbeat{
			testHeartbeats()[0],
			testHeartbeats()[1],
		})

		return []heartbeat.Result{}, errors.New("error")
	})

	// run
	_, err = handle(t.Context(), []heartbeat.Heartbeat{
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

	err = db.Close()
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868367.219124-12-file-undefined-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)
}

func TestWithQueue_InvalidResults(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	opt := offline.WithQueue(f.Name())

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, testHeartbeats())

		return []heartbeat.Result{
			{
				Status: 201,
				ID:     "C3D4E5F6-A7B8-4901-C345-678901CDEFGH",
			},
			{
				Status: 500,
				ID:     "D4E5F6A7-B8C9-4012-D456-789012DEFGHI",
			},
			{
				Status: 429,
				Errors: []string{"Too many heartbeats"},
			},
		}, nil
	})

	// run
	results, err := handle(t.Context(), testHeartbeats())
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
			ID:     "C3D4E5F6-A7B8-4901-C345-678901CDEFGH",
		},
		{
			Status: 500,
			ID:     "D4E5F6A7-B8C9-4012-D456-789012DEFGHI",
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

	err = db.Close()
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	assert.Len(t, stored, 2)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Equal(t, "1592868394.084354-14-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[1].ID)
	assert.JSONEq(t, string(dataJs), stored[1].Heartbeat)
}

func TestWithQueue_HandleLeftovers(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	opt := offline.WithQueue(f.Name())

	handle := opt(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		assert.Equal(t, hh, testHeartbeats())

		return []heartbeat.Result{
			{
				Status: 201,
				ID:     "E5F6A7B8-C9D0-4123-E567-890123EFGHIJ",
			},
		}, nil
	})

	// run
	results, err := handle(t.Context(), testHeartbeats())
	require.NoError(t, err)

	// check
	assert.Equal(t, []heartbeat.Result{
		{
			Status: 201,
			ID:     "E5F6A7B8-C9D0-4123-E567-890123EFGHIJ",
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

	err = db.Close()
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Equal(t, "1592868394.084354-14-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[1].ID)
	assert.JSONEq(t, string(dataJs), stored[1].Heartbeat)
}

func TestWithSync(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-12-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	opt := offline.WithSync(f.Name(), offline.SyncMaxDefault)

	handle := opt(func(_ context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return []heartbeat.Result{
			{
				Status: http.StatusCreated,
				ID:     "F6A7B8C9-D0E1-4234-F678-901234FGHIJK",
			},
			{
				Status: http.StatusCreated,
				ID:     "A7B8C9D0-E1F2-4345-A789-012345GHIJKL",
			},
		}, nil
	})

	// run
	results, err := handle(t.Context(), nil)
	require.NoError(t, err)

	// check
	assert.Nil(t, results)

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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 0)
}

func TestSync_MultipleRequests(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	const totalHeartbeats = 26

	for i := 0; i < totalHeartbeats; i++ {
		h := heartbeat.Heartbeat{
			Entity:     fmt.Sprintf("/tmp/main-%02d.go", i),
			EntityType: heartbeat.FileType,
			Time:       float64(1592868367 + i*10),
			UserAgent:  "wakatime/test",
		}
		data, marshalErr := json.Marshal(h)
		require.NoError(t, marshalErr)

		insertHeartbeatRecord(t, db, "heartbeats", heartbeatRecord{
			ID:        h.ID(),
			Heartbeat: string(data),
		})
	}

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 1000)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		expectedLen := totalHeartbeats - ((numCalls - 1) * offline.SendLimit)
		if expectedLen > offline.SendLimit {
			expectedLen = offline.SendLimit
		}

		assert.Len(t, hh, expectedLen)

		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{
				Status: http.StatusCreated,
				ID:     "C9D0E1F2-A3B4-4567-C901-234567IJKLMN",
			}
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 0)

	expectedCalls := int(math.Ceil(float64(totalHeartbeats) / float64(offline.SendLimit)))

	assert.Equal(t, expectedCalls, numCalls)
}

func TestSync_APIError(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-12-file-undefined-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 3)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Equal(t, []heartbeat.Heartbeat{
			testHeartbeats()[1],
			testHeartbeats()[0],
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 2)

	assert.Equal(t, "1592868367.219124-12-file-undefined-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)

	assert.Equal(t, 1, numCalls)
}

func TestSync_APIErrorBulkNested(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-12-file-undefined-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: string(dataGo),
		},
		{
			ID:        "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: string(dataPy),
		},
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: string(dataJs),
		},
	})

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 3)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Equal(t, []heartbeat.Heartbeat{
			testHeartbeats()[2],
			testHeartbeats()[1],
			testHeartbeats()[0],
		}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
				ID:     "B4C5D6E7-F8A9-4012-B456-789012NOPQRS",
			},
			{
				Status: 401,
				Errors: []string{"An error."},
			},
			{
				Status: 201,
				ID:     "B4C5D6E7-F8A9-4012-B456-789012NOPQRS",
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 1)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[0].ID)
	assert.JSONEq(t, string(dataPy), stored[0].Heartbeat)

	assert.Equal(t, 1, numCalls)
}

func TestSync_InvalidResults(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 1000)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		// first request
		if numCalls == 1 {
			require.Len(t, hh, 3)
			assert.Equal(t, []heartbeat.Heartbeat{
				testHeartbeats()[2],
				testHeartbeats()[1],
				testHeartbeats()[0],
			}, hh)

			return []heartbeat.Result{
				{
					Status: 201,
					ID:     "D0E1F2A3-B4C5-4678-D012-345678JKLMNO",
				},
				// any non 201/202/400 status results will be retried.
				{
					Status: 429,
					Errors: []string{"Too many heartbeats"},
					ID:     "E1F2A3B4-C5D6-4789-E123-456789KLMNOP",
				},
				// 400 status results will be discarded
				{
					Status: 400,
					ID:     "F2A3B4C5-D6E7-4890-F234-567890LMNOPQ",
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
				Status: 201,
				ID:     "A3B4C5D6-E7F8-4901-A345-678901MNOPQR",
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 0)

	assert.Equal(t, 2, numCalls)
}

func TestSync_SyncLimit(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
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

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 1)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Equal(t, []heartbeat.Heartbeat{testHeartbeats()[1]}, hh)

		return []heartbeat.Result{
			{
				Status: 201,
				ID:     "B4C5D6E7-F8A9-4012-B456-789012NOPQRS",
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 1)

	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, 1, numCalls)
}

func TestSync_SyncLimitAcrossMultipleBatches(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	for i := 0; i < 40; i++ {
		h := heartbeat.Heartbeat{
			Entity:     fmt.Sprintf("/tmp/limit-%02d.go", i),
			EntityType: heartbeat.FileType,
			Time:       float64(1592868367 + i*10),
			UserAgent:  "wakatime/test",
		}
		data, marshalErr := json.Marshal(h)
		require.NoError(t, marshalErr)

		insertHeartbeatRecord(t, db, "heartbeats", heartbeatRecord{
			ID:        h.ID(),
			Heartbeat: string(data),
		})
	}

	require.NoError(t, db.Close())

	syncFn := offline.Sync(t.Context(), f.Name(), 30)

	var (
		numCalls int
		sent     int
	)

	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++
		previouslySent := sent
		sent += len(hh)

		expectedLen := offline.SendLimit
		if remaining := 30 - previouslySent; remaining < expectedLen {
			expectedLen = remaining
		}

		assert.Len(t, hh, expectedLen)

		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{Status: http.StatusCreated, ID: fmt.Sprintf("created-%d-%d", numCalls, i)}
		}

		return results, nil
	})
	require.NoError(t, err)
	assert.Equal(t, int(math.Ceil(float64(30)/float64(offline.SendLimit))), numCalls)
	assert.Equal(t, 30, sent)

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
	require.NoError(t, db.Close())

	require.Len(t, stored, 10)
}

func TestSync_SyncUnlimited(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
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

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), math.MaxInt32)

	var numCalls int

	// run
	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++

		assert.Len(t, hh, 2)

		return []heartbeat.Result{
			{
				Status: 201,
				ID:     "84DA67BB-14ED-432F-A141-F667196CEDC2",
			},
			{
				Status: 201,
				ID:     "96CD6874-F014-40B5-A860-7E53B23DA59E",
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

	err = db.Close()
	require.NoError(t, err)

	require.Len(t, stored, 0)

	assert.Equal(t, 1, numCalls)
}

func TestSync_DeletesDuplicates(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	records := make([]heartbeatRecord, 0, 4)

	for _, h := range []heartbeat.Heartbeat{
		{
			Entity:     "/tmp/a.go",
			EntityType: heartbeat.FileType,
			Time:       1000,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/dup.go",
			EntityType: heartbeat.FileType,
			Time:       3000,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/dup.go",
			EntityType: heartbeat.FileType,
			Time:       3000.5,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/b.go",
			EntityType: heartbeat.FileType,
			Time:       4000,
			UserAgent:  "wakatime/test",
		},
	} {
		data, marshalErr := json.Marshal(h)
		require.NoError(t, marshalErr)

		records = append(records, heartbeatRecord{
			ID:        h.ID(),
			Heartbeat: string(data),
		})
	}

	insertHeartbeatRecords(t, db, "heartbeats", records)

	err = db.Close()
	require.NoError(t, err)

	syncFn := offline.Sync(t.Context(), f.Name(), 5)

	var (
		numCalls  int
		totalSent int
	)

	err = syncFn(func(_ context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		numCalls++
		totalSent += len(hh)

		results := make([]heartbeat.Result, len(hh))
		for i := range hh {
			results[i] = heartbeat.Result{
				Status: http.StatusCreated,
				ID:     fmt.Sprintf("id-%d-%d", numCalls, i),
			}
		}

		return results, nil
	})
	require.NoError(t, err)

	assert.Equal(t, 1, numCalls)
	assert.Equal(t, 3, totalSent)

	count, err := offline.CountHeartbeats(t.Context(), f.Name())
	require.NoError(t, err)
	assert.Zero(t, count)
}

func TestSync_RequeuesHeartbeatMissingResultByAssociation(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	tx, err := db.Begin(true)
	require.NoError(t, err)

	hh := []heartbeat.Heartbeat{
		{
			Entity:     "/tmp/a.go",
			EntityType: heartbeat.FileType,
			Time:       1000,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/b.go",
			EntityType: heartbeat.FileType,
			Time:       2000,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/c.go",
			EntityType: heartbeat.FileType,
			Time:       3000,
			UserAgent:  "wakatime/test",
		},
	}

	q := offline.NewQueue(tx)
	require.NoError(t, q.PushMany(hh))
	require.NoError(t, tx.Commit())
	require.NoError(t, db.Close())

	syncFn := offline.Sync(t.Context(), f.Name(), 3)

	var sent []heartbeat.Heartbeat

	err = syncFn(func(_ context.Context, batch []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		require.Len(t, batch, 3)
		sent = append([]heartbeat.Heartbeat(nil), batch...)

		return []heartbeat.Result{
			{
				Status:    http.StatusCreated,
				ID:        "id-0",
				Heartbeat: batch[0],
			},
			{
				Status:    http.StatusCreated,
				ID:        "id-2",
				Heartbeat: batch[2],
			},
		}, nil
	})
	require.NoError(t, err)

	stored, err := offline.ReadHeartbeats(t.Context(), f.Name(), 10)
	require.NoError(t, err)
	require.Len(t, stored, 1)
	assert.Equal(t, sent[1].ID(), stored[0].ID())
}

func TestCountHeartbeats(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	insertHeartbeatRecords(t, db, "heartbeats", []heartbeatRecord{
		{
			ID:        "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
			Heartbeat: "heartbeat_go",
		},
		{
			ID:        "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false",
			Heartbeat: "heartbeat_py",
		},
		{
			ID:        "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false",
			Heartbeat: "heartbeat_js",
		},
	})

	err = db.Close()
	require.NoError(t, err)

	count, err := offline.CountHeartbeats(t.Context(), f.Name())
	require.NoError(t, err)

	assert.Equal(t, count, 3)
}

func TestCountHeartbeats_Empty(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	count, err := offline.CountHeartbeats(t.Context(), f.Name())
	require.NoError(t, err)

	assert.Equal(t, count, 0)
}

func TestCountHeartbeats_CorruptDBReturnsErrOpenDB(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	queuePath := f.Name()
	require.NoError(t, f.Close())

	db, err := bolt.Open(queuePath, 0600, nil)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	corruptBoltPageFlags(t, queuePath, 3, 0x0a)

	count, err := offline.CountHeartbeats(t.Context(), queuePath)

	var erropen offline.ErrOpenDB
	require.ErrorAs(t, err, &erropen)
	assert.Zero(t, count)
	assert.True(t, erropen.Reset)
	assert.FileExists(t, erropen.BackupFilepath)
	assert.NoFileExists(t, queuePath)
	assert.Contains(t, err.Error(), "panicked:")
	assert.Contains(t, err.Error(), "page 3")

	count, err = offline.CountHeartbeats(t.Context(), queuePath)
	require.NoError(t, err)
	assert.Zero(t, count)
	assert.FileExists(t, queuePath)
}

func TestCountHeartbeats_InvalidDBResetsQueue(t *testing.T) {
	queuePath := filepath.Join(t.TempDir(), "offline_heartbeats.bdb")
	require.NoError(t, os.WriteFile(queuePath, []byte("not a bolt db"), 0600))

	count, err := offline.CountHeartbeats(t.Context(), queuePath)

	var erropen offline.ErrOpenDB
	require.ErrorAs(t, err, &erropen)
	assert.Zero(t, count)
	assert.True(t, erropen.Reset)
	assert.FileExists(t, erropen.BackupFilepath)
	assert.NoFileExists(t, queuePath)

	count, err = offline.CountHeartbeats(t.Context(), queuePath)
	require.NoError(t, err)
	assert.Zero(t, count)
	assert.FileExists(t, queuePath)
}

func TestWithQueue_CorruptDBResetsAndSavesHeartbeats(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	queuePath := f.Name()
	require.NoError(t, f.Close())

	db, err := bolt.Open(queuePath, 0600, nil)
	require.NoError(t, err)
	require.NoError(t, db.Close())

	corruptBoltPageFlags(t, queuePath, 3, 0x0a)

	h := heartbeat.Heartbeat{
		Entity:     "/tmp/main.go",
		EntityType: heartbeat.FileType,
		Time:       1,
		UserAgent:  "wakatime/test",
	}

	handle := offline.WithQueue(queuePath)(func(context.Context, []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
		return nil, errors.New("api unavailable")
	})

	_, err = handle(t.Context(), []heartbeat.Heartbeat{h})
	require.EqualError(t, err, "api unavailable")

	count, err := offline.CountHeartbeats(t.Context(), queuePath)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	hh, err := offline.ReadHeartbeats(t.Context(), queuePath, 1)
	require.NoError(t, err)
	require.Len(t, hh, 1)
	assert.Equal(t, h.ID(), hh[0].ID())
}

func TestReadHeartbeats(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
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

	err = db.Close()
	require.NoError(t, err)

	hh, err := offline.ReadHeartbeats(t.Context(), f.Name(), offline.PrintMaxDefault)
	require.NoError(t, err)

	assert.Len(t, hh, 2)
}

func TestReadHeartbeats_WithLimit(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
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

	err = db.Close()
	require.NoError(t, err)

	hh, err := offline.ReadHeartbeats(t.Context(), f.Name(), 1)
	require.NoError(t, err)

	assert.Len(t, hh, 1)
}

func TestReadHeartbeats_Empty(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	hh, err := offline.ReadHeartbeats(t.Context(), f.Name(), offline.PrintMaxDefault)
	require.NoError(t, err)

	assert.Len(t, hh, 0)
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

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	err = json.Unmarshal(dataPy, &heartbeatPy)
	require.NoError(t, err)

	var heartbeatJs heartbeat.Heartbeat

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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

func TestQueue_PopMany(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	defer func() {
		err = db.Close()
		require.NoError(t, err)
	}()

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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
	assert.Equal(t, []heartbeat.Heartbeat{
		testHeartbeats()[2],
		testHeartbeats()[1],
	}, hh)

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
	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)
}

func TestQueue_PushMany(t *testing.T) {
	// setup
	db, cleanup := initDB(t)
	defer cleanup()

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	insertHeartbeatRecord(t, db, "test_bucket", heartbeatRecord{
		ID:        "1592868367.219124-1-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true",
		Heartbeat: string(dataGo),
	})

	var heartbeatPy heartbeat.Heartbeat

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	err = json.Unmarshal(dataPy, &heartbeatPy)
	require.NoError(t, err)

	var heartbeatJs heartbeat.Heartbeat

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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

	assert.Equal(t, "1592868367.219124-1-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)

	assert.Equal(t, "1592868386.079084-13-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)

	assert.Equal(t, "1592868394.084354-14-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[2].ID)
	assert.JSONEq(t, string(dataJs), stored[2].Heartbeat)
}

func TestQueue_PushMany_PreservesAITokensAndSession(t *testing.T) {
	db, cleanup := initDB(t)
	defer cleanup()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"

	expected := heartbeat.Heartbeat{
		AISession:      "session-123",
		AIOutputTokens: 17,
		Category:       heartbeat.AICodingCategory.String(),
		Entity:         "Claude session",
		EntityType:     heartbeat.AppType,
		IsWrite:        heartbeat.PointerTo(false),
		AIPromptLength: 42,
		Time:           1770000000,
		UserAgent:      "Claude/2.1.45 plugin/0.0.1",
	}

	require.NoError(t, q.PushMany([]heartbeat.Heartbeat{expected}))
	require.NoError(t, tx.Commit())

	tx, err = db.Begin(true)
	require.NoError(t, err)

	defer tx.Rollback()

	q = offline.NewQueue(tx)
	q.Bucket = "test_bucket"

	got, err := q.ReadMany(1)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, expected, got[0])
}

func TestQueue_ReadMany(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	defer func() {
		err = db.Close()
		require.NoError(t, err)
	}()

	dataGo, err := os.ReadFile("testdata/heartbeat_go.json")
	require.NoError(t, err)

	dataPy, err := os.ReadFile("testdata/heartbeat_py.json")
	require.NoError(t, err)

	dataJs, err := os.ReadFile("testdata/heartbeat_js.json")
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
	hh, err := q.ReadMany(2)
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

	assert.Len(t, stored, 3)

	assert.Equal(t, "1592868367.219124-file-coding-wakatime-cli-heartbeat-/tmp/main.go-true", stored[0].ID)
	assert.Equal(t, "1592868386.079084-file-debugging-wakatime-summary-/tmp/main.py-false", stored[1].ID)
	assert.Equal(t, "1592868394.084354-file-building-wakatime-todaygoal-/tmp/main.js-false", stored[2].ID)

	assert.JSONEq(t, string(dataGo), stored[0].Heartbeat)
	assert.JSONEq(t, string(dataPy), stored[1].Heartbeat)
	assert.JSONEq(t, string(dataJs), stored[2].Heartbeat)
}

func TestQueue_ReadMany_Empty(t *testing.T) {
	// setup
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	defer f.Close()

	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	defer func() {
		err = db.Close()
		require.NoError(t, err)
	}()

	tx, err := db.Begin(true)
	require.NoError(t, err)

	// run
	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	hh, err := q.ReadMany(10)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	// check
	assert.Len(t, hh, 0)
}

func TestQueue_DeleteDuplicates(t *testing.T) {
	db, cleanup := initDB(t)
	defer cleanup()

	hh := []heartbeat.Heartbeat{
		{
			Entity:     "/tmp/dup.go",
			EntityType: heartbeat.FileType,
			Time:       1000,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/dup.go",
			EntityType: heartbeat.FileType,
			Time:       1000.5,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/dup.go",
			EntityType: heartbeat.FileType,
			Time:       1001.2,
			UserAgent:  "wakatime/test",
		},
		{
			Entity:     "/tmp/other.go",
			EntityType: heartbeat.FileType,
			Time:       1000.4,
			UserAgent:  "wakatime/test",
		},
	}

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	err = q.PushMany(hh)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	tx, err = db.Begin(true)
	require.NoError(t, err)

	q = offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	deleted, err := q.DeleteDuplicates()
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	remaining, err := q.ReadMany(10)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	assert.Len(t, remaining, 3)
	assert.Contains(t, remaining, hh[0])
	assert.NotContains(t, remaining, hh[1])
	assert.Contains(t, remaining, hh[2])
	assert.Contains(t, remaining, hh[3])
}

func TestQueue_DeleteDuplicates_ChecksAllKeptTimesWithSymmetricWindow(t *testing.T) {
	db, cleanup := initDB(t)
	defer cleanup()

	records := []heartbeatRecord{
		{
			ID:        "a",
			Heartbeat: `{"entity":"/tmp/dup.go","type":"file","time":1002,"user_agent":"wakatime/test"}`,
		},
		{
			ID:        "b",
			Heartbeat: `{"entity":"/tmp/dup.go","type":"file","time":1000,"user_agent":"wakatime/test"}`,
		},
		{
			ID:        "c",
			Heartbeat: `{"entity":"/tmp/dup.go","type":"file","time":1001.1,"user_agent":"wakatime/test"}`,
		},
	}

	insertHeartbeatRecords(t, db, "test_bucket", records)

	tx, err := db.Begin(true)
	require.NoError(t, err)

	q := offline.NewQueue(tx)
	q.Bucket = "test_bucket"
	deleted, err := q.DeleteDuplicates()
	require.NoError(t, err)
	assert.Equal(t, 1, deleted)

	remaining, err := q.ReadMany(10)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	assert.Len(t, remaining, 2)
	assert.Contains(t, remaining, heartbeat.Heartbeat{
		Entity:     "/tmp/dup.go",
		EntityType: heartbeat.FileType,
		Time:       1002,
		UserAgent:  "wakatime/test",
	})
	assert.Contains(t, remaining, heartbeat.Heartbeat{
		Entity:     "/tmp/dup.go",
		EntityType: heartbeat.FileType,
		Time:       1000,
		UserAgent:  "wakatime/test",
	})
	assert.NotContains(t, remaining, heartbeat.Heartbeat{
		Entity:     "/tmp/dup.go",
		EntityType: heartbeat.FileType,
		Time:       1001.1,
		UserAgent:  "wakatime/test",
	})
}

func initDB(t *testing.T) (*bolt.DB, func()) {
	// create tmp file
	f, err := os.CreateTemp(t.TempDir(), "")
	require.NoError(t, err)

	// init db
	db, err := bolt.Open(f.Name(), 0600, nil)
	require.NoError(t, err)

	return db, func() {
		defer f.Close()
		defer func() {
			err = db.Close()
			require.NoError(t, err)
		}()
	}
}

func testHeartbeats() []heartbeat.Heartbeat {
	return []heartbeat.Heartbeat{
		{
			Branch:         heartbeat.PointerTo("heartbeat"),
			Category:       heartbeat.UndefinedCategory.String(),
			CursorPosition: heartbeat.PointerTo(12),
			Dependencies:   []string{"dep1", "dep2"},
			Entity:         "/tmp/main.go",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.PointerTo(true),
			Language:       heartbeat.PointerTo("Go"),
			LineNumber:     heartbeat.PointerTo(42),
			Lines:          heartbeat.PointerTo(100),
			Project:        heartbeat.PointerTo("wakatime-cli"),
			Time:           1592868367.219124,
			UserAgent:      "wakatime/13.0.6",
		},
		{
			Branch:         heartbeat.PointerTo("summary"),
			Category:       heartbeat.DebuggingCategory.String(),
			CursorPosition: heartbeat.PointerTo(13),
			Dependencies:   []string{"dep3", "dep4"},
			Entity:         "/tmp/main.py",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.PointerTo(false),
			Language:       heartbeat.PointerTo("Python"),
			LineNumber:     heartbeat.PointerTo(43),
			Lines:          heartbeat.PointerTo(101),
			Project:        heartbeat.PointerTo("wakatime"),
			Time:           1592868386.079084,
			UserAgent:      "wakatime/13.0.7",
		},
		{
			Branch:         heartbeat.PointerTo("todaygoal"),
			Category:       heartbeat.BuildingCategory.String(),
			CursorPosition: heartbeat.PointerTo(14),
			Dependencies:   []string{"dep5", "dep6"},
			Entity:         "/tmp/main.js",
			EntityType:     heartbeat.FileType,
			IsWrite:        heartbeat.PointerTo(false),
			Language:       heartbeat.PointerTo("JavaScript"),
			LineNumber:     heartbeat.PointerTo(44),
			Lines:          heartbeat.PointerTo(102),
			Project:        heartbeat.PointerTo("wakatime"),
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
			return fmt.Errorf("failed put heartbeat: %s", err)
		}

		return nil
	})
	require.NoError(t, err)
}

func corruptBoltPageFlags(t *testing.T, path string, pageID uint64, flags uint16) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_WRONLY, 0600)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, f.Close())
	}()

	var data [2]byte
	binary.LittleEndian.PutUint16(data[:], flags)

	offset := int64(pageID)*int64(os.Getpagesize()) + 8
	n, err := f.WriteAt(data[:], offset)
	require.NoError(t, err)
	require.Equal(t, len(data), n)
}
