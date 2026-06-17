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
	"github.com/wakatime/wakatime-cli/pkg/ini"

	// Register the pure-Go SQLite driver used to read OpenCode session databases.
	_ "modernc.org/sqlite"
)

// OpenCode contains params for detecting heartbeats from OpenCode session logs.
type OpenCode ParserConfig

const (
	openCodeRecentMessageRowLimit = 5000
	openCodeRecentPartRowLimit    = 5000
)

type (
	openCodeSessionInfo struct {
		ID        string `json:"id"`
		Directory string `json:"directory"`
		Version   string `json:"version"`
		ParentID  string `json:"parentID"`
		Time      struct {
			Created int64 `json:"created"`
			Updated int64 `json:"updated"`
		} `json:"time"`
	}

	openCodeMessageInfo struct {
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
			Input  int64 `json:"input"`
			Output int64 `json:"output"`
		} `json:"tokens"`
		Time struct {
			Created int64 `json:"created"`
		} `json:"time"`
	}

	openCodePart struct {
		ID        string             `json:"id"`
		MessageID string             `json:"messageID"`
		SessionID string             `json:"sessionID"`
		Type      string             `json:"type"`
		Text      string             `json:"text"`
		Ignored   bool               `json:"ignored"`
		Tool      string             `json:"tool"`
		State     *openCodeToolState `json:"state"`
	}

	openCodeToolState struct {
		Status   string          `json:"status"`
		Input    json.RawMessage `json:"input"`
		Metadata json.RawMessage `json:"metadata"`
	}

	openCodeMessageWithParts struct {
		info  openCodeMessageInfo
		parts []openCodePart
	}

	openCodeEditInput struct {
		FilePath  string `json:"filePath"`
		OldString string `json:"oldString"`
		NewString string `json:"newString"`
	}

	openCodeWriteInput struct {
		FilePath string `json:"filePath"`
		Content  string `json:"content"`
	}

	openCodeWriteMetadata struct {
		Exists bool `json:"exists"`
	}

	openCodeApplyPatchInput struct {
		PatchText string `json:"patchText"`
	}

	openCodeFileDiff struct {
		File      string `json:"file"`
		FilePath  string `json:"filePath"`
		Patch     string `json:"patch"`
		Additions int    `json:"additions"`
		Deletions int    `json:"deletions"`
	}

	openCodeEditMetadata struct {
		FileDiff *openCodeFileDiff `json:"filediff"`
	}

	openCodeApplyPatchMetadata struct {
		Files []struct {
			FilePath  string `json:"filePath"`
			Patch     string `json:"patch"`
			Additions int    `json:"additions"`
			Deletions int    `json:"deletions"`
		} `json:"files"`
	}
)

// Parse parses OpenCode legacy storage first, then falls back to the SQLite db.
func (g OpenCode) Parse(ctx context.Context) (Heartbeats, error) {
	legacy, err := g.parseLegacyStorage(ctx)
	if err != nil {
		return nil, err
	}

	if len(legacy) > 0 {
		return legacy, nil
	}

	return g.parseSQLite(ctx)
}

func (g OpenCode) parseLegacyStorage(ctx context.Context) (Heartbeats, error) {
	dataRoots, err := openCodeDataRoots(ctx)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, dataRoot := range dataRoots {
		parsed, err := g.parseLegacyRoot(dataRoot)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g OpenCode) parseLegacyRoot(dataRoot string) (Heartbeats, error) {
	sessionsDir := filepath.Join(dataRoot, "storage", "session")
	if _, err := os.Stat(sessionsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat OpenCode sessions directory %q: %s", sessionsDir, err)
	}

	var sessionPaths []string

	err := filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			return nil
		}

		session, err := g.readSessionInfo(path)
		if err != nil {
			return err
		}

		if g.After.IsZero() || session.Time.Updated == 0 {
			sessionPaths = append(sessionPaths, path)
			return nil
		}

		if !time.UnixMilli(session.Time.Updated).Before(g.After) {
			sessionPaths = append(sessionPaths, path)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk OpenCode sessions directory %q: %s", sessionsDir, err)
	}

	var heartbeats Heartbeats

	for _, sessionPath := range sessionPaths {
		parsed, err := g.parseLegacySession(sessionPath)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g OpenCode) parseLegacySession(sessionPath string) (Heartbeats, error) {
	session, err := g.readSessionInfo(sessionPath)
	if err != nil {
		return nil, err
	}

	baseStorageDir := filepath.Dir(filepath.Dir(filepath.Dir(sessionPath)))
	messagesDir := filepath.Join(baseStorageDir, "message", session.ID)

	messageEntries, err := os.ReadDir(messagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read OpenCode messages directory %q: %s", messagesDir, err)
	}

	var messages []openCodeMessageWithParts

	for _, entry := range messageEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		messagePath := filepath.Join(messagesDir, entry.Name())

		message, err := g.readMessageInfo(messagePath)
		if err != nil {
			return nil, err
		}

		parts, err := g.readParts(baseStorageDir, message.ID)
		if err != nil {
			return nil, err
		}

		messages = append(messages, openCodeMessageWithParts{
			info:  message,
			parts: parts,
		})
	}

	return g.sessionHeartbeats(session, messages), nil
}

func (g OpenCode) parseSQLite(ctx context.Context) (Heartbeats, error) {
	dbPaths, err := openCodeSQLiteDBPaths(ctx)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, dbPath := range dbPaths {
		if !openCodeSQLiteDBModifiedAfter(dbPath, g.After) {
			continue
		}

		parsed, err := g.parseSQLiteDB(ctx, dbPath)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g OpenCode) parseSQLiteDB(ctx context.Context, dbPath string) (Heartbeats, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed opening OpenCode sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	sessions, err := queryOpenCodeSQLiteSessions(ctx, db, dbPath)
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

	if err := queryOpenCodeSQLiteParts(ctx, db, dbPath, messagesBySession, messageIDs); err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for sessionID, messages := range messagesBySession {
		session, ok := sessions[sessionID]
		if !ok {
			session = openCodeSessionInfo{ID: sessionID}
		}

		heartbeats = append(heartbeats, g.sessionHeartbeats(session, messages)...)
	}

	return heartbeats, nil
}

func queryOpenCodeSQLiteSessions(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
) (map[string]openCodeSessionInfo, error) {
	rows, err := db.QueryContext(ctx, `
SELECT id, COALESCE(directory, ''), COALESCE(version, '')
FROM session;
`)
	if err != nil {
		return nil, fmt.Errorf("failed querying OpenCode sqlite sessions %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	sessions := make(map[string]openCodeSessionInfo)

	for rows.Next() {
		var session openCodeSessionInfo
		if err := rows.Scan(&session.ID, &session.Directory, &session.Version); err != nil {
			return nil, fmt.Errorf("failed scanning OpenCode sqlite session row: %s", err)
		}

		sessions[session.ID] = session
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed reading OpenCode sqlite sessions %q: %s", dbPath, err)
	}

	return sessions, nil
}

func (g OpenCode) querySQLiteMessages(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
) (map[string][]openCodeMessageWithParts, map[string]string, error) {
	rows, err := db.QueryContext(ctx, `
WITH recentOpenCodeMessages AS (
	SELECT id, session_id, CAST(data AS TEXT) AS data, time_created
	FROM message
	ORDER BY rowid DESC
	LIMIT ?
)
SELECT id, session_id, data, time_created
FROM recentOpenCodeMessages
WHERE time_created >= ?
  AND json_valid(data)
ORDER BY time_created ASC, id ASC;
`, openCodeRecentMessageRowLimit, g.afterUnixMilli())
	if err != nil {
		return nil, nil, fmt.Errorf("failed querying OpenCode sqlite messages %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	messagesBySession := make(map[string][]openCodeMessageWithParts)
	messageIDs := make(map[string]string)

	for rows.Next() {
		var (
			id        string
			sessionID string
			data      string
			createdAt int64
		)

		if err := rows.Scan(&id, &sessionID, &data, &createdAt); err != nil {
			return nil, nil, fmt.Errorf("failed scanning OpenCode sqlite message row: %s", err)
		}

		var message openCodeMessageInfo
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &message); err != nil {
			continue
		}

		message.ID = firstNonEmptyString(message.ID, id)

		message.SessionID = firstNonEmptyString(message.SessionID, sessionID)
		if message.Time.Created == 0 {
			message.Time.Created = createdAt
		}

		if message.Time.Created == 0 || time.UnixMilli(message.Time.Created).Before(g.After) {
			continue
		}

		messagesBySession[message.SessionID] = append(messagesBySession[message.SessionID], openCodeMessageWithParts{
			info: message,
		})
		messageIDs[message.ID] = message.SessionID
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed reading OpenCode sqlite messages %q: %s", dbPath, err)
	}

	if !g.After.IsZero() {
		if err := g.querySQLiteSeedMessages(ctx, db, dbPath, messagesBySession); err != nil {
			return nil, nil, err
		}
	}

	return messagesBySession, messageIDs, nil
}

func (g OpenCode) querySQLiteSeedMessages(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
	messagesBySession map[string][]openCodeMessageWithParts,
) error {
	if len(messagesBySession) == 0 {
		return nil
	}

	rows, err := db.QueryContext(ctx, `
WITH recentOpenCodeSeedMessages AS (
	SELECT id, session_id, CAST(data AS TEXT) AS data, time_created
	FROM message
	ORDER BY rowid DESC
	LIMIT ?
)
SELECT id, session_id, data, time_created
FROM recentOpenCodeSeedMessages
WHERE time_created < ?
  AND json_valid(data)
ORDER BY time_created DESC, id DESC;
`, openCodeRecentMessageRowLimit, g.afterUnixMilli())
	if err != nil {
		return fmt.Errorf("failed querying OpenCode sqlite seed messages %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	remaining := make(map[string]struct{}, len(messagesBySession))
	for sessionID := range messagesBySession {
		remaining[sessionID] = struct{}{}
	}

	for rows.Next() {
		if len(remaining) == 0 {
			break
		}

		var (
			id        string
			sessionID string
			data      string
			createdAt int64
		)

		if err := rows.Scan(&id, &sessionID, &data, &createdAt); err != nil {
			return fmt.Errorf("failed scanning OpenCode sqlite seed message row: %s", err)
		}

		if _, ok := remaining[sessionID]; !ok {
			continue
		}

		var message openCodeMessageInfo
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &message); err != nil {
			continue
		}

		message.ID = firstNonEmptyString(message.ID, id)

		message.SessionID = firstNonEmptyString(message.SessionID, sessionID)
		if message.Time.Created == 0 {
			message.Time.Created = createdAt
		}

		if message.SessionID == "" || message.Time.Created == 0 {
			continue
		}

		messagesBySession[message.SessionID] = append(messagesBySession[message.SessionID], openCodeMessageWithParts{
			info: message,
		})
		delete(remaining, message.SessionID)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed reading OpenCode sqlite seed messages %q: %s", dbPath, err)
	}

	return nil
}

func queryOpenCodeSQLiteParts(
	ctx context.Context,
	db *sql.DB,
	dbPath string,
	messagesBySession map[string][]openCodeMessageWithParts,
	messageIDs map[string]string,
) error {
	rows, err := db.QueryContext(ctx, `
WITH recentOpenCodeParts AS (
	SELECT id, message_id, session_id, CAST(data AS TEXT) AS data, time_created
	FROM part
	ORDER BY rowid DESC
	LIMIT ?
)
SELECT id, message_id, session_id, data
FROM recentOpenCodeParts
WHERE json_valid(data)
ORDER BY time_created ASC, id ASC;
`, openCodeRecentPartRowLimit)
	if err != nil {
		return fmt.Errorf("failed querying OpenCode sqlite parts %q: %s", dbPath, err)
	}
	defer rows.Close() // nolint:errcheck

	partsByMessage := make(map[string][]openCodePart)

	for rows.Next() {
		var (
			id        string
			messageID string
			sessionID string
			data      string
		)

		if err := rows.Scan(&id, &messageID, &sessionID, &data); err != nil {
			return fmt.Errorf("failed scanning OpenCode sqlite part row: %s", err)
		}

		if _, ok := messageIDs[messageID]; !ok {
			continue
		}

		var part openCodePart
		if err := json.Unmarshal([]byte(strings.TrimSpace(data)), &part); err != nil {
			continue
		}

		part.ID = firstNonEmptyString(part.ID, id)
		part.MessageID = firstNonEmptyString(part.MessageID, messageID)
		part.SessionID = firstNonEmptyString(part.SessionID, sessionID)

		partsByMessage[part.MessageID] = append(partsByMessage[part.MessageID], part)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed reading OpenCode sqlite parts %q: %s", dbPath, err)
	}

	for sessionID, messages := range messagesBySession {
		for i := range messages {
			messages[i].parts = partsByMessage[messages[i].info.ID]
		}

		messagesBySession[sessionID] = messages
	}

	return nil
}

func (g OpenCode) sessionHeartbeats(
	session openCodeSessionInfo,
	messages []openCodeMessageWithParts,
) Heartbeats {
	if len(messages) == 0 {
		return nil
	}

	slices.SortFunc(messages, func(a, b openCodeMessageWithParts) int {
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
			tokens.CurrentInput = message.info.Tokens.Input
			tokens.CurrentOutput = message.info.Tokens.Output
		}

		if message.info.Time.Created == 0 || (!g.After.IsZero() && messageTime.Before(g.After)) {
			tokens.LastInput = tokens.CurrentInput
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
		tokens.LastOutput = tokens.CurrentOutput
	}

	return heartbeats
}

func openCodeDataRoots(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	return []string{
		filepath.Join(home, ".local", "share", "opencode"),
		filepath.Join(home, "Library", "Application Support", "opencode"),
		filepath.Join(home, "AppData", "Local", "opencode"),
		filepath.Join(home, "AppData", "Roaming", "opencode"),
	}, nil
}

func openCodeSQLiteDBPaths(ctx context.Context) ([]string, error) {
	dataRoots, err := openCodeDataRoots(ctx)
	if err != nil {
		return nil, err
	}

	var dbPaths []string

	for _, dataRoot := range dataRoots {
		entries, err := filepath.Glob(filepath.Join(dataRoot, "opencode*.db"))
		if err != nil {
			return nil, fmt.Errorf("failed globbing OpenCode sqlite dbs in %q: %s", dataRoot, err)
		}

		for _, candidate := range entries {
			if _, err := os.Stat(candidate); err == nil {
				dbPaths = append(dbPaths, candidate)
			}
		}
	}

	return dbPaths, nil
}

func openCodeSQLiteDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return info.ModTime().After(after)
}

func (g OpenCode) afterUnixMilli() int64 {
	if g.After.IsZero() {
		return 0
	}

	return g.After.UnixMilli()
}

func (OpenCode) readSessionInfo(sessionPath string) (openCodeSessionInfo, error) {
	// nolint:gosec // Reads a session file discovered inside the trusted OpenCode storage directory.
	contents, err := os.ReadFile(filepath.Clean(sessionPath))
	if err != nil {
		return openCodeSessionInfo{}, fmt.Errorf("failed to read OpenCode session %q: %s", sessionPath, err)
	}

	var session openCodeSessionInfo
	if err := json.Unmarshal(contents, &session); err != nil {
		return openCodeSessionInfo{}, fmt.Errorf("failed to parse OpenCode session %q: %s", sessionPath, err)
	}

	return session, nil
}

func (OpenCode) readMessageInfo(messagePath string) (openCodeMessageInfo, error) {
	// nolint:gosec // Reads a message file discovered inside the trusted OpenCode storage directory.
	contents, err := os.ReadFile(filepath.Clean(messagePath))
	if err != nil {
		return openCodeMessageInfo{}, fmt.Errorf("failed to read OpenCode message %q: %s", messagePath, err)
	}

	var message openCodeMessageInfo
	if err := json.Unmarshal(contents, &message); err != nil {
		return openCodeMessageInfo{}, fmt.Errorf("failed to parse OpenCode message %q: %s", messagePath, err)
	}

	return message, nil
}

func (OpenCode) readParts(baseStorageDir string, messageID string) ([]openCodePart, error) {
	partsDir := filepath.Join(baseStorageDir, "part", messageID)

	entries, err := os.ReadDir(partsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read OpenCode parts directory %q: %s", partsDir, err)
	}

	var parts []openCodePart

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		partPath := filepath.Join(partsDir, entry.Name())
		// nolint:gosec // Reads a part file discovered inside the trusted OpenCode storage directory.
		contents, err := os.ReadFile(filepath.Clean(partPath))
		if err != nil {
			return nil, fmt.Errorf("failed to read OpenCode part %q: %s", partPath, err)
		}

		var part openCodePart
		if err := json.Unmarshal(contents, &part); err != nil {
			return nil, fmt.Errorf("failed to parse OpenCode part %q: %s", partPath, err)
		}

		parts = append(parts, part)
	}

	slices.SortFunc(parts, func(a, b openCodePart) int {
		return strings.Compare(a.ID, b.ID)
	})

	return parts, nil
}

func (g OpenCode) userHeartbeat(
	entity string,
	version string,
	model string,
	sessionID string,
	cwd string,
	message openCodeMessageWithParts,
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
		float64(time.UnixMilli(message.info.Time.Created).Unix()),
		aiUserAgentWithAgentPrefix(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, version), model),
	)
	h.AIPromptLength = prompt

	return &h
}

func (g OpenCode) assistantHeartbeat(
	entity string,
	version string,
	model string,
	sessionID string,
	cwd string,
	message openCodeMessageWithParts,
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
		float64(time.UnixMilli(message.info.Time.Created).Unix()),
		aiUserAgentWithAgentPrefix(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, version), model),
	)

	return &h
}

func (g OpenCode) toolHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	part openCodePart,
) Heartbeats {
	if part.Type != "tool" || part.State == nil || part.State.Status != "completed" {
		return nil
	}

	switch part.Tool {
	case "edit":
		return g.editHeartbeats(version, model, sessionID, cwd, timestamp, *part.State)
	case "write":
		return g.writeHeartbeats(version, model, sessionID, cwd, timestamp, *part.State)
	case "apply_patch":
		return g.applyPatchHeartbeats(version, model, sessionID, cwd, timestamp, *part.State)
	default:
		return nil
	}
}

func (g OpenCode) editHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	state openCodeToolState,
) Heartbeats {
	var input openCodeEditInput
	if err := json.Unmarshal(state.Input, &input); err != nil {
		return nil
	}

	filePath := openCodeResolvePath(cwd, input.FilePath)
	lineChanges := countStringLines(input.NewString) - countStringLines(input.OldString)

	var metadata openCodeEditMetadata
	if err := json.Unmarshal(state.Metadata, &metadata); err == nil && metadata.FileDiff != nil {
		diff := metadata.FileDiff
		if diff.Additions != 0 || diff.Deletions != 0 {
			lineChanges = diff.Additions - diff.Deletions
		}

		filePath = firstNonEmptyString(
			openCodeResolvePath(cwd, diff.FilePath),
			openCodeResolvePath(cwd, diff.File),
			filePath,
		)
	}

	return Heartbeats{g.fileHeartbeat(version, model, sessionID, filePath, lineChanges, timestamp)}
}

func (g OpenCode) writeHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	state openCodeToolState,
) Heartbeats {
	var input openCodeWriteInput
	if err := json.Unmarshal(state.Input, &input); err != nil {
		return nil
	}

	filePath := openCodeResolvePath(cwd, input.FilePath)

	lineChanges := 0

	var metadata openCodeWriteMetadata
	if err := json.Unmarshal(state.Metadata, &metadata); err != nil || !metadata.Exists {
		lineChanges = countStringLines(input.Content)
	}

	return Heartbeats{g.fileHeartbeat(version, model, sessionID, filePath, lineChanges, timestamp)}
}

func (g OpenCode) applyPatchHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	timestamp time.Time,
	state openCodeToolState,
) Heartbeats {
	var metadata openCodeApplyPatchMetadata
	if err := json.Unmarshal(state.Metadata, &metadata); err == nil && len(metadata.Files) > 0 {
		var heartbeats Heartbeats

		for _, file := range metadata.Files {
			filePath := openCodeResolvePath(cwd, file.FilePath)
			heartbeats = append(heartbeats, g.fileHeartbeat(
				version,
				model,
				sessionID,
				filePath,
				file.Additions-file.Deletions,
				timestamp,
			))
		}

		return heartbeats
	}

	var input openCodeApplyPatchInput
	if err := json.Unmarshal(state.Input, &input); err != nil {
		return nil
	}

	return g.patchHeartbeats(version, model, sessionID, cwd, input.PatchText, timestamp)
}

func (g OpenCode) patchHeartbeats(
	version string,
	model string,
	sessionID string,
	cwd string,
	patchText string,
	timestamp time.Time,
) Heartbeats {
	var heartbeats Heartbeats

	var (
		currentFile string
		additions   int
		deletions   int
	)

	lines := strings.Split(patchText, "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "*** Update File: "),
			strings.HasPrefix(line, "*** Add File: "),
			strings.HasPrefix(line, "*** Delete File: "):
			if currentFile != "" {
				heartbeats = append(heartbeats, g.fileHeartbeat(
					version,
					model,
					sessionID,
					currentFile,
					additions-deletions,
					timestamp,
				))
			}

			currentFile = openCodePatchFilePath(cwd, line)
			additions = 0
			deletions = 0
		case currentFile != "" && strings.HasPrefix(line, "+"):
			additions++
		case currentFile != "" && strings.HasPrefix(line, "-"):
			deletions++
		}
	}

	if currentFile != "" {
		heartbeats = append(heartbeats, g.fileHeartbeat(
			version,
			model,
			sessionID,
			currentFile,
			additions-deletions,
			timestamp,
		))
	}

	return heartbeats
}

func (g OpenCode) fileHeartbeat(
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
		float64(timestamp.Unix()),
		aiUserAgentWithAgentPrefix(filePath, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, version), model),
	)
}

func openCodePatchFilePath(cwd string, line string) string {
	prefixes := []string{
		"*** Update File: ",
		"*** Add File: ",
		"*** Delete File: ",
	}
	for _, prefix := range prefixes {
		if file, ok := strings.CutPrefix(line, prefix); ok {
			return openCodeResolvePath(cwd, file)
		}
	}

	return ""
}

func openCodeResolvePath(base string, filePath string) string {
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

// Name returns its id.
func (OpenCode) Name() string {
	return "OpenCode"
}
