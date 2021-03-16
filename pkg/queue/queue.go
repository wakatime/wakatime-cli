package queue

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	"github.com/mitchellh/go-homedir"
	bolt "go.etcd.io/bbolt"
)

const (
	dbFilename = ".wakatime.bdb"
	dbBucket   = "heartbeats"
)

// QueueFilepath returns the path for offline queue db file.
func QueueFilepath() (string, error) {
	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		p, err := homedir.Expand(home)
		if err != nil {
			return "", fmt.Errorf("failed parsing WAKATIME_HOME environment variable: %s", err)
		}

		return filepath.Join(p, dbFilename), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return filepath.Join(home, dbFilename), nil
}

// Queue is a db client to temporarily store heartbeats in bolt db, in case heartbeat
// sending to wakatime api is not possible. Transaction handling is left to the user
// via the passed in transaction.
type Queue struct {
	tx *bolt.Tx
}

// NewQueue creates a new instance of Queue.
func NewQueue(tx *bolt.Tx) *Queue {
	return &Queue{
		tx: tx,
	}
}

// Delete removes heartbeats with the specified ids from the db.
func (q *Queue) Delete(ids []string) error {
	b := q.tx.Bucket([]byte(dbBucket))
	if b == nil {
		return fmt.Errorf("failed to load bucket %q", dbBucket)
	}

	for _, id := range ids {
		if err := b.Delete([]byte(id)); err != nil {
			return fmt.Errorf("failed to delete key %q: %s", id, err)
		}
	}

	return nil
}

// LoadMany retrieves heartbeats with the specified ids from db.
func (q *Queue) LoadMany(limit int) ([]string, []heartbeat.Heartbeat, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(dbBucket))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load/create bucket: %s", err)
	}

	var (
		ids        []string
		heartbeats []heartbeat.Heartbeat
	)

	// load values
	c := b.Cursor()

	for key, value := c.First(); key != nil; key, value = c.Next() {
		if len(ids) >= limit {
			break
		}

		var h heartbeat.Heartbeat

		err := json.Unmarshal(value, &h)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to json unmarshal heartbeat data: %s", err)
		}

		ids = append(ids, string(key))
		heartbeats = append(heartbeats, h)
	}

	return ids, heartbeats, nil
}

// StoreMany stores the provided heartbeats in the db.
func (q *Queue) StoreMany(hh []heartbeat.Heartbeat) error {
	b, err := q.tx.CreateBucketIfNotExists([]byte(dbBucket))
	if err != nil {
		return fmt.Errorf("failed to load/create bucket: %s", err)
	}

	for _, h := range hh {
		data, err := json.Marshal(h)
		if err != nil {
			return fmt.Errorf("failed to json marshal heartbeat: %s", err)
		}

		err = b.Put([]byte(h.ID()), data)
		if err != nil {
			return fmt.Errorf("failed to store heartbeat with id %q: %s", h.ID(), err)
		}
	}

	return nil
}
