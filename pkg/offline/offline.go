package offline

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	_ "github.com/mattn/go-sqlite3" // not used directly
	"github.com/mitchellh/go-homedir"
	jww "github.com/spf13/jwalterweatherman"
)

const (
	tableName = "heartbeat_2"
)

// QueueFilepath returns the path to the offline queue db file.
func QueueFilepath() (string, error) {
	dir := os.Getenv("WAKATIME_HOME")

	var err error
	if dir == "" {
		dir, err = os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to retrieve user's home dir: %s", err)
		}
	}

	expanded, err := homedir.Expand(dir)
	if err != nil {
		return "", fmt.Errorf("failed to expand offline queue folder path: %s", err)
	}

	return path.Join(expanded, ".wakatime.db"), nil
}

// WithQueue initializes and returns a heartbeat handle option, which can be
// used in a heartbeat processing pipeline for automatic handling of failures
// of heartbeat sending to the API. Upon inability to send due to missing or
// failing connection to API, failed sending or errors returned by API, the
// heartbeats will be temporarily stored in a sqlite DB and sending will be
// retried at next usages of the wakatime cli.
func WithQueue(filepath string, syncLimit int) (heartbeat.HandleOption, error) {
	conn, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %s", err)
	}

	_, err = conn.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (id TEXT, heartbeat TEXT)", tableName))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize db: %s", err)
	}

	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			// start transaction
			tx, err := conn.Begin()
			if err != nil {
				jww.ERROR.Fatalf("failed to start offline queue db transaction: %s", err)
				return next(hh)
			}
			// nolint
			defer tx.Rollback()

			queue := NewQueue(tx)

			queued, err := queue.PopMany(syncLimit)
			if err != nil {
				jww.ERROR.Fatalf("failed to pop heartbeat(s) from offline queue: %s", err)
			}

			if len(queued) > 0 {
				jww.DEBUG.Printf("include %d heartbeat(s) from offline queue", len(queued))
				hh = append(hh, queued...)
			}

			results, err := next(hh)
			if err != nil {
				jww.DEBUG.Printf("api error: %s", err)
				jww.DEBUG.Printf("pushing %d heartbeat(s) to offline queue", len(hh))

				// push to queue on any err
				queueErr := queue.PushMany(hh)
				if queueErr != nil {
					jww.ERROR.Fatalf("failed to push heartbeat(s) to queue: %s", queueErr)
				}

				// commit transaction
				if err := tx.Commit(); err != nil {
					jww.ERROR.Fatalf("failed to commit offline queue db transaction: %s", err)
				}

				return nil, err
			}

			for _, result := range results {
				// push to queue on invalid result status codes
				if result.Status != http.StatusCreated &&
					result.Status != http.StatusAccepted &&
					result.Status != http.StatusBadRequest {
					queueErr := queue.PushMany([]heartbeat.Heartbeat{result.Heartbeat})
					if queueErr != nil {
						jww.ERROR.Fatalf("failed to push invalid result heartbeat to queue: %s", queueErr)
					}
				}
			}

			// commit transaction
			if err = tx.Commit(); err != nil {
				jww.ERROR.Fatalf("failed to commit offline queue db transaction: %s", err)
			}

			return results, nil
		}
	}, nil
}

// DB is a minimal database connection interface satisfied by both sql.DB and sql.Tx.
type DB interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (*sql.Stmt, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Queue is a queue to temporarily store heartbeats.
type Queue struct {
	conn DB
}

// NewQueue creates a new Queue instance.
func NewQueue(conn DB) *Queue {
	return &Queue{
		conn: conn,
	}
}

// PushMany adds multiple heartbeats to the queue.
func (q *Queue) PushMany(hh []heartbeat.Heartbeat) error {
	stmt, err := q.conn.Prepare(fmt.Sprintf("INSERT INTO %s VALUES ($1, $2);", tableName))
	if err != nil {
		return fmt.Errorf("failed to prepare db statement: %s", err)
	}
	defer stmt.Close()

	for _, h := range hh {
		data, err := json.Marshal(h)
		if err != nil {
			return fmt.Errorf("failed to json encode heartbeat: %s", err)
		}

		result, err := stmt.Exec(h.ID(), data)
		if err != nil {
			return fmt.Errorf("failed to execute db query: %s", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("checking number of affected rows failed: %s", err)
		}

		if affected != 1 {
			return fmt.Errorf("unexpected number of affected rows. got: %d, want: %d", affected, 1)
		}
	}

	return nil
}

// PopMany takes multiple heartbeats from the queue.
func (q *Queue) PopMany(limit int) ([]heartbeat.Heartbeat, error) {
	rows, err := q.conn.Query(fmt.Sprintf("SELECT id, heartbeat FROM %s LIMIT $1;", tableName), limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute select db query: %s", err)
	}

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
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %s", err)
		}

		ids = append(ids, id)

		var h heartbeat.Heartbeat

		err = json.Unmarshal([]byte(data), &h)
		if err != nil {
			return nil, fmt.Errorf("failed to parse heartbeat json data: %s", err)
		}

		heartbeats = append(heartbeats, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row error: %s", err)
	}

	for _, id := range ids {
		_, err = q.conn.Exec(fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName), id)
		if err != nil {
			return nil, fmt.Errorf("failed to execute delete db query: %s", err)
		}
	}

	return heartbeats, nil
}
