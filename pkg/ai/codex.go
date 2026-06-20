package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Codex contains params for detecting heartbeats from Codex session transcripts.
type Codex ParserConfig

type (
	codexSessionState struct {
		cwd     string
		entity  string
		id      string
		source  string
		version string
	}

	codexParseState struct {
		heartbeats                    Heartbeats
		tokens                        heartbeat.AITokens
		subscriptionPlan              string
		model                         string
		reasoningEffort               string
		lastAgentMessageTime          time.Time
		lastResponseItemAssistantTime time.Time
		lastResponseItemUserTime      time.Time
	}

	codexSessionMeta struct {
		Type    string `json:"type"`
		Payload *struct {
			ID      *string `json:"id"`
			Cwd     *string `json:"cwd"`
			Source  *string `json:"source"`
			Version *string `json:"cli_version"`
		} `json:"payload"`
	}

	codexPayload struct {
		Type              *string                     `json:"type"`
		Name              *string                     `json:"name"`
		Input             *string                     `json:"input"`
		Message           *string                     `json:"message"`
		Role              *string                     `json:"role"`
		Status            *string                     `json:"status"`
		Content           []codexContentItem          `json:"content"`
		Cwd               *string                     `json:"cwd"`
		Info              *codexPayloadTokenCountInfo `json:"info"`
		Limits            *codexPayloadRateLimits     `json:"rate_limits"`
		Model             *string                     `json:"model"`
		Effort            *string                     `json:"effort"`
		CollaborationMode *struct {
			Settings *struct {
				Model           *string `json:"model"`
				ReasoningEffort *string `json:"reasoning_effort"`
			} `json:"settings"`
		} `json:"collaboration_mode"`
	}

	codexPayloadRateLimits struct {
		PlanType *string `json:"plan_type"`
	}

	codexPayloadTokenCountInfo struct {
		TotalTokenUsage    *codexPayloadTokenCountInfoUsage `json:"total_token_usage"`
		ModelContextWindow *int                             `json:"model_context_window"`
		TotalTokens        *int                             `json:"total_tokens"`
	}

	codexPayloadTokenCountInfoUsage struct {
		InputTokens           *int `json:"input_tokens"`
		CachedInputTokens     *int `json:"cached_input_tokens"`
		OutputTokens          *int `json:"output_tokens"`
		ReasoningOutputTokens *int `json:"reasoning_output_tokens"`
		TotalTokens           *int `json:"total_tokens"`
	}

	codexContentItem struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	codexLogLine struct {
		Timestamp time.Time     `json:"timestamp"`
		Type      string        `json:"type"`
		Payload   *codexPayload `json:"payload"`
	}
)

// Parse parses the Codex JSONL session transcript logs for ai heartbeats.
func (g Codex) Parse(ctx context.Context) (Heartbeats, error) {
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
			logger.Warnf("failed parsing codex transcript %q: %s", transcript, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Codex) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	sessionsDir := filepath.Join(home, ".codex", "sessions")
	if _, err := os.Stat(sessionsDir); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat .codex sessions directory: %s", err)
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
		return nil, fmt.Errorf("failed to walk .codex sessions directory: %s", err)
	}

	return transcripts, nil
}

func (g Codex) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open codex transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	session, err := g.readSessionState(logger, fh, transcript)
	if err != nil {
		return nil, err
	}

	scanner, err := codexScanner(fh, transcript)
	if err != nil {
		return nil, err
	}

	state := codexParseState{}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleTranscriptLine(logger, transcript, scanner.Bytes(), session, &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading codex transcript %q: %s", transcript, err)
	}

	state.applySubscriptionPlan()

	return state.heartbeats, nil
}

func (g Codex) readSessionState(logger *log.Logger, fh *os.File, transcript string) (codexSessionState, error) {
	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return codexSessionState{}, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
	}

	state := codexSessionState{
		entity: appHeartbeatEntity("Codex", transcript),
		id:     g.sessionIDFromPath(transcript),
	}

	if len(firstLine) > 0 {
		var sessionMeta *codexSessionMeta
		if err := json.Unmarshal(firstLine, &sessionMeta); err != nil {
			logger.Debugf("failed parsing codex session metadata from %q: %s", transcript, err)
		} else {
			g.updateSessionInfo(&state, sessionMeta)
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return codexSessionState{}, fmt.Errorf("failed to rewind codex transcript %q: %s", transcript, err)
	}

	return state, nil
}

func codexScanner(fh *os.File, transcript string) (*bufio.Scanner, error) {
	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat codex transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek codex transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
		}
	}

	return scanner, nil
}

func (g Codex) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	session codexSessionState,
	state *codexParseState,
) {
	if len(line) == 0 {
		return
	}

	var logLine codexLogLine
	if err := json.Unmarshal(line, &logLine); err != nil {
		logger.Warnf("failed parsing codex transcript line from %q: %s", transcript, err)
		logger.Debugf("failed parsing codex transcript line: %s", line)

		return
	}

	state.tokens = g.codexTokenCounts(logLine, state.tokens, g.After)
	state.trackSubscriptionPlan(logLine)
	state.trackModel(logLine)
	state.trackUserMessage(logLine)

	if logLine.Timestamp.IsZero() ||
		logLine.Timestamp.Before(g.After) ||
		state.shouldSkipAgentMessage(logLine) ||
		state.shouldSkipAssistantMessage(logLine) ||
		state.shouldSkipUserMessage(logLine) {
		return
	}

	if logLine.Payload == nil {
		return
	}

	aiHeartbeats := g.getHeartbeats(
		logLine.Timestamp,
		session.entity,
		session.id,
		session.version,
		codexAgentVersion(state.model, state.reasoningEffort),
		session.source,
		session.cwd,
		g.UserAgents,
		g.FallbackUserAgent,
		*logLine.Payload,
		state.tokens,
	)
	if len(aiHeartbeats) == 0 {
		return
	}

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
	state.heartbeats = append(state.heartbeats, aiHeartbeats...)
	state.trackAgentMessage(logLine)
	state.trackAssistantMessage(logLine)
}

func (s *codexParseState) trackModel(logLine codexLogLine) {
	if logLine.Type != "turn_context" || logLine.Payload == nil {
		return
	}

	payload := logLine.Payload
	model := ""
	reasoningEffort := ""

	if payload.Model != nil {
		model = strings.TrimSpace(*payload.Model)
	}

	if payload.Effort != nil {
		reasoningEffort = strings.TrimSpace(*payload.Effort)
	}

	if payload.CollaborationMode != nil && payload.CollaborationMode.Settings != nil {
		settings := payload.CollaborationMode.Settings
		if model == "" && settings.Model != nil {
			model = strings.TrimSpace(*settings.Model)
		}

		if reasoningEffort == "" && settings.ReasoningEffort != nil {
			reasoningEffort = strings.TrimSpace(*settings.ReasoningEffort)
		}
	}

	if model != "" {
		s.model = model
	}

	if reasoningEffort != "" {
		s.reasoningEffort = reasoningEffort
	}
}

func codexAgentVersion(model string, reasoningEffort string) string {
	model = strings.TrimSpace(model)
	reasoningEffort = strings.TrimSpace(reasoningEffort)

	if model == "" {
		return ""
	}

	if reasoningEffort == "" {
		return model
	}

	return model + "-" + reasoningEffort
}

func (s *codexParseState) trackSubscriptionPlan(logLine codexLogLine) {
	if logLine.Payload == nil || logLine.Payload.Type == nil || *logLine.Payload.Type != "token_count" ||
		logLine.Payload.Limits == nil || logLine.Payload.Limits.PlanType == nil ||
		strings.TrimSpace(*logLine.Payload.Limits.PlanType) == "" {
		return
	}

	s.subscriptionPlan = strings.TrimSpace(*logLine.Payload.Limits.PlanType)
}

func (s *codexParseState) applySubscriptionPlan() {
	if s.subscriptionPlan == "" {
		return
	}

	for i := range s.heartbeats {
		s.heartbeats[i].AISubscriptionPlan = s.subscriptionPlan
	}
}

func (s *codexParseState) trackUserMessage(logLine codexLogLine) {
	if logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "message" &&
		logLine.Payload.Role != nil && *logLine.Payload.Role == "user" {
		s.lastResponseItemUserTime = logLine.Timestamp
	}
}

func (s *codexParseState) trackAssistantMessage(logLine codexLogLine) {
	if logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "message" &&
		logLine.Payload.Role != nil && *logLine.Payload.Role == "assistant" {
		s.lastResponseItemAssistantTime = logLine.Timestamp
	}
}

func (s *codexParseState) trackAgentMessage(logLine codexLogLine) {
	if logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "agent_message" {
		s.lastAgentMessageTime = logLine.Timestamp
	}
}

func (s codexParseState) shouldSkipUserMessage(logLine codexLogLine) bool {
	return logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "user_message" &&
		withinOneSecondAfter(logLine.Timestamp, s.lastResponseItemUserTime)
}

func (s codexParseState) shouldSkipAssistantMessage(logLine codexLogLine) bool {
	return logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "message" &&
		logLine.Payload.Role != nil && *logLine.Payload.Role == "assistant" &&
		withinOneSecondAfter(logLine.Timestamp, s.lastAgentMessageTime)
}

func (s codexParseState) shouldSkipAgentMessage(logLine codexLogLine) bool {
	return logLine.Payload != nil && logLine.Payload.Type != nil && *logLine.Payload.Type == "agent_message" &&
		withinOneSecondAfter(logLine.Timestamp, s.lastResponseItemAssistantTime)
}

func withinOneSecondAfter(timestamp time.Time, previous time.Time) bool {
	return !previous.IsZero() && !timestamp.Before(previous) && timestamp.Sub(previous) <= time.Second
}

func (Codex) updateSessionInfo(state *codexSessionState, sessionMeta *codexSessionMeta) {
	if sessionMeta == nil || sessionMeta.Type != "session_meta" || sessionMeta.Payload == nil {
		return
	}

	if sessionMeta.Payload.ID != nil && *sessionMeta.Payload.ID != "" {
		state.id = *sessionMeta.Payload.ID
	}

	if sessionMeta.Payload.Cwd != nil && *sessionMeta.Payload.Cwd != "" {
		state.cwd = *sessionMeta.Payload.Cwd
	}

	if sessionMeta.Payload.Version != nil && *sessionMeta.Payload.Version != "" {
		state.version = *sessionMeta.Payload.Version
	}

	if sessionMeta.Payload.Source != nil && *sessionMeta.Payload.Source != "" {
		state.source = *sessionMeta.Payload.Source
	}
}

func (g Codex) getHeartbeats(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	agentVersion string,
	source string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	payload codexPayload,
	tokens heartbeat.AITokens,
) Heartbeats {
	if payload.Type != nil && *payload.Type == "message" && payload.Role != nil {
		if heartbeat := g.messageHeartbeat(
			timestamp,
			sessionEntity,
			sessionID,
			version,
			agentVersion,
			source,
			cwd,
			userAgents,
			fallbackUserAgent,
			payload,
			tokens,
		); heartbeat != nil {
			return Heartbeats{*heartbeat}
		}
	}

	if payload.Type != nil && *payload.Type == "user_message" && payload.Message != nil {
		if heartbeat := g.userMessageHeartbeat(
			timestamp,
			sessionEntity,
			sessionID,
			version,
			agentVersion,
			source,
			cwd,
			userAgents,
			fallbackUserAgent,
			*payload.Message,
			tokens,
		); heartbeat != nil {
			return Heartbeats{*heartbeat}
		}
	}

	if payload.Type != nil && *payload.Type == "agent_message" && payload.Message != nil {
		if heartbeat := g.agentMessageHeartbeat(
			timestamp,
			sessionEntity,
			sessionID,
			version,
			agentVersion,
			source,
			cwd,
			userAgents,
			fallbackUserAgent,
			*payload.Message,
			tokens,
		); heartbeat != nil {
			return Heartbeats{*heartbeat}
		}
	}

	var heartbeats Heartbeats
	for _, input := range codexPatchInputs(payload) {
		heartbeats = append(heartbeats, g.patchHeartbeats(
			timestamp,
			version,
			agentVersion,
			source,
			cwd,
			userAgents,
			fallbackUserAgent,
			input,
			sessionID,
			tokens,
		)...)
	}

	return heartbeats
}

func codexPatchInputs(payload codexPayload) []string {
	if payload.Name == nil || payload.Input == nil {
		return nil
	}

	switch *payload.Name {
	case "apply_patch":
		return []string{*payload.Input}
	case "exec":
		return codexExecPatchInputs(*payload.Input)
	default:
		return nil
	}
}

func codexExecPatchInputs(input string) []string {
	if !strings.Contains(input, "tools.apply_patch") {
		return nil
	}

	const (
		beginPatch = "*** Begin Patch"
		endPatch   = "*** End Patch"
	)

	var patches []string

	for {
		begin := strings.Index(input, beginPatch)
		if begin == -1 {
			break
		}

		input = input[begin:]

		end := strings.Index(input, endPatch)
		if end == -1 {
			break
		}

		end += len(endPatch)
		encoded := input[:end]
		input = input[end:]

		if strings.Contains(encoded, "\n") {
			patches = append(patches, encoded)
			continue
		}

		decoded, err := strconv.Unquote(`"` + encoded + `"`)
		if err == nil {
			patches = append(patches, decoded)
		}
	}

	return patches
}

func (g Codex) patchHeartbeats(
	timestamp time.Time,
	version string,
	agentVersion string,
	source string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	input string,
	sessionID string,
	tokens heartbeat.AITokens,
) Heartbeats {
	var heartbeats Heartbeats

	var (
		currentFile string
		additions   int
		deletions   int
	)

	lines := strings.Split(input, "\n")
	for i := range lines {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "*** ") {
			if currentFile != "" {
				heartbeats = append(heartbeats, g.heartbeat(
					currentFile,
					sessionID,
					timestamp,
					version,
					agentVersion,
					source,
					userAgents,
					fallbackUserAgent,
					additions,
					deletions,
					tokens,
				))
			}

			currentFile = codexFilePath(cwd, line)
			additions = 0
			deletions = 0
		} else if currentFile != "" {
			if strings.HasPrefix(line, "+") {
				additions++
			} else if strings.HasPrefix(line, "-") {
				deletions++
			}
		}
	}

	if currentFile != "" {
		heartbeats = append(heartbeats, g.heartbeat(
			currentFile,
			sessionID,
			timestamp,
			version,
			agentVersion,
			source,
			userAgents,
			fallbackUserAgent,
			additions,
			deletions,
			tokens,
		))
	}

	return heartbeats
}

func (g Codex) messageHeartbeat(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	agentVersion string,
	source string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	payload codexPayload,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	var (
		entity       = sessionEntity
		expectedType string
		lineChanges  int
		promptChars  int
	)

	switch *payload.Role {
	case "user":
		expectedType = "input_text"
	case "assistant":
		expectedType = "output_text"
	default:
		return nil
	}

	for _, item := range payload.Content {
		text := item.Text
		if *payload.Role == "user" {
			text = codexUserMessageText(text)
		}

		if item.Type != expectedType || strings.TrimSpace(text) == "" {
			continue
		}

		if *payload.Role == "user" {
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
		float64(timestamp.Unix()),
		g.userAgent(entity, agentVersion, version, source, userAgents, fallbackUserAgent),
	)
	if *payload.Role == "user" && promptChars > 0 {
		h.AIPromptLength = promptChars
	}

	return &h
}

func (g Codex) agentMessageHeartbeat(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	agentVersion string,
	source string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	message string,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	role := "assistant"

	return g.messageHeartbeat(
		timestamp,
		sessionEntity,
		sessionID,
		version,
		agentVersion,
		source,
		cwd,
		userAgents,
		fallbackUserAgent,
		codexPayload{
			Role: &role,
			Content: []codexContentItem{
				{
					Type: "output_text",
					Text: message,
				},
			},
		},
		tokens,
	)
}

func (g Codex) userMessageHeartbeat(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
	agentVersion string,
	source string,
	cwd string,
	userAgents map[string]string,
	fallbackUserAgent string,
	message string,
	tokens heartbeat.AITokens,
) *heartbeat.Heartbeat {
	text := codexUserMessageText(message)
	if strings.TrimSpace(text) == "" {
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
		float64(timestamp.Unix()),
		g.userAgent(sessionEntity, agentVersion, version, source, userAgents, fallbackUserAgent),
	)
	h.AIPromptLength = len([]rune(text))

	return &h
}

func codexUserMessageText(text string) string {
	trimmed := codexStripHarnessPrefix(text)
	if trimmed == "" {
		return ""
	}

	const requestPrefix = "## My request for Codex:"
	if strings.Contains(trimmed, "# Context from my IDE setup:") {
		if _, request, ok := strings.Cut(trimmed, requestPrefix); ok {
			return strings.TrimSpace(request)
		}
	}

	return trimmed
}

func codexStripHarnessPrefix(text string) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}

	for strings.HasPrefix(trimmed, "<") {
		closeIndex := strings.Index(trimmed, ">")
		if closeIndex <= 1 {
			return ""
		}

		openTag := trimmed[1:closeIndex]
		if strings.HasPrefix(openTag, "/") {
			return ""
		}

		closeTag := "</" + openTag + ">"

		blockEnd := strings.Index(trimmed, closeTag)
		if blockEnd < 0 {
			return ""
		}

		trimmed = strings.TrimSpace(trimmed[blockEnd+len(closeTag):])
	}

	return trimmed
}

func codexFilePath(cwd string, line string) string {
	prefixes := []string{
		"*** Update File: ",
		"*** Add File: ",
		"*** Delete File: ",
	}
	for _, prefix := range prefixes {
		if file, ok := strings.CutPrefix(line, prefix); ok {
			if file == "" || filepath.IsAbs(file) || strings.HasPrefix(file, "/") {
				return file
			}

			return filepath.Join(cwd, file)
		}
	}

	return ""
}

func (Codex) userAgent(
	entity string,
	agentVersion string,
	version string,
	source string,
	userAgents map[string]string,
	fallbackUserAgent string,
) string {
	return aiUserAgentWithModelAndEditor(
		entity,
		userAgents,
		fallbackUserAgent,
		agentVersion,
		"",
		codexSourceEditor(source, version),
	)
}

func codexSourceEditor(source string, version string) string {
	source = strings.TrimSpace(strings.ToLower(source))
	if source == "" {
		return ""
	}

	switch source {
	case "cli":
		if version == "" {
			version = "unknown"
		}

		return "codex-cli/" + version
	default:
		product := codexSourceProduct(source)
		if product == "" {
			return ""
		}

		return product + "-wakatime/unknown"
	}
}

func codexSourceProduct(source string) string {
	source = strings.TrimSpace(strings.ToLower(source))
	source = strings.NewReplacer("/", "-", "\\", "-", " ", "-").Replace(source)
	source = strings.Trim(source, "-")

	return source
}

func (g Codex) heartbeat(
	currentFile string,
	sessionID string,
	timestamp time.Time,
	version string,
	agentVersion string,
	source string,
	userAgents map[string]string,
	fallbackUserAgent string,
	additions int,
	deletions int,
	tokens heartbeat.AITokens,
) heartbeat.Heartbeat {
	return heartbeat.NewWithAITokens(
		heartbeat.PointerTo(additions-deletions),
		sessionID,
		tokens,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		currentFile,
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
		g.userAgent(currentFile, agentVersion, version, source, userAgents, fallbackUserAgent),
	)
}

func (Codex) codexTokenCounts(line codexLogLine, tokens heartbeat.AITokens, after time.Time) heartbeat.AITokens {
	if line.Payload == nil {
		return tokens
	}

	payload := *line.Payload
	if payload.Type == nil || *payload.Type != "token_count" || payload.Info == nil {
		return tokens
	}

	info := *payload.Info
	if info.TotalTokenUsage != nil && info.TotalTokenUsage.InputTokens != nil && info.TotalTokenUsage.OutputTokens != nil {
		tokens.CurrentInput = int64(*info.TotalTokenUsage.InputTokens)
		tokens.CurrentOutput = int64(*info.TotalTokenUsage.OutputTokens)
	}

	if line.Timestamp.IsZero() || line.Timestamp.Before(after) {
		tokens.LastInput = tokens.CurrentInput
		tokens.LastOutput = tokens.CurrentOutput
	}

	return tokens
}

func (Codex) sessionIDFromPath(path string) string {
	base := strings.TrimPrefix(filepath.Base(path), "rollout-")
	base = strings.TrimSuffix(base, filepath.Ext(base))

	parts := strings.SplitN(base, "-", 6)
	if len(parts) == 6 {
		return parts[5]
	}

	return base
}

// Name returns its id.
func (Codex) Name() string {
	return "Codex"
}
