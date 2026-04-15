package ai

import (
	"bufio"
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

// Codex contains params for detecting heartbeats from Codex session transcripts.
type Codex ParserConfig

type (
	codexSessionMeta struct {
		Type    string `json:"type"`
		Payload *struct {
			ID      *string `json:"id"`
			Cwd     *string `json:"cwd"`
			Version *string `json:"cli_version"`
		} `json:"payload"`
	}

	codexPayload struct {
		Type    *string                     `json:"type"`
		Name    *string                     `json:"name"`
		Input   *string                     `json:"input"`
		Role    *string                     `json:"role"`
		Status  *string                     `json:"status"`
		Content []codexContentItem          `json:"content"`
		Cwd     *string                     `json:"cwd"`
		Info    *codexPayloadTokenCountInfo `json:"info"`
	}

	codexPayloadTokenCountInfo struct {
		TotalTokenUsage    *codexPayloadTokenCountInfoUsage `json:"total_token_usage"`
		LastTokenUsage     *codexPayloadTokenCountInfoUsage `json:"last_token_usage"`
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
			return nil, err
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

	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
	}

	cwd := ""
	sessionEntity := appHeartbeatEntity("Codex", transcript)
	sessionID := g.sessionIDFromPath(transcript)
	version := ""

	if len(firstLine) > 0 {
		var sessionMeta *codexSessionMeta
		if err := json.Unmarshal(firstLine, &sessionMeta); err != nil {
			logger.Debugf("failed parsing codex session metadata from %q: %s", transcript, err)
		} else {
			cwd, version, sessionID = g.sessionInfo(cwd, version, sessionID, sessionMeta)
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind codex transcript %q: %s", transcript, err)
	}

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

	var (
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
	)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var logLine codexLogLine
		if err := json.Unmarshal(line, &logLine); err != nil {
			logger.Warnf("failed parsing codex transcript line from %q: %s", transcript, err)
			logger.Debugf("failed parsing codex transcript line: %s", line)

			continue
		}

		tokens = g.codexTokenCounts(logLine, tokens)

		if logLine.Timestamp.IsZero() || logLine.Timestamp.Before(g.After) {
			tokens.LastInput = tokens.CurrentInput
			tokens.LastOutput = tokens.CurrentOutput

			continue
		}

		var aiHeartbeats Heartbeats
		if logLine.Payload != nil {
			aiHeartbeats = g.getHeartbeats(
				logLine.Timestamp,
				sessionEntity,
				sessionID,
				version,
				cwd,
				g.UserAgents,
				g.FallbackUserAgent,
				*logLine.Payload,
				tokens,
			)
		}

		if len(aiHeartbeats) == 0 {
			continue
		}

		tokens.LastInput = tokens.CurrentInput
		tokens.LastOutput = tokens.CurrentOutput

		heartbeats = append(heartbeats, aiHeartbeats...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading codex transcript %q: %s", transcript, err)
	}

	return heartbeats, nil
}

func (Codex) sessionInfo(
	cwd string,
	version string,
	sessionID string,
	sessionMeta *codexSessionMeta,
) (string, string, string) {
	if sessionMeta == nil || sessionMeta.Type != "session_meta" || sessionMeta.Payload == nil {
		return cwd, version, sessionID
	}

	if sessionMeta.Payload.ID != nil && *sessionMeta.Payload.ID != "" {
		sessionID = *sessionMeta.Payload.ID
	}

	if sessionMeta.Payload.Cwd != nil && *sessionMeta.Payload.Cwd != "" {
		cwd = *sessionMeta.Payload.Cwd
	}

	if sessionMeta.Payload.Version != nil && *sessionMeta.Payload.Version != "" {
		version = *sessionMeta.Payload.Version
	}

	return cwd, version, sessionID
}

func (g Codex) getHeartbeats(
	timestamp time.Time,
	sessionEntity string,
	sessionID string,
	version string,
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
			cwd,
			userAgents,
			fallbackUserAgent,
			payload,
			tokens,
		); heartbeat != nil {
			return Heartbeats{*heartbeat}
		}
	}

	if payload.Name == nil || *payload.Name != "apply_patch" || payload.Input == nil {
		return nil
	}

	return g.patchHeartbeats(timestamp, version, cwd, userAgents, fallbackUserAgent, *payload.Input, sessionID, tokens)
}

func (g Codex) patchHeartbeats(
	timestamp time.Time,
	version string,
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
		if item.Type != expectedType || strings.TrimSpace(item.Text) == "" {
			continue
		}

		if *payload.Role == "user" {
			if strings.HasPrefix(strings.TrimSpace(item.Text), "<") {
				continue
			}

			promptChars += len([]rune(item.Text))
		}

		lineChanges += countStringLines(item.Text)
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
		aiUserAgent(entity, userAgents, fallbackUserAgent, aiPlugin(g, version)),
	)
	if *payload.Role == "user" && promptChars > 0 {
		h.AIPromptLength = promptChars
	}

	return &h
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

func (g Codex) heartbeat(
	currentFile string,
	sessionID string,
	timestamp time.Time,
	version string,
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
		aiUserAgent(currentFile, userAgents, fallbackUserAgent, aiPlugin(g, version)),
	)
}

func (Codex) codexTokenCounts(line codexLogLine, previous heartbeat.AITokens) heartbeat.AITokens {
	if line.Payload == nil {
		return previous
	}

	payload := *line.Payload
	if payload.Type == nil || *payload.Type != "token_count" || payload.Info == nil {
		return previous
	}

	info := *payload.Info

	lastInput := previous.LastInput
	lastOutput := previous.LastOutput

	var (
		currentInput  int64
		currentOutput int64
	)

	if lastInput == 0 && lastOutput == 0 &&
		info.LastTokenUsage != nil &&
		info.LastTokenUsage.InputTokens != nil &&
		info.LastTokenUsage.OutputTokens != nil {
		lastInput = int64(*info.LastTokenUsage.InputTokens)
		lastOutput = int64(*info.LastTokenUsage.OutputTokens)
	}

	if info.TotalTokenUsage != nil && info.TotalTokenUsage.InputTokens != nil && info.TotalTokenUsage.OutputTokens != nil {
		currentInput = int64(*info.TotalTokenUsage.InputTokens)
		currentOutput = int64(*info.TotalTokenUsage.OutputTokens)
	} else if info.TotalTokens != nil {
		currentOutput = int64(*info.TotalTokens)
	}

	return heartbeat.AITokens{
		LastInput:     lastInput,
		LastOutput:    lastOutput,
		CurrentInput:  currentInput,
		CurrentOutput: currentOutput,
	}
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
