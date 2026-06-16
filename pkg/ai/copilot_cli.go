package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	pathutil "path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

var copilotCLIPlanFilePattern = regexp.MustCompile(`(?:^|/)\.copilot/session-state/[^/]+/plan\.md$`)

type (
	copilotCLIEvent struct {
		Type      string          `json:"type"`
		Data      json.RawMessage `json:"data"`
		Timestamp time.Time       `json:"timestamp"`
	}

	copilotCLIParseState struct {
		sessionID        string
		sessionEntity    string
		cwd              string
		gitRoot          string
		cliVersion       string
		agentVersion     string
		model            string
		tokens           heartbeat.AITokens
		toolNames        map[string]string
		fileHeartbeatIDs map[string]struct{}
		heartbeats       []copilotTimedHeartbeat
	}

	copilotCLIStartData struct {
		SessionID      string    `json:"sessionId"`
		CopilotVersion string    `json:"copilotVersion"`
		StartTime      time.Time `json:"startTime"`
		Context        struct {
			Cwd     string `json:"cwd"`
			GitRoot string `json:"gitRoot"`
		} `json:"context"`
	}

	copilotCLIModelChangeData struct {
		NewModel string `json:"newModel"`
	}

	copilotCLIUserMessageData struct {
		Content string `json:"content"`
		Message string `json:"message"`
	}

	copilotCLIAssistantMessageData struct {
		Model        string `json:"model"`
		OutputTokens int64  `json:"outputTokens"`
	}

	copilotCLIToolExecutionStartData struct {
		ToolCallID string `json:"toolCallId"`
		ToolName   string `json:"toolName"`
		Name       string `json:"name"`
	}

	copilotCLIToolExecutionCompleteData struct {
		ToolCallID      string                   `json:"toolCallId"`
		ToolName        string                   `json:"toolName"`
		Success         *bool                    `json:"success"`
		ToolTelemetry   *copilotCLIToolTelemetry `json:"toolTelemetry"`
		ResultTelemetry *copilotCLIToolTelemetry `json:"toolResultTelemetry"`
		Result          *copilotCLIToolResult    `json:"result"`
	}

	copilotCLIToolResult struct {
		ToolTelemetry *copilotCLIToolTelemetry `json:"toolTelemetry"`
	}

	copilotCLIToolTelemetry struct {
		Properties           copilotCLIToolProperties            `json:"properties"`
		RestrictedProperties *copilotCLIToolRestrictedProperties `json:"restrictedProperties"`
		Metrics              copilotCLIToolMetrics               `json:"metrics"`
	}

	copilotCLIToolProperties struct {
		CodeBlocks json.RawMessage `json:"codeBlocks"`
	}

	copilotCLIToolRestrictedProperties struct {
		FilePaths    json.RawMessage `json:"filePaths"`
		AddedPaths   json.RawMessage `json:"addedPaths"`
		DeletedPaths json.RawMessage `json:"deletedPaths"`
	}

	copilotCLIToolMetrics struct {
		LinesAdded   *int `json:"linesAdded"`
		LinesRemoved *int `json:"linesRemoved"`
	}

	copilotCLICodeBlock struct {
		LinesAdded   int `json:"linesAdded"`
		LinesRemoved int `json:"linesRemoved"`
	}

	copilotCLIShutdownData struct {
		CurrentModel string `json:"currentModel"`
		TokenDetails struct {
			Input struct {
				TokenCount *int64 `json:"tokenCount"`
			} `json:"input"`
			Output struct {
				TokenCount *int64 `json:"tokenCount"`
			} `json:"output"`
		} `json:"tokenDetails"`
		CodeChanges *struct {
			LinesAdded    int      `json:"linesAdded"`
			LinesRemoved  int      `json:"linesRemoved"`
			FilesModified []string `json:"filesModified"`
		} `json:"codeChanges"`
	}
)

func (g Copilot) cliHeartbeats(ctx context.Context) ([]copilotTimedHeartbeat, error) {
	transcripts, err := g.cliTranscriptPaths(ctx)
	if err != nil {
		return nil, err
	}

	if len(transcripts) == 0 {
		return nil, nil
	}

	logger := log.Extract(ctx)
	logger.Debugf("Found %d Copilot CLI transcript logs modified after %s", len(transcripts), g.After)

	var timed []copilotTimedHeartbeat

	for _, transcript := range transcripts {
		parsed, err := g.parseCLITranscript(ctx, transcript)
		if err != nil {
			logger.Warnf("failed parsing copilot cli transcript %q: %s", transcript, err)
			continue
		}

		timed = append(timed, parsed...)
	}

	return timed, nil
}

func (g Copilot) cliTranscriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	sessionsDir := filepath.Join(home, ".copilot", "session-state")

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed reading copilot cli session state directory %q: %s", sessionsDir, err)
	}

	var transcripts []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		path := filepath.Join(sessionsDir, entry.Name(), "events.jsonl")

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed statting copilot cli transcript %q: %s", path, err)
		}

		if info.ModTime().Before(g.After) {
			continue
		}

		transcripts = append(transcripts, path)
	}

	sort.Strings(transcripts)

	return transcripts, nil
}

func (g Copilot) parseCLITranscript(ctx context.Context, transcript string) ([]copilotTimedHeartbeat, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed opening copilot cli transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	state := copilotCLIParseState{
		sessionID:        filepath.Base(filepath.Dir(transcript)),
		toolNames:        make(map[string]string),
		fileHeartbeatIDs: make(map[string]struct{}),
	}
	state.sessionEntity = appHeartbeatEntity(g.Name(), state.sessionID)

	logger := log.Extract(ctx)

	if err := g.readCLISessionState(logger, fh, transcript, &state); err != nil {
		return nil, err
	}

	scanner, err := copilotCLIScanner(fh, transcript)
	if err != nil {
		return nil, err
	}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleCLITranscriptLine(logger, transcript, scanner.Bytes(), &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading copilot cli transcript %q: %s", transcript, err)
	}

	return state.heartbeats, nil
}

func (g Copilot) readCLISessionState(
	logger *log.Logger,
	fh *os.File,
	transcript string,
	state *copilotCLIParseState,
) error {
	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return fmt.Errorf("failed to read copilot cli transcript %q: %s", transcript, err)
	}

	if len(firstLine) > 0 {
		var event copilotCLIEvent
		if err := json.Unmarshal(firstLine, &event); err != nil {
			logger.Debugf("failed parsing copilot cli session metadata from %q: %s", transcript, err)
		} else if event.Type == "session.start" {
			g.handleCLIStart(event, state)
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to rewind copilot cli transcript %q: %s", transcript, err)
	}

	return nil
}

func copilotCLIScanner(fh *os.File, transcript string) (*bufio.Scanner, error) {
	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat copilot cli transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek copilot cli transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read copilot cli transcript %q: %s", transcript, err)
		}
	}

	return scanner, nil
}

func (g Copilot) handleCLITranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	state *copilotCLIParseState,
) {
	if len(line) == 0 {
		return
	}

	var event copilotCLIEvent
	if err := json.Unmarshal(line, &event); err != nil {
		logger.Warnf("failed parsing copilot cli transcript line from %q: %s", transcript, err)
		logger.Debugf("failed parsing copilot cli transcript line: %s", line)

		return
	}

	switch event.Type {
	case "session.start":
		g.handleCLIStart(event, state)
	case "session.model_change":
		g.handleCLIModelChange(event, state)
	case "user.message":
		g.handleCLIUserMessage(event, state)
	case "assistant.message":
		g.handleCLIAssistantMessage(event, state)
	case "tool.execution_start":
		g.handleCLIToolExecutionStart(event, state)
	case "tool.execution_complete":
		g.handleCLIToolExecutionComplete(event, state)
	case "session.shutdown":
		g.handleCLIShutdown(event, state)
	}
}

func (g Copilot) handleCLIStart(event copilotCLIEvent, state *copilotCLIParseState) {
	var data copilotCLIStartData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	if data.SessionID != "" {
		state.sessionID = data.SessionID
		state.sessionEntity = appHeartbeatEntity(g.Name(), data.SessionID)
	}

	state.cliVersion = firstNonEmptyString(data.CopilotVersion, state.cliVersion)
	state.agentVersion = firstNonEmptyString(data.CopilotVersion, state.agentVersion)
	state.cwd = firstNonEmptyString(data.Context.Cwd, state.cwd)
	state.gitRoot = firstNonEmptyString(data.Context.GitRoot, state.gitRoot)
}

func (Copilot) handleCLIModelChange(event copilotCLIEvent, state *copilotCLIParseState) {
	var data copilotCLIModelChangeData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	if data.NewModel != "" {
		state.model = data.NewModel
	}
}

func (g Copilot) handleCLIUserMessage(event copilotCLIEvent, state *copilotCLIParseState) {
	if event.Timestamp.IsZero() || event.Timestamp.Before(g.After) {
		return
	}

	var data copilotCLIUserMessageData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	message := firstNonEmptyString(data.Content, data.Message)
	if strings.TrimSpace(message) == "" {
		return
	}

	state.heartbeats = append(state.heartbeats, copilotTimedHeartbeat{
		timestamp: event.Timestamp,
		heartbeat: g.cliAppHeartbeat(
			state.sessionEntity,
			state.sessionID,
			nil,
			promptLength(message),
			event.Timestamp,
			state.projectPath(),
			state.model,
			state.cliVersion,
			state.agentVersion,
		),
	})
}

func (g Copilot) handleCLIAssistantMessage(event copilotCLIEvent, state *copilotCLIParseState) {
	var data copilotCLIAssistantMessageData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	if data.Model != "" {
		state.model = data.Model
	}

	assignTokens := data.OutputTokens > 0
	if assignTokens {
		state.tokens.CurrentOutput += data.OutputTokens
	}

	if event.Timestamp.IsZero() || event.Timestamp.Before(g.After) {
		if assignTokens {
			state.tokens = g.advanceTokens(state.tokens)
		}

		return
	}

	state.heartbeats = append(state.heartbeats, copilotTimedHeartbeat{
		timestamp: event.Timestamp,
		heartbeat: g.cliAppHeartbeat(
			state.sessionEntity,
			state.sessionID,
			g.tokensForFirstHeartbeat(assignTokens, state.tokens),
			0,
			event.Timestamp,
			state.projectPath(),
			state.model,
			state.cliVersion,
			state.agentVersion,
		),
	})

	if assignTokens {
		state.tokens = g.advanceTokens(state.tokens)
	}
}

func (Copilot) handleCLIToolExecutionStart(event copilotCLIEvent, state *copilotCLIParseState) {
	var data copilotCLIToolExecutionStartData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	toolName := firstNonEmptyString(data.ToolName, data.Name)
	if data.ToolCallID != "" && toolName != "" {
		state.toolNames[data.ToolCallID] = toolName
	}
}

func (g Copilot) handleCLIToolExecutionComplete(event copilotCLIEvent, state *copilotCLIParseState) {
	if event.Timestamp.IsZero() {
		return
	}

	var data copilotCLIToolExecutionCompleteData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	if data.Success != nil && !*data.Success {
		return
	}

	toolName := firstNonEmptyString(data.ToolName, state.toolNames[data.ToolCallID])

	telemetry := data.telemetry()
	if telemetry == nil || !telemetry.hasWriteSignals(toolName) {
		return
	}

	paths := telemetry.filePaths()
	if len(paths) == 0 {
		return
	}

	lineChanges := telemetry.lineChanges(len(paths))
	for i, path := range paths {
		if !g.shouldTrackCLIPath(path) {
			continue
		}

		if event.Timestamp.Before(g.After) {
			state.trackCLIFilePath(path)
			continue
		}

		timestamp := event.Timestamp.Add(time.Duration(i) * time.Millisecond)
		lineChange := lineChanges.value(i)
		state.appendFileHeartbeat(g, path, timestamp, lineChange)
	}
}

func (g Copilot) handleCLIShutdown(event copilotCLIEvent, state *copilotCLIParseState) {
	var data copilotCLIShutdownData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return
	}

	if data.CurrentModel != "" {
		state.model = data.CurrentModel
	}

	assignTokens := false

	if data.TokenDetails.Input.TokenCount != nil {
		state.tokens.CurrentInput = *data.TokenDetails.Input.TokenCount
		assignTokens = true
	}

	if data.TokenDetails.Output.TokenCount != nil {
		state.tokens.CurrentOutput = *data.TokenDetails.Output.TokenCount
		assignTokens = true
	}

	if event.Timestamp.IsZero() || event.Timestamp.Before(g.After) {
		if assignTokens {
			state.tokens = g.advanceTokens(state.tokens)
		}

		return
	}

	hasTokenDelta := assignTokens && g.hasTokenDelta(state.tokens)

	hasCodeChanges := data.CodeChanges != nil && len(data.CodeChanges.FilesModified) > 0
	if hasTokenDelta || hasCodeChanges || len(state.heartbeats) > 0 {
		state.heartbeats = append(state.heartbeats, copilotTimedHeartbeat{
			timestamp: event.Timestamp,
			heartbeat: g.cliAppHeartbeat(
				state.sessionEntity,
				state.sessionID,
				g.tokensForFirstHeartbeat(hasTokenDelta, state.tokens),
				0,
				event.Timestamp,
				state.projectPath(),
				state.model,
				state.cliVersion,
				state.agentVersion,
			),
		})
	}

	if assignTokens {
		state.tokens = g.advanceTokens(state.tokens)
	}

	if data.CodeChanges == nil {
		return
	}

	paths := data.CodeChanges.FilesModified

	lineChanges := copilotCLIOptionalLineChanges{}
	if len(paths) == 1 {
		lineChanges = copilotCLIOptionalLineChanges{
			values: []*int{heartbeat.PointerTo(data.CodeChanges.LinesAdded - data.CodeChanges.LinesRemoved)},
		}
	}

	for i, path := range paths {
		if !g.shouldTrackCLIPath(path) {
			continue
		}

		if _, found := state.fileHeartbeatIDs[copilotCLIPathID(path)]; found {
			continue
		}

		timestamp := event.Timestamp.Add(time.Duration(i+1) * time.Millisecond)
		state.appendFileHeartbeat(g, path, timestamp, lineChanges.value(i))
	}
}

func (data copilotCLIToolExecutionCompleteData) telemetry() *copilotCLIToolTelemetry {
	if data.ToolTelemetry != nil {
		return data.ToolTelemetry
	}

	if data.ResultTelemetry != nil {
		return data.ResultTelemetry
	}

	if data.Result != nil {
		return data.Result.ToolTelemetry
	}

	return nil
}

func (state copilotCLIParseState) projectPath() string {
	return firstNonEmptyString(state.gitRoot, state.cwd)
}

func (state *copilotCLIParseState) appendFileHeartbeat(
	g Copilot,
	path string,
	timestamp time.Time,
	lineChanges *int,
) {
	state.trackCLIFilePath(path)
	state.heartbeats = append(state.heartbeats, copilotTimedHeartbeat{
		timestamp: timestamp,
		heartbeat: g.cliFileHeartbeat(
			path,
			state.sessionID,
			nil,
			timestamp,
			state.model,
			state.cliVersion,
			state.agentVersion,
			true,
			lineChanges,
		),
	})
}

func (state *copilotCLIParseState) trackCLIFilePath(path string) {
	id := copilotCLIPathID(path)
	if id == "" {
		return
	}

	state.fileHeartbeatIDs[id] = struct{}{}
}

func (t copilotCLIToolTelemetry) hasWriteSignals(toolName string) bool {
	switch toolName {
	case "apply_patch", "edit", "write":
		return true
	}

	if t.Metrics.LinesAdded != nil || t.Metrics.LinesRemoved != nil {
		return true
	}

	if t.RestrictedProperties == nil {
		return false
	}

	return len(decodeCopilotCLIStringArray(t.RestrictedProperties.AddedPaths)) > 0 ||
		len(decodeCopilotCLIStringArray(t.RestrictedProperties.DeletedPaths)) > 0
}

func (t copilotCLIToolTelemetry) filePaths() []string {
	if t.RestrictedProperties == nil {
		return nil
	}

	paths := decodeCopilotCLIStringArray(t.RestrictedProperties.FilePaths)
	if len(paths) > 0 {
		return uniqueNonEmptyStrings(paths)
	}

	paths = append(paths, decodeCopilotCLIStringArray(t.RestrictedProperties.AddedPaths)...)
	paths = append(paths, decodeCopilotCLIStringArray(t.RestrictedProperties.DeletedPaths)...)

	return uniqueNonEmptyStrings(paths)
}

func (t copilotCLIToolTelemetry) lineChanges(pathCount int) copilotCLIOptionalLineChanges {
	blocks := decodeCopilotCLICodeBlocks(t.Properties.CodeBlocks)
	if len(blocks) == pathCount {
		values := make([]*int, 0, len(blocks))
		for _, block := range blocks {
			values = append(values, heartbeat.PointerTo(block.LinesAdded-block.LinesRemoved))
		}

		return copilotCLIOptionalLineChanges{values: values}
	}

	if pathCount == 1 {
		if value, ok := t.Metrics.lineChanges(); ok {
			return copilotCLIOptionalLineChanges{values: []*int{heartbeat.PointerTo(value)}}
		}
	}

	return copilotCLIOptionalLineChanges{}
}

func (m copilotCLIToolMetrics) lineChanges() (int, bool) {
	if m.LinesAdded == nil && m.LinesRemoved == nil {
		return 0, false
	}

	var additions int
	if m.LinesAdded != nil {
		additions = *m.LinesAdded
	}

	var deletions int
	if m.LinesRemoved != nil {
		deletions = *m.LinesRemoved
	}

	return additions - deletions, true
}

type copilotCLIOptionalLineChanges struct {
	values []*int
}

func (c copilotCLIOptionalLineChanges) value(index int) *int {
	if index < 0 || index >= len(c.values) {
		return nil
	}

	return c.values[index]
}

func decodeCopilotCLIStringArray(raw json.RawMessage) []string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var values []string
	if err := json.Unmarshal(raw, &values); err == nil {
		return values
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil || encoded == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(encoded), &values); err != nil {
		return nil
	}

	return values
}

func decodeCopilotCLICodeBlocks(raw json.RawMessage) []copilotCLICodeBlock {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}

	var blocks []copilotCLICodeBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		return blocks
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err != nil || encoded == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(encoded), &blocks); err != nil {
		return nil
	}

	return blocks
}

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}

		if _, found := seen[value]; found {
			continue
		}

		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func (g Copilot) shouldTrackCLIPath(path string) bool {
	return strings.TrimSpace(path) != "" && !g.isCLIPlanFile(path)
}

func (Copilot) isCLIPlanFile(path string) bool {
	normalized := copilotCLIPathID(path)
	return copilotCLIPlanFilePattern.MatchString(normalized)
}

func copilotCLIPathID(path string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(path), "\\", "/")
	if normalized == "" {
		return ""
	}

	return pathutil.Clean(normalized)
}

func (g Copilot) cliAppHeartbeat(
	entity string,
	sessionID string,
	aiTokens *heartbeat.AITokens,
	promptLength int,
	timestamp time.Time,
	projectPath string,
	model string,
	cliVersion string,
	agentVersion string,
) heartbeat.Heartbeat {
	if aiTokens != nil {
		h := heartbeat.NewWithAITokens(
			nil,
			sessionID,
			*aiTokens,
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
			float64(timestamp.UnixMilli())/1000,
			g.cliUserAgent(entity, model, cliVersion, agentVersion),
		)

		h.AIPromptLength = promptLength

		return h
	}

	h := heartbeat.New(
		nil,
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
		float64(timestamp.UnixMilli())/1000,
		g.cliUserAgent(entity, model, cliVersion, agentVersion),
	)
	h.AISession = sessionID
	h.AIPromptLength = promptLength

	return h
}

func (g Copilot) cliFileHeartbeat(
	path string,
	sessionID string,
	aiTokens *heartbeat.AITokens,
	timestamp time.Time,
	model string,
	cliVersion string,
	agentVersion string,
	isWrite bool,
	lineChanges *int,
) heartbeat.Heartbeat {
	if aiTokens != nil {
		return heartbeat.NewWithAITokens(
			lineChanges,
			sessionID,
			*aiTokens,
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			path,
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
			g.cliUserAgent(path, model, cliVersion, agentVersion),
		)
	}

	h := heartbeat.New(
		lineChanges,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		path,
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
		g.cliUserAgent(path, model, cliVersion, agentVersion),
	)
	h.AISession = sessionID

	return h
}

func (g Copilot) cliUserAgent(entity string, model string, cliVersion string, agentVersion string) string {
	parser := aiPlugin(g, "")

	existing := g.FallbackUserAgent
	if fromHeartbeat, found := g.UserAgents[entity]; found && fromHeartbeat != "" {
		existing = fromHeartbeat
	}

	existing = userAgentWithPrependedEditor(existing, copilotCLIHarnessUserAgent(cliVersion, agentVersion))

	return userAgentWithPrependedAgent(
		userAgentWithPrependedParser(existing, parser),
		aiAgentUserAgentToken(model),
	)
}

func copilotCLIHarnessUserAgent(cliVersion string, agentVersion string) string {
	if cliVersion == "" && agentVersion == "" {
		return ""
	}

	if cliVersion == "" {
		cliVersion = "unknown"
	}

	if agentVersion == "" {
		agentVersion = "unknown"
	}

	return "github-copilot-cli/" + cliVersion + " copilot/" + agentVersion
}
