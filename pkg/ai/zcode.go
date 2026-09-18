//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	// Register the pure-Go SQLite driver used to read ZCode session databases.
	_ "modernc.org/sqlite"
)

// ZCode contains parameters for detecting heartbeats from ZCode logs.
type ZCode ParserConfig

type (
	zCodeSessionInfo struct {
		ID        string
		Directory string
		Version   string
	}

	zCodeMessageInfo struct {
		ID         string `json:"id"`
		SessionID  string `json:"sessionID"`
		Role       string `json:"role"`
		ParentID   string `json:"parentID"`
		ProviderID string `json:"providerID"`
		ModelID    string `json:"modelID"`
		Path       *struct {
			Cwd  string `json:"cwd"`
			Root string `json:"root"`
		} `json:"path"`
		Tokens *struct {
			Input     int64 `json:"input"`
			Output    int64 `json:"output"`
			Reasoning int64 `json:"reasoning"`
			Cache     struct {
				Read  int64 `json:"read"`
				Write int64 `json:"write"`
			} `json:"cache"`
		} `json:"tokens"`
		Time struct {
			Created int64 `json:"created"`
		} `json:"time"`
	}

	zCodePart struct {
		ID        string          `json:"id"`
		MessageID string          `json:"messageID"`
		SessionID string          `json:"sessionID"`
		Type      string          `json:"type"`
		Text      string          `json:"text"`
		Ignored   bool            `json:"ignored"`
		Tool      string          `json:"tool"`
		State     *zCodeToolState `json:"state"`
	}

	zCodeToolState struct {
		Status   string          `json:"status"`
		Input    json.RawMessage `json:"input"`
		Output   string          `json:"output"`
		Metadata json.RawMessage `json:"metadata"`
	}

	zCodeMessageWithParts struct {
		info  zCodeMessageInfo
		parts []zCodePart
	}

	zCodeEditInput struct {
		FilePath  string  `json:"file_path"`
		OldString *string `json:"old_string"`
		NewString *string `json:"new_string"`
	}

	zCodeWriteInput struct {
		FilePath string  `json:"file_path"`
		Content  *string `json:"content"`
	}

	zCodeToolMetadata struct {
		Display *struct {
			Additions *int   `json:"additions"`
			Deletions *int   `json:"deletions"`
			FilePath  string `json:"filePath"`
		} `json:"display"`
		ReadFileState *struct {
			Path string `json:"path"`
		} `json:"readFileState"`
	}
)

// Parse parses ZCode logs for AI heartbeats.
func (g ZCode) Parse(ctx context.Context) (Heartbeats, error) {
	return g.parseSQLite(ctx)
}

func (g ZCode) parseSQLite(ctx context.Context) (Heartbeats, error) {
	dbPath, err := zCodeSQLiteDBPath(ctx)
	if err != nil {
		return nil, err
	}

	if dbPath == "" || !zCodeSQLiteDBModifiedAfter(dbPath, g.After) {
		return nil, nil
	}

	return g.parseSQLiteDB(ctx, dbPath)
}

func (g ZCode) parseSQLiteDB(ctx context.Context, dbPath string) (Heartbeats, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening ZCode sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	sessions, err := querySQLiteSessions(ctx, db, dbPath)
	if err != nil {
		return nil, err
	}

	messagesBySession, messageIDs, err := g.querySQLiteMessages(ctx, db, dbPath)
	if err != nil {
		return nil, err
	}

	if len(messageIDs) == 0 {
		return nil, nil
	}

	if err := g.querySQLiteParts(ctx, db, dbPath, messagesBySession, messageIDs); err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for sessionID, messages := range messagesBySession {
		session, ok := sessions[sessionID]
		if !ok {
			session = zCodeSessionInfo{ID: sessionID}
		}

		heartbeats = append(heartbeats, g.sessionHeartbeats(session, messages)...)
	}

	return heartbeats, nil
}

func querySQLiteSessions(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
) (map[string]zCodeSessionInfo, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, COALESCE(directory, ''), COALESCE(version, '')
FROM session;
`)
	if err != nil {
		return nil, fmt.Errorf("failed querying ZCode sqlite sessions %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	sessions := make(map[string]zCodeSessionInfo)

	for rows.Next() {
		var session zCodeSessionInfo
		if err := rows.Scan(&session.ID, &session.Directory, &session.Version); err != nil {
			return nil, fmt.Errorf("failed scanning ZCode sqlite session row: %s", err)
		}

		sessions[session.ID] = session
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading ZCode sqlite sessions %q: %s", dbPath, err)
	}

	return sessions, nil
}

func (g ZCode) querySQLiteMessages(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
) (map[string][]zCodeMessageWithParts, map[string]struct{}, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, session_id, CAST(data AS TEXT) AS data, time_created
FROM message
WHERE time_created >= ?
  AND json_valid(data)
ORDER BY time_created ASC, id ASC;
`, g.afterUnixMilli())
	if err != nil {
		return nil, nil, fmt.Errorf("failed querying ZCode sqlite messages %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	messagesBySession := make(map[string][]zCodeMessageWithParts)
	messageIDs := make(map[string]struct{})

	for rows.Next() {
		var (
			id        string
			sessionID string
			data      string
			createdAt int64
		)

		if err := rows.Scan(&id, &sessionID, &data, &createdAt); err != nil {
			return nil, nil, fmt.Errorf("failed scanning ZCode sqlite message row: %s", err)
		}

		var message zCodeMessageInfo
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &message); err != nil {
			continue
		}

		message.ID = firstNonEmptyString(message.ID, id)

		message.SessionID = firstNonEmptyString(message.SessionID, sessionID)
		if message.Time.Created == 0 {
			message.Time.Created = createdAt
		}

		if message.Time.Created == 0 ||
			!timestampAtOrAfterCutoff(time.UnixMilli(message.Time.Created), g.After) {
			continue
		}

		messagesBySession[message.SessionID] = append(messagesBySession[message.SessionID], zCodeMessageWithParts{
			info: message,
		})
		messageIDs[message.ID] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed reading ZCode sqlite messages %q: %s", dbPath, err)
	}

	return messagesBySession, messageIDs, nil
}

func (g ZCode) querySQLiteParts(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
	messagesBySession map[string][]zCodeMessageWithParts,
	messageIDs map[string]struct{},
) error {
	rows, err := db.QueryContext(ctx, `
SELECT
	p.id,
	p.message_id,
	p.session_id,
	CAST(p.data AS TEXT) AS data
FROM part AS p
JOIN message AS m ON m.id = p.message_id
WHERE m.time_created >= ?
  AND json_valid(p.data)
ORDER BY p.time_created ASC, p.id ASC;
`, g.afterUnixMilli())
	if err != nil {
		return fmt.Errorf("failed querying ZCode sqlite parts %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	partsByMessage := make(map[string][]zCodePart)

	for rows.Next() {
		var (
			id        string
			messageID string
			sessionID string
			data      string
		)

		if err := rows.Scan(&id, &messageID, &sessionID, &data); err != nil {
			return fmt.Errorf("failed scanning ZCode sqlite part row: %s", err)
		}

		if _, ok := messageIDs[messageID]; !ok {
			continue
		}

		var part zCodePart
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &part); err != nil {
			continue
		}

		part.ID = firstNonEmptyString(part.ID, id)
		part.MessageID = firstNonEmptyString(part.MessageID, messageID)
		part.SessionID = firstNonEmptyString(part.SessionID, sessionID)

		partsByMessage[part.MessageID] = append(partsByMessage[part.MessageID], part)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed reading ZCode sqlite parts %q: %s", dbPath, err)
	}

	for sessionID, messages := range messagesBySession {
		for i := range messages {
			messages[i].parts = partsByMessage[messages[i].info.ID]
		}

		messagesBySession[sessionID] = messages
	}

	return nil
}

func (g ZCode) sessionHeartbeats(
	session zCodeSessionInfo,
	messages []zCodeMessageWithParts,
) Heartbeats {
	if len(messages) == 0 {
		return nil
	}

	slices.SortFunc(messages, func(a, b zCodeMessageWithParts) int {
		return int(a.info.Time.Created - b.info.Time.Created)
	})

	sessionEntity := appHeartbeatEntity(g.Name(), session.ID)
	sessionCwd := session.Directory

	var (
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
	)

	for _, message := range messages {
		messageTime := time.UnixMilli(message.info.Time.Created)

		cwd := sessionCwd
		if message.info.Path != nil {
			cwd = firstNonEmptyString(message.info.Path.Cwd, message.info.Path.Root, sessionCwd)
		}

		if message.info.Tokens != nil {
			// ZCode records per-model-request usage: each assistant message
			// carries the input, output, reasoning and cache numbers of the
			// single request that produced it (verified 1:1 against the
			// model_usage table). Context growth makes input and cache reads
			// look near-monotonic across a session, but they are not session
			// counters, so values must never be differenced against the
			// previous message.
			//
			// tokens.input includes cache reads; WakaTime has a cached-read
			// bucket but no cache-write bucket, so regular input only
			// subtracts cache reads and cache creation stays in regular
			// input. Current/Last exist only to fit the delta interface of
			// heartbeat.NewWithAITokens.
			tokens.CurrentInput = tokens.LastInput + max(
				message.info.Tokens.Input-message.info.Tokens.Cache.Read,
				0,
			)
			tokens.CurrentCachedInput = tokens.LastCachedInput + max(message.info.Tokens.Cache.Read, 0)
			tokens.CurrentOutput = tokens.LastOutput + max(message.info.Tokens.Output, 0) +
				max(message.info.Tokens.Reasoning, 0)
		}

		if message.info.Time.Created == 0 ||
			(!g.After.IsZero() && !timestampAtOrAfterCutoff(messageTime, g.After)) {
			tokens.LastInput = tokens.CurrentInput
			tokens.LastCachedInput = tokens.CurrentCachedInput
			tokens.LastOutput = tokens.CurrentOutput

			continue
		}

		switch message.info.Role {
		case "user":
			if hb := g.userHeartbeat(
				sessionEntity,
				session.Version,
				message.info.ModelID,
				session.ID,
				cwd,
				message,
				tokens,
			); hb != nil {
				heartbeats = append(heartbeats, *hb)
			}
		case "assistant":
			if hb := g.assistantHeartbeat(
				sessionEntity,
				session.Version,
				message.info.ModelID,
				session.ID,
				cwd,
				message,
				tokens,
			); hb != nil {
				heartbeats = append(heartbeats, *hb)
			}

			for _, part := range message.parts {
				hbs := g.toolHeartbeats(
					session.Version,
					message.info.ModelID,
					session.ID,
					cwd,
					messageTime,
					part,
				)
				if len(hbs) > 0 {
					heartbeats = append(heartbeats, hbs...)
				}
			}
		}

		tokens.LastInput = tokens.CurrentInput
		tokens.LastCachedInput = tokens.CurrentCachedInput
		tokens.LastOutput = tokens.CurrentOutput
	}

	return heartbeats
}

func zCodeSQLiteDBPath(ctx context.Context) (string, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return "", err
	}

	dbPath := filepath.Join(home, ".zcode", "cli", "db", "db.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", fmt.Errorf("failed to stat ZCode sqlite db %q: %s", dbPath, err)
	}

	return dbPath, nil
}

func zCodeSQLiteDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	for _, path := range []string{dbPath, dbPath + "-wal"} {
		info, err := os.Stat(path)
		if err == nil && timestampAtOrAfterCutoff(info.ModTime(), after) {
			return true
		}
	}

	return false
}

func (g ZCode) afterUnixMilli() int64 {
	if g.After.IsZero() {
		return 0
	}

	return g.After.UnixMilli()
}

func (g ZCode) userHeartbeat(
	entity string,
	version string,
	model string,
	sessionID string,
	cwd string,
	message zCodeMessageWithParts,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	prompt := 0

	for _, part := range message.parts {
		if part.Type != "text" || part.Ignored || strings.TrimSpace(part.Text) == "" {
			continue
		}

		prompt += promptLength(part.Text)
	}

	if prompt == 0 {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		nil,
		sessionID,
		tokens,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		heartbeat.AppType,
		nil,
		false,
		heartbeat.PointerTo(false),
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		cwd,
		heartbeatTimestamp(time.UnixMilli(message.info.Time.Created)),
		aiUserAgentWithModelAndEditor(
			entity,
			g.UserAgents,
			g.FallbackUserAgent,
			model,
			"",
			"zcode-cli/"+unknownIfEmpty(version),
		),
	)
	h.AIPromptLength = prompt

	return &h
}

func (g ZCode) assistantHeartbeat(
	entity string,
	version string,
	model string,
	sessionID string,
	cwd string,
	message zCodeMessageWithParts,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	hasSignal := false

	for _, part := range message.parts {
		switch part.Type {
		case "text", "reasoning", "tool", "step-finish":
			hasSignal = true
		}

		if hasSignal {
			break
		}
	}

	if !hasSignal && message.info.Tokens == nil {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		nil,
		sessionID,
		tokens,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		heartbeat.AppType,
		nil,
		false,
		heartbeat.PointerTo(false),
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		cwd,
		heartbeatTimestamp(time.UnixMilli(message.info.Time.Created)),
		aiUserAgentWithModelAndEditor(
			entity,
			g.UserAgents,
			g.FallbackUserAgent,
			model,
			"",
			"zcode-cli/"+unknownIfEmpty(version),
		),
	)

	return &h
}

func (g ZCode) toolHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	part zCodePart,
) Heartbeats {
	if part.Type != "tool" || part.State == nil || part.State.Status != "completed" {
		return nil
	}

	switch strings.ToLower(part.Tool) {
	case "edit":
		return g.editHeartbeats(version, model, sessionID, cwd, timestamp, *part.State)
	case "write":
		return g.writeHeartbeats(version, model, sessionID, cwd, timestamp, *part.State)
	default:
		return nil
	}
}

func (g ZCode) editHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	state zCodeToolState,
) Heartbeats {
	var input zCodeEditInput
	if err := json.Unmarshal(state.Input, &input); err != nil {
		return nil
	}

	var metadata zCodeToolMetadata
	_ = json.Unmarshal(state.Metadata, &metadata)

	filePath := zCodeResolvePath(cwd, zCodeToolFilePath(metadata, input.FilePath))
	if filePath == "" {
		return nil
	}

	if lineChanges, ok := zCodeMetadataLineChanges(metadata); ok {
		return Heartbeats{g.fileHeartbeat(version, model, sessionID, filePath, lineChanges, timestamp)}
	}

	if input.OldString == nil && input.NewString == nil {
		return nil
	}

	var oldString, newString string
	if input.OldString != nil {
		oldString = *input.OldString
	}
	if input.NewString != nil {
		newString = *input.NewString
	}

	lineChanges := countStringLines(newString) - countStringLines(oldString)

	return Heartbeats{g.fileHeartbeat(version, model, sessionID, filePath, lineChanges, timestamp)}
}

func (g ZCode) writeHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	state zCodeToolState,
) Heartbeats {
	var input zCodeWriteInput
	if err := json.Unmarshal(state.Input, &input); err != nil {
		return nil
	}

	var metadata zCodeToolMetadata
	_ = json.Unmarshal(state.Metadata, &metadata)

	filePath := zCodeResolvePath(cwd, zCodeToolFilePath(metadata, input.FilePath))
	if filePath == "" {
		return nil
	}

	if lineChanges, ok := zCodeMetadataLineChanges(metadata); ok {
		return Heartbeats{g.fileHeartbeat(version, model, sessionID, filePath, lineChanges, timestamp)}
	}

	if input.Content == nil || !zCodeWriteCreated(state.Output) {
		return nil
	}

	return Heartbeats{
		g.fileHeartbeat(version, model, sessionID, filePath, countStringLines(*input.Content), timestamp),
	}
}

func zCodeToolFilePath(metadata zCodeToolMetadata, inputPath string) string {
	var displayPath, readPath string
	if metadata.Display != nil {
		displayPath = metadata.Display.FilePath
	}
	if metadata.ReadFileState != nil {
		readPath = metadata.ReadFileState.Path
	}

	return firstNonEmptyString(displayPath, inputPath, readPath)
}

func zCodeMetadataLineChanges(metadata zCodeToolMetadata) (int, bool) {
	if metadata.Display == nil ||
		(metadata.Display.Additions == nil && metadata.Display.Deletions == nil) {
		return 0, false
	}

	var additions, deletions int
	if metadata.Display.Additions != nil {
		additions = *metadata.Display.Additions
	}
	if metadata.Display.Deletions != nil {
		deletions = *metadata.Display.Deletions
	}

	return additions - deletions, true
}

func zCodeWriteCreated(output string) bool {
	return strings.HasPrefix(strings.TrimSpace(output), "File created successfully at:")
}

func (g ZCode) fileHeartbeat(
	version string,
	model string,
	sessionID string,
	filePath string,
	lineChanges int,
	timestamp time.Time,
) heartbeat.Heartbeat {
	return heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		sessionID,
		heartbeat.AITokens{},
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		filePath,
		heartbeat.FileType,
		nil,
		false,
		heartbeat.PointerTo(true),
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		"",
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModelAndEditor(
			filePath,
			g.UserAgents,
			g.FallbackUserAgent,
			model,
			"",
			"zcode-cli/"+unknownIfEmpty(version),
		),
	)
}

func zCodeResolvePath(base string, filePath string) string {
	switch {
	case filePath == "":
		return ""
	case filepath.IsAbs(filePath), strings.HasPrefix(filePath, "/"):
		return filePath
	case base == "":
		return filePath
	case strings.HasPrefix(base, "/") && !strings.Contains(base, `\`):
		return pathpkg.Join(base, filePath)
	default:
		return filepath.Join(base, filePath)
	}
}

// Name returns the ZCode parser name.
func (ZCode) Name() string { return "ZCode" }
