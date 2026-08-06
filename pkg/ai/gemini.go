package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

// Gemini contains params for detecting heartbeats from Gemini and Antigravity
// session logs.
type Gemini ParserConfig

type (
	geminiSession struct {
		SessionID   string          `json:"sessionId"`
		StartTime   string          `json:"startTime"`
		LastUpdated string          `json:"lastUpdated"`
		Messages    []geminiMessage `json:"messages"`
	}

	geminiSessionUpdate struct {
		SessionID   *string          `json:"sessionId"`
		StartTime   *string          `json:"startTime"`
		LastUpdated *string          `json:"lastUpdated"`
		Messages    *[]geminiMessage `json:"messages"`
	}

	geminiJSONLControl struct {
		Set      *geminiSessionUpdate `json:"$set"`
		RewindTo *string              `json:"$rewindTo"`
	}

	geminiMessage struct {
		ID        string              `json:"id"`
		Type      string              `json:"type"`
		Timestamp string              `json:"timestamp"`
		Content   json.RawMessage     `json:"content"`
		Tokens    *geminiMessageToken `json:"tokens"`
		Model     string              `json:"model"`
		ToolCalls []geminiToolCall    `json:"toolCalls"`
	}

	geminiMessageToken struct {
		Input    int64 `json:"input"`
		Output   int64 `json:"output"`
		Cached   int64 `json:"cached"`
		Thoughts int64 `json:"thoughts"`
		Tool     int64 `json:"tool"`
		Total    int64 `json:"total"`
	}

	geminiToolCall struct {
		ID            string          `json:"id"`
		Name          string          `json:"name"`
		DisplayName   string          `json:"displayName"`
		Args          json.RawMessage `json:"args"`
		Result        json.RawMessage `json:"result"`
		ResultDisplay json.RawMessage `json:"resultDisplay"`
		Status        string          `json:"status"`
		Timestamp     string          `json:"timestamp"`
	}

	geminiProjectsRegistry struct {
		Projects map[string]string `json:"projects"`
	}

	geminiTextPart struct {
		Text string `json:"text"`
	}

	geminiWriteArgs struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	geminiReplaceArgs struct {
		FilePath  string `json:"file_path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
	}

	geminiResultDisplay struct {
		FilePath string `json:"filePath"`
		DiffStat *struct {
			ModelAddedLines   int `json:"model_added_lines"`
			ModelRemovedLines int `json:"model_removed_lines"`
		} `json:"diffStat"`
	}

	geminiPromptLog struct {
		SessionID string `json:"sessionId"`
		Type      string `json:"type"`
		Message   string `json:"message"`
		Timestamp string `json:"timestamp"`
	}

	geminiProjectPaths struct {
		bySlug map[string]string
	}
)

// Parse parses Gemini and Antigravity session logs for ai heartbeats.
func (g Gemini) Parse(ctx context.Context) (Heartbeats, error) {
	transcripts, err := g.transcriptPaths(ctx)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	if len(transcripts) > 0 {
		projectPaths, err := g.projectPaths(ctx)
		if err != nil {
			return nil, err
		}

		for slug, sessionPaths := range transcripts {
			for _, transcript := range sessionPaths {
				parsed, err := g.parseTranscript(transcript, projectPaths.bySlug[slug])
				if err != nil {
					return nil, err
				}

				heartbeats = append(heartbeats, parsed...)
			}
		}
	}

	antigravityHeartbeats, err := g.parseAntigravity(ctx)
	if err != nil {
		return nil, err
	}

	heartbeats = append(heartbeats, antigravityHeartbeats...)

	return heartbeats, nil
}

func (g Gemini) transcriptPaths(ctx context.Context) (map[string][]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	tmpDir := filepath.Join(home, ".gemini", "tmp")
	if _, err := os.Stat(tmpDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat .gemini tmp directory: %s", err)
	}

	transcripts := make(map[string][]string)

	err = filepath.WalkDir(tmpDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() {
			return nil
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".json" && ext != ".jsonl" {
			return nil
		}

		rel, err := filepath.Rel(tmpDir, path)
		if err != nil {
			return nil
		}

		parts := strings.Split(filepath.Clean(rel), string(filepath.Separator))
		if len(parts) < 3 || parts[1] != "chats" {
			return nil
		}

		// Main-agent sessions use the session- prefix. Subagent sessions are
		// nested below chats/<parent-session>/ and use their bare session ID.
		if len(parts) == 3 && !strings.HasPrefix(entry.Name(), "session-") {
			return nil
		}

		info, err := entry.Info()
		if err != nil || !timestampAtOrAfterCutoff(info.ModTime(), g.After) {
			return nil
		}

		slug := parts[0]
		if slug == "." || slug == string(filepath.Separator) {
			return nil
		}

		transcripts[slug] = append(transcripts[slug], path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk .gemini tmp directory: %s", err)
	}

	// Gemini leaves the legacy .json file in place when migrating a resumed
	// session to .jsonl. Prefer the JSONL copy with the same basename so the
	// migrated history is not emitted twice.
	for slug, paths := range transcripts {
		jsonlBases := make(map[string]struct{})

		for _, path := range paths {
			if filepath.Ext(path) == ".jsonl" {
				jsonlBases[strings.TrimSuffix(path, ".jsonl")] = struct{}{}
			}
		}

		filtered := paths[:0]
		for _, path := range paths {
			if filepath.Ext(path) == ".json" {
				if _, migrated := jsonlBases[strings.TrimSuffix(path, ".json")]; migrated {
					continue
				}
			}

			filtered = append(filtered, path)
		}

		transcripts[slug] = filtered
	}

	return transcripts, nil
}

func (Gemini) projectPaths(ctx context.Context) (geminiProjectPaths, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return geminiProjectPaths{}, fmt.Errorf("failed to find user home dir: %s", err)
	}

	tmpDir := filepath.Join(home, ".gemini", "tmp")
	projectPaths := geminiProjectPaths{bySlug: map[string]string{}}

	entries, err := os.ReadDir(tmpDir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			slug := entry.Name()
			rootPath := filepath.Join(tmpDir, slug, ".project_root")

			contents, err := os.ReadFile(filepath.Clean(rootPath))
			if err != nil {
				continue
			}

			projectPath := strings.TrimSpace(string(contents))
			if projectPath != "" {
				projectPaths.bySlug[slug] = projectPath
			}
		}
	} else if !os.IsNotExist(err) {
		return geminiProjectPaths{}, fmt.Errorf("failed to read .gemini tmp directory: %s", err)
	}

	projectsPath := filepath.Join(home, ".gemini", "projects.json")

	contents, err := os.ReadFile(filepath.Clean(projectsPath))
	if err != nil {
		if os.IsNotExist(err) {
			return projectPaths, nil
		}

		return geminiProjectPaths{}, fmt.Errorf("failed to read .gemini projects registry: %s", err)
	}

	var registry geminiProjectsRegistry
	if err := json.Unmarshal(contents, &registry); err != nil {
		return geminiProjectPaths{}, fmt.Errorf("failed to parse .gemini projects registry: %s", err)
	}

	for cwd, slug := range registry.Projects {
		if cwd == "" || slug == "" {
			continue
		}

		if _, found := projectPaths.bySlug[slug]; !found {
			projectPaths.bySlug[slug] = cwd
		}
	}

	return projectPaths, nil
}

func (g Gemini) parseTranscript(transcript string, projectPath string) (Heartbeats, error) {
	session, err := loadGeminiSession(transcript)
	if err != nil {
		return nil, err
	}

	sessionID := session.SessionID
	if sessionID == "" {
		sessionID = strings.TrimSuffix(filepath.Base(transcript), filepath.Ext(transcript))
	}

	sessionEntity := appHeartbeatEntity(g.Name(), sessionID)
	projectPath = g.projectPathFromLogs(transcript, projectPath)

	state := heartbeat.AITokens{}

	var (
		heartbeats     Heartbeats
		sawUserMessage bool
	)

	for _, message := range session.Messages {
		messageTime := parseGeminiTime(message.Timestamp)
		if message.Tokens != nil {
			state.CurrentInput = message.Tokens.Input - message.Tokens.Cached
			if state.CurrentInput < 0 {
				state.CurrentInput = 0
			}

			state.CurrentCachedInput = max(message.Tokens.Cached, 0)
			state.CurrentOutput = message.Tokens.Output
		}

		if messageTime.IsZero() || !timestampAtOrAfterCutoff(messageTime, g.After) {
			state.LastInput = state.CurrentInput
			state.LastCachedInput = state.CurrentCachedInput
			state.LastOutput = state.CurrentOutput

			continue
		}

		messageHeartbeats := g.messageHeartbeats(
			messageTime,
			sessionEntity,
			sessionID,
			projectPath,
			message,
			state,
		)
		if len(messageHeartbeats) == 0 {
			state.LastInput = state.CurrentInput
			state.LastCachedInput = state.CurrentCachedInput
			state.LastOutput = state.CurrentOutput

			continue
		}

		if message.Type == "user" {
			sawUserMessage = true
		}

		state.LastInput = state.CurrentInput
		state.LastCachedInput = state.CurrentCachedInput
		state.LastOutput = state.CurrentOutput

		heartbeats = append(heartbeats, messageHeartbeats...)
	}

	if sawUserMessage {
		return heartbeats, nil
	}

	fallbackPrompts, err := g.promptHeartbeatsFromLogs(transcript, sessionEntity, sessionID, projectPath)
	if err != nil {
		return nil, err
	}

	return append(fallbackPrompts, heartbeats...), nil
}

func loadGeminiSession(transcript string) (geminiSession, error) {
	if filepath.Ext(transcript) == ".json" {
		contents, err := os.ReadFile(filepath.Clean(transcript))
		if err != nil {
			return geminiSession{}, fmt.Errorf("failed to read Gemini session %q: %s", transcript, err)
		}

		var session geminiSession
		if err := json.Unmarshal(contents, &session); err != nil {
			return geminiSession{}, fmt.Errorf("failed to parse Gemini session %q: %s", transcript, err)
		}

		return session, nil
	}

	//nolint:gosec // Reads a transcript path discovered below the Gemini storage directory.
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return geminiSession{}, fmt.Errorf("failed to read Gemini session %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	var session geminiSession

	messageIndexes := make(map[string]int)

	rebuildMessageIndexes := func() {
		clear(messageIndexes)

		for i, message := range session.Messages {
			messageIndexes[message.ID] = i
		}
	}

	setMessages := func(messages []geminiMessage) {
		session.Messages = nil

		clear(messageIndexes)

		for _, message := range messages {
			if index, ok := messageIndexes[message.ID]; ok {
				session.Messages[index] = message
				continue
			}

			messageIndexes[message.ID] = len(session.Messages)
			session.Messages = append(session.Messages, message)
		}
	}

	upsertMessage := func(message geminiMessage) {
		if index, ok := messageIndexes[message.ID]; ok {
			session.Messages[index] = message
			return
		}

		messageIndexes[message.ID] = len(session.Messages)
		session.Messages = append(session.Messages, message)
	}

	applyUpdate := func(update geminiSessionUpdate) {
		if update.SessionID != nil {
			session.SessionID = *update.SessionID
		}

		if update.StartTime != nil {
			session.StartTime = *update.StartTime
		}

		if update.LastUpdated != nil {
			session.LastUpdated = *update.LastUpdated
		}

		if update.Messages != nil {
			setMessages(*update.Messages)
		}
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		var control geminiJSONLControl
		if json.Unmarshal(line, &control) != nil {
			continue
		}

		switch {
		case control.RewindTo != nil:
			index, ok := messageIndexes[*control.RewindTo]
			if !ok {
				session.Messages = nil
			} else {
				session.Messages = session.Messages[:index]
			}

			rebuildMessageIndexes()
		case control.Set != nil:
			applyUpdate(*control.Set)
		default:
			var message geminiMessage
			if json.Unmarshal(line, &message) == nil && message.ID != "" && message.Type != "" {
				upsertMessage(message)
				continue
			}

			var update geminiSessionUpdate
			if json.Unmarshal(line, &update) == nil {
				applyUpdate(update)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return geminiSession{}, fmt.Errorf("failed to read Gemini session %q: %s", transcript, err)
	}

	if session.SessionID == "" {
		return geminiSession{}, fmt.Errorf("failed to parse Gemini session %q: missing sessionId", transcript)
	}

	return session, nil
}

func (g Gemini) messageHeartbeats(
	messageTime time.Time,
	sessionEntity string,
	sessionID string,
	projectPath string,
	message geminiMessage,
	tokens heartbeat.AITokens,
) Heartbeats {
	var heartbeats Heartbeats

	assignTokens := g.hasTokenDelta(tokens)

	if messageHeartbeat := g.appHeartbeat(
		messageTime,
		sessionEntity,
		sessionID,
		projectPath,
		message,
		g.tokensForFirstHeartbeat(assignTokens, tokens),
	); messageHeartbeat != nil {
		heartbeats = append(heartbeats, *messageHeartbeat)
		assignTokens = false
	}

	if message.Type != "gemini" || len(message.ToolCalls) == 0 {
		return heartbeats
	}

	toolCalls := append([]geminiToolCall(nil), message.ToolCalls...)
	slices.SortFunc(toolCalls, func(a, b geminiToolCall) int {
		return strings.Compare(a.Timestamp, b.Timestamp)
	})

	for _, toolCall := range toolCalls {
		hb, ok := g.toolHeartbeat(
			messageTime,
			sessionID,
			projectPath,
			message.Model,
			toolCall,
			g.tokensForFirstHeartbeat(assignTokens, tokens),
		)
		if !ok {
			continue
		}

		heartbeats = append(heartbeats, hb)
		assignTokens = false
	}

	return heartbeats
}

func (g Gemini) appHeartbeat(
	timestamp time.Time,
	entity string,
	sessionID string,
	projectPath string,
	message geminiMessage,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	text := geminiText(message.Content)
	if strings.TrimSpace(text) == "" {
		return nil
	}

	role := message.Type
	if role != "user" && role != "gemini" {
		return nil
	}

	aiTokens := heartbeat.AITokens{}
	if tokens != nil {
		aiTokens = *tokens
	}

	h := heartbeat.NewWithAITokens(
		nil,
		sessionID,
		aiTokens,
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
		projectPath,
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, message.Model, ""),
	)
	if role == "user" {
		h.AIPromptLength = promptLength(text)
	}

	return &h
}

func (g Gemini) toolHeartbeat(
	defaultTime time.Time,
	sessionID string,
	projectPath string,
	model string,
	toolCall geminiToolCall,
	tokens *heartbeat.AITokens,
) (heartbeat.Heartbeat, bool) {
	if toolCall.Status != "" && toolCall.Status != "success" {
		return heartbeat.Heartbeat{}, false
	}

	timestamp := parseGeminiTime(toolCall.Timestamp)
	if timestamp.IsZero() {
		timestamp = defaultTime
	}

	toolName := strings.TrimSpace(toolCall.Name)
	if toolName == "" {
		toolName = strings.TrimSpace(toolCall.DisplayName)
	}

	toolName = strings.ToLower(toolName)

	aiTokens := heartbeat.AITokens{}
	if tokens != nil {
		aiTokens = *tokens
	}

	switch toolName {
	case "writefile", "write_file":
		var args geminiWriteArgs
		if err := json.Unmarshal(toolCall.Args, &args); err != nil || args.FilePath == "" {
			return heartbeat.Heartbeat{}, false
		}

		filePath := g.toolFilePath(projectPath, toolCall.ResultDisplay, args.FilePath)

		return heartbeat.NewWithAITokens(
			heartbeat.PointerTo(countStringLines(args.Content)),
			sessionID,
			aiTokens,
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
			aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, model, ""),
		), true
	case "edit", "replace":
		var args geminiReplaceArgs
		if err := json.Unmarshal(toolCall.Args, &args); err != nil || args.FilePath == "" {
			return heartbeat.Heartbeat{}, false
		}

		filePath := g.toolFilePath(projectPath, toolCall.ResultDisplay, args.FilePath)
		lineChanges := g.replaceLineChanges(toolCall.ResultDisplay, args.OldString, args.NewString)

		return heartbeat.NewWithAITokens(
			heartbeat.PointerTo(lineChanges),
			sessionID,
			aiTokens,
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
			aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, model, ""),
		), true
	default:
		return heartbeat.Heartbeat{}, false
	}
}

func (g Gemini) promptHeartbeatsFromLogs(
	transcript string,
	sessionEntity string,
	sessionID string,
	projectPath string,
) (Heartbeats, error) {
	logsPath := filepath.Join(geminiProjectTempDir(transcript), "logs.json")

	contents, err := os.ReadFile(filepath.Clean(logsPath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read Gemini logs %q: %s", logsPath, err)
	}

	var logs []geminiPromptLog
	if err := json.Unmarshal(contents, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini logs %q: %s", logsPath, err)
	}

	var heartbeats Heartbeats

	for _, entry := range logs {
		if entry.SessionID != sessionID || entry.Type != "user" {
			continue
		}

		timestamp := parseGeminiTime(entry.Timestamp)
		if timestamp.IsZero() || !timestampAtOrAfterCutoff(timestamp, g.After) || strings.TrimSpace(entry.Message) == "" {
			continue
		}

		h := heartbeat.NewWithAITokens(
			nil,
			sessionID,
			heartbeat.AITokens{},
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			sessionEntity,
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
			projectPath,
			heartbeatTimestamp(timestamp),
			aiUserAgentWithModel(sessionEntity, g.UserAgents, g.FallbackUserAgent, "", ""),
		)
		h.AIPromptLength = promptLength(entry.Message)
		heartbeats = append(heartbeats, h)
	}

	return heartbeats, nil
}

func (Gemini) projectPathFromLogs(transcript string, fallback string) string {
	if fallback != "" {
		return fallback
	}

	rootPath := filepath.Join(geminiProjectTempDir(transcript), ".project_root")

	contents, err := os.ReadFile(filepath.Clean(rootPath))
	if err == nil {
		if projectPath := strings.TrimSpace(string(contents)); projectPath != "" {
			return projectPath
		}
	}

	return fallback
}

func geminiProjectTempDir(transcript string) string {
	dir := filepath.Dir(transcript)

	for {
		if filepath.Base(dir) == "chats" {
			return filepath.Dir(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return filepath.Dir(filepath.Dir(transcript))
		}

		dir = parent
	}
}

func (Gemini) toolFilePath(projectPath string, resultDisplay json.RawMessage, filePath string) string {
	var display geminiResultDisplay
	if len(resultDisplay) > 0 && string(resultDisplay) != "null" {
		if err := json.Unmarshal(resultDisplay, &display); err == nil && display.FilePath != "" {
			return display.FilePath
		}
	}

	return geminiResolvePath(projectPath, filePath)
}

func (Gemini) replaceLineChanges(resultDisplay json.RawMessage, oldString string, newString string) int {
	var display geminiResultDisplay
	if len(resultDisplay) > 0 && string(resultDisplay) != "null" {
		if err := json.Unmarshal(resultDisplay, &display); err == nil && display.DiffStat != nil {
			return display.DiffStat.ModelAddedLines - display.DiffStat.ModelRemovedLines
		}
	}

	return countStringLines(newString) - countStringLines(oldString)
}

func (Gemini) hasTokenDelta(tokens heartbeat.AITokens) bool {
	input := tokens.CurrentInput - tokens.LastInput
	if input < 0 {
		input = 0
	}

	cachedInput := tokens.CurrentCachedInput - tokens.LastCachedInput
	if cachedInput < 0 {
		cachedInput = 0
	}

	output := tokens.CurrentOutput - tokens.LastOutput
	if output < 0 {
		output = 0
	}

	return input > 0 || cachedInput > 0 || output > 0
}

func (Gemini) tokensForFirstHeartbeat(assign bool, tokens heartbeat.AITokens) *heartbeat.AITokens {
	if !assign {
		return nil
	}

	copy := tokens

	return &copy
}

func geminiText(content json.RawMessage) string {
	if len(content) == 0 || string(content) == "null" {
		return ""
	}

	var str string
	if err := json.Unmarshal(content, &str); err == nil {
		return str
	}

	var parts []geminiTextPart
	if err := json.Unmarshal(content, &parts); err == nil {
		var builder strings.Builder
		for _, part := range parts {
			builder.WriteString(part.Text)
		}

		return builder.String()
	}

	return ""
}

func parseGeminiTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}

	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z07:00",
	} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}

	return time.Time{}
}

func geminiResolvePath(projectPath string, filePath string) string {
	if filePath == "" {
		return ""
	}

	if filepath.IsAbs(filePath) {
		return filePath
	}

	if projectPath == "" {
		return filePath
	}

	return filepath.Join(projectPath, filePath)
}

// Name returns its id.
func (Gemini) Name() string {
	return "Gemini"
}
