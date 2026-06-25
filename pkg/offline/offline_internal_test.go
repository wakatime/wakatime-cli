package offline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	bolt "go.etcd.io/bbolt"
)

func TestQueueMethodsReadOnlyTransactionErrors(t *testing.T) {
	db := openTestQueueDB(t)
	defer db.Close()

	require.NoError(t, db.View(func(tx *bolt.Tx) error {
		queue := NewQueue(tx)

		_, err := queue.Count()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create/load bucket")

		_, err = queue.PopMany(1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create/load bucket")

		err = queue.PushMany([]heartbeat.Heartbeat{{Entity: "main.go"}})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create/load bucket")

		_, err = queue.ReadMany(1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create/load bucket")

		_, err = queue.DeleteDuplicates()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create/load bucket")

		return nil
	}))
}

func TestQueueInvalidJSONErrors(t *testing.T) {
	tests := map[string]func(*Queue) error{
		"pop": func(queue *Queue) error {
			_, err := queue.PopMany(1)
			return err
		},
		"read": func(queue *Queue) error {
			_, err := queue.ReadMany(1)
			return err
		},
		"dedupe": func(queue *Queue) error {
			_, err := queue.DeleteDuplicates()
			return err
		},
	}

	for name, run := range tests {
		t.Run(name, func(t *testing.T) {
			db := openTestQueueDB(t)
			defer db.Close()

			require.NoError(t, db.Update(func(tx *bolt.Tx) error {
				bucket, err := tx.CreateBucketIfNotExists([]byte(dbBucket))
				require.NoError(t, err)

				return bucket.Put([]byte("1"), []byte("{"))
			}))

			require.NoError(t, db.Update(func(tx *bolt.Tx) error {
				err := run(NewQueue(tx))
				require.Error(t, err)
				assert.Contains(t, err.Error(), "failed to json unmarshal heartbeat data")

				return nil
			}))
		})
	}
}

func TestQueueDeleteDuplicatesRemovesNearbyHeartbeat(t *testing.T) {
	db := openTestQueueDB(t)
	defer db.Close()

	h1 := heartbeat.Heartbeat{Entity: "main.go", EntityType: heartbeat.FileType, Time: 10, UserAgent: "test/1"}
	h2 := heartbeat.Heartbeat{Entity: "main.go", EntityType: heartbeat.FileType, Time: 10.5, UserAgent: "test/1"}
	h3 := heartbeat.Heartbeat{Entity: "main.go", EntityType: heartbeat.FileType, Time: 30, UserAgent: "test/1"}

	require.NoError(t, db.Update(func(tx *bolt.Tx) error {
		queue := NewQueue(tx)
		require.NoError(t, queue.PushMany([]heartbeat.Heartbeat{h1, h2, h3}))

		deleted, err := queue.DeleteDuplicates()
		require.NoError(t, err)
		assert.Equal(t, 1, deleted)

		remaining, err := queue.ReadMany(10)
		require.NoError(t, err)
		assert.Len(t, remaining, 2)

		return nil
	}))
}

func TestResetCorruptDBHelpers(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.bdb")
	backup, err := resetCorruptDB(missing)
	require.NoError(t, err)
	assert.Empty(t, backup)

	resetErr := resetCorruptDBError(missing, assert.AnError)
	assert.True(t, resetErr.Reset)
	assert.Contains(t, resetErr.Error(), "corrupt db file already removed")

	dbPath := filepath.Join(t.TempDir(), "offline.bdb")
	require.NoError(t, os.WriteFile(dbPath, []byte("corrupt"), 0600))
	backup, err = resetCorruptDB(dbPath)
	require.NoError(t, err)
	assert.NotEmpty(t, backup)
	assert.NoFileExists(t, dbPath)
	assert.FileExists(t, backup)
}

func openTestQueueDB(t *testing.T) *bolt.DB {
	t.Helper()

	db, err := bolt.Open(filepath.Join(t.TempDir(), "offline.bdb"), 0600, nil)
	require.NoError(t, err)

	return db
}
