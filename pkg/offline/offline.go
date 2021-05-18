package offline

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

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

// WithQueue initializes and returns a heartbeat handle option, which can be
// used in a heartbeat processing pipeline for automatic handling of failures
// of heartbeat sending to the API. Upon inability to send due to missing or
// failing connection to API, failed sending or errors returned by API, the
// heartbeats will be temporarily stored in a DB and sending will be retried
// at next usages of the wakatime cli.
func WithQueue(filepath string, syncLimit int) (heartbeat.HandleOption, error) {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute offline queue")

			db, err := bolt.Open(filepath, 0600, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to open db connection: %s", err)
			}

			defer db.Close()

			// start transaction
			tx, err := db.Begin(true)
			if err != nil {
				log.Errorf("failed to start offline queue db transaction: %s", err)
				return next(hh)
			}

			// nolint
			defer tx.Rollback()

			queue, err := NewQueue(tx)
			if err != nil {
				return nil, fmt.Errorf("failed initialize new queue: %s", err)
			}

			queued, err := queue.PopMany(syncLimit)
			if err != nil {
				log.Errorf("failed to pop heartbeat(s) from offline queue: %s", err)
			}

			if len(queued) > 0 {
				log.Debugf("include %d heartbeat(s) from offline queue", len(queued))
				hh = append(hh, queued...)
			}

			results, err := next(hh)
			if err != nil {
				log.Debugf("api error: %s", err)
				log.Debugf("pushing %d heartbeat(s) to offline queue", len(hh))

				// push to queue on any err
				queueErr := queue.PushMany(hh)
				if queueErr != nil {
					log.Errorf("failed to push heartbeat(s) to queue: %s", queueErr)
				}

				// commit transaction
				if err := tx.Commit(); err != nil {
					log.Errorf("failed to commit offline queue db transaction: %s", err)
				}

				return nil, err
			}

			// push heartbeats with invalid result status codes to queue
			var withInvalidStatus []heartbeat.Heartbeat

			for n, result := range results {
				if n >= len(hh) {
					log.Warnln("results from api not matching heartbeats sent")
					break
				}

				if result.Status != http.StatusCreated &&
					result.Status != http.StatusAccepted &&
					result.Status != http.StatusBadRequest {
					withInvalidStatus = append(withInvalidStatus, hh[n])
				}
			}

			if len(withInvalidStatus) > 0 {
				log.Debugf("pushing %d heartbeat(s) with invalid result to offline queue", len(withInvalidStatus))

				err = queue.PushMany(withInvalidStatus)
				if err != nil {
					log.Errorf("failed to push invalid result heartbeat(s) to queue: %s", err)
				}
			}

			// handle leftovers
			leftovers := len(hh) - len(results)
			if leftovers > 0 {
				log.Warnf("Missing %d results from api.", leftovers)

				start := len(hh) - leftovers

				queueErr := queue.PushMany(hh[start:])
				if queueErr != nil {
					log.Errorf("failed to push leftover heartbeat to queue: %s", queueErr)
				}
			}

			// commit transaction
			if err = tx.Commit(); err != nil {
				log.Errorf("failed to commit offline queue db transaction: %s", err)
			}

			return results, nil
		}
	}, nil
}

// Queue is a db client to temporarily store heartbeats in bolt db, in case heartbeat
// sending to wakatime api is not possible. Transaction handling is left to the user
// via the passed in transaction.
type Queue struct {
	Bucket string
	tx     *bolt.Tx
}

// NewQueue creates a new instance of Queue.
func NewQueue(tx *bolt.Tx) (*Queue, error) {
	return &Queue{
		Bucket: dbBucket,
		tx:     tx,
	}, nil
}

// PopMany retrieves heartbeats with the specified ids from db.
func (q *Queue) PopMany(limit int) ([]heartbeat.Heartbeat, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return nil, fmt.Errorf("failed to create/load bucket: %s", err)
	}

	var (
		heartbeats []heartbeat.Heartbeat
		ids        []string
	)

	// load values
	c := b.Cursor()

	for key, value := c.First(); key != nil; key, value = c.Next() {
		if len(heartbeats) >= limit {
			break
		}

		var h heartbeat.Heartbeat

		err := json.Unmarshal(value, &h)
		if err != nil {
			return nil, fmt.Errorf("failed to json unmarshal heartbeat data: %s", err)
		}

		heartbeats = append(heartbeats, h)
		ids = append(ids, string(key))
	}

	for _, id := range ids {
		if err := b.Delete([]byte(id)); err != nil {
			return nil, fmt.Errorf("failed to delete key %q: %s", id, err)
		}
	}

	return heartbeats, nil
}

// PushMany stores the provided heartbeats in the db.
func (q *Queue) PushMany(hh []heartbeat.Heartbeat) error {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return fmt.Errorf("failed to create/load bucket: %s", err)
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
