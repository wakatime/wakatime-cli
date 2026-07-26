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
	"sort"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// GrokBuild contains params for detecting heartbeats from Grok Build (x.ai CLI) sessions.
type GrokBuild ParserConfig

type (
	grokBuildSummaryInfo struct {
		ID  string `json:"id"`
		Cwd string `json:"cwd"`
	}

	grokBuildSummary struct {
		Info           grokBuildSummaryInfo `json:"info"`
		CurrentModelID string               `json:"current_model_id"`
		GitRootDir     string               `json:"git_root_dir"`
		UpdatedAt      time.Time            `json:"updated_at"`
		LastActiveAt   time.Time            `json:"last_active_at"`
	}

	grokBuildVersionFile struct {
		Version string `json:"version"`
	}

	grokBuildHunk struct {
		HunkID       string    `json:"hunkId"`
		FilePath     string    `json:"filePath"`
		LinesAdded   int       `json:"linesAdded"`
		LinesRemoved int       `json:"linesRemoved"`
		AuthorType   string    `json:"authorType"`
		SourceType   string    `json:"sourceType"`
		EventType    string    `json:"eventType"`
		SessionID    string    `json:"sessionId"`
		PromptIndex  *int      `json:"promptIndex"`
		Timestamp    time.Time `json:"timestamp"`
	}

	grokBuildContentMeta struct {
		BashCommand string `json:"bash_command"`
	}

	grokBuildUpdateContent struct {
		Type string                `json:"type"`
		Text string                `json:"text"`
		Meta *grokBuildContentMeta `json:"_meta"`
	}

	grokBuildUpdateMeta struct {
		ModelID          string `json:"modelId"`
		PromptIndex      *int   `json:"promptIndex"`
		AgentTimestampMs int64  `json:"agentTimestampMs"`
		EventID          string `json:"eventId"`
		HostTurn         bool   `json:"hostTurn"`
	}

	grokBuildUsage struct {
		InputTokens  int64 `json:"inputTokens"`
		OutputTokens int64 `json:"outputTokens"`
	}

	grokBuildUpdate struct {
		SessionUpdate string                  `json:"sessionUpdate"`
		Content       *grokBuildUpdateContent `json:"content"`
		Usage         *grokBuildUsage         `json:"usage"`
		Meta          *grokBuildUpdateMeta    `json:"_meta"`
	}

	grokBuildUpdateParams struct {
		SessionID string               `json:"sessionId"`
		Update    grokBuildUpdate      `json:"update"`
		Meta      *grokBuildUpdateMeta `json:"_meta"`
	}

	grokBuildUpdateLine struct {
		Timestamp int64                 `json:"timestamp"`
		Method    string                `json:"method"`
		Params    grokBuildUpdateParams `json:"params"`
	}

	grokBuildSession struct {
		cwd     string
		id      string
		model   string
		version string
	}

	grokBuildPendingPrompt struct {
		text        string
		promptIndex *int
		ts          time.Time
		model       string
	}

	grokBuildParseState struct {
		heartbeats          Heartbeats
		tokens              heartbeat.AITokens
		model               string
		modelsByPromptIndex map[int]string
		agentHunks          map[string]bool
		pending             *grokBuildPendingPrompt
	}
)

// Parse parses Grok Build session directories for ai heartbeats.
func (g GrokBuild) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	sessions, err := g.sessionDirs(ctx)
	if err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d Grok Build sessions modified after %s for %s", len(sessions), g.After, g.Name())

	version := g.cliVersion(ctx)
	var heartbeats Heartbeats

	for _, sessionDir := range sessions {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		parsed, err := g.parseSession(ctx, sessionDir, version)
		if err != nil {
			logger.Warnf("failed parsing Grok Build session %q: %s", sessionDir, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	sort.SliceStable(heartbeats, func(i, j int) bool {
		return heartbeats[i].Time < heartbeats[j].Time
	})

	return heartbeats, nil
}

func (g GrokBuild) sessionDirs(ctx context.Context) ([]string, error) {
	grokHome, err := grokBuildHomeDir(ctx)
	if err != nil {
		return nil, err
	}

	sessionsRoot := filepath.Join(grokHome, "sessions")
	if _, err := os.Stat(sessionsRoot); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat .grok sessions directory: %s", err)
	}

	var sessions []string

	cwdEntries, err := os.ReadDir(sessionsRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to read .grok sessions directory: %s", err)
	}

	for _, cwdEntry := range cwdEntries {
		if !cwdEntry.IsDir() {
			continue
		}

		cwdDir := filepath.Join(sessionsRoot, cwdEntry.Name())

		sessionEntries, err := os.ReadDir(cwdDir)
		if err != nil {
			continue
		}

		for _, sessionEntry := range sessionEntries {
			if !sessionEntry.IsDir() {
				continue
			}

			sessionDir := filepath.Join(cwdDir, sessionEntry.Name())
			if !g.sessionModifiedAfter(sessionDir) {
				continue
			}

			sessions = append(sessions, sessionDir)
		}
	}

	return sessions, nil
}

func (g GrokBuild) sessionModifiedAfter(sessionDir string) bool {
	for _, name := range []string{"summary.json", "hunk_records.jsonl", "updates.jsonl"} {
		info, err := os.Stat(filepath.Join(sessionDir, name))
		if err != nil {
			continue
		}

		if !info.ModTime().Before(g.After) {
			return true
		}
	}

	return false
}

func (g GrokBuild) cliVersion(ctx context.Context) string {
	grokHome, err := grokBuildHomeDir(ctx)
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(grokHome, "version.json")) //nolint:gosec
	if err != nil {
		return ""
	}

	var versionFile grokBuildVersionFile
	if err := json.Unmarshal(data, &versionFile); err != nil {
		return ""
	}

	return strings.TrimSpace(versionFile.Version)
}

func (g GrokBuild) parseSession(ctx context.Context, sessionDir string, version string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	session, err := g.readSummary(sessionDir, version)
	if err != nil {
		return nil, err
	}

	state := grokBuildParseState{
		model:               session.model,
		modelsByPromptIndex: make(map[int]string),
		agentHunks:          make(map[string]bool),
	}

	if err := g.parseUpdates(ctx, logger, sessionDir, session, &state); err != nil {
		logger.Warnf("failed parsing Grok Build updates for session %q: %s", sessionDir, err)
	}

	if err := g.parseHunks(ctx, logger, sessionDir, session, &state); err != nil {
		logger.Warnf("failed parsing Grok Build hunks for session %q: %s", sessionDir, err)
	}

	return state.heartbeats, nil
}

func (g GrokBuild) readSummary(sessionDir string, version string) (grokBuildSession, error) {
	summaryPath := filepath.Join(sessionDir, "summary.json")

	data, err := os.ReadFile(filepath.Clean(summaryPath)) //nolint:gosec
	if err != nil {
		return grokBuildSession{}, fmt.Errorf("failed to read summary %q: %s", summaryPath, err)
	}

	var summary grokBuildSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return grokBuildSession{}, fmt.Errorf("failed to parse summary %q: %s", summaryPath, err)
	}

	sessionID := strings.TrimSpace(summary.Info.ID)
	if sessionID == "" {
		sessionID = filepath.Base(sessionDir)
	}

	cwd := firstNonEmptyString(strings.TrimSpace(summary.GitRootDir), strings.TrimSpace(summary.Info.Cwd))
	cwd = strings.TrimRight(cwd, `/\`)
	cwd = filepath.FromSlash(cwd)

	return grokBuildSession{
		cwd:     cwd,
		id:      sessionID,
		model:   strings.TrimSpace(summary.CurrentModelID),
		version: version,
	}, nil
}

func (g GrokBuild) parseUpdates(
	ctx context.Context,
	logger *log.Logger,
	sessionDir string,
	session grokBuildSession,
	state *grokBuildParseState,
) error {
	path := filepath.Join(sessionDir, "updates.jsonl")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to stat updates %q: %s", path, err)
	}

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open updates %q: %s", path, err)
	}
	defer fh.Close() //nolint:errcheck,gosec

	reader := bufio.NewReader(fh)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line, readErr := grokBuildReadJSONLLine(reader, maxTranscriptLineSize)
		if errors.Is(readErr, errGrokBuildOversizedLine) {
			logger.Warnf("skipping oversized Grok Build updates line in %q", path)
			continue
		}

		if errors.Is(readErr, io.EOF) {
			if len(line) == 0 {
				break
			}
		} else if readErr != nil {
			return fmt.Errorf("failed reading updates %q: %s", path, readErr)
		}

		if len(line) == 0 {
			if errors.Is(readErr, io.EOF) {
				break
			}

			continue
		}

		updateLine, ok := grokBuildDecodeUpdateLine(line)
		if !ok {
			logger.Warnf("failed parsing Grok Build updates line from %q", path)
			if errors.Is(readErr, io.EOF) {
				break
			}

			continue
		}

		g.handleUpdateLine(updateLine, session, state)

		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	g.flushPendingPrompt(session, state)

	return nil
}

func (g GrokBuild) handleUpdateLine(
	line grokBuildUpdateLine,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	update := line.Params.Update
	ts := grokBuildEventTime(line)

	if update.Meta != nil && update.Meta.ModelID != "" {
		state.model = update.Meta.ModelID
		if update.Meta.PromptIndex != nil {
			state.modelsByPromptIndex[*update.Meta.PromptIndex] = update.Meta.ModelID
		}
	}

	switch update.SessionUpdate {
	case "user_message_chunk":
		g.handleUserMessage(ts, update, session, state)
	default:
		g.flushPendingPrompt(session, state)

		if update.SessionUpdate == "turn_completed" {
			g.handleTurnCompleted(ts, update, state)
		}
	}
}

func (g GrokBuild) handleUserMessage(
	ts time.Time,
	update grokBuildUpdate,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	// Excluded chunks must not open a new prompt or emit heartbeats, and must not
	// interrupt consecutive user text accumulation (unlike upstream NotUserMessage,
	// which flushes for prompt-index reconstruction).
	if update.Meta != nil && update.Meta.HostTurn {
		return
	}

	if update.Content == nil {
		return
	}

	if update.Content.Meta != nil && strings.TrimSpace(update.Content.Meta.BashCommand) != "" {
		return
	}

	if !strings.EqualFold(strings.TrimSpace(update.Content.Type), "text") {
		// Non-text user chunks (e.g. images) end any in-progress accumulation.
		g.flushPendingPrompt(session, state)
		return
	}

	text := update.Content.Text
	if strings.TrimSpace(text) == "" {
		return
	}

	var promptIndex *int
	if update.Meta != nil {
		promptIndex = update.Meta.PromptIndex
	}

	model := firstNonEmptyString(state.model, session.model)

	if state.pending != nil {
		newRun := false
		if promptIndex != nil || state.pending.promptIndex != nil {
			newRun = !grokBuildSamePromptIndex(state.pending.promptIndex, promptIndex)
		}

		if newRun {
			g.flushPendingPrompt(session, state)
		}
	}

	if state.pending == nil {
		state.pending = &grokBuildPendingPrompt{
			promptIndex: promptIndex,
			ts:          ts,
			model:       model,
		}
	}

	state.pending.text += text

	if !ts.IsZero() && (state.pending.ts.IsZero() || ts.After(state.pending.ts)) {
		state.pending.ts = ts
	}

	if model != "" {
		state.pending.model = model
	}

	if promptIndex != nil {
		state.pending.promptIndex = promptIndex
	}
}

func (g GrokBuild) flushPendingPrompt(session grokBuildSession, state *grokBuildParseState) {
	if state.pending == nil {
		return
	}

	pending := state.pending
	state.pending = nil

	length := promptLength(pending.text)
	if length == 0 {
		return
	}

	if pending.ts.IsZero() || pending.ts.Before(g.After) {
		return
	}

	entity := appHeartbeatEntity(g.Name(), session.id)
	model := firstNonEmptyString(pending.model, state.model, session.model)

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
		float64(pending.ts.UnixMilli())/1000,
		g.userAgent(entity, model, session.version),
	)
	h.AIPromptLength = length

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
	state.heartbeats = append(state.heartbeats, h)
}

func (g GrokBuild) handleTurnCompleted(
	ts time.Time,
	update grokBuildUpdate,
	state *grokBuildParseState,
) {
	if update.Usage == nil {
		return
	}

	state.tokens.CurrentInput = state.tokens.LastInput + update.Usage.InputTokens
	state.tokens.CurrentOutput = state.tokens.LastOutput + update.Usage.OutputTokens

	if ts.IsZero() || ts.Before(g.After) {
		state.tokens.LastInput = state.tokens.CurrentInput
		state.tokens.LastOutput = state.tokens.CurrentOutput

		return
	}

	inputTokens := state.tokens.CurrentInput - state.tokens.LastInput
	if inputTokens < 0 {
		inputTokens = 0
	}

	outputTokens := state.tokens.CurrentOutput - state.tokens.LastOutput
	if outputTokens < 0 {
		outputTokens = 0
	}

	if len(state.heartbeats) > 0 {
		i := len(state.heartbeats) - 1
		state.heartbeats[i].AIInputTokens += inputTokens
		state.heartbeats[i].AIOutputTokens += outputTokens
	}

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
}

func (g GrokBuild) parseHunks(
	ctx context.Context,
	logger *log.Logger,
	sessionDir string,
	session grokBuildSession,
	state *grokBuildParseState,
) error {
	path := filepath.Join(sessionDir, "hunk_records.jsonl")
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to stat hunk records %q: %s", path, err)
	}

	grokHome, _ := grokBuildHomeDir(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open hunk records %q: %s", path, err)
	}
	defer fh.Close() //nolint:errcheck,gosec

	reader := bufio.NewReader(fh)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line, readErr := grokBuildReadJSONLLine(reader, maxTranscriptLineSize)
		if errors.Is(readErr, errGrokBuildOversizedLine) {
			logger.Warnf("skipping oversized Grok Build hunk line in %q", path)
			continue
		}

		if errors.Is(readErr, io.EOF) {
			if len(line) == 0 {
				break
			}
		} else if readErr != nil {
			return fmt.Errorf("failed reading hunk records %q: %s", path, readErr)
		}

		if len(line) == 0 {
			if errors.Is(readErr, io.EOF) {
				break
			}

			continue
		}

		var hunk grokBuildHunk
		if unmarshalErr := json.Unmarshal(line, &hunk); unmarshalErr != nil {
			logger.Warnf("failed parsing Grok Build hunk line from %q: %s", path, unmarshalErr)
			if errors.Is(readErr, io.EOF) {
				break
			}

			continue
		}

		if hb := g.hunkHeartbeat(hunk, session, state, grokHome); hb != nil {
			state.heartbeats = append(state.heartbeats, *hb)
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	return nil
}

func (g GrokBuild) hunkHeartbeat(
	hunk grokBuildHunk,
	session grokBuildSession,
	state *grokBuildParseState,
	grokHome string,
) *heartbeat.Heartbeat {
	eventType := strings.ToLower(strings.TrimSpace(hunk.EventType))
	hunkID := strings.TrimSpace(hunk.HunkID)
	isAgent := strings.EqualFold(strings.TrimSpace(hunk.AuthorType), "agent")
	isRemoved := eventType == "removed"

	if isRemoved {
		if hunkID == "" || !state.agentHunks[hunkID] {
			return nil
		}

		delete(state.agentHunks, hunkID)
	} else {
		if !isAgent {
			return nil
		}

		if hunkID != "" {
			state.agentHunks[hunkID] = true
		}
	}

	filePath := filepath.FromSlash(strings.TrimSpace(hunk.FilePath))
	if filePath == "" {
		return nil
	}

	if grokBuildShouldSkipPath(filePath, grokHome) {
		return nil
	}

	if hunk.Timestamp.IsZero() || hunk.Timestamp.Before(g.After) {
		return nil
	}

	lineChanges := hunk.LinesAdded - hunk.LinesRemoved
	model := g.hunkModel(hunk, session, state)

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		session.id,
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
		float64(hunk.Timestamp.UnixMilli())/1000,
		g.userAgent(filePath, model, session.version),
	)

	return &h
}

func (g GrokBuild) hunkModel(
	hunk grokBuildHunk,
	session grokBuildSession,
	state *grokBuildParseState,
) string {
	if hunk.PromptIndex != nil {
		if model, ok := state.modelsByPromptIndex[*hunk.PromptIndex]; ok && model != "" {
			return model
		}
	}

	return firstNonEmptyString(state.model, session.model)
}

func (g GrokBuild) userAgent(entity string, model string, version string) string {
	return aiUserAgentWithModelAndEditor(
		entity,
		g.UserAgents,
		g.FallbackUserAgent,
		model,
		"",
		aiPlugin(g, version),
	)
}

func grokBuildHomeDir(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	if configured := strings.TrimSpace(os.Getenv("GROK_HOME")); configured != "" {
		return grokBuildResolveDir(configured, home)
	}

	return filepath.Join(home, ".grok"), nil
}

func grokBuildResolveDir(path string, home string) (string, error) {
	path = strings.TrimSpace(path)

	switch {
	case path == "~":
		path = home
	case strings.HasPrefix(path, "~/"), strings.HasPrefix(path, `~\`):
		path = filepath.Join(home, path[2:])
	}

	return filepath.Abs(path)
}

var errGrokBuildOversizedLine = errors.New("oversized jsonl line")

// grokBuildReadJSONLLine reads one JSONL record with a hard size bound.
// Oversized records are consumed through their terminating newline and return
// errGrokBuildOversizedLine so callers can skip and continue.
func grokBuildReadJSONLLine(reader *bufio.Reader, maxSize int) ([]byte, error) {
	var buf []byte

	for {
		chunk, err := reader.ReadSlice('\n')
		hasNewline := len(chunk) > 0 && chunk[len(chunk)-1] == '\n'
		content := chunk

		if hasNewline {
			content = chunk[:len(chunk)-1]
		}

		if len(buf)+len(content) > maxSize {
			if !hasNewline {
				if discardErr := grokBuildDiscardThroughNewline(reader); discardErr != nil && !errors.Is(discardErr, io.EOF) {
					return nil, discardErr
				}
			}

			return nil, errGrokBuildOversizedLine
		}

		if len(content) > 0 {
			buf = append(buf, content...)
		}

		if hasNewline {
			return buf, nil
		}

		if errors.Is(err, io.EOF) {
			if len(buf) == 0 {
				return nil, io.EOF
			}

			return buf, io.EOF
		}

		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}

		if err != nil {
			return nil, err
		}
	}
}

func grokBuildDiscardThroughNewline(reader *bufio.Reader) error {
	for {
		_, err := reader.ReadSlice('\n')
		if err == nil {
			return nil
		}

		if errors.Is(err, io.EOF) {
			return io.EOF
		}

		if !errors.Is(err, bufio.ErrBufferFull) {
			return err
		}
	}
}

func grokBuildDecodeUpdateLine(line []byte) (grokBuildUpdateLine, bool) {
	var envelope grokBuildUpdateLine
	if err := json.Unmarshal(line, &envelope); err == nil && strings.TrimSpace(envelope.Method) != "" {
		return envelope, true
	}

	var params grokBuildUpdateParams
	if err := json.Unmarshal(line, &params); err == nil && strings.TrimSpace(params.Update.SessionUpdate) != "" {
		return grokBuildUpdateLine{Params: params}, true
	}

	return grokBuildUpdateLine{}, false
}

func grokBuildSamePromptIndex(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	return *a == *b
}

func grokBuildEventTime(line grokBuildUpdateLine) time.Time {
	meta := line.Params.Meta
	if meta == nil {
		meta = line.Params.Update.Meta
	}

	if meta != nil && meta.AgentTimestampMs > 0 {
		return time.UnixMilli(meta.AgentTimestampMs).UTC()
	}

	if line.Timestamp > 0 {
		// Prefer seconds; values from Grok are unix seconds.
		if line.Timestamp > 1_000_000_000_000 {
			return time.UnixMilli(line.Timestamp).UTC()
		}

		return time.Unix(line.Timestamp, 0).UTC()
	}

	return time.Time{}
}

func grokBuildShouldSkipPath(path string, grokHome string) bool {
	clean := filepath.Clean(path)

	if grokHome != "" {
		rel, err := filepath.Rel(filepath.Clean(grokHome), clean)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
			return true
		}
	}

	return strings.Contains(filepath.ToSlash(clean), "/.grok/")
}

// Name returns its id.
func (GrokBuild) Name() string {
	return "Grok Build"
}
