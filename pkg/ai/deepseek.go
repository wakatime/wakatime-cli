package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// DeepSeek contains parameters for detecting heartbeats from DeepSeek Harness session transcripts.
type DeepSeek ParserConfig

type (
	dshSessionState struct {
		cwd        string
		id         string
		seedLength int
	}

	dshParseState struct {
		eventIndex              int
		heartbeats              Heartbeats
		model                   string
		pendingPromptHeartbeats []int
		pendingTools            map[string]dshToolCall
		pendingUsage            *dshPendingUsage
		tokens                  heartbeat.AITokens
	}

	dshSessionHeader struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Cwd        string `json:"cwd"`
		SeedLength int    `json:"seedLength"`
	}

	dshMessageSource struct {
		Kind     string `json:"kind"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
		CallID   string `json:"callId"`
	}

	dshContentBlock struct {
		Type       string            `json:"type"`
		Text       string            `json:"text"`
		ToolCallID string            `json:"toolCallId"`
		IsError    bool              `json:"isError"`
		Content    []dshContentBlock `json:"content"`
	}

	dshMessage struct {
		Role    string            `json:"role"`
		Content []dshContentBlock `json:"content"`
		Source  *dshMessageSource `json:"source"`
	}

	dshUsage struct {
		InputTokens      *int64 `json:"inputTokens"`
		OutputTokens     *int64 `json:"outputTokens"`
		CacheReadTokens  *int64 `json:"cacheReadTokens"`
		CacheWriteTokens *int64 `json:"cacheWriteTokens"`
	}

	dshChunk struct {
		Type   string           `json:"type"`
		Usage  *dshUsage        `json:"usage"`
		Reason *dshFinishReason `json:"reason"`
	}

	dshFinishReason struct {
		Kind string `json:"kind"`
	}

	dshRequestHeader struct {
		Config *dshRequestConfig `json:"config"`
	}

	dshRequestConfig struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}

	dshEventError struct {
		Name string `json:"name"`
		Code string `json:"code"`
	}

	dshEventData struct {
		Turn       int               `json:"turn"`
		Step       int               `json:"step"`
		Role       string            `json:"role"`
		Content    []dshContentBlock `json:"content"`
		Source     *dshMessageSource `json:"source"`
		Provenance *dshMessageSource `json:"provenance"`
		Message    *dshMessage       `json:"message"`
		Usage      *dshUsage         `json:"usage"`
		Chunk      *dshChunk         `json:"chunk"`
		CallID     string            `json:"callId"`
		Name       string            `json:"name"`
		Arguments  json.RawMessage   `json:"arguments"`
		IsError    bool              `json:"isError"`
		Provider   string            `json:"provider"`
		Model      string            `json:"model"`
		Header     *dshRequestHeader `json:"header"`
		Error      *dshEventError    `json:"error"`
	}

	dshLogLine struct {
		Type string       `json:"type"`
		Seq  *int         `json:"seq"`
		Time int64        `json:"time"`
		Data dshEventData `json:"data"`
	}

	dshToolCall struct {
		arguments map[string]interface{}
		name      string
	}

	dshPendingUsage struct {
		turn      int
		step      int
		timestamp time.Time
		usage     dshUsage
	}

	dshTranscriptReader struct {
		closer io.Closer
		reader io.Reader
	}
)

// Parse parses DeepSeek Harness JSONL session transcripts for AI heartbeats.
func (g DeepSeek) Parse(ctx context.Context) (Heartbeats, error) {
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

func (g DeepSeek) transcriptPaths(ctx context.Context) ([]string, error) {
	sessionsDir, err := dshSessionsDir(ctx)
	if err != nil {
		return nil, err
	}

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

		if entry.IsDir() || !dshTranscriptName(entry.Name()) {
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

func dshSessionsDir(ctx context.Context) (string, error) {
	root := strings.TrimSpace(os.Getenv("DSH_HOME"))
	if root == "" {
		home, err := ini.UserHomeDir(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to find user home dir: %s", err)
		}

		root = filepath.Join(home, ".dsh")
	}

	sessionsDir, err := filepath.Abs(filepath.Join(root, "sessions"))
	if err != nil {
		return "", fmt.Errorf("failed to resolve DSH_HOME: %s", err)
	}

	return sessionsDir, nil
}

func dshTranscriptName(name string) bool {
	return name == "session.jsonl" || name == "session.jsonl.zstd"
}

func (g DeepSeek) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	stream, err := dshOpenTranscript(transcript)
	if err != nil {
		return nil, err
	}
	defer stream.closer.Close() // nolint:errcheck

	scanner := bufio.NewScanner(stream.reader)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	session := dshSessionState{id: filepath.Base(filepath.Dir(transcript))}
	if scanner.Scan() {
		g.readSessionHeader(logger, transcript, scanner.Bytes(), &session)
	}

	state := dshParseState{pendingTools: make(map[string]dshToolCall)}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleTranscriptLine(logger, transcript, scanner.Bytes(), session, &state)
	}

	// DSH appends independent zstd frames. A concurrent reader can reach an incomplete
	// final frame after decoding all previously committed lines, which are still safe to use.
	if err := scanner.Err(); err != nil &&
		(!strings.HasSuffix(transcript, ".zstd") || !errors.Is(err, io.ErrUnexpectedEOF)) {
		return nil, fmt.Errorf("failed reading deepseek harness transcript %q: %s", transcript, err)
	}

	g.commitPendingUsage(session, &state)

	return state.heartbeats, nil
}

func dshOpenTranscript(transcript string) (dshTranscriptReader, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return dshTranscriptReader{}, fmt.Errorf("failed to open deepseek harness transcript %q: %s", transcript, err)
	}

	if !strings.HasSuffix(transcript, ".zstd") {
		return dshTranscriptReader{closer: fh, reader: fh}, nil
	}

	decoder, err := zstd.NewReader(fh)
	if err != nil {
		fh.Close() // nolint:errcheck,gosec

		return dshTranscriptReader{}, fmt.Errorf(
			"failed to create zstd reader for deepseek harness transcript %q: %s", transcript, err,
		)
	}

	return dshTranscriptReader{
		closer: dshCombinedCloser{decoder: decoder, file: fh},
		reader: decoder,
	}, nil
}

type dshCombinedCloser struct {
	decoder *zstd.Decoder
	file    *os.File
}

func (c dshCombinedCloser) Close() error {
	c.decoder.Close()

	return c.file.Close()
}

func (DeepSeek) readSessionHeader(
	logger *log.Logger,
	transcript string,
	line []byte,
	session *dshSessionState,
) {
	if len(line) == 0 {
		return
	}

	var header dshSessionHeader
	if err := json.Unmarshal(line, &header); err != nil {
		logger.Warnf("failed parsing deepseek harness session header from %q: %s", transcript, err)

		return
	}

	if header.Type != "session" {
		logger.Warnf("missing deepseek harness session header in %q", transcript)

		return
	}

	if header.ID != "" {
		session.id = header.ID
	}

	if header.Cwd != "" {
		session.cwd = header.Cwd
	}

	if header.SeedLength > 0 {
		session.seedLength = header.SeedLength
	}
}

func (g DeepSeek) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	session dshSessionState,
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

	eventIndex := state.eventIndex
	state.eventIndex++

	if dshSeedEvent(logLine.Seq, eventIndex, session.seedLength) {
		return
	}

	timestamp := time.UnixMilli(logLine.Time).UTC()

	switch logLine.Type {
	case "request/header", "request/context":
		g.trackModel(logLine, state)
	case "user/message":
		g.handleUserMessage(timestamp, logLine, session, state)
	case "assistant/chunk":
		g.handleAssistantChunk(timestamp, logLine, session, state)
	case "assistant/message":
		g.handleAssistantMessage(timestamp, logLine, session, state)
	case "tool/call":
		state.handleToolCall(logLine)
	case "tool/result":
		g.handleToolResult(timestamp, logLine, session, state)
	case "compaction/summary":
		g.commitUsage(logLine.Data.Usage, timestamp, session, state)
	}
}

func dshSeedEvent(seq *int, eventIndex int, seedLength int) bool {
	if seedLength <= 0 {
		return false
	}

	if seq != nil {
		return *seq < seedLength
	}

	return eventIndex < seedLength
}

func (g DeepSeek) trackModel(logLine dshLogLine, state *dshParseState) {
	var model string

	if logLine.Type == "request/header" && logLine.Data.Header != nil && logLine.Data.Header.Config != nil {
		model = logLine.Data.Header.Config.Model
	} else {
		model = logLine.Data.Model
	}

	g.setModel(state, model)
}

func (g DeepSeek) setModel(state *dshParseState, model string) {
	model = strings.TrimSpace(model)
	if model == "" {
		return
	}

	state.model = model

	for _, i := range state.pendingPromptHeartbeats {
		if i < 0 || i >= len(state.heartbeats) {
			continue
		}

		entity := state.heartbeats[i].Entity
		state.heartbeats[i].UserAgent = aiUserAgentWithModel(
			entity,
			g.UserAgents,
			g.FallbackUserAgent,
			model,
			"",
		)
	}

	state.pendingPromptHeartbeats = nil
}

func (g DeepSeek) handleUserMessage(
	timestamp time.Time,
	logLine dshLogLine,
	session dshSessionState,
	state *dshParseState,
) {
	if logLine.Time <= 0 || !timestampAtOrAfterCutoff(timestamp, g.After) ||
		logLine.Data.Source == nil || logLine.Data.Source.Kind != "user" {
		return
	}

	var promptChars int

	for _, block := range logLine.Data.Content {
		if block.Type == "text" {
			promptChars += promptLength(block.Text)
		}
	}

	if promptChars == 0 {
		return
	}

	h := g.appHeartbeat(timestamp, session, state)
	h.AIPromptLength = promptChars
	state.appendHeartbeat(h)
	// The request header follows user/message, so update this heartbeat when that header arrives.
	state.pendingPromptHeartbeats = append(state.pendingPromptHeartbeats, len(state.heartbeats)-1)
}

func (g DeepSeek) handleAssistantChunk(
	timestamp time.Time,
	logLine dshLogLine,
	session dshSessionState,
	state *dshParseState,
) {
	chunk := logLine.Data.Chunk
	if chunk == nil {
		return
	}

	switch chunk.Type {
	case "usage":
		if chunk.Usage == nil {
			return
		}

		if state.pendingUsage != nil &&
			(state.pendingUsage.turn != logLine.Data.Turn || state.pendingUsage.step != logLine.Data.Step) {
			g.commitPendingUsage(session, state)
		}

		state.pendingUsage = &dshPendingUsage{
			turn:      logLine.Data.Turn,
			step:      logLine.Data.Step,
			timestamp: timestamp,
			usage:     *chunk.Usage,
		}
	case "finish":
		if chunk.Reason != nil && (chunk.Reason.Kind == "error" || chunk.Reason.Kind == "aborted") {
			g.commitPendingUsage(session, state)
		}
	}
}

func (g DeepSeek) handleAssistantMessage(
	timestamp time.Time,
	logLine dshLogLine,
	session dshSessionState,
	state *dshParseState,
) {
	message := logLine.Data.Message
	if message == nil {
		message = &dshMessage{Role: logLine.Data.Role, Content: logLine.Data.Content, Source: logLine.Data.Source}
	}

	modelSource := message.Source
	if modelSource == nil {
		modelSource = logLine.Data.Provenance
	}

	if modelSource != nil && modelSource.Model != "" {
		g.setModel(state, modelSource.Model)
	}

	if logLine.Time > 0 && timestampAtOrAfterCutoff(timestamp, g.After) {
		if h := g.assistantHeartbeat(timestamp, session, state, message); h != nil {
			state.appendHeartbeat(*h)
		}
	}

	usage := logLine.Data.Usage
	usageTimestamp := timestamp

	if state.pendingUsage != nil && state.pendingUsage.turn == logLine.Data.Turn &&
		state.pendingUsage.step == logLine.Data.Step {
		if usage == nil {
			usage = &state.pendingUsage.usage
			usageTimestamp = state.pendingUsage.timestamp
		}

		state.pendingUsage = nil
	}

	g.commitUsage(usage, usageTimestamp, session, state)
}

func (s *dshParseState) handleToolCall(logLine dshLogLine) {
	if logLine.Data.CallID == "" || logLine.Data.Name == "" {
		return
	}

	s.pendingTools[logLine.Data.CallID] = dshToolCall{
		arguments: dshToolArguments(logLine.Data.Arguments),
		name:      logLine.Data.Name,
	}
}

func (g DeepSeek) handleToolResult(
	timestamp time.Time,
	logLine dshLogLine,
	session dshSessionState,
	state *dshParseState,
) {
	callID, failed := dshToolResult(logLine.Data)
	if callID == "" {
		return
	}

	call, ok := state.pendingTools[callID]
	if !ok {
		return
	}

	delete(state.pendingTools, callID)

	if failed || logLine.Time <= 0 || !timestampAtOrAfterCutoff(timestamp, g.After) {
		return
	}

	filePath, isWrite, lineChanges := dshToolHeartbeatInfo(call, session.cwd)
	if filePath == "" || dshDirectoryViewResult(call, logLine.Data, filePath) {
		return
	}

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		session.id,
		state.tokens,
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
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, state.model, ""),
	)
	state.appendHeartbeat(h)
}

func dshToolResult(data dshEventData) (string, bool) {
	failed := data.Error != nil || data.IsError
	callID := data.CallID

	if data.Message == nil {
		return callID, failed
	}

	if callID == "" && data.Message.Source != nil {
		callID = data.Message.Source.CallID
	}

	for _, block := range data.Message.Content {
		if block.Type != "tool-result" {
			continue
		}

		if callID == "" {
			callID = block.ToolCallID
		}

		failed = failed || block.IsError
	}

	return callID, failed
}

func dshDirectoryViewResult(call dshToolCall, data dshEventData, filePath string) bool {
	if !strings.EqualFold(strings.TrimSpace(call.name), "str_replace_editor") ||
		!strings.EqualFold(strings.TrimSpace(piStringValue(call.arguments, "command")), "view") {
		return false
	}

	if dshContentHasDirectoryListing(data.Content) ||
		(data.Message != nil && dshContentHasDirectoryListing(data.Message.Content)) {
		return true
	}

	info, err := os.Stat(filepath.FromSlash(filePath))

	return err == nil && info.IsDir()
}

func dshContentHasDirectoryListing(blocks []dshContentBlock) bool {
	for _, block := range blocks {
		if block.Type == "text" && strings.HasPrefix(
			strings.TrimSpace(block.Text),
			"Here're the files and directories up to 2 levels deep in ",
		) {
			return true
		}

		if dshContentHasDirectoryListing(block.Content) {
			return true
		}
	}

	return false
}

func (g DeepSeek) assistantHeartbeat(
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
	message *dshMessage,
) *heartbeat.Heartbeat {
	if message == nil || message.Role != "assistant" {
		return nil
	}

	var lineChanges int

	for _, block := range message.Content {
		if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
			lineChanges += countStringLines(block.Text)
		}
	}

	if lineChanges == 0 {
		return nil
	}

	h := g.appHeartbeat(timestamp, session, state)

	return &h
}

func (g DeepSeek) appHeartbeat(
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
) heartbeat.Heartbeat {
	entity := g.Name()

	return heartbeat.NewWithAITokens(
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
}

func (s *dshParseState) appendHeartbeat(h heartbeat.Heartbeat) {
	s.heartbeats = append(s.heartbeats, h)
	s.advanceTokens()
}

func (g DeepSeek) commitPendingUsage(session dshSessionState, state *dshParseState) {
	if state.pendingUsage == nil {
		return
	}

	pending := state.pendingUsage
	state.pendingUsage = nil
	g.commitUsage(&pending.usage, pending.timestamp, session, state)
}

func (g DeepSeek) commitUsage(
	usage *dshUsage,
	timestamp time.Time,
	session dshSessionState,
	state *dshParseState,
) {
	if usage != nil && !timestamp.IsZero() && timestampAtOrAfterCutoff(timestamp, g.After) &&
		len(state.heartbeats) == 0 && dshUsageHasTokens(*usage) {
		state.appendHeartbeat(g.appHeartbeat(timestamp, session, state))
	}

	state.commitUsage(usage, timestamp, g.After)
}

func (s *dshParseState) commitUsage(usage *dshUsage, timestamp time.Time, after time.Time) {
	if usage == nil {
		return
	}

	s.tokens.CurrentInput += dshTokenValue(usage.InputTokens) + dshTokenValue(usage.CacheWriteTokens)
	s.tokens.CurrentCachedInput += dshTokenValue(usage.CacheReadTokens)
	s.tokens.CurrentOutput += dshTokenValue(usage.OutputTokens)

	if timestamp.IsZero() || !timestampAtOrAfterCutoff(timestamp, after) {
		s.advanceTokens()

		return
	}

	s.attachTokensToLastHeartbeat()
}

func dshTokenValue(value *int64) int64 {
	if value == nil || *value < 0 {
		return 0
	}

	return *value
}

func dshUsageHasTokens(usage dshUsage) bool {
	return dshTokenValue(usage.InputTokens) > 0 || dshTokenValue(usage.OutputTokens) > 0 ||
		dshTokenValue(usage.CacheReadTokens) > 0 || dshTokenValue(usage.CacheWriteTokens) > 0
}

func (s *dshParseState) attachTokensToLastHeartbeat() {
	input := max(s.tokens.CurrentInput-s.tokens.LastInput, 0)
	cachedInput := max(s.tokens.CurrentCachedInput-s.tokens.LastCachedInput, 0)
	output := max(s.tokens.CurrentOutput-s.tokens.LastOutput, 0)

	if len(s.heartbeats) > 0 {
		i := len(s.heartbeats) - 1
		s.heartbeats[i].AIInputTokens += input
		s.heartbeats[i].AICachedInputTokens += cachedInput
		s.heartbeats[i].AIOutputTokens += output
	}

	s.advanceTokens()
}

func (s *dshParseState) advanceTokens() {
	s.tokens.LastInput = s.tokens.CurrentInput
	s.tokens.LastCachedInput = s.tokens.CurrentCachedInput
	s.tokens.LastOutput = s.tokens.CurrentOutput
}

func dshToolArguments(raw json.RawMessage) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}

	var arguments map[string]interface{}
	if err := json.Unmarshal(raw, &arguments); err == nil {
		return arguments
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil {
		return nil
	}

	if err := json.Unmarshal([]byte(encoded), &arguments); err != nil {
		return nil
	}

	return arguments
}

func dshToolHeartbeatInfo(call dshToolCall, cwd string) (string, bool, int) {
	name := strings.ToLower(strings.TrimSpace(call.name))

	switch name {
	case "read", "read_image":
		return dshToolPath(call.arguments, cwd, "file_path"), false, 0
	case "write":
		return dshToolPath(call.arguments, cwd, "file_path"), true,
			dshTextLineCount(piStringValue(call.arguments, "content"))
	case "edit":
		return dshToolPath(call.arguments, cwd, "file_path"), true,
			dshReplacementLineChanges(call.arguments, "old_string", "new_string")
	case "str_replace_editor":
		return dshStrReplaceEditorInfo(call.arguments, cwd)
	default:
		return "", false, 0
	}
}

func dshStrReplaceEditorInfo(arguments map[string]interface{}, cwd string) (string, bool, int) {
	path := dshToolPath(arguments, cwd, "path")

	switch strings.ToLower(piStringValue(arguments, "command")) {
	case "view":
		return path, false, 0
	case "create":
		return path, true, dshTextLineCount(piStringValue(arguments, "file_text"))
	case "insert":
		return path, true, dshTextLineCount(piStringValue(arguments, "new_str"))
	case "str_replace":
		return path, true, dshReplacementLineChanges(arguments, "old_str", "new_str")
	default:
		return "", false, 0
	}
}

func dshToolPath(arguments map[string]interface{}, cwd string, key string) string {
	path := strings.TrimSpace(piStringValue(arguments, key))
	if path == "" {
		return ""
	}

	if !filepath.IsAbs(path) && cwd != "" {
		path = filepath.Join(cwd, path)
	}

	return filepath.ToSlash(filepath.Clean(path))
}

func dshReplacementLineChanges(arguments map[string]interface{}, oldKey string, newKey string) int {
	return dshTextLineCount(piStringValue(arguments, newKey)) -
		dshTextLineCount(piStringValue(arguments, oldKey))
}

func dshTextLineCount(text string) int {
	if text == "" {
		return 0
	}

	return countStringLines(text)
}

// Name returns the parser name.
func (DeepSeek) Name() string { return "DeepSeek Harness" }
