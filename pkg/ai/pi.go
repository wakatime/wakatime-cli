package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Pi contains params for detecting heartbeats from Pi session JSONL logs.
type Pi ParserConfig

type (
	piSessionState struct {
		cwd     string
		entity  string
		id      string
		version string
	}

	piParseState struct {
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
		toolCalls  map[string]piToolCall
		version    string
	}

	piSessionHeader struct {
		Type string  `json:"type"`
		ID   *string `json:"id"`
		Cwd  *string `json:"cwd"`
	}

	piUsage struct {
		Input      *int64 `json:"input"`
		Output     *int64 `json:"output"`
		CacheRead  *int64 `json:"cacheRead"`
		CacheWrite *int64 `json:"cacheWrite"`
	}

	piContentBlock struct {
		Type      string                 `json:"type"`
		Text      string                 `json:"text"`
		Name      string                 `json:"name"`
		ID        string                 `json:"id"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	piMessage struct {
		Role       string           `json:"role"`
		Content    []piContentBlock `json:"content"`
		Provider   string           `json:"provider"`
		Model      string           `json:"model"`
		ToolCallID string           `json:"toolCallId"`
		ToolName   string           `json:"toolName"`
		Usage      *piUsage         `json:"usage"`
		IsError    bool             `json:"isError"`
		Details    json.RawMessage  `json:"details"`
	}

	piLogLine struct {
		Timestamp time.Time  `json:"timestamp"`
		Type      string     `json:"type"`
		Message   *piMessage `json:"message"`
		Provider  string     `json:"provider"`
		ModelID   string     `json:"modelId"`
	}

	piToolCall struct {
		Name      string
		Arguments map[string]interface{}
	}
)

var piSuccessPathPattern = regexp.MustCompile(`(?i)\b(?:in|to)\s+([~/A-Za-z]:[^\s]+|/[^\s]+)`)

// Parse parses the Pi JSONL session logs for ai heartbeats.
func (g Pi) Parse(ctx context.Context) (Heartbeats, error) {
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
			logger.Warnf("failed parsing pi transcript %q: %s", transcript, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Pi) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	sessionsDir := filepath.Join(home, ".pi", "agent", "sessions")
	if _, err := os.Stat(sessionsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat .pi sessions directory: %s", err)
	}

	var transcripts []string

	err = filepath.WalkDir(sessionsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" {
			return nil
		}

		info, err := entry.Info()
		if err != nil || info.ModTime().Before(g.After) {
			return nil
		}

		transcripts = append(transcripts, path)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk .pi sessions directory: %s", err)
	}

	return transcripts, nil
}

func (g Pi) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open pi transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	session, err := g.readSessionState(logger, fh, transcript)
	if err != nil {
		return nil, err
	}

	scanner, err := piScanner(fh, transcript)
	if err != nil {
		return nil, err
	}

	state := piParseState{
		toolCalls: make(map[string]piToolCall),
		version:   session.version,
	}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleTranscriptLine(logger, transcript, scanner.Bytes(), session, &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading pi transcript %q: %s", transcript, err)
	}

	return state.heartbeats, nil
}

func (g Pi) readSessionState(logger *log.Logger, fh *os.File, transcript string) (piSessionState, error) {
	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return piSessionState{}, fmt.Errorf("failed to read pi transcript %q: %s", transcript, err)
	}

	state := piSessionState{
		entity: appHeartbeatEntity("Pi", transcript),
		id:     g.sessionIDFromPath(transcript),
	}

	if len(firstLine) > 0 {
		var sessionHeader *piSessionHeader
		if err := json.Unmarshal(firstLine, &sessionHeader); err != nil {
			logger.Debugf("failed parsing pi session metadata from %q: %s", transcript, err)
		} else {
			state.cwd, state.id = g.sessionInfo(state.cwd, state.id, sessionHeader)
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return piSessionState{}, fmt.Errorf("failed to rewind pi transcript %q: %s", transcript, err)
	}

	return state, nil
}

func piScanner(fh *os.File, transcript string) (*bufio.Scanner, error) {
	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat pi transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek pi transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read pi transcript %q: %s", transcript, err)
		}
	}

	return scanner, nil
}

func (g Pi) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	session piSessionState,
	state *piParseState,
) {
	if len(line) == 0 {
		return
	}

	var logLine piLogLine
	if err := json.Unmarshal(line, &logLine); err != nil {
		logger.Warnf("failed parsing pi transcript line from %q: %s", transcript, err)
		logger.Debugf("failed parsing pi transcript line: %s", line)

		return
	}

	state.tokens = g.piTokenCounts(logLine, state.tokens, g.After)
	state.version = g.sessionVersion(state.version, logLine)

	if logLine.Timestamp.IsZero() || logLine.Timestamp.Before(g.After) {
		return
	}

	aiHeartbeats := g.getHeartbeats(
		logLine.Timestamp,
		session.entity,
		session.id,
		state.version,
		session.cwd,
		g.UserAgents,
		g.FallbackUserAgent,
		logLine,
		state,
	)
	if len(aiHeartbeats) == 0 {
		return
	}

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
	state.heartbeats = append(state.heartbeats, aiHeartbeats...)
}

func (Pi) sessionInfo(cwd string, sessionID string, header *piSessionHeader) (string, string) {
	if header == nil || header.Type != "session" {
		return cwd, sessionID
	}

	if header.ID != nil && *header.ID != "" {
		sessionID = *header.ID
	}

	if header.Cwd != nil && *header.Cwd != "" {
		cwd = *header.Cwd
	}

	return cwd, sessionID
}

func (Pi) sessionVersion(version string, logLine piLogLine) string {
	switch {
	case logLine.Type == "model_change" && logLine.ModelID != "":
		return logLine.ModelID
	case logLine.Message != nil && logLine.Message.Model != "":
		return logLine.Message.Model
	case logLine.Type == "model_change" && logLine.Provider != "":
		return logLine.Provider
	case logLine.Message != nil && logLine.Message.Provider != "":
		return logLine.Message.Provider
	default:
		return version
	}
}

func (g Pi) getHeartbeats(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	logLine piLogLine,
	state *piParseState,
) Heartbeats {
	if logLine.Message == nil {
		return nil
	}

	msg := *logLine.Message

	if msg.Role == "user" || msg.Role == "assistant" {
		g.trackToolCalls(msg, state.toolCalls)

		if heartbeat := g.messageHeartbeat(
			timestamp,
			sessionEntity,
			sessionID,
			version,
			cwd,
			userAgents,
			fallbackUserAgent,
			msg,
			state.tokens,
		); heartbeat != nil {
			return Heartbeats{*heartbeat}
		}
	}

	if msg.Role != "toolResult" || msg.IsError {
		return nil
	}

	if heartbeat := g.toolResultHeartbeat(
		timestamp,
		sessionID,
		version,
		cwd,
		userAgents,
		fallbackUserAgent,
		msg,
		state.toolCalls[msg.ToolCallID],
		state.tokens,
	); heartbeat != nil {
		return Heartbeats{*heartbeat}
	}

	return nil
}

func (Pi) trackToolCalls(msg piMessage, toolCalls map[string]piToolCall) {
	for _, block := range msg.Content {
		if block.Type != "toolCall" || block.ID == "" || block.Name == "" {
			continue
		}

		toolCalls[block.ID] = piToolCall{
			Name:      block.Name,
			Arguments: block.Arguments,
		}
	}
}

func (g Pi) messageHeartbeat(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	msg piMessage,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	var (
		lineChanges int
		promptChars int
	)

	for _, block := range msg.Content {
		if block.Type != "text" || strings.TrimSpace(block.Text) == "" {
			continue
		}

		text := block.Text
		if msg.Role == "user" {
			text = piUserMessageText(text)
			if text == "" {
				continue
			}

			promptChars += len([]rune(text))
		}

		lineChanges += countStringLines(text)
	}

	if lineChanges == 0 {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		nil,
		sessionID,
		tokens,
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
		cwd,
		float64(timestamp.UnixMilli())/1000,
		aiUserAgentWithAgentPrefix(sessionEntity, userAgents, fallbackUserAgent, aiPlugin(g, ""), version),
	)
	if msg.Role == "user" && promptChars > 0 {
		h.AIPromptLength = promptChars
	}

	return &h
}

func piUserMessageText(text string) string {
	return strings.TrimSpace(text)
}

func (g Pi) toolResultHeartbeat(
	timestamp time.Time,
	sessionID string,
	version string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	msg piMessage,
	call piToolCall,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	toolName := firstNonEmptyString(msg.ToolName, call.Name)
	if toolName == "" {
		return nil
	}

	filePath := piToolPath(msg, call.Arguments, cwd)
	if filePath == "" {
		return nil
	}

	lineChanges, isWrite := g.toolLineChanges(toolName, msg, call.Arguments)

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		sessionID,
		tokens,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		filePath,
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
		float64(timestamp.UnixMilli())/1000,
		aiUserAgentWithAgentPrefix(filePath, userAgents, fallbackUserAgent, aiPlugin(g, ""), version),
	)

	return &h
}

func (Pi) piTokenCounts(line piLogLine, tokens heartbeat.AITokens, after time.Time) heartbeat.AITokens {
	if line.Message == nil || line.Message.Role != "assistant" || line.Message.Usage == nil {
		return tokens
	}

	usage := *line.Message.Usage

	inputTokens := int64(0)
	if usage.Input != nil {
		inputTokens = *usage.Input
	}

	if usage.CacheRead != nil {
		inputTokens += *usage.CacheRead
	}

	if usage.CacheWrite != nil {
		inputTokens += *usage.CacheWrite
	}

	outputTokens := int64(0)
	if usage.Output != nil {
		outputTokens = *usage.Output
	}

	tokens.CurrentInput = tokens.LastInput + inputTokens
	tokens.CurrentOutput = tokens.LastOutput + outputTokens

	if line.Timestamp.IsZero() || line.Timestamp.Before(after) {
		tokens.LastInput = tokens.CurrentInput
		tokens.LastOutput = tokens.CurrentOutput
	}

	return tokens
}

func piToolPath(msg piMessage, args map[string]interface{}, cwd string) string {
	for _, key := range []string{"filePath", "path", "targetFile", "target_file"} {
		if value := piStringValue(args, key); value != "" {
			return piAbsPath(value, cwd)
		}
	}

	for _, block := range msg.Content {
		if block.Type != "text" {
			continue
		}

		if matches := piSuccessPathPattern.FindStringSubmatch(block.Text); len(matches) == 2 {
			return piAbsPath(matches[1], cwd)
		}
	}

	return ""
}

func (g Pi) toolLineChanges(toolName string, msg piMessage, args map[string]interface{}) (int, bool) {
	if !piIsWriteTool(toolName) {
		return 0, false
	}

	if diff := g.diffFromDetails(msg.Details); diff != "" {
		return piDiffLineChanges(diff), true
	}

	for _, key := range []string{"content", "text", "newText", "new_text"} {
		if value := piStringValue(args, key); value != "" {
			return countStringLines(value), true
		}
	}

	return 0, true
}

func piIsWriteTool(name string) bool {
	switch strings.ToLower(name) {
	case "edit", "write", "multiedit", "patch":
		return true
	default:
		return false
	}
}

func (Pi) diffFromDetails(details json.RawMessage) string {
	if len(details) == 0 || string(details) == "null" {
		return ""
	}

	var payload struct {
		Diff string `json:"diff"`
	}
	if err := json.Unmarshal(details, &payload); err != nil {
		return ""
	}

	return payload.Diff
}

func piDiffLineChanges(diff string) int {
	additions := 0
	deletions := 0

	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+++"), strings.HasPrefix(line, "---"), strings.HasPrefix(line, "@@"):
			continue
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			deletions++
		}
	}

	return additions - deletions
}

func piStringValue(values map[string]interface{}, key string) string {
	if values == nil {
		return ""
	}

	value, found := values[key]
	if !found {
		return ""
	}

	str, ok := value.(string)
	if !ok {
		return ""
	}

	return str
}

func piAbsPath(path string, cwd string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	if filepath.IsAbs(path) {
		return path
	}

	if cwd == "" {
		return path
	}

	return filepath.Join(cwd, path)
}

func (Pi) sessionIDFromPath(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if _, id, ok := strings.Cut(base, "_"); ok && id != "" {
		return id
	}

	return base
}

// Name returns its id.
func (Pi) Name() string {
	return "Pi"
}
