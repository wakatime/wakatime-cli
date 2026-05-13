package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	bolt "go.etcd.io/bbolt"
)

const (
	// dbFilename is the default bolt db filename.
	dbFilename = "offline_heartbeats.bdb"
	// dbBucket is the standard bolt db bucket name.
	dbBucket = "heartbeats"
	// maxRequeueAttempts defines the maximum number of attempts to requeue heartbeats,
	// which could not successfully be sent to the WakaTime API.
	maxRequeueAttempts = 3
	// PrintMaxDefault is the default maximum number of heartbeats to print.
	PrintMaxDefault = 10
	// RateLimitDefaultSeconds is the default seconds between sending heartbeats
	// to the API. If not enough time has passed, heartbeats are saved to the offline queue.
	RateLimitDefaultSeconds = 120
	// SendLimit is the maximum number of heartbeats, which will be sent at once
	// to the WakaTime API.
	SendLimit = 10
	// SyncMaxDefault is the default maximum number of heartbeats from the
	// offline queue, which will be synced upon sending heartbeats to the API.
	SyncMaxDefault = 1000
	// readDedupeWindowSeconds defines the max distance between two heartbeats
	// for the same entity to be considered duplicates during read cleanup.
	readDedupeWindowSeconds = 1
)

// WithQueue initializes and returns a heartbeat handle option, which can be
// used in a heartbeat processing pipeline for automatic handling of failures
// of heartbeat sending to the API. Upon inability to send due to missing or
// failing connection to API, failed sending or errors returned by API, the
// heartbeats will be temporarily stored in a DB and sending will be retried
// at next usages of the wakatime cli.
func WithQueue(filepath string) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			// logger.Debugf("execute offline queue with file %s", filepath)

			if len(hh) == 0 {
				logger.Debugln("abort execution, as there are no heartbeats ready for sending")

				return nil, nil
			}

			results, err := next(ctx, hh)
			if err != nil {
				logger.Debugf("pushing %d heartbeat(s) to queue after error: %s", len(hh), err)

				requeueErr := pushHeartbeatsWithRetry(ctx, filepath, hh)
				if requeueErr != nil {
					return nil, fmt.Errorf(
						"failed to push heartbeats to queue: %s",
						requeueErr,
					)
				}

				return nil, err
			}

			_, err = handleResults(ctx, filepath, results, hh)
			if err != nil {
				return nil, fmt.Errorf("failed to handle results: %s", err)
			}

			return results, nil
		}
	}
}

// QueueFilepath returns the path for offline queue db file. If
// the resource directory cannot be detected, it defaults to the
// current directory.
func QueueFilepath(ctx context.Context, v *viper.Viper) (string, error) {
	paramFile := vipertools.GetString(v, "offline-queue-file")
	if paramFile != "" {
		p, err := homedir.Expand(paramFile)
		if err != nil {
			return "", fmt.Errorf("failed expanding offline-queue-file param: %s", err)
		}

		return p, nil
	}

	folder, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return dbFilename, fmt.Errorf("failed getting resource directory, defaulting to current directory: %s", err)
	}

	return filepath.Join(folder, dbFilename), nil
}

// WithSync initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to pop heartbeats
// from offline queue and send the heartbeats to WakaTime API.
func WithSync(filepath string, syncLimit int) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, _ []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			// logger.Debugf("execute offline sync with file %s", filepath)
			err := Sync(ctx, filepath, syncLimit)(next)
			if err != nil {
				return nil, fmt.Errorf("failed to sync offline heartbeats: %s", err)
			}

			return nil, nil
		}
	}
}

// Sync returns a function to send queued heartbeats to the WakaTime API.
func Sync(ctx context.Context, filepath string, syncLimit int) func(next heartbeat.Handle) error {
	return func(next heartbeat.Handle) error {
		var (
			alreadySent int
			run         int
		)

		logger := log.Extract(ctx)

		for {
			run++

			if alreadySent >= syncLimit {
				break
			}

			var num = SendLimit

			if alreadySent+SendLimit > syncLimit {
				num = syncLimit - alreadySent
				alreadySent += num
			}

			hh, err := popHeartbeats(ctx, filepath, num)
			if err != nil {
				return fmt.Errorf("failed to fetch heartbeat from offline queue: %s", err)
			}

			if len(hh) == 0 {
				logger.Debugln("no queued heartbeats ready for sending")

				break
			}

			logger.Debugf("send %d heartbeats on sync run %d", len(hh), run)

			results, err := next(ctx, hh)
			if err != nil {
				requeueErr := pushHeartbeatsWithRetry(ctx, filepath, hh)
				if requeueErr != nil {
					logger.Warnf("failed to push heartbeats to queue after api error: %s", requeueErr)
				}

				return err
			}

			stopSending, err := handleResults(ctx, filepath, results, hh)
			if err != nil {
				return fmt.Errorf("failed to handle heartbeats api results: %s", err)
			}

			if stopSending {
				break
			}

			alreadySent += len(hh)
		}

		return nil
	}
}

func handleResults(
	ctx context.Context,
	filepath string,
	results []heartbeat.Result,
	hh []heartbeat.Heartbeat,
) (bool, error) {
	var (
		err               error
		withInvalidStatus []heartbeat.Heartbeat
		stopSending       bool
	)

	logger := log.Extract(ctx)

	if len(results) < len(hh) {
		logger.Warnf("expected %d results from api but received %d", len(hh), len(results))
	}

	handled := make(map[string]int, len(results))

	// push heartbeats with invalid result status codes to queue
	for n, result := range results {
		h, ok := resultHeartbeat(result, hh, n)
		if !ok {
			stopSending = true

			break
		}

		handled[h.ID()]++

		if result.Status == http.StatusBadRequest {
			serialized, jsonErr := json.Marshal(h)
			if jsonErr != nil {
				logger.Warnf(
					"failed to json marshal heartbeat: %s. heartbeat: %#v",
					jsonErr,
					h,
				)
			}

			logger.Debugf("heartbeat result status bad request: %s", string(serialized))

			continue
		}

		if result.Status < http.StatusOK || result.Status > 299 {
			withInvalidStatus = append(withInvalidStatus, h)

			logger.Debugf("heartbeat %d has invalid status code %d: %s", n, result.Status, strings.Join(result.Errors, ", "))
		}
	}

	if len(withInvalidStatus) > 0 {
		logger.Debugf("pushing %d heartbeat(s) with invalid result to queue", len(withInvalidStatus))

		stopSending = true

		err = pushHeartbeatsWithRetry(ctx, filepath, withInvalidStatus)
		if err != nil {
			logger.Warnf("failed to push heartbeats with invalid status to queue: %s", err)
		}
	}

	// handle leftover heartbeats
	leftovers := missingResultHeartbeats(hh, handled)
	if len(leftovers) > 0 {
		logger.Warnf("missing %d results from api.", len(leftovers))

		err = pushHeartbeatsWithRetry(ctx, filepath, leftovers)
		if err != nil {
			logger.Warnf("failed to push leftover heartbeats to queue: %s", err)
		}
	}

	return err != nil && !stopSending, err
}

func resultHeartbeat(result heartbeat.Result, hh []heartbeat.Heartbeat, n int) (heartbeat.Heartbeat, bool) {
	if result.Heartbeat.Entity != "" || result.Heartbeat.Time != 0 {
		return result.Heartbeat, true
	}

	if n >= len(hh) {
		return heartbeat.Heartbeat{}, false
	}

	return hh[n], true
}

func missingResultHeartbeats(hh []heartbeat.Heartbeat, handled map[string]int) []heartbeat.Heartbeat {
	var missing []heartbeat.Heartbeat

	for _, h := range hh {
		id := h.ID()
		if handled[id] > 0 {
			handled[id]--
			continue
		}

		missing = append(missing, h)
	}

	return missing
}

func popHeartbeats(ctx context.Context, filepath string, limit int) ([]heartbeat.Heartbeat, error) {
	db, close, err := openDB(ctx, filepath)
	if err != nil {
		return nil, err
	}

	defer close()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, fmt.Errorf("failed to start db transaction: %s", err)
	}

	queue := NewQueue(tx)
	logger := log.Extract(ctx)

	_, err = queue.DeleteDuplicates()
	if err != nil {
		logger.Errorf("failed to delete duplicate offline heartbeats: %s", err)

		_ = tx.Rollback()

		return nil, err
	}

	queued, err := queue.PopMany(limit)
	if err != nil {
		errrb := tx.Rollback()
		if errrb != nil {
			logger.Errorf("failed to rollback transaction: %s", errrb)
		}

		return nil, fmt.Errorf("failed to pop heartbeat(s) from queue: %s", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit db transaction: %s", err)
	}

	return queued, nil
}

func pushHeartbeatsWithRetry(ctx context.Context, filepath string, hh []heartbeat.Heartbeat) error {
	var (
		count int
		err   error
	)

	logger := log.Extract(ctx)

	for {
		if count >= maxRequeueAttempts {
			serialized, jsonErr := json.Marshal(hh)
			if jsonErr != nil {
				logger.Warnf("failed to json marshal heartbeats: %s. heartbeats: %#v", jsonErr, hh)
			}

			return fmt.Errorf(
				"abort requeuing after %d unsuccessful attempts: %s. heartbeats: %s",
				count,
				err,
				string(serialized),
			)
		}

		err = pushHeartbeats(ctx, filepath, hh)
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

func pushHeartbeats(ctx context.Context, filepath string, hh []heartbeat.Heartbeat) error {
	db, close, err := openDB(ctx, filepath)
	if err != nil {
		return err
	}

	defer close()

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
func CountHeartbeats(ctx context.Context, filepath string) (int, error) {
	db, close, err := openDB(ctx, filepath)
	if err != nil {
		return 0, err
	}

	defer close()

	tx, err := db.Begin(true)
	if err != nil {
		return 0, fmt.Errorf("failed to start db transaction: %s", err)
	}

	logger := log.Extract(ctx)

	defer func() {
		err := tx.Rollback()
		if err != nil {
			logger.Errorf("failed to rollback transaction: %s", err)
		}
	}()

	queue := NewQueue(tx)

	count, err := queue.Count()
	if err != nil {
		return 0, fmt.Errorf("failed to count heartbeats: %s", err)
	}

	return count, nil
}

// ReadHeartbeats reads the informed heartbeats in the offline db.
func ReadHeartbeats(ctx context.Context, filepath string, limit int) ([]heartbeat.Heartbeat, error) {
	db, close, err := openDB(ctx, filepath)
	if err != nil {
		return nil, err
	}

	defer close()

	tx, err := db.Begin(true)
	if err != nil {
		return nil, fmt.Errorf("failed to start db transaction: %s", err)
	}

	queue := NewQueue(tx)
	logger := log.Extract(ctx)

	hh, err := queue.ReadMany(limit)
	if err != nil {
		logger.Errorf("failed to read offline heartbeats: %s", err)

		_ = tx.Rollback()

		return nil, err
	}

	err = tx.Rollback()
	if err != nil {
		logger.Warnf("failed to rollback transaction: %s", err)
	}

	return hh, nil
}

// openDB opens a connection to the offline db.
// It returns the pointer to bolt.DB, a function to close the connection and an error.
// Although named parameters should be avoided, this func uses them to access inside the deferred function and set an error.
func openDB(ctx context.Context, filepath string) (db *bolt.DB, _ func(), err error) {
	defer func() {
		if r := recover(); r != nil {
			err = ErrOpenDB{Err: fmt.Errorf("panicked: %v", r)}
		}
	}()

	db, err = bolt.Open(filepath, 0644, &bolt.Options{Timeout: 30 * time.Second})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open db file: %s", err)
	}

	return db, func() {
		logger := log.Extract(ctx)

		// recover from panic when closing db
		defer func() {
			if r := recover(); r != nil {
				logger.Warnf("panicked: failed to close db file: %v", r)
			}
		}()

		if err := db.Close(); err != nil {
			logger.Debugf("failed to close db file: %s", err)
		}
	}, err
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

// Count returns the total number of heartbeats in the offline db.
func (q *Queue) Count() (int, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return 0, fmt.Errorf("failed to create/load bucket: %s", err)
	}

	return b.Stats().KeyN, nil
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

// ReadMany reads heartbeats from db without deleting them.
func (q *Queue) ReadMany(limit int) ([]heartbeat.Heartbeat, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return nil, fmt.Errorf("failed to create/load bucket: %s", err)
	}

	var heartbeats = make([]heartbeat.Heartbeat, 0)

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
	}

	return heartbeats, nil
}

// DeleteDuplicates cleanups and deletes duplicate heartbeats from the offline db.
func (q *Queue) DeleteDuplicates() (int, error) {
	b, err := q.tx.CreateBucketIfNotExists([]byte(q.Bucket))
	if err != nil {
		return 0, fmt.Errorf("failed to create/load bucket: %s", err)
	}

	kept := make(map[string][]float64)
	deleted := 0

	c := b.Cursor()
	for key, value := c.First(); key != nil; key, value = c.Next() {
		var h heartbeat.Heartbeat

		err := json.Unmarshal(value, &h)
		if err != nil {
			return 0, fmt.Errorf("failed to json unmarshal heartbeat data: %s", err)
		}

		isDuplicate := false

		for _, keptTime := range kept[h.Entity] {
			if math.Abs(h.Time-keptTime) <= readDedupeWindowSeconds {
				isDuplicate = true
				break
			}
		}

		if isDuplicate {
			if err := c.Delete(); err != nil {
				return 0, fmt.Errorf("failed to delete duplicate heartbeat with key %q: %s", string(key), err)
			}

			deleted++

			continue
		}

		kept[h.Entity] = append(kept[h.Entity], h.Time)
	}

	return deleted, nil
}
