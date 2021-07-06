package offline

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/mitchellh/go-homedir"
	bolt "go.etcd.io/bbolt"
)

const (
	// SyncMaxDefault is the default maximum number of heartbeats from the
	// offline queue, which will be synced upon sending heartbeats to the API.
	SyncMaxDefault = 1000
)

const (
	// dbFilename is the default bolt db filename.
	dbFilename = ".wakatime.bdb"
	// dbBucket is the standard bolt db bucket name.
	dbBucket = "heartbeats"
	// maxRequeueAttempts defines the maximum number of attempts to requeue heartbeats,
	// which could not successfully be sent to the WakaTime API.
	maxRequeueAttempts = 3
	// sendLimit is the maximum number of heartbeats, which will be sent at once
	// to the WakaTime API.
	sendLimit = 24
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
			log.Debugf("execute offline queue with file %s", filepath)

			if len(hh) == 0 {
				log.Debugln("abort execution, as there are no heartbeats ready for sending")

				return nil, nil
			}

			results, err := next(hh)
			if err != nil {
				log.Debugf("pushing %d heartbeat(s) to queue due to error", len(hh))

				requeueErr := pushHeartbeatsWithRetry(filepath, hh)
				if requeueErr != nil {
					log.Errorf("failed to push heatbeats to queue after api error: %s", requeueErr)
				}

				return nil, err
			}

			err = handleResults(filepath, results, hh)
			if err != nil {
				return nil, fmt.Errorf("failed to handle results: %s", err)
			}

			return results, nil
		}
	}, nil
}

// Sync returns a function to send queued heartbeats to the WakaTime API.
func Sync(filepath string, syncLimit int) func(next heartbeat.Handle) error {
	return func(next heartbeat.Handle) error {
		log.Debugf("execute offline sync with file %s", filepath)

		var (
			alreadySent int
			run         int
		)

		for {
			run++

			if alreadySent >= syncLimit {
				break
			}

			var num = sendLimit

			if alreadySent+sendLimit > syncLimit {
				num = syncLimit - alreadySent
				alreadySent += num
			}

			hh, err := popHeartbeats(filepath, num)
			if err != nil {
				return fmt.Errorf("failed to fetch heartbeat from offline queue: %s", err)
			}

			if len(hh) == 0 {
				log.Debugln("no queued heartbeats ready for sending")

				break
			}

			log.Debugf("send %d heartbeats on sync run %d", len(hh), run)

			results, err := next(hh)
			if err != nil {
				requeueErr := pushHeartbeatsWithRetry(filepath, hh)
				if requeueErr != nil {
					log.Warnf("failed to push heatbeats to queue after api error: %s", requeueErr)
				}

				return err
			}

			err = handleResults(filepath, results, hh)
			if err != nil {
				return fmt.Errorf("failed to handle heatbeats api results: %s", err)
			}
		}

		return nil
	}
}

func handleResults(filepath string, results []heartbeat.Result, hh []heartbeat.Heartbeat) error {
	var (
		err               error
		withInvalidStatus []heartbeat.Heartbeat
	)

	// push heartbeats with invalid result status codes to queue
	for n, result := range results {
		if n >= len(hh) {
			log.Warnln("results from api not matching heartbeats sent")
			break
		}

		if result.Status == http.StatusBadRequest {
			serialized, jsonErr := json.Marshal(result.Heartbeat)
			if jsonErr != nil {
				log.Warnf(
					"failed to json marshal heartbeat: %s. heartbeat: %#v",
					jsonErr,
					result.Heartbeat,
				)
			}

			log.Debugf("heartbeat result status bad request: %s", string(serialized))

			continue
		}

		if result.Status != http.StatusCreated &&
			result.Status != http.StatusAccepted {
			withInvalidStatus = append(withInvalidStatus, hh[n])
		}
	}

	if len(withInvalidStatus) > 0 {
		log.Debugf("pushing %d heartbeat(s) with invalid result to queue", len(withInvalidStatus))

		err = pushHeartbeatsWithRetry(filepath, withInvalidStatus)
		if err != nil {
			log.Warnf("failed to push heatbeats with invalid status to queue: %s", err)
		}
	}

	// handle leftover heartbeats
	leftovers := len(hh) - len(results)
	if leftovers > 0 {
		log.Warnf("missing %d results from api.", leftovers)

		start := len(hh) - leftovers

		err = pushHeartbeatsWithRetry(filepath, hh[start:])
		if err != nil {
			log.Warnf("failed to push leftover heatbeats to queue: %s", err)
		}
	}

	return err
}

func popHeartbeats(filepath string, limit int) ([]heartbeat.Heartbeat, error) {
	db, err := bolt.Open(filepath, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %s", err)
	}

	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, fmt.Errorf("failed to start db transaction: %s", err)
	}

	queue := NewQueue(tx)

	queued, err := queue.PopMany(limit)
	if err != nil {
		errrb := tx.Rollback()
		if errrb != nil {
			log.Errorf("failed to rollback transaction: %s", errrb)
		}

		return nil, fmt.Errorf("failed to pop heartbeat(s) from queue: %s", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit db transaction: %s", err)
	}

	return queued, nil
}

func pushHeartbeatsWithRetry(filepath string, hh []heartbeat.Heartbeat) error {
	var (
		count int
		err   error
	)

	for {
		if count >= maxRequeueAttempts {
			serialized, jsonErr := json.Marshal(hh)
			if jsonErr != nil {
				log.Warnf("failed to json marshal heartbeats: %s. heartbeats: %#v", jsonErr, hh)
			}

			return fmt.Errorf(
				"abort requeuing after %d unsuccessful attempts: %s. heartbeats: %s",
				count,
				err,
				string(serialized),
			)
		}

		err = pushHeartbeats(filepath, hh)
		if err != nil {
			count++

			sleepSeconds := math.Pow(2, float64(count))

			time.Sleep(time.Duration(sleepSeconds) * time.Second)

			continue
		}

		break
	}

	return nil
}

func pushHeartbeats(filepath string, hh []heartbeat.Heartbeat) error {
	db, err := bolt.Open(filepath, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %s", err)
	}

	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return fmt.Errorf("failed to start db transaction: %s", err)
	}

	queue := NewQueue(tx)

	err = queue.PushMany(hh)
	if err != nil {
		return fmt.Errorf("failed to push heartbeat(s) to queue: %s", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit db transaction: %s", err)
	}

	return nil
}

// CountHeartbeats returns the total number of heartbeats in the offline db.
func CountHeartbeats(filepath string) (int, error) {
	db, err := bolt.Open(filepath, 0600, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to open db connection: %s", err)
	}

	defer db.Close()

	tx, err := db.Begin(true)
	if err != nil {
		return 0, fmt.Errorf("failed to start db transaction: %s", err)
	}

	queue := NewQueue(tx)

	count, err := queue.Count()
	if err != nil {
		log.Errorf("failed to count offline heartbeats: %s", err)

		_ = tx.Rollback()

		return count, err
	}

	err = tx.Rollback()
	if err != nil {
		log.Warnf("failed to rollback transaction: %s", err)
	}

	return count, nil
}

// Queue is a db client to temporarily store heartbeats in bolt db, in case heartbeat
// sending to wakatime api is not possible. Transaction handling is left to the user
// via the passed in transaction.
type Queue struct {
	Bucket string
	tx     *bolt.Tx
}

// NewQueue creates a new instance of Queue.
func NewQueue(tx *bolt.Tx) *Queue {
	return &Queue{
		Bucket: dbBucket,
		tx:     tx,
	}
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

// Count returns the total number of heartbeats in the offline db.
func (q *Queue) Count() (int, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return 0, fmt.Errorf("failed to create/load bucket: %s", err)
	}

	return b.Stats().KeyN, nil
}
