package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
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
	codexSessionState struct {
		cwd      string
		entity   string
		id       string
		source   string
		version  string
		parentID string
		created  time.Time
		// internal is set for Codex's own background sessions, such as guardian
		// approval reviews, which are not user coding activity.
		internal bool
	}

	codexParseState struct {
		heartbeats                    Heartbeats
		usageHeartbeat                func(time.Time, int64, int64, int64) heartbeat.Heartbeat
		tokens                        heartbeat.AITokens
		lastUsageIdentity             string
		seenUsage                     map[string]bool
		usageOwner                    string
		pendingPatches                map[string]codexPendingPatch
		subscriptionPlan              string
		model                         string
		reasoningEffort               string
		lastAgentMessageTime          time.Time
		lastResponseItemAssistantTime time.Time
		lastResponseItemUserTime      time.Time
	}

	codexPendingPatch struct {
		inputs            []string
		version           string
		agentVersion      string
		source            string
		cwd               string
		userAgents        map[string]string
		fallbackUserAgent string
		sessionID         string
	}

	codexSessionMeta struct {
		Type      string    `json:"type"`
		Timestamp time.Time `json:"timestamp"`
		Payload   *struct {
			ID             *string         `json:"id"`
			Cwd            *string         `json:"cwd"`
			Source         json.RawMessage `json:"source"`
			ForkedFromID   string          `json:"forked_from_id"`
			ParentThreadID string          `json:"parent_thread_id"`
			ThreadSource   string          `json:"thread_source"`
			Version        *string         `json:"cli_version"`
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
		CallID  *string         `json:"call_id"`
		Success *bool           `json:"success"`
		Output  json.RawMessage `json:"output"`
	}

	codexPayloadRateLimits struct {
		PlanType *string `json:"plan_type"`
	}

	codexPayloadTokenCountInfo struct {
		LastTokenUsage     *codexPayloadTokenCountInfoUsage `json:"last_token_usage"`
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

type codexSeenUsageKey struct{}

type codexParentSourcesKey struct{}

// Parse parses the Codex JSONL session transcript logs for ai heartbeats.
func (g Codex) Parse(ctx context.Context) (Heartbeats, error) {
	ctx = context.WithValue(ctx, codexSeenUsageKey{}, make(map[string]bool))
	ctx = context.WithValue(ctx, codexParentSourcesKey{}, make(map[string]string))
	logger := log.Extract(ctx)

	transcripts, err := g.discoverTranscripts(ctx)
	if err != nil {
		return nil, err
	}

	g.pruneSources(transcripts)

	if len(transcripts) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d transcript logs modified after %s for %s", len(transcripts), g.After, g.Name())

	g.planTranscriptReads(transcripts)

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		if transcript.skip {
			logger.Debugf("skipping unchanged codex transcript %q", transcript.path)
			continue
		}

		parsed, err := g.parseTranscriptFrom(ctx, transcript.path, transcript.start)
		if err != nil {
			logger.Warnf("failed parsing codex transcript %q: %s", transcript.path, err)
			continue
		}

		g.recordSource(transcript)

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

// codexTranscript is a discovered transcript and how much of it to read.
type codexTranscript struct {
	path string
	// owner is the root session ID. Forks and subagents share their parent's
	// owner, and token usage is deduplicated across files with the same owner.
	owner  string
	source checkpointSource
	skip   bool
	start  int64
}

// planTranscriptReads uses the checkpoint's per-file sizes so a sync only
// reads what changed. A transcript seen for the first time is read in full,
// which keeps backfills and syncs after downtime complete. A transcript that
// grew resumes one line-size window before its previous end, so every new
// byte is read and the preceding context restores the model, cwd and token
// baselines. An unchanged transcript is skipped, unless another transcript
// with the same owner changed, since usage is deduplicated across those.
func (g Codex) planTranscriptReads(transcripts []codexTranscript) {
	if g.checkpoint == nil {
		return
	}

	changedOwners := make(map[string]bool)
	unchanged := make([]bool, len(transcripts))

	for i, transcript := range transcripts {
		previous, ok := g.checkpoint.Sources[transcript.path]
		if !ok || previous.Size > transcript.source.Size {
			// New, or rewritten shorter: read the whole file.
			changedOwners[transcript.owner] = true
			continue
		}

		if previous.Size == transcript.source.Size && previous.Modified.Equal(transcript.source.Modified) {
			unchanged[i] = true
		} else {
			changedOwners[transcript.owner] = true
		}

		transcripts[i].start = max(previous.Size-maxTranscriptLineSize, 0)
	}

	for i := range transcripts {
		transcripts[i].skip = unchanged[i] && !changedOwners[transcripts[i].owner]
	}
}

// pruneSources keeps the checkpoint file small. A transcript that is not
// discovered has not changed since the discovery cutoff, so it is not read;
// if it changes later it is simply read in full once.
func (g Codex) pruneSources(transcripts []codexTranscript) {
	if g.checkpoint == nil || len(g.checkpoint.Sources) == 0 {
		return
	}

	discovered := make(map[string]bool, len(transcripts))
	for _, transcript := range transcripts {
		discovered[transcript.path] = true
	}

	maps.DeleteFunc(g.checkpoint.Sources, func(path string, _ checkpointSource) bool { return !discovered[path] })
}

func (g Codex) recordSource(transcript codexTranscript) {
	if g.checkpoint == nil {
		return
	}

	if g.checkpoint.Sources == nil {
		g.checkpoint.Sources = make(map[string]checkpointSource)
	}

	// The snapshot was taken before reading, so a concurrent append is reread.
	g.checkpoint.Sources[transcript.path] = transcript.source
}

func codexRoots(ctx context.Context) ([]string, error) {
	sessionsDir, err := codexSessionsDir(ctx)
	if err != nil {
		return nil, err
	}

	roots := []string{filepath.Dir(sessionsDir)}

	// Launcher discovery is optional: CODEX_HOME must work without a user home.
	if home, err := ini.UserHomeDir(ctx); err == nil {
		// Launcher homes can retain unique sessions alongside copies of ~/.codex.
		primary := filepath.Join(home, ".codex")
		for _, launcher := range []string{filepath.Join(home, ".buzz")} {
			relative, err := filepath.Rel(launcher, roots[0])
			if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				roots = append([]string{primary}, roots...)
				break
			}
		}
	}

	for _, home := range wslHomes(ctx) {
		roots = append(roots, filepath.Join(home, ".codex"))
	}

	return roots, nil
}

func (g Codex) transcriptPaths(ctx context.Context) ([]string, error) {
	transcripts, err := g.discoverTranscripts(ctx)
	if err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(transcripts))
	for _, transcript := range transcripts {
		paths = append(paths, transcript.path)
	}

	return paths, nil
}

func (g Codex) discoverTranscripts(ctx context.Context) ([]codexTranscript, error) {
	roots, err := codexRoots(ctx)
	if err != nil {
		return nil, err
	}

	var transcripts []codexTranscript

	seen := make(map[string]bool)

	for _, root := range roots {
		for _, dir := range []string{"sessions", "archived_sessions"} {
			base := filepath.Join(root, dir)
			if _, err := os.Stat(base); os.IsNotExist(err) {
				continue
			}

			err := filepath.WalkDir(base, func(path string, entry os.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}

				if entry.IsDir() || filepath.Ext(path) != ".jsonl" {
					return nil
				}

				info, err := entry.Info()
				if err != nil || !timestampAtOrAfterCutoff(info.ModTime(), g.After) {
					return nil
				}

				id := g.sessionIDFromPath(path)

				file, err := os.Open(filepath.Clean(path)) // nolint:gosec
				if err != nil {
					return err
				}

				meta, err := g.readSessionState(log.Extract(ctx), file, path)
				_ = file.Close()

				owner := id

				if err == nil && meta.id != "" {
					id = meta.id
					owner = firstNonEmptyString(meta.parentID, meta.id)
				}

				if !seen[id] {
					seen[id] = true

					transcripts = append(transcripts, codexTranscript{
						path:   path,
						owner:  owner,
						source: checkpointSource{Size: info.Size(), Modified: info.ModTime()},
					})
				}

				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}

	return transcripts, nil
}

// parentSource returns the source of the nearest ancestor session that has
// one, following parent links through nested subagents.
func (g Codex) parentSource(ctx context.Context, parentID string) string {
	cache, _ := ctx.Value(codexParentSourcesKey{}).(map[string]string)

	roots, err := codexRoots(ctx)
	if err != nil {
		return ""
	}

	visited := make(map[string]bool)

	for id := parentID; id != "" && !visited[id] && len(visited) < 8; {
		visited[id] = true

		if source, ok := cache[id]; ok {
			return source
		}

		session, found := g.findSession(log.Extract(ctx), roots, id)
		if !found {
			return ""
		}

		if session.source != "" {
			if cache != nil {
				cache[parentID] = session.source
			}

			return session.source
		}

		id = session.parentID
	}

	return ""
}

func (g Codex) findSession(logger *log.Logger, roots []string, id string) (codexSessionState, bool) {
	if strings.ContainsAny(id, `*?[]\/`) {
		return codexSessionState{}, false
	}

	name := "rollout-*-" + id + ".jsonl"

	for _, root := range roots {
		for _, pattern := range []string{
			filepath.Join(root, "sessions", "*", "*", "*", name),
			filepath.Join(root, "archived_sessions", name),
		} {
			matches, _ := filepath.Glob(pattern)
			for _, match := range matches {
				fh, err := os.Open(filepath.Clean(match)) // nolint:gosec
				if err != nil {
					continue
				}

				session, err := g.readSessionState(logger, fh, match)
				_ = fh.Close()

				if err == nil {
					return session, true
				}
			}
		}
	}

	return codexSessionState{}, false
}

func codexSessionsDir(ctx context.Context) (string, error) {
	if configuredHome := os.Getenv("CODEX_HOME"); configuredHome != "" {
		sessionsDir, err := filepath.Abs(filepath.Join(configuredHome, "sessions"))
		if err != nil {
			return "", fmt.Errorf("failed to resolve CODEX_HOME: %s", err)
		}

		return sessionsDir, nil
	}

	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	return filepath.Join(home, ".codex", "sessions"), nil
}

func (g Codex) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	return g.parseTranscriptFrom(ctx, transcript, 0)
}

// parseTranscriptFrom parses transcript starting at the first full line at or
// after start. Session metadata always comes from the first line.
func (g Codex) parseTranscriptFrom(ctx context.Context, transcript string, start int64) (Heartbeats, error) {
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

	g.After = ParserConfig(g).sessionAfter(session.id)

	if session.internal {
		logger.Debugf("skipping internal codex session %q", transcript)
		return nil, nil
	}

	if session.source == "" && session.parentID != "" {
		// Subagents only record their parent thread, so inherit its editor.
		session.source = g.parentSource(ctx, session.parentID)
	}

	reader, err := codexReader(fh, transcript, start)
	if err != nil {
		return nil, err
	}

	seen, _ := ctx.Value(codexSeenUsageKey{}).(map[string]bool)
	state := codexParseState{seenUsage: seen, usageOwner: firstNonEmptyString(session.parentID, session.id),
		pendingPatches: make(map[string]codexPendingPatch),
	}

	if g.checkpoint != nil {
		state.usageHeartbeat = func(timestamp time.Time, input, cached, output int64) heartbeat.Heartbeat {
			event := genericAIEvent{sessionID: session.id, timestamp: timestamp, cwd: session.cwd,
				model: state.model, input: input, cachedInput: cached, output: output, tokensFound: true}
			h := genericAIEventHeartbeats(g, ParserConfig(g), event)[0]
			h.Entity = session.entity
			h.UserAgent = g.userAgent(session.entity, codexAgentVersion(state.model, state.reasoningEffort),
				session.version, session.source, g.UserAgents, g.FallbackUserAgent)

			return h
		}
	}

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Lines over the size limit, such as huge tool outputs, are skipped
		// rather than ending the scan, so later activity is still read.
		line, readErr := grokBuildReadJSONLLine(reader, maxTranscriptLineSize)
		if errors.Is(readErr, errGrokBuildLineTooLong) {
			logger.Debugf("skipping oversized codex transcript line in %q", transcript)
			continue
		}

		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return nil, fmt.Errorf("failed reading codex transcript %q: %s", transcript, readErr)
		}

		g.handleTranscriptLine(logger, transcript, line, &session, &state)

		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	state.applySubscriptionPlan()

	return translateWSLHeartbeats(transcript, state.heartbeats), nil
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

func codexReader(fh *os.File, transcript string, start int64) (*bufio.Reader, error) {
	if start <= 0 {
		if _, err := fh.Seek(0, io.SeekStart); err != nil {
			return nil, fmt.Errorf("failed to seek codex transcript %q: %s", transcript, err)
		}

		return bufio.NewReaderSize(fh, 64*1024), nil
	}

	// Start one byte early and discard through the next newline. That drops a
	// partial line, but keeps a full line that begins exactly at start.
	if _, err := fh.Seek(start-1, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek codex transcript %q: %s", transcript, err)
	}

	reader := bufio.NewReaderSize(fh, 64*1024)
	if err := grokBuildDiscardJSONLLine(reader); err != nil && !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
	}

	return reader, nil
}

func (g Codex) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	session *codexSessionState,
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

	// A turn can run in a different directory than the session started in,
	// for example a worktree, so relative patch paths follow the latest cwd.
	if logLine.Type == "turn_context" && logLine.Payload != nil && logLine.Payload.Cwd != nil &&
		strings.TrimSpace(*logLine.Payload.Cwd) != "" {
		session.cwd = strings.TrimSpace(*logLine.Payload.Cwd)
	}

	after := g.After
	if session.parentID != "" && session.created.After(after) {
		after = session.created
	}

	state.trackTokenCount(logLine, after)

	if session.parentID != "" && !session.created.IsZero() && logLine.Timestamp.Before(session.created) {
		state.trackModel(logLine)
		return
	}

	state.trackSubscriptionPlan(logLine)
	state.trackModel(logLine)
	state.trackUserMessage(logLine)

	if patchHeartbeats, handled := g.handlePendingPatch(logLine, *session, state); handled {
		if !logLine.Timestamp.IsZero() && timestampAtOrAfterCutoff(logLine.Timestamp, g.After) {
			state.heartbeats = append(state.heartbeats, patchHeartbeats...)
		}

		return
	}

	if logLine.Timestamp.IsZero() ||
		!timestampAtOrAfterCutoff(logLine.Timestamp, g.After) ||
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
		heartbeat.AITokens{},
	)
	if len(aiHeartbeats) == 0 {
		return
	}

	state.heartbeats = append(state.heartbeats, aiHeartbeats...)
	state.trackAgentMessage(logLine)
	state.trackAssistantMessage(logLine)
}

func (s *codexParseState) trackTokenCount(logLine codexLogLine, after time.Time) {
	if logLine.Payload == nil || logLine.Payload.Type == nil || *logLine.Payload.Type != "token_count" ||
		logLine.Payload.Info == nil {
		return
	}

	info := logLine.Payload.Info

	identity, _ := json.Marshal(info)
	if string(identity) == s.lastUsageIdentity {
		return
	}

	s.lastUsageIdentity = string(identity)
	usage := info.TotalTokenUsage

	cumulative := usage != nil && usage.InputTokens != nil && usage.OutputTokens != nil
	if !cumulative {
		usage = info.LastTokenUsage
	}

	if usage == nil {
		return
	}

	value := func(p *int) int64 {
		if p == nil {
			return 0
		}

		return max(int64(*p), 0)
	}
	cachedInputTokens := value(usage.CachedInputTokens)
	input := max(value(usage.InputTokens)-cachedInputTokens, 0)

	output := value(usage.OutputTokens)
	if cumulative {
		s.tokens.CurrentInput, s.tokens.CurrentCachedInput, s.tokens.CurrentOutput = input, cachedInputTokens, output
	} else {
		s.tokens.CurrentInput += input
		s.tokens.CurrentCachedInput += cachedInputTokens
		s.tokens.CurrentOutput += output
	}

	duplicate := false

	if cumulative && s.seenUsage != nil && !logLine.Timestamp.IsZero() &&
		timestampAtOrAfterCutoff(logLine.Timestamp, after) {
		raw, _ := json.Marshal(usage)
		key := s.usageOwner + ":" + logLine.Timestamp.Format(time.RFC3339Nano) + ":" + string(raw)
		duplicate = s.seenUsage[key]
		s.seenUsage[key] = true
	}

	if duplicate || logLine.Timestamp.IsZero() || !timestampAtOrAfterCutoff(logLine.Timestamp, after) {
		s.tokens.LastInput = s.tokens.CurrentInput
		s.tokens.LastCachedInput = s.tokens.CurrentCachedInput
		s.tokens.LastOutput = s.tokens.CurrentOutput

		return
	}

	inputTokens := s.tokens.CurrentInput - s.tokens.LastInput
	if inputTokens < 0 {
		inputTokens = 0
	}

	cachedInputTokens = s.tokens.CurrentCachedInput - s.tokens.LastCachedInput
	if cachedInputTokens < 0 {
		cachedInputTokens = 0
	}

	outputTokens := s.tokens.CurrentOutput - s.tokens.LastOutput
	if outputTokens < 0 {
		outputTokens = 0
	}

	if s.usageHeartbeat != nil && (inputTokens > 0 || cachedInputTokens > 0 || outputTokens > 0) {
		// Usage can arrive in a later sync than the activity that triggered it.
		// Give it its own timestamp so advancing a session cannot discard it.
		s.heartbeats = append(s.heartbeats,
			s.usageHeartbeat(logLine.Timestamp, inputTokens, cachedInputTokens, outputTokens))
	} else if s.usageHeartbeat == nil && len(s.heartbeats) > 0 {
		i := len(s.heartbeats) - 1
		s.heartbeats[i].AIInputTokens += inputTokens
		s.heartbeats[i].AICachedInputTokens += cachedInputTokens
		s.heartbeats[i].AIOutputTokens += outputTokens
	}

	s.tokens.LastInput = s.tokens.CurrentInput
	s.tokens.LastCachedInput = s.tokens.CurrentCachedInput
	s.tokens.LastOutput = s.tokens.CurrentOutput
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

// Codex reasoning effort is an inference setting, not part of the billable
// model identity. Keep the user-agent model token stable so server-side model
// pricing can match it even when the effort changes within a session.
func codexAgentVersion(model string, _ string) string {
	model = strings.TrimSpace(model)

	if model == "" {
		return ""
	}

	return model
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

	state.parentID = firstNonEmptyString(sessionMeta.Payload.ForkedFromID, sessionMeta.Payload.ParentThreadID)
	state.created = sessionMeta.Timestamp
	state.internal = sessionMeta.Payload.ThreadSource == "guardian_review"

	var source string
	if json.Unmarshal(sessionMeta.Payload.Source, &source) == nil {
		state.source = source
		return
	}

	// Object sources look like {"internal":"guardian"} or {"subagent":...},
	// where subagent is a string ("review") or an object such as
	// {"thread_spawn":{"parent_thread_id":"..."}} or {"other":"guardian"}.
	var nested struct {
		Internal string          `json:"internal"`
		Subagent json.RawMessage `json:"subagent"`
	}
	if json.Unmarshal(sessionMeta.Payload.Source, &nested) != nil {
		return
	}

	if nested.Internal != "" {
		state.internal = true
	}

	var subagent struct {
		ThreadSpawn struct {
			ParentThreadID string `json:"parent_thread_id"`
		} `json:"thread_spawn"`
		Other string `json:"other"`
	}
	if len(nested.Subagent) > 0 && json.Unmarshal(nested.Subagent, &subagent) == nil {
		state.parentID = firstNonEmptyString(state.parentID, subagent.ThreadSpawn.ParentThreadID)

		if strings.EqualFold(strings.TrimSpace(subagent.Other), "guardian") {
			state.internal = true
		}
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

func (g Codex) handlePendingPatch(
	logLine codexLogLine,
	session codexSessionState,
	state *codexParseState,
) (Heartbeats, bool) {
	if logLine.Payload == nil || logLine.Payload.Type == nil {
		return nil, false
	}

	payload := *logLine.Payload
	if payload.CallID == nil || strings.TrimSpace(*payload.CallID) == "" {
		return nil, false
	}

	callID := strings.TrimSpace(*payload.CallID)

	inputs := codexPatchInputs(payload)
	if len(inputs) > 0 {
		state.pendingPatches[callID] = codexPendingPatch{
			inputs:            inputs,
			version:           session.version,
			agentVersion:      codexAgentVersion(state.model, state.reasoningEffort),
			source:            session.source,
			cwd:               session.cwd,
			userAgents:        g.UserAgents,
			fallbackUserAgent: g.FallbackUserAgent,
			sessionID:         session.id,
		}

		return nil, true
	}

	pending, ok := state.pendingPatches[callID]
	if !ok {
		return nil, false
	}

	switch *payload.Type {
	case "patch_apply_end":
		delete(state.pendingPatches, callID)

		if payload.Success == nil || !*payload.Success {
			return nil, true
		}
	case "custom_tool_call_output":
		delete(state.pendingPatches, callID)

		if !codexToolCallSucceeded(payload.Output) {
			return nil, true
		}
	default:
		return nil, false
	}

	var heartbeats Heartbeats
	for _, input := range pending.inputs {
		heartbeats = append(heartbeats, g.patchHeartbeats(
			logLine.Timestamp,
			pending.version,
			pending.agentVersion,
			pending.source,
			pending.cwd,
			pending.userAgents,
			pending.fallbackUserAgent,
			input,
			pending.sessionID,
			heartbeat.AITokens{},
		)...)
	}

	return heartbeats, true
}

func codexToolCallSucceeded(output json.RawMessage) bool {
	if len(output) == 0 || string(output) == "null" {
		return true
	}

	var items []codexContentItem
	if err := json.Unmarshal(output, &items); err != nil {
		var text string
		if err := json.Unmarshal(output, &text); err != nil {
			return true
		}

		items = []codexContentItem{{Text: text}}
	}

	texts := make([]string, 0, len(items))

	for _, item := range items {
		// Desktop wraps tool results as {"content":[...],"isError":false}.
		var structured struct {
			IsError *bool              `json:"isError"`
			Content []codexContentItem `json:"content"`
		}
		if json.Unmarshal([]byte(strings.TrimSpace(item.Text)), &structured) == nil && structured.IsError != nil {
			if *structured.IsError {
				return false
			}

			for _, content := range structured.Content {
				texts = append(texts, content.Text)
			}

			continue
		}

		texts = append(texts, item.Text)
	}

	// apply_patch reports this once the patch is on disk, so a later failing
	// command in the same exec script must not discard the write.
	for _, text := range texts {
		if strings.Contains(strings.ToLower(text), "success. updated the following files") {
			return true
		}
	}

	for _, text := range texts {
		text = strings.ToLower(text)
		if strings.Contains(text, "failed") ||
			strings.Contains(text, "error") ||
			strings.Contains(text, "invalid context") ||
			strings.Contains(text, "invalid patch") {
			return false
		}
	}

	return true
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
	if !strings.Contains(input, "tools.apply_patch(") {
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

		patches = append(patches, codexDecodePatch(encoded))
	}

	return patches
}

func codexDecodePatch(encoded string) string {
	var decoded strings.Builder
	decoded.Grow(len(encoded))

	for i := 0; i < len(encoded); i++ {
		if encoded[i] != '\\' || i+1 >= len(encoded) {
			decoded.WriteByte(encoded[i])
			continue
		}

		next := encoded[i+1]
		switch next {
		case 'n':
			decoded.WriteByte('\n')

			i++
		case 'r':
			decoded.WriteByte('\r')

			i++
		case 't':
			decoded.WriteByte('\t')

			i++
		case 'b':
			decoded.WriteByte('\b')

			i++
		case 'f':
			decoded.WriteByte('\f')

			i++
		case 'v':
			decoded.WriteByte('\v')

			i++
		case '\\', '"', '\'', '/':
			decoded.WriteByte(next)

			i++
		case 'x':
			if value, consumed, ok := codexDecodeHexEscape(encoded[i+2:], 2); ok {
				decoded.WriteRune(value)

				i += consumed + 1
			} else {
				decoded.WriteByte(encoded[i])
			}
		case 'u':
			if value, consumed, ok := codexDecodeHexEscape(encoded[i+2:], 4); ok {
				decoded.WriteRune(value)

				i += consumed + 1
			} else {
				decoded.WriteByte(encoded[i])
			}
		default:
			decoded.WriteByte(encoded[i])
		}
	}

	return decoded.String()
}

func codexDecodeHexEscape(encoded string, size int) (rune, int, bool) {
	if len(encoded) < size {
		return 0, 0, false
	}

	var value rune

	for i := 0; i < size; i++ {
		var digit rune

		switch char := encoded[i]; {
		case char >= '0' && char <= '9':
			digit = rune(char - '0')
		case char >= 'a' && char <= 'f':
			digit = rune(char-'a') + 10
		case char >= 'A' && char <= 'F':
			digit = rune(char-'A') + 10
		default:
			return 0, 0, false
		}

		value = value*16 + digit
	}

	return value, size, true
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
		isDeleted   bool
		additions   int
		deletions   int
	)

	lines := strings.Split(input, "\n")
	for i := range lines {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		if moveFile := codexMoveFilePath(cwd, line); moveFile != "" {
			if currentFile != "" {
				currentFile = moveFile
			}

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
				heartbeats[len(heartbeats)-1].IsUnsavedEntity = isDeleted
				tokens.LastInput = tokens.CurrentInput
				tokens.LastCachedInput = tokens.CurrentCachedInput
				tokens.LastOutput = tokens.CurrentOutput
			}

			currentFile = codexFilePath(cwd, line)
			isDeleted = strings.HasPrefix(line, "*** Delete File: ")
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
		heartbeats[len(heartbeats)-1].IsUnsavedEntity = isDeleted
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
		heartbeatTimestamp(timestamp),
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
		heartbeatTimestamp(timestamp),
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

func codexMoveFilePath(cwd string, line string) string {
	file, ok := strings.CutPrefix(line, "*** Move to: ")
	if !ok {
		return ""
	}

	if file == "" || filepath.IsAbs(file) || strings.HasPrefix(file, "/") {
		return file
	}

	return filepath.Join(cwd, file)
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

		if version == "" {
			version = "unknown"
		}

		return "codex-" + product + "/" + version
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
		heartbeatTimestamp(timestamp),
		g.userAgent(currentFile, agentVersion, version, source, userAgents, fallbackUserAgent),
	)
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
