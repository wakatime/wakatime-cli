package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
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
		FilePath     string    `json:"filePath"`
		LinesAdded   int       `json:"linesAdded"`
		LinesRemoved int       `json:"linesRemoved"`
		AuthorType   string    `json:"authorType"`
		SourceType   string    `json:"sourceType"`
		EventType    string    `json:"eventType"`
		SessionID    string    `json:"sessionId"`
		Timestamp    time.Time `json:"timestamp"`
	}

	grokBuildUpdateContent struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	grokBuildUpdateMeta struct {
		ModelID          string `json:"modelId"`
		PromptIndex      *int   `json:"promptIndex"`
		AgentTimestampMs int64  `json:"agentTimestampMs"`
		EventID          string `json:"eventId"`
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

	grokBuildParseState struct {
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
		model      string
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
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	sessionsRoot := filepath.Join(home, ".grok", "sessions")
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
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(home, ".grok", "version.json")) //nolint:gosec
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

	state := grokBuildParseState{model: session.model}

	if err := g.parseUpdates(ctx, logger, sessionDir, session, &state); err != nil {
		return nil, err
	}

	if err := g.parseHunks(ctx, logger, sessionDir, session, &state); err != nil {
		return nil, err
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

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var updateLine grokBuildUpdateLine
		if err := json.Unmarshal(line, &updateLine); err != nil {
			logger.Warnf("failed parsing Grok Build updates line from %q: %s", path, err)
			continue
		}

		g.handleUpdateLine(updateLine, session, state)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed reading updates %q: %s", path, err)
	}

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
	}

	switch update.SessionUpdate {
	case "user_message_chunk":
		g.handleUserMessage(ts, update, session, state)
	case "turn_completed":
		g.handleTurnCompleted(ts, update, state)
	}
}

func (g GrokBuild) handleUserMessage(
	ts time.Time,
	update grokBuildUpdate,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	if update.Content == nil {
		return
	}

	length := promptLength(update.Content.Text)
	if length == 0 {
		return
	}

	if ts.IsZero() || ts.Before(g.After) {
		return
	}

	entity := appHeartbeatEntity(g.Name(), session.id)
	model := firstNonEmptyString(state.model, session.model)

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
		float64(ts.UnixMilli())/1000,
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

	home, _ := ini.UserHomeDir(ctx)
	grokHome := filepath.Join(home, ".grok")

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return fmt.Errorf("failed to open hunk records %q: %s", path, err)
	}
	defer fh.Close() //nolint:errcheck,gosec

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var hunk grokBuildHunk
		if err := json.Unmarshal(line, &hunk); err != nil {
			logger.Warnf("failed parsing Grok Build hunk line from %q: %s", path, err)
			continue
		}

		if hb := g.hunkHeartbeat(hunk, session, state.model, grokHome); hb != nil {
			state.heartbeats = append(state.heartbeats, *hb)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed reading hunk records %q: %s", path, err)
	}

	return nil
}

func (g GrokBuild) hunkHeartbeat(
	hunk grokBuildHunk,
	session grokBuildSession,
	model string,
	grokHome string,
) *heartbeat.Heartbeat {
	if !strings.EqualFold(strings.TrimSpace(hunk.AuthorType), "agent") {
		return nil
	}

	filePath := strings.TrimSpace(hunk.FilePath)
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
	model = firstNonEmptyString(model, session.model)

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
