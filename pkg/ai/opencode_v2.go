//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Read sessions with v2 messages, including archived and child sessions. Check
// ownership before applying the time cutoff so frozen legacy copies lose.
func (g OpenCode) parseSQLiteV2(ctx context.Context, db *sql.DB) (Heartbeats, map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, COALESCE(directory, ''), COALESCE(version, '') FROM session
 WHERE EXISTS (SELECT 1 FROM session_message WHERE session_id = session.id)`)
	if err != nil {
		return nil, nil, err
	}

	sessions := make(map[string]openCodeSessionInfo)
	migrated := make(map[string]bool)

	for rows.Next() {
		var session openCodeSessionInfo
		if err := rows.Scan(&session.ID, &session.Directory, &session.Version); err != nil {
			log.Extract(ctx).Debugf("skipping OpenCode v2 session row: %s", err)
			continue
		}

		sessions[session.ID] = session
		migrated[session.ID] = true
	}

	err = rows.Err()
	_ = rows.Close()

	if err != nil {
		return nil, nil, err
	}

	rows, err = db.QueryContext(ctx, `SELECT id, session_id, type, time_created, CAST(data AS TEXT)
 FROM session_message WHERE time_created >= ? ORDER BY time_created, session_id, seq`, g.afterUnixMilli())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close() // nolint:errcheck

	messages := make(map[string][]openCodeMessageWithParts)

	for rows.Next() {
		var (
			id, sessionID, kind, raw string
			created                  int64
		)

		if err := rows.Scan(&id, &sessionID, &kind, &created, &raw); err != nil {
			log.Extract(ctx).Debugf("skipping OpenCode v2 message row: %s", err)
			continue
		}

		if err := aiSQLiteRead(ctx, id, sessionID, raw); err != nil {
			return nil, nil, err
		}

		message, ok := openCodeV2Message(id, sessionID, kind, created, raw)
		if ok {
			messages[sessionID] = append(messages[sessionID], message)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	var heartbeats Heartbeats

	for id, messages := range messages {
		session := sessions[id]
		session.ID = id
		heartbeats = append(heartbeats, g.sessionHeartbeats(session, messages)...)
	}

	return heartbeats, migrated, nil
}

func openCodeV2Message(id, sessionID, kind string, created int64, raw string) (openCodeMessageWithParts, bool) {
	var payload struct {
		openCodeMessageInfo
		Text  string `json:"text"`
		Model struct {
			ID         string `json:"id"`
			ProviderID string `json:"providerID"`
		} `json:"model"`
		Content []json.RawMessage `json:"content"`
	}
	if json.Unmarshal([]byte(raw), &payload) != nil {
		return openCodeMessageWithParts{}, false
	}

	message := openCodeMessageWithParts{info: payload.openCodeMessageInfo}
	message.info.ID, message.info.SessionID, message.info.Role = id, sessionID, kind
	message.info.Time.Created = created
	message.info.ModelID = payload.Model.ID
	message.info.ProviderID = payload.Model.ProviderID

	switch kind {
	case "user":
		message.parts = []openCodePart{{Type: "text", Text: payload.Text}}
	case "assistant":
		for _, rawContent := range payload.Content {
			var content struct {
				openCodePart
				Name string `json:"name"`
			}
			if json.Unmarshal(rawContent, &content) != nil {
				continue
			}

			part := content.openCodePart
			part.Tool = firstNonEmptyString(content.Name, part.Tool)
			message.parts = append(message.parts, part)
		}
	default:
		return message, false
	}

	return message, true
}

// V2 built-in tools store their output in state.structured, with file/status
// fields instead of the legacy metadata filePath/type fields.
func (g OpenCode) v2ToolFileHeartbeats(
	version, model, sessionID, cwd string,
	timestamp time.Time,
	state openCodeToolState,
) Heartbeats {
	var output struct {
		Files []struct {
			File      string `json:"file"`
			Status    string `json:"status"`
			Additions int    `json:"additions"`
			Deletions int    `json:"deletions"`
		} `json:"files"`
	}
	if json.Unmarshal(state.Structured, &output) != nil {
		return nil
	}

	var hh Heartbeats

	for _, file := range output.Files {
		if file.File == "" {
			continue
		}

		h := g.fileHeartbeat(version, model, sessionID, openCodeResolvePath(cwd, file.File),
			file.Additions-file.Deletions, timestamp)
		h.IsUnsavedEntity = file.Status == "deleted"
		hh = append(hh, h)
	}

	return hh
}
