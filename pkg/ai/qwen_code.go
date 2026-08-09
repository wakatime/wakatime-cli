package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// QwenCode contains params for detecting heartbeats from Qwen Code session JSONL logs.
type QwenCode ParserConfig

type (
	qwenCodeParseState struct {
		cwd        string
		heartbeats Heartbeats
		sessionID  string
		model      string
		tokens     heartbeat.AITokens
		toolCalls  map[string]qwenCodeToolCall
		toolQueue  []qwenCodeToolCall
		version    string
	}

	qwenCodeRecord struct {
		CWD            string                  `json:"cwd"`
		ForkedFrom     *struct{}               `json:"forkedFrom"`
		Message        qwenCodeMessage         `json:"message"`
		Model          string                  `json:"model"`
		SessionID      string                  `json:"sessionId"`
		Subtype        string                  `json:"subtype"`
		Timestamp      time.Time               `json:"timestamp"`
		ToolCallResult *qwenCodeToolCallResult `json:"toolCallResult"`
		Type           string                  `json:"type"`
		UsageMetadata  *qwenCodeUsageMetadata  `json:"usageMetadata"`
		Version        string                  `json:"version"`
	}

	qwenCodeMessage struct {
		Parts []qwenCodePart `json:"parts"`
		Role  string         `json:"role"`
	}

	qwenCodePart struct {
		FunctionCall     *qwenCodeFunctionCall     `json:"functionCall"`
		FunctionResponse *qwenCodeFunctionResponse `json:"functionResponse"`
		Text             string                    `json:"text"`
		Thought          bool                      `json:"thought"`
	}

	qwenCodeFunctionCall struct {
		Args map[string]interface{} `json:"args"`
		ID   string                 `json:"id"`
		Name string                 `json:"name"`
	}

	qwenCodeFunctionResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	qwenCodeToolCallResult struct {
		CallID        string          `json:"callId"`
		ResultDisplay json.RawMessage `json:"resultDisplay"`
		Status        string          `json:"status"`
	}

	qwenCodeUsageMetadata struct {
		CachedContentTokenCount int64 `json:"cachedContentTokenCount"`
		CandidatesTokenCount    int64 `json:"candidatesTokenCount"`
		PromptTokenCount        int64 `json:"promptTokenCount"`
		ThoughtsTokenCount      int64 `json:"thoughtsTokenCount"`
		ToolUsePromptTokenCount int64 `json:"toolUsePromptTokenCount"`
	}

	qwenCodeToolCall struct {
		Args map[string]interface{}
		ID   string
		Name string
	}

	qwenCodeResultDisplay struct {
		FileName string `json:"fileName"`
		DiffStat *struct {
			ModelAddedLines   int `json:"model_added_lines"`
			ModelRemovedLines int `json:"model_removed_lines"`
		} `json:"diffStat"`
	}

	qwenCodeSettings struct {
		Advanced *struct {
			RuntimeOutputDir string `json:"runtimeOutputDir"`
		} `json:"advanced"`
	}
)

// Parse parses the Qwen Code JSONL session logs for ai heartbeats.
func (g QwenCode) Parse(ctx context.Context) (Heartbeats, error) {
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
			logger.Warnf("failed parsing qwen code transcript %q: %s", transcript, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g QwenCode) transcriptPaths(ctx context.Context) ([]string, error) {
	runtimeDir, err := qwenCodeRuntimeDir(ctx)
	if err != nil {
		return nil, err
	}

	projectsDir := filepath.Join(runtimeDir, "projects")
	if _, err := os.Stat(projectsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat qwen code projects directory: %s", err)
	}

	var transcripts []string

	err = filepath.WalkDir(projectsDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if entry.IsDir() || filepath.Ext(entry.Name()) != ".jsonl" || filepath.Base(filepath.Dir(path)) != "chats" {
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
		return nil, fmt.Errorf("failed to walk qwen code projects directory: %s", err)
	}

	return transcripts, nil
}

func qwenCodeRuntimeDir(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	if runtimeDir := strings.TrimSpace(os.Getenv("QWEN_RUNTIME_DIR")); runtimeDir != "" {
		return qwenCodeResolveDir(runtimeDir, home)
	}

	qwenHome := filepath.Join(home, ".qwen")
	if configuredHome := strings.TrimSpace(os.Getenv("QWEN_HOME")); configuredHome != "" {
		qwenHome, err = qwenCodeResolveDir(configuredHome, home)
		if err != nil {
			return "", err
		}
	}

	// Relative runtimeOutputDir settings are project-specific and cannot be resolved without the originating cwd.
	contents, err := os.ReadFile(filepath.Clean(filepath.Join(qwenHome, "settings.json")))
	if err == nil {
		var settings qwenCodeSettings
		if json.Unmarshal(qwenCodeStripJSONComments(contents), &settings) == nil && settings.Advanced != nil {
			if runtimeDir, ok := qwenCodeConfiguredRuntimeDir(settings.Advanced.RuntimeOutputDir, home); ok {
				return runtimeDir, nil
			}
		}
	}

	return qwenHome, nil
}

func qwenCodeResolveDir(path string, home string) (string, error) {
	path = strings.TrimSpace(path)
	switch {
	case path == "~":
		path = home
	case strings.HasPrefix(path, "~/"), strings.HasPrefix(path, `~\`):
		path = filepath.Join(home, path[2:])
	}

	return filepath.Abs(path)
}

func qwenCodeConfiguredRuntimeDir(path string, home string) (string, bool) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", false
	}

	if path == "~" || strings.HasPrefix(path, "~/") || strings.HasPrefix(path, `~\`) {
		resolved, err := qwenCodeResolveDir(path, home)

		return resolved, err == nil
	}

	if filepath.IsAbs(path) {
		return filepath.Clean(path), true
	}

	return "", false
}

func qwenCodeStripJSONComments(contents []byte) []byte {
	stripped := append([]byte(nil), contents...)
	inString := false
	escaped := false

	for i := 0; i < len(stripped); i++ {
		switch {
		case inString:
			switch {
			case escaped:
				escaped = false
			case stripped[i] == '\\':
				escaped = true
			case stripped[i] == '"':
				inString = false
			}
		case stripped[i] == '"':
			inString = true
		case stripped[i] == '/' && i+1 < len(stripped) && stripped[i+1] == '/':
			for ; i < len(stripped) && stripped[i] != '\n' && stripped[i] != '\r'; i++ {
				stripped[i] = ' '
			}

			i--
		case stripped[i] == '/' && i+1 < len(stripped) && stripped[i+1] == '*':
			stripped[i] = ' '
			stripped[i+1] = ' '
			i += 2

			for ; i < len(stripped); i++ {
				if stripped[i] == '*' && i+1 < len(stripped) && stripped[i+1] == '/' {
					stripped[i] = ' '
					stripped[i+1] = ' '
					i++

					break
				}

				if stripped[i] != '\n' && stripped[i] != '\r' {
					stripped[i] = ' '
				}
			}
		}
	}

	return stripped
}

func (call *qwenCodeFunctionCall) UnmarshalJSON(contents []byte) error {
	var fields struct {
		Args json.RawMessage `json:"args"`
		ID   string          `json:"id"`
		Name string          `json:"name"`
	}
	if err := json.Unmarshal(contents, &fields); err != nil {
		return err
	}

	call.Args = qwenCodeFunctionCallArgs(fields.Args)
	call.ID = fields.ID
	call.Name = fields.Name

	return nil
}

func qwenCodeFunctionCallArgs(contents json.RawMessage) map[string]interface{} {
	var args map[string]interface{}
	if json.Unmarshal(contents, &args) == nil {
		return args
	}

	var encoded string
	if json.Unmarshal(contents, &encoded) == nil {
		_ = json.Unmarshal([]byte(encoded), &args)
	}

	return args
}

func (g QwenCode) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open qwen code transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	state := qwenCodeParseState{
		sessionID: strings.TrimSuffix(filepath.Base(transcript), filepath.Ext(transcript)),
		toolCalls: make(map[string]qwenCodeToolCall),
	}

	scanner, err := qwenCodeScanner(fh, transcript)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleTranscriptLine(logger, transcript, scanner.Bytes(), &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading qwen code transcript %q: %s", transcript, err)
	}

	return state.heartbeats, nil
}

func qwenCodeScanner(fh *os.File, transcript string) (*bufio.Scanner, error) {
	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat qwen code transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek qwen code transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read qwen code transcript %q: %s", transcript, err)
		}
	}

	return scanner, nil
}

func (g QwenCode) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	state *qwenCodeParseState,
) {
	if len(line) == 0 {
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(line))

	for {
		var record qwenCodeRecord
		if err := decoder.Decode(&record); err != nil {
			if err != io.EOF {
				logger.Warnf("failed parsing qwen code transcript line from %q: %s", transcript, err)
				logger.Debugf("failed parsing qwen code transcript line: %s", line)
			}

			return
		}

		g.handleTranscriptRecord(record, state)
	}
}

func (g QwenCode) handleTranscriptRecord(record qwenCodeRecord, state *qwenCodeParseState) {
	state.cwd = firstNonEmptyString(record.CWD, state.cwd)
	state.sessionID = firstNonEmptyString(record.SessionID, state.sessionID)
	state.model = firstNonEmptyString(record.Model, state.model)
	state.version = firstNonEmptyString(record.Version, state.version)

	if record.ForkedFrom != nil {
		return
	}

	g.trackToolCalls(record.Message, state)
	state.tokens = g.tokenCounts(record, state.tokens)

	if record.Timestamp.IsZero() || !timestampAtOrAfterCutoff(record.Timestamp, g.After) {
		state.tokens.LastInput = state.tokens.CurrentInput
		state.tokens.LastCachedInput = state.tokens.CurrentCachedInput
		state.tokens.LastOutput = state.tokens.CurrentOutput

		return
	}

	heartbeats := g.getHeartbeats(record, state)
	if len(heartbeats) == 0 {
		return
	}

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastCachedInput = state.tokens.CurrentCachedInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
	state.heartbeats = append(state.heartbeats, heartbeats...)
}

func (g QwenCode) getHeartbeats(record qwenCodeRecord, state *qwenCodeParseState) Heartbeats {
	switch record.Type {
	case "user", "assistant":
		if h := g.messageHeartbeat(record, *state); h != nil {
			return Heartbeats{*h}
		}
	case "tool_result":
		if h := g.toolResultHeartbeat(record, state); h != nil {
			return Heartbeats{*h}
		}
	}

	return nil
}

func (QwenCode) trackToolCalls(message qwenCodeMessage, state *qwenCodeParseState) {
	for _, part := range message.Parts {
		if part.FunctionCall == nil || part.FunctionCall.Name == "" {
			continue
		}

		call := qwenCodeToolCall{
			Args: part.FunctionCall.Args,
			ID:   part.FunctionCall.ID,
			Name: part.FunctionCall.Name,
		}
		if call.ID != "" {
			state.toolCalls[call.ID] = call
		}

		state.toolQueue = append(state.toolQueue, call)
	}
}

func (g QwenCode) messageHeartbeat(record qwenCodeRecord, state qwenCodeParseState) *heartbeat.Heartbeat {
	if record.Type != "user" && record.Type != "assistant" {
		return nil
	}

	if record.Type == "user" && (record.Subtype == "cron" || record.Subtype == "notification") {
		return nil
	}

	var text strings.Builder

	for _, part := range record.Message.Parts {
		if part.Thought || strings.TrimSpace(part.Text) == "" {
			continue
		}

		text.WriteString(part.Text)
	}

	content := text.String()
	if strings.TrimSpace(content) == "" {
		return nil
	}

	entity := appHeartbeatEntity(g.Name(), state.sessionID)

	h := heartbeat.NewWithAITokens(
		nil,
		state.sessionID,
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
		state.cwd,
		float64(record.Timestamp.UnixMilli())/1000,
		aiUserAgentWithModelAndEditor(
			entity, g.UserAgents, g.FallbackUserAgent, state.model, "", "qwen-code-cli/"+unknownIfEmpty(state.version)),
	)
	if record.Type == "user" {
		h.AIPromptLength = promptLength(content)
	}

	return &h
}

func (g QwenCode) toolResultHeartbeat(record qwenCodeRecord, state *qwenCodeParseState) *heartbeat.Heartbeat {
	result := record.ToolCallResult
	if result == nil {
		return nil
	}

	call, found := qwenCodePopToolCall(record, state)
	if !found {
		return nil
	}

	if result.Status != "" && result.Status != "success" {
		return nil
	}

	filePath := qwenCodeToolPath(call.Args, result.ResultDisplay, state.cwd)
	if filePath == "" {
		return nil
	}

	lineChanges, isWrite, ok := qwenCodeToolLineChanges(call, result.ResultDisplay)
	if !ok {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		state.sessionID,
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
		float64(record.Timestamp.UnixMilli())/1000,
		aiUserAgentWithModelAndEditor(
			filePath, g.UserAgents, g.FallbackUserAgent, state.model, "", "qwen-code-cli/"+unknownIfEmpty(state.version)),
	)

	return &h
}

func qwenCodePopToolCall(record qwenCodeRecord, state *qwenCodeParseState) (qwenCodeToolCall, bool) {
	result := record.ToolCallResult
	if result == nil {
		return qwenCodeToolCall{}, false
	}

	callID := firstNonEmptyString(result.CallID, qwenCodeFunctionResponseID(record.Message))
	if call, found := state.toolCalls[callID]; found {
		delete(state.toolCalls, callID)

		if queued, removed := qwenCodeRemoveQueuedToolCall(state, func(queued qwenCodeToolCall) bool {
			return queued.ID == callID
		}); removed {
			return queued, true
		}

		return call, true
	}

	name := qwenCodeFunctionResponseName(record.Message)
	if name == "" {
		return qwenCodeToolCall{}, false
	}

	return qwenCodeRemoveQueuedToolCall(state, func(queued qwenCodeToolCall) bool {
		return strings.EqualFold(queued.Name, name)
	})
}

func qwenCodeRemoveQueuedToolCall(
	state *qwenCodeParseState,
	match func(qwenCodeToolCall) bool,
) (qwenCodeToolCall, bool) {
	for i, call := range state.toolQueue {
		if !match(call) {
			continue
		}

		state.toolQueue = append(state.toolQueue[:i], state.toolQueue[i+1:]...)
		if call.ID != "" {
			delete(state.toolCalls, call.ID)
		}

		return call, true
	}

	return qwenCodeToolCall{}, false
}

func qwenCodeFunctionResponseName(message qwenCodeMessage) string {
	for _, part := range message.Parts {
		if part.FunctionResponse != nil && part.FunctionResponse.Name != "" {
			return part.FunctionResponse.Name
		}
	}

	return ""
}

func qwenCodeFunctionResponseID(message qwenCodeMessage) string {
	for _, part := range message.Parts {
		if part.FunctionResponse != nil && part.FunctionResponse.ID != "" {
			return part.FunctionResponse.ID
		}
	}

	return ""
}

func (QwenCode) tokenCounts(record qwenCodeRecord, tokens heartbeat.AITokens) heartbeat.AITokens {
	if record.Type != "assistant" || record.UsageMetadata == nil {
		return tokens
	}

	cachedInput := max(record.UsageMetadata.CachedContentTokenCount, 0)

	freshInput := record.UsageMetadata.PromptTokenCount - cachedInput
	if freshInput < 0 {
		freshInput = 0
	}

	tokens.CurrentInput = tokens.LastInput + freshInput + max(record.UsageMetadata.ToolUsePromptTokenCount, 0)
	tokens.CurrentCachedInput = tokens.LastCachedInput + cachedInput
	tokens.CurrentOutput = tokens.LastOutput + max(record.UsageMetadata.CandidatesTokenCount, 0) +
		max(record.UsageMetadata.ThoughtsTokenCount, 0)

	return tokens
}

func qwenCodeToolPath(args map[string]interface{}, resultDisplay json.RawMessage, cwd string) string {
	for _, key := range []string{"file_path", "path"} {
		if value := qwenCodeStringValue(args, key); value != "" {
			return qwenCodeAbsPath(value, cwd)
		}
	}

	var display qwenCodeResultDisplay
	if len(resultDisplay) > 0 && string(resultDisplay) != "null" {
		if err := json.Unmarshal(resultDisplay, &display); err == nil && display.FileName != "" {
			return qwenCodeAbsPath(display.FileName, cwd)
		}
	}

	return ""
}

func qwenCodeToolLineChanges(call qwenCodeToolCall, resultDisplay json.RawMessage) (int, bool, bool) {
	switch strings.ToLower(call.Name) {
	case "read_file":
		return 0, false, true
	case "edit", "replace":
		if lineChanges, ok := qwenCodeDiffLineChanges(resultDisplay); ok {
			return lineChanges, true, true
		}

		return countStringLines(qwenCodeStringValue(call.Args, "new_string")) -
			countStringLines(qwenCodeStringValue(call.Args, "old_string")), true, true
	case "write_file":
		if lineChanges, ok := qwenCodeDiffLineChanges(resultDisplay); ok {
			return lineChanges, true, true
		}

		return countStringLines(qwenCodeStringValue(call.Args, "content")), true, true
	default:
		return 0, false, false
	}
}

func qwenCodeDiffLineChanges(resultDisplay json.RawMessage) (int, bool) {
	if len(resultDisplay) == 0 || string(resultDisplay) == "null" {
		return 0, false
	}

	var display qwenCodeResultDisplay
	if err := json.Unmarshal(resultDisplay, &display); err != nil || display.DiffStat == nil {
		return 0, false
	}

	return display.DiffStat.ModelAddedLines - display.DiffStat.ModelRemovedLines, true
}

func qwenCodeStringValue(values map[string]interface{}, key string) string {
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

func qwenCodeAbsPath(path string, cwd string) string {
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

// Name returns its id.
func (QwenCode) Name() string {
	return "Qwen Code"
}
