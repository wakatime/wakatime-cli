package offline

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

const (
	tableName = "heartbeat_2"
)

// Queue is a queue to temporarily store heartbeats.
type Queue struct {
	conn *sql.DB
}

// NewQueue creates a new Queue instance.
func NewQueue(conn *sql.DB) *Queue {
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
