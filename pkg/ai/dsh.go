package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Dsh contains parameters for detecting heartbeats from DeepSeek Harness session logs.
type Dsh ParserConfig

type (
	dshSessionState struct {
		cwd string
		id  string
	}

	dshParseState struct {
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
		model      string
		toolCalls  map[string]dshToolCall
	}

	dshContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	dshMessageSource struct {
		Kind     string `json:"kind"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}

	dshHeader struct {
		Config *dshMessageSource `json:"config"`
	}

	dshMessage struct {
		Role    string            `json:"role"`
		Content []dshContentBlock `json:"content"`
		Source  *dshMessageSource `json:"source"`
	}

	dshUsage struct {
		InputTokens     *int64 `json:"inputTokens"`
		OutputTokens    *int64 `json:"outputTokens"`
		CacheReadTokens *int64 `json:"cacheReadTokens"`
	}

	dshEventData struct {
		Role      string            `json:"role"`
		Content   []dshContentBlock `json:"content"`
		Message   *dshMessage       `json:"message"`
		Usage     *dshUsage         `json:"usage"`
		CallID    string            `json:"callId"`
		Name      string            `json:"name"`
		Arguments json.RawMessage   `json:"arguments"`
		Provider  string            `json:"provider"`
		Model     string            `json:"model"`
		Header    *dshHeader        `json:"header"`
		ID        string            `json:"id"`
		Cwd       string            `json:"cwd"`
	}

	dshLogLine struct {
		Type string       `json:"type"`
		Time int64        `json:"time"` // epoch milliseconds.
		Data dshEventData `json:"data"`
		ID   string       `json:"id"`
		Cwd  string       `json:"cwd"`
	}

	dshToolCall struct {
		Name      string
		Arguments map[string]interface{}
	}
)

// Parse parses the DeepSeek Harness session logs for ai heartbeats.
func (g Dsh) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	transcripts, err := g.transcriptPaths(ctx)
	if err != nil {
		return nil, err
	}

	if len(transcripts) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d transcript logs modified after %s for %s", len(transcripts), g.After, g.Name())

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		parsed, err := g.parseTranscript(ctx, transcript)
		if err != nil {
			logger.Warnf("failed parsing deepseek harness transcript %q: %s", transcript, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

// Name returns its id.
func (Dsh) Name() string { return "DeepSeek Harness" }

func (g Dsh) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	sessionsDir := filepath.Join(envOrDefault("DSH_HOME", filepath.Join(home, ".dsh")), "sessions")
	if _, err := os.Stat(sessionsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat .dsh sessions directory: %s", err)
	}

	var transcripts []string

	err = filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl.zstd") {
			return nil
		}

		info, err := entry.Info()
		if err != nil || !timestampAtOrAfterCutoff(info.ModTime(), g.After) {
			return nil
		}

		transcripts = append(transcripts, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk .dsh sessions directory: %s", err)
	}

	return transcripts, nil
}

func (g Dsh) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open deepseek harness transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	decoder, err := zstd.NewReader(fh)
	if err != nil {
		return nil, fmt.Errorf("failed to create zstd reader for deepseek harness transcript %q: %s", transcript, err)
	}
	defer decoder.Close()

	session := dshSessionState{id: dshSessionIDFromPath(transcript)}
	state := dshParseState{toolCalls: make(map[string]dshToolCall)}

	scanner := bufio.NewScanner(decoder)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		g.handleTranscriptLine(logger, transcript, line, &session, &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading deepseek harness transcript %q: %s", transcript, err)
	}

	return state.heartbeats, nil
}

func (g Dsh) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	session *dshSessionState,
	state *dshParseState,
) {
	if len(line) == 0 {
		return
	}

	var logLine dshLogLine
	if err := json.Unmarshal(line, &logLine); err != nil {
		logger.Warnf("failed parsing deepseek harness transcript line from %q: %s", transcript, err)
		logger.Debugf("failed parsing deepseek harness transcript line: %s", line)

		return
	}

	timestamp := time.UnixMilli(logLine.Time).UTC()

	switch logLine.Type {
	case "session":
		if id := firstNonEmptyString(logLine.ID, logLine.Data.ID); id != "" {
			session.id = strings.TrimPrefix(id, "session-")
		}

		if cwd := firstNonEmptyString(logLine.Cwd, logLine.Data.Cwd); cwd != "" {
			session.cwd = cwd
		}
	case "request/header":
		if header := logLine.Data.Header; header != nil && header.Config != nil && header.Config.Model != "" {
			state.model = header.Config.Model
		}
	case "request/context":
		if logLine.Data.Model != "" {
			state.model = logLine.Data.Model
		}
	case "user/message":
		if logLine.Time <= 0 || !timestampAtOrAfterCutoff(timestamp, g.After) {
			return
		}

		g.userMessageHeartbeat(timestamp, *session, state, logLine.Data)
	case "assistant/message":
		msg := logLine.Data.Message
		if msg == nil {
			msg = &dshMessage{Role: logLine.Data.Role, Content: logLine.Data.Content}
		}

		if msg.Source != nil && msg.Source.Model != "" {
			state.model = msg.Source.Model
		}

		state.tokens = g.tokenCounts(logLine.Data.Usage, state.tokens, timestamp, g.After)

		if logLine.Time <= 0 || !timestampAtOrAfterCutoff(timestamp, g.After) {
			return
		}

		g.assistantMessageHeartbeat(timestamp, *session, state, msg)
	case "tool/call":
		if logLine.Time <= 0 || !timestampAtOrAfterCutoff(timestamp, g.After) {
			return
		}

		g.toolCallHeartbeat(timestamp, *session, state, logLine.Data)
	}
}

func (g Dsh) userMessageHeartbeat(
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
	data dshEventData,
) {
	var (
		lineChanges int
		promptChars int
	)

	for _, block := range data.Content {
		if block.Type != "text" || strings.TrimSpace(block.Text) == "" {
			continue
		}

		lineChanges += countStringLines(block.Text)
		promptChars += promptLength(block.Text)
	}

	if lineChanges == 0 {
		return
	}

	entity := g.Name()

	h := heartbeat.NewWithAITokens(
		nil,
		session.id,
		state.tokens,
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
		session.cwd,
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, state.model, ""),
	)

	if promptChars > 0 {
		h.AIPromptLength = promptChars
	}

	state.heartbeats = append(state.heartbeats, h)
}

func (g Dsh) assistantMessageHeartbeat(
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
	msg *dshMessage,
) {
	if msg == nil {
		return
	}

	var lineChanges int

	for _, block := range msg.Content {
		if block.Type != "text" || strings.TrimSpace(block.Text) == "" {
			continue
		}

		lineChanges += countStringLines(block.Text)
	}

	if lineChanges == 0 {
		return
	}

	entity := g.Name()

	h := heartbeat.NewWithAITokens(
		nil,
		session.id,
		state.tokens,
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
		session.cwd,
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, state.model, ""),
	)

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastCachedInput = state.tokens.CurrentCachedInput
	state.tokens.LastOutput = state.tokens.CurrentOutput

	state.heartbeats = append(state.heartbeats, h)
}

func (g Dsh) toolCallHeartbeat(
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
	data dshEventData,
) {
	args := dshToolArguments(data.Arguments)

	if data.CallID != "" && data.Name != "" {
		state.toolCalls[data.CallID] = dshToolCall{Name: data.Name, Arguments: args}
	}

	filePath := dshToolPath(args, session.cwd)
	if filePath == "" {
		return
	}

	isWrite := dshIsWriteTool(data.Name)

	lineChanges := 0
	if isWrite {
		lineChanges = dshWriteLineChanges(args)
	}

	entity := filePath

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		session.id,
		state.tokens,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		heartbeat.FileType,
		nil,
		false,
		heartbeat.PointerTo(isWrite),
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
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, state.model, ""),
	)

	state.heartbeats = append(state.heartbeats, h)
}

func (Dsh) tokenCounts(usage *dshUsage, tokens heartbeat.AITokens, timestamp time.Time, after time.Time) heartbeat.AITokens {
	if usage == nil {
		return tokens
	}

	inputTokens := int64(0)
	if usage.InputTokens != nil {
		inputTokens = *usage.InputTokens
	}

	cachedInputTokens := int64(0)
	if usage.CacheReadTokens != nil {
		cachedInputTokens = max(*usage.CacheReadTokens, 0)
	}

	outputTokens := int64(0)
	if usage.OutputTokens != nil {
		outputTokens = *usage.OutputTokens
	}

	tokens.CurrentInput = tokens.LastInput + inputTokens
	tokens.CurrentCachedInput = tokens.LastCachedInput + cachedInputTokens
	tokens.CurrentOutput = tokens.LastOutput + outputTokens

	if !timestampAtOrAfterCutoff(timestamp, after) {
		tokens.LastInput = tokens.CurrentInput
		tokens.LastCachedInput = tokens.CurrentCachedInput
		tokens.LastOutput = tokens.CurrentOutput
	}

	return tokens
}

// dshToolArguments decodes tool call arguments, which may be a JSON object or a
// JSON-encoded string wrapping one.
func dshToolArguments(raw json.RawMessage) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err == nil {
		return obj
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil
	}

	var inner map[string]interface{}
	if err := json.Unmarshal([]byte(encoded), &inner); err != nil {
		return nil
	}

	return inner
}

func dshToolPath(args map[string]interface{}, cwd string) string {
	// Note that file operations use "file_path"; the generic "path" key
	// belongs to search tools like grep and glob and holds a directory.
	for _, key := range []string{"file_path", "filePath"} {
		if value := piStringValue(args, key); value != "" {
			return piAbsPath(value, cwd)
		}
	}

	return ""
}

func dshIsWriteTool(name string) bool {
	switch strings.ToLower(name) {
	case "edit", "write", "multiedit", "patch", "apply_patch":
		return true
	default:
		return false
	}
}

func dshWriteLineChanges(args map[string]interface{}) int {
	for _, key := range []string{"content", "new_string", "newString", "text"} {
		if value := piStringValue(args, key); value != "" {
			return countStringLines(value)
		}
	}

	return 0
}

func dshSessionIDFromPath(path string) string {
	id := filepath.Base(filepath.Dir(path))

	return strings.TrimPrefix(id, "session-")
}
