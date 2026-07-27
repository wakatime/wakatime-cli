package ai

import (
	"bufio"
	"bytes"
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

const grokBuildXAIUpdateMethod = "_x.ai/session/update"

// GrokBuild contains params for detecting heartbeats from Grok Build sessions.
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
	}

	grokBuildVersion struct {
		Version       string `json:"version"`
		StableVersion string `json:"stable_version"`
	}

	grokBuildSession struct {
		cwd     string
		id      string
		model   string
		version string
	}

	grokBuildUpdateEnvelope struct {
		Timestamp int64           `json:"timestamp"`
		Method    string          `json:"method"`
		Params    json.RawMessage `json:"params"`
	}

	grokBuildUpdateParams struct {
		SessionID string               `json:"sessionId"`
		Update    grokBuildUpdate      `json:"update"`
		Meta      *grokBuildUpdateMeta `json:"_meta"`
	}

	grokBuildUpdate struct {
		SessionUpdate     string                  `json:"sessionUpdate"`
		Content           *grokBuildUpdateContent `json:"content"`
		Usage             *grokBuildUsage         `json:"usage"`
		Meta              *grokBuildUpdateMeta    `json:"_meta"`
		TargetPromptIndex *int                    `json:"target_prompt_index"`
	}

	grokBuildUpdateContent struct {
		Type *string               `json:"type"`
		Text *string               `json:"text"`
		Meta *grokBuildContentMeta `json:"_meta"`
	}

	grokBuildContentMeta struct {
		BashCommand *string `json:"bash_command"`
	}

	grokBuildUpdateMeta struct {
		ModelID          string `json:"modelId"`
		PromptIndex      *int   `json:"promptIndex"`
		AgentTimestampMS int64  `json:"agentTimestampMs"`
		HostTurn         *bool  `json:"hostTurn"`
	}

	grokBuildUsage struct {
		InputTokens  *int64 `json:"inputTokens"`
		OutputTokens *int64 `json:"outputTokens"`
	}

	grokBuildUpdateEvent struct {
		timestamp int64
		isXAI     bool
		params    grokBuildUpdateParams
	}

	grokBuildPendingPrompt struct {
		text        string
		promptIndex *int
		timestamp   time.Time
		model       string
		counts      bool
	}

	grokBuildPrompt struct {
		length       int
		timestamp    time.Time
		model        string
		inputTokens  int64
		outputTokens int64
	}

	grokBuildModelSnapshot struct {
		timestamp time.Time
		model     string
	}

	grokBuildHunk struct {
		HunkID        string    `json:"hunkId"`
		FilePath      string    `json:"filePath"`
		LinesAdded    int       `json:"linesAdded"`
		LinesRemoved  int       `json:"linesRemoved"`
		AuthorType    string    `json:"authorType"`
		EventType     string    `json:"eventType"`
		RemovalReason string    `json:"removalReason"`
		PromptIndex   *int      `json:"promptIndex"`
		Timestamp     time.Time `json:"timestamp"`
	}

	grokBuildParseState struct {
		prompts        []grokBuildPrompt
		hunkHeartbeats Heartbeats
		tokens         heartbeat.AITokens
		model          string
		// Prompt indexes can be reused after a rewind, so retain model history.
		modelsByPromptIndex map[int][]grokBuildModelSnapshot
		// Removal records omit author and model while negating the whole mixed-author hunk.
		hunkAIContributions map[string]map[string]int
		pending             *grokBuildPendingPrompt
		seenPromptIndex     bool
	}
)

// Parse parses Grok Build session transcripts for AI heartbeats.
func (g GrokBuild) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	grokHome, err := grokBuildHomeDir(ctx)
	if err != nil {
		return nil, err
	}

	sessionDirs, err := g.sessionDirs(ctx, grokHome)
	if err != nil {
		return nil, err
	}

	if len(sessionDirs) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d Grok Build sessions modified after %s", len(sessionDirs), g.After)
	version := grokBuildCLIVersion(logger, grokHome)

	var heartbeats Heartbeats

	for _, sessionDir := range sessionDirs {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		parsed, parseErr := g.parseSession(ctx, logger, grokHome, sessionDir, version)
		if parseErr != nil {
			if errors.Is(parseErr, context.Canceled) || errors.Is(parseErr, context.DeadlineExceeded) {
				return nil, parseErr
			}

			logger.Warnf("failed parsing Grok Build session %q: %s", sessionDir, parseErr)

			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	sort.SliceStable(heartbeats, func(i, j int) bool {
		return heartbeats[i].Time < heartbeats[j].Time
	})

	return heartbeats, nil
}

func (g GrokBuild) sessionDirs(ctx context.Context, grokHome string) ([]string, error) {
	sessionsRoot := filepath.Join(grokHome, "sessions")
	if _, err := os.Stat(sessionsRoot); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to stat Grok Build sessions directory: %s", err)
	}

	var sessionDirs []string

	err := filepath.WalkDir(sessionsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if entry.IsDir() || entry.Name() != "summary.json" {
			return nil
		}

		sessionDir := filepath.Dir(path)

		modified, modifiedErr := grokBuildSessionModifiedAtOrAfter(sessionDir, g.After)
		if modifiedErr != nil {
			return modifiedErr
		}

		if modified {
			sessionDirs = append(sessionDirs, sessionDir)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk Grok Build sessions directory: %s", err)
	}

	sort.Strings(sessionDirs)

	return sessionDirs, nil
}

func grokBuildSessionModifiedAtOrAfter(sessionDir string, after time.Time) (bool, error) {
	for _, name := range []string{"summary.json", "updates.jsonl", "hunk_records.jsonl"} {
		path := filepath.Join(sessionDir, name)

		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return false, fmt.Errorf("failed to stat Grok Build session file %q: %s", path, err)
		}

		if timestampAtOrAfterCutoff(info.ModTime(), after) {
			return true, nil
		}
	}

	return false, nil
}

func (g GrokBuild) parseSession(
	ctx context.Context,
	logger *log.Logger,
	grokHome string,
	sessionDir string,
	version string,
) (Heartbeats, error) {
	session, err := grokBuildReadSession(sessionDir, version)
	if err != nil {
		return nil, err
	}

	state := grokBuildParseState{
		model:               session.model,
		modelsByPromptIndex: make(map[int][]grokBuildModelSnapshot),
		hunkAIContributions: make(map[string]map[string]int),
	}

	if err := g.parseUpdates(ctx, logger, sessionDir, session, &state); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logger.Warnf("failed parsing Grok Build updates for session %q: %s", sessionDir, err)
	}

	if err := g.parseHunks(ctx, logger, grokHome, sessionDir, session, &state); err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		logger.Warnf("failed parsing Grok Build hunks for session %q: %s", sessionDir, err)
	}

	heartbeats := append(g.promptHeartbeats(session, state.prompts), state.hunkHeartbeats...)
	sort.SliceStable(heartbeats, func(i, j int) bool {
		return heartbeats[i].Time < heartbeats[j].Time
	})

	return heartbeats, nil
}

func grokBuildReadSession(sessionDir string, version string) (grokBuildSession, error) {
	path := filepath.Join(sessionDir, "summary.json")

	data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec
	if err != nil {
		return grokBuildSession{}, fmt.Errorf("failed to read Grok Build summary %q: %s", path, err)
	}

	var summary grokBuildSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		return grokBuildSession{}, fmt.Errorf("failed to parse Grok Build summary %q: %s", path, err)
	}

	sessionID := strings.TrimSpace(summary.Info.ID)
	if sessionID == "" {
		sessionID = filepath.Base(sessionDir)
	}

	cwd := firstNonEmptyString(strings.TrimSpace(summary.GitRootDir), strings.TrimSpace(summary.Info.Cwd))
	if cwd != "" {
		cwd = filepath.Clean(filepath.FromSlash(cwd))
	}

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
	defer grokBuildFlushPrompt(session, state)

	path := filepath.Join(sessionDir, "updates.jsonl")

	return grokBuildWalkJSONL(
		ctx,
		logger,
		path,
		"updates",
		func(line []byte) error {
			event, err := grokBuildDecodeUpdate(line)
			if err != nil {
				return err
			}

			g.handleUpdate(event, session, state)

			return nil
		},
		func() {
			grokBuildFlushPrompt(session, state)
		},
	)
}

func (g GrokBuild) handleUpdate(
	event grokBuildUpdateEvent,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	update := event.params.Update
	timestamp := grokBuildEventTime(event)
	g.trackUpdateModel(timestamp, update, session, state)

	switch {
	case !event.isXAI && update.SessionUpdate == "user_message_chunk":
		grokBuildHandleUserMessage(timestamp, update, session, state)
	case event.isXAI && update.SessionUpdate == "rewind_marker":
		grokBuildFlushPrompt(session, state)
	default:
		grokBuildFlushPrompt(session, state)

		if update.SessionUpdate == "turn_completed" {
			g.handleTurnCompleted(timestamp, update, state)
		}
	}
}

func (GrokBuild) trackUpdateModel(
	timestamp time.Time,
	update grokBuildUpdate,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	if update.Meta == nil {
		return
	}

	model := strings.TrimSpace(update.Meta.ModelID)
	if model != "" {
		state.model = model
	}

	if update.Meta.PromptIndex == nil {
		return
	}

	model = firstNonEmptyString(model, state.model, session.model)
	if model != "" {
		promptIndex := *update.Meta.PromptIndex

		history := state.modelsByPromptIndex[promptIndex]
		if len(history) == 0 || history[len(history)-1].model != model {
			state.modelsByPromptIndex[promptIndex] = append(history, grokBuildModelSnapshot{
				timestamp: timestamp,
				model:     model,
			})
		}
	}
}

func grokBuildHandleUserMessage(
	timestamp time.Time,
	update grokBuildUpdate,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	if !grokBuildIsPromptText(update) {
		grokBuildFlushPrompt(session, state)
		return
	}

	var promptIndex *int
	if update.Meta != nil {
		promptIndex = update.Meta.PromptIndex
	}

	if promptIndex != nil {
		state.seenPromptIndex = true
	}

	counts := !state.seenPromptIndex || promptIndex != nil

	newRun := state.pending == nil
	if state.pending != nil && (state.seenPromptIndex || promptIndex != nil) {
		newRun = !grokBuildSamePromptIndex(state.pending.promptIndex, promptIndex)
	}

	if newRun {
		grokBuildFlushPrompt(session, state)
		state.pending = &grokBuildPendingPrompt{
			promptIndex: promptIndex,
			timestamp:   timestamp,
			model:       firstNonEmptyString(state.model, session.model),
			counts:      counts,
		}
	}

	state.pending.text += *update.Content.Text

	if state.pending.timestamp.IsZero() && !timestamp.IsZero() {
		state.pending.timestamp = timestamp
	}

	if state.model != "" {
		state.pending.model = state.model
	}

	if state.pending.promptIndex == nil && promptIndex != nil {
		state.pending.promptIndex = promptIndex
		state.pending.counts = true
	}
}

func grokBuildIsPromptText(update grokBuildUpdate) bool {
	if update.Meta != nil && update.Meta.HostTurn != nil && *update.Meta.HostTurn {
		return false
	}

	if update.Content == nil || update.Content.Type == nil || update.Content.Text == nil {
		return false
	}

	if *update.Content.Type != "text" {
		return false
	}

	return update.Content.Meta == nil || update.Content.Meta.BashCommand == nil
}

func grokBuildFlushPrompt(session grokBuildSession, state *grokBuildParseState) {
	if state.pending == nil {
		return
	}

	pending := state.pending
	state.pending = nil

	if !pending.counts {
		return
	}

	text := strings.TrimSpace(pending.text)
	if text == "" {
		return
	}

	model := pending.model
	if pending.promptIndex != nil {
		model = firstNonEmptyString(
			grokBuildModelAtPromptIndex(*pending.promptIndex, pending.timestamp, state),
			model,
		)
	}

	state.prompts = append(state.prompts, grokBuildPrompt{
		length:    promptLength(text),
		timestamp: pending.timestamp,
		model:     firstNonEmptyString(model, state.model, session.model),
	})
}

func (g GrokBuild) handleTurnCompleted(
	timestamp time.Time,
	update grokBuildUpdate,
	state *grokBuildParseState,
) {
	if update.Usage == nil {
		return
	}

	inputTokens := grokBuildNonNegativeTokenCount(update.Usage.InputTokens)
	outputTokens := grokBuildNonNegativeTokenCount(update.Usage.OutputTokens)
	// Turn usage is incremental. Convert it to Codex-style cumulative state so
	// the heartbeat receives only the delta since the previous token event.
	state.tokens.CurrentInput = state.tokens.LastInput + inputTokens
	state.tokens.CurrentOutput = state.tokens.LastOutput + outputTokens

	inputDelta := state.tokens.CurrentInput - state.tokens.LastInput
	outputDelta := state.tokens.CurrentOutput - state.tokens.LastOutput

	if !timestamp.IsZero() && timestampAtOrAfterCutoff(timestamp, g.After) && len(state.prompts) > 0 {
		prompt := &state.prompts[len(state.prompts)-1]
		if timestampAtOrAfterCutoff(prompt.timestamp, g.After) {
			prompt.inputTokens += inputDelta
			prompt.outputTokens += outputDelta
		}
	}

	state.tokens.LastInput = state.tokens.CurrentInput
	state.tokens.LastOutput = state.tokens.CurrentOutput
}

func grokBuildNonNegativeTokenCount(value *int64) int64 {
	if value == nil || *value < 0 {
		return 0
	}

	return *value
}

func (g GrokBuild) promptHeartbeats(session grokBuildSession, prompts []grokBuildPrompt) Heartbeats {
	heartbeats := make(Heartbeats, 0, len(prompts))
	entity := appHeartbeatEntity(g.Name(), session.id)

	for _, prompt := range prompts {
		if prompt.timestamp.IsZero() || !timestampAtOrAfterCutoff(prompt.timestamp, g.After) {
			continue
		}

		h := heartbeat.NewWithAITokens(
			nil,
			session.id,
			heartbeat.AITokens{
				CurrentInput:  prompt.inputTokens,
				CurrentOutput: prompt.outputTokens,
			},
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
			heartbeatTimestamp(prompt.timestamp),
			g.userAgent(entity, prompt.model, session.version),
		)
		h.AIPromptLength = prompt.length
		heartbeats = append(heartbeats, h)
	}

	return heartbeats
}

func (g GrokBuild) parseHunks(
	ctx context.Context,
	logger *log.Logger,
	grokHome string,
	sessionDir string,
	session grokBuildSession,
	state *grokBuildParseState,
) error {
	path := filepath.Join(sessionDir, "hunk_records.jsonl")

	return grokBuildWalkJSONL(
		ctx,
		logger,
		path,
		"hunks",
		func(line []byte) error {
			var hunk grokBuildHunk
			if err := json.Unmarshal(line, &hunk); err != nil {
				return err
			}

			g.handleHunk(hunk, grokHome, session, state)

			return nil
		},
		nil,
	)
}

func (g GrokBuild) handleHunk(
	hunk grokBuildHunk,
	grokHome string,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	eventType := strings.ToLower(strings.TrimSpace(hunk.EventType))
	if eventType == "removed" {
		g.handleRemovedHunk(hunk, grokHome, session, state)
		return
	}

	if eventType != "added" && eventType != "updated" {
		return
	}

	if !strings.EqualFold(strings.TrimSpace(hunk.AuthorType), "agent") {
		return
	}

	filePath := grokBuildNativePath(hunk.FilePath)
	if filePath == "" || grokBuildShouldSkipPath(filePath, grokHome) || hunk.Timestamp.IsZero() {
		return
	}

	lineChanges := hunk.LinesAdded - hunk.LinesRemoved
	model := grokBuildHunkModel(hunk, session, state)
	hunkID := strings.TrimSpace(hunk.HunkID)

	if hunkID != "" {
		contributions := state.hunkAIContributions[hunkID]
		if contributions == nil {
			contributions = make(map[string]int)
			state.hunkAIContributions[hunkID] = contributions
		}

		contributions[model] += lineChanges
	}

	if timestampAtOrAfterCutoff(hunk.Timestamp, g.After) {
		state.hunkHeartbeats = append(
			state.hunkHeartbeats,
			g.hunkHeartbeat(filePath, lineChanges, hunk.Timestamp, model, session),
		)
	}
}

func (g GrokBuild) handleRemovedHunk(
	hunk grokBuildHunk,
	grokHome string,
	session grokBuildSession,
	state *grokBuildParseState,
) {
	hunkID := strings.TrimSpace(hunk.HunkID)
	contributions := state.hunkAIContributions[hunkID]
	delete(state.hunkAIContributions, hunkID)

	if strings.EqualFold(strings.TrimSpace(hunk.RemovalReason), "accepted") {
		return
	}

	if len(contributions) == 0 || hunk.Timestamp.IsZero() || !timestampAtOrAfterCutoff(hunk.Timestamp, g.After) {
		return
	}

	filePath := grokBuildNativePath(hunk.FilePath)
	if filePath == "" || grokBuildShouldSkipPath(filePath, grokHome) {
		return
	}

	models := make([]string, 0, len(contributions))
	for model, lineChanges := range contributions {
		if lineChanges != 0 {
			models = append(models, model)
		}
	}

	sort.Strings(models)

	for _, model := range models {
		state.hunkHeartbeats = append(
			state.hunkHeartbeats,
			g.hunkHeartbeat(filePath, -contributions[model], hunk.Timestamp, model, session),
		)
	}
}

func grokBuildHunkModel(
	hunk grokBuildHunk,
	session grokBuildSession,
	state *grokBuildParseState,
) string {
	if hunk.PromptIndex != nil {
		// Updated records retain the hunk's creation timestamp, so it cannot
		// identify which generation produced the update after a rewind reuses a
		// prompt index. Prefer the latest known generation for updated records.
		if strings.EqualFold(strings.TrimSpace(hunk.EventType), "updated") {
			if model := grokBuildLatestModelAtPromptIndex(*hunk.PromptIndex, state); model != "" {
				return model
			}
		}

		if model := grokBuildModelAtPromptIndex(*hunk.PromptIndex, hunk.Timestamp, state); model != "" {
			return model
		}
	}

	return firstNonEmptyString(state.model, session.model)
}

func grokBuildLatestModelAtPromptIndex(promptIndex int, state *grokBuildParseState) string {
	history := state.modelsByPromptIndex[promptIndex]
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].model != "" {
			return history[i].model
		}
	}

	return ""
}

func grokBuildModelAtPromptIndex(
	promptIndex int,
	timestamp time.Time,
	state *grokBuildParseState,
) string {
	history := state.modelsByPromptIndex[promptIndex]
	if len(history) == 0 {
		return ""
	}

	model := ""

	var matchedAt time.Time

	for _, snapshot := range history {
		if snapshot.timestamp.IsZero() {
			if model == "" {
				model = snapshot.model
			}

			continue
		}

		if !timestamp.IsZero() && snapshot.timestamp.After(timestamp) {
			continue
		}

		if matchedAt.IsZero() || snapshot.timestamp.After(matchedAt) || snapshot.timestamp.Equal(matchedAt) {
			model = snapshot.model
			matchedAt = snapshot.timestamp
		}
	}

	if model == "" {
		return history[0].model
	}

	return model
}

func (g GrokBuild) hunkHeartbeat(
	filePath string,
	lineChanges int,
	timestamp time.Time,
	model string,
	session grokBuildSession,
) heartbeat.Heartbeat {
	return heartbeat.NewWithAITokens(
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
		session.cwd,
		heartbeatTimestamp(timestamp),
		g.userAgent(filePath, model, session.version),
	)
}

func (g GrokBuild) userAgent(entity string, model string, version string) string {
	return aiUserAgentWithModelAndEditor(
		entity,
		g.UserAgents,
		g.FallbackUserAgent,
		model,
		"",
		grokBuildUserAgentProduct(version),
	)
}

func grokBuildUserAgentProduct(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		version = "unknown"
	}

	return "grok-build/" + version
}

func grokBuildDecodeUpdate(line []byte) (grokBuildUpdateEvent, error) {
	var envelope grokBuildUpdateEnvelope
	if err := json.Unmarshal(line, &envelope); err != nil {
		return grokBuildUpdateEvent{}, err
	}

	var params grokBuildUpdateParams
	if len(envelope.Params) > 0 && string(envelope.Params) != "null" {
		if err := json.Unmarshal(envelope.Params, &params); err != nil {
			return grokBuildUpdateEvent{}, err
		}

		if strings.TrimSpace(params.Update.SessionUpdate) == "" {
			return grokBuildUpdateEvent{}, errors.New("missing Grok Build session update type")
		}

		return grokBuildUpdateEvent{
			timestamp: envelope.Timestamp,
			isXAI:     envelope.Method == grokBuildXAIUpdateMethod,
			params:    params,
		}, nil
	}

	if err := json.Unmarshal(line, &params); err != nil {
		return grokBuildUpdateEvent{}, err
	}

	if strings.TrimSpace(params.Update.SessionUpdate) == "" {
		return grokBuildUpdateEvent{}, errors.New("missing Grok Build session update type")
	}

	return grokBuildUpdateEvent{params: params}, nil
}

func grokBuildEventTime(event grokBuildUpdateEvent) time.Time {
	meta := event.params.Meta
	if meta == nil {
		meta = event.params.Update.Meta
	}

	if meta != nil && meta.AgentTimestampMS > 0 {
		return time.UnixMilli(meta.AgentTimestampMS).UTC()
	}

	if event.timestamp > 1_000_000_000_000 {
		return time.UnixMilli(event.timestamp).UTC()
	}

	if event.timestamp > 0 {
		return time.Unix(event.timestamp, 0).UTC()
	}

	return time.Time{}
}

func grokBuildSamePromptIndex(a *int, b *int) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return *a == *b
}

func grokBuildNativePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	return filepath.Clean(filepath.FromSlash(path))
}

func grokBuildShouldSkipPath(path string, grokHome string) bool {
	cleanPath := filepath.Clean(path)
	cleanHome := filepath.Clean(grokHome)

	relative, err := filepath.Rel(cleanHome, cleanPath)
	if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return true
	}

	return strings.Contains(filepath.ToSlash(cleanPath), "/.grok/")
}

func grokBuildHomeDir(ctx context.Context) (string, error) {
	configured := strings.TrimSpace(os.Getenv("GROK_HOME"))
	if configured == "" {
		home, err := ini.UserHomeDir(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to find user home dir: %s", err)
		}

		return filepath.Join(home, ".grok"), nil
	}

	if filepath.IsAbs(configured) {
		return filepath.Clean(configured), nil
	}

	if configured == "~" || strings.HasPrefix(configured, "~/") || strings.HasPrefix(configured, `~\`) {
		home, err := ini.UserHomeDir(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to expand GROK_HOME: %s", err)
		}

		if configured == "~" {
			return home, nil
		}

		configured = filepath.Join(home, configured[2:])
	}

	resolved, err := filepath.Abs(configured)
	if err != nil {
		return "", fmt.Errorf("failed to resolve GROK_HOME %q: %s", configured, err)
	}

	return resolved, nil
}

func grokBuildCLIVersion(logger *log.Logger, grokHome string) string {
	path := filepath.Join(grokHome, "version.json")

	data, err := os.ReadFile(filepath.Clean(path)) //nolint:gosec
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Debugf("failed to read Grok Build version file %q: %s", path, err)
		}

		return ""
	}

	var version grokBuildVersion
	if err := json.Unmarshal(data, &version); err != nil {
		logger.Debugf("failed to parse Grok Build version file %q: %s", path, err)
		return ""
	}

	return firstNonEmptyString(strings.TrimSpace(version.Version), strings.TrimSpace(version.StableVersion))
}

func grokBuildWalkJSONL(
	ctx context.Context,
	logger *log.Logger,
	path string,
	kind string,
	handle func([]byte) error,
	handleSkipped func(),
) error {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("failed to open Grok Build %s %q: %s", kind, path, err)
	}
	defer fh.Close() //nolint:errcheck,gosec

	reader := bufio.NewReader(fh)

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		line, readErr := grokBuildReadJSONLLine(reader, maxTranscriptLineSize)
		if errors.Is(readErr, errGrokBuildLineTooLong) {
			logger.Warnf("skipping oversized Grok Build %s line in %q", kind, path)

			if handleSkipped != nil {
				handleSkipped()
			}

			continue
		}

		if errors.Is(readErr, io.EOF) && len(line) == 0 {
			break
		}

		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return fmt.Errorf("failed reading Grok Build %s %q: %s", kind, path, readErr)
		}

		if len(bytes.TrimSpace(line)) > 0 {
			if err := handle(line); err != nil {
				logger.Warnf("failed parsing Grok Build %s line from %q: %s", kind, path, err)

				if handleSkipped != nil {
					handleSkipped()
				}
			}
		}

		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	return nil
}

var errGrokBuildLineTooLong = errors.New("grok build JSONL line exceeds size limit")

func grokBuildReadJSONLLine(reader *bufio.Reader, maxSize int) ([]byte, error) {
	var line []byte

	for {
		fragment, err := reader.ReadSlice('\n')
		hasNewline := len(fragment) > 0 && fragment[len(fragment)-1] == '\n'

		content := fragment
		if hasNewline {
			content = fragment[:len(fragment)-1]
		}

		if len(line)+len(content) > maxSize {
			if !hasNewline {
				if discardErr := grokBuildDiscardJSONLLine(reader); discardErr != nil && !errors.Is(discardErr, io.EOF) {
					return nil, discardErr
				}
			}

			return nil, errGrokBuildLineTooLong
		}

		line = append(line, content...)
		if hasNewline {
			return bytesTrimSuffix(line, '\r'), nil
		}

		if errors.Is(err, io.EOF) {
			if len(line) == 0 {
				return nil, io.EOF
			}

			return bytesTrimSuffix(line, '\r'), io.EOF
		}

		if errors.Is(err, bufio.ErrBufferFull) {
			continue
		}

		if err != nil {
			return nil, err
		}
	}
}

func grokBuildDiscardJSONLLine(reader *bufio.Reader) error {
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

func bytesTrimSuffix(value []byte, suffix byte) []byte {
	if len(value) > 0 && value[len(value)-1] == suffix {
		return value[:len(value)-1]
	}

	return value
}

// Name returns its id.
func (GrokBuild) Name() string {
	return "Grok Build"
}
