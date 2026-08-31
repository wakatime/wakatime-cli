package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// claudeTaskOutputPathPattern matches Claude Code's per-session sub-agent task
// output files, e.g. /private/tmp/claude-501/-Users-foo-bar/<session-uuid>/tasks/<id>.output.
// These are internal scratch artifacts produced by the Task tool, not user code,
// so we ignore them when emitting file heartbeats. Paths are normalized to forward
// slashes before matching so the same pattern works on Windows.
var claudeTaskOutputPathPattern = regexp.MustCompile(`(?:^|/)claude-\d+/(?:[^/]+/)+tasks/[^/]+\.output$`)

// Claude contains params for detecting heartbeats from Claude transcripts.
type Claude ParserConfig

type (
	claudeUsage struct {
		InputTokens              *int `json:"input_tokens"`
		CacheCreationInputTokens *int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     *int `json:"cache_read_input_tokens"`
		OutputTokens             *int `json:"output_tokens"`
		TotalTokens              *int `json:"total_tokens"`
	}

	claudeMessage struct {
		ID      string                   `json:"id"`
		Role    string                   `json:"role"`
		Model   string                   `json:"model"`
		Effort  string                   `json:"effort"`
		Usage   *claudeUsage             `json:"usage"`
		Content claudeMessageContentList `json:"content"`
	}

	claudeMessageContent struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	claudeMessageContentList []claudeMessageContent

	structuredPatch struct {
		NewLines int `json:"newLines"`
		OldLines int `json:"oldLines"`
	}

	contentValue struct {
		String *string
		Array  *[]json.RawMessage
	}

	toolUseResultFile struct {
		Content  *contentValue `json:"content"`
		FilePath *string       `json:"filePath"`
	}

	toolUseResult struct {
		Type            *string            `json:"type"`
		File            *toolUseResultFile `json:"file"`
		Content         *contentValue      `json:"content"`
		Stdout          *contentValue      `json:"stdout"`
		Stderr          *contentValue      `json:"stderr"`
		Result          *contentValue      `json:"result"`
		CodeText        *contentValue      `json:"codeText"`
		FilePath        *string            `json:"filePath"`
		OriginalFile    *string            `json:"originalFile"`
		OldString       *string            `json:"oldString"`
		NewString       *string            `json:"newString"`
		StructuredPatch *[]structuredPatch `json:"structuredPatch"`
		Raw             map[string]json.RawMessage
	}

	toolUseResultValue struct {
		Object *toolUseResult
		String *string
		Array  *contentValue
	}

	claudeLogLine struct {
		Timestamp     time.Time           `json:"timestamp"`
		SessionID     string              `json:"sessionId"`
		Version       string              `json:"version"`
		ToolUseResult *toolUseResultValue `json:"toolUseResult"`
		Usage         *claudeUsage        `json:"usage"`
		Message       *claudeMessage      `json:"message"`
		IsSideChain   *bool               `json:"isSidechain"`
		PromptID      *string             `json:"promptId"`
		Type          *string             `json:"type"`
		Cwd           *string             `json:"cwd"`
	}

	claudeConfig struct {
		OAuthAccount *claudeOAuthAccount `json:"oauthAccount"`
	}

	claudeOAuthAccount struct {
		OrganizationType string `json:"organizationType"`
		SubscriptionType string `json:"subscriptionType"`
	}

	// claudeLastMessage tracks the most recent message's token contribution
	// so that streaming duplicates (same message.id logged multiple times)
	// replace rather than accumulate.
	claudeLastMessage struct {
		ID                string
		InputTokens       int64
		CachedInputTokens int64
		OutputTokens      int64
	}
)

func (m *claudeMessage) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := claudeJSONUnmarshal(data, &raw); err != nil {
		return err
	}

	claudeDecodeField(raw, "id", &m.ID)
	claudeDecodeField(raw, "role", &m.Role)
	claudeDecodeField(raw, "model", &m.Model)
	claudeDecodeField(raw, "effort", &m.Effort)
	claudeDecodeField(raw, "usage", &m.Usage)
	claudeDecodeField(raw, "content", &m.Content)

	return nil
}

func (v *contentValue) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}

	var str string
	if err := claudeJSONUnmarshal(data, &str); err == nil {
		v.String = &str

		return nil
	}

	var arr []json.RawMessage
	if err := claudeJSONUnmarshal(data, &arr); err == nil {
		v.Array = &arr

		return nil
	}

	return nil
}

func (r *toolUseResult) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := claudeJSONUnmarshal(data, &raw); err != nil {
		return err
	}

	r.Raw = raw
	claudeDecodeField(raw, "type", &r.Type)
	claudeDecodeField(raw, "file", &r.File)
	claudeDecodeField(raw, "content", &r.Content)
	claudeDecodeField(raw, "stdout", &r.Stdout)
	claudeDecodeField(raw, "stderr", &r.Stderr)
	claudeDecodeField(raw, "result", &r.Result)
	claudeDecodeField(raw, "codeText", &r.CodeText)
	claudeDecodeField(raw, "filePath", &r.FilePath)
	claudeDecodeField(raw, "originalFile", &r.OriginalFile)
	claudeDecodeField(raw, "oldString", &r.OldString)
	claudeDecodeField(raw, "newString", &r.NewString)
	claudeDecodeField(raw, "structuredPatch", &r.StructuredPatch)

	return nil
}

func (v *toolUseResultValue) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}

	var result toolUseResult
	if err := claudeJSONUnmarshal(data, &result); err == nil {
		v.Object = &result

		return nil
	}

	var str string
	if err := claudeJSONUnmarshal(data, &str); err == nil {
		v.String = &str

		return nil
	}

	var arr []json.RawMessage
	if err := claudeJSONUnmarshal(data, &arr); err == nil {
		v.Array = &contentValue{Array: &arr}

		return nil
	}

	return nil
}

func (c *claudeMessageContentList) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}

	var rawItems []any
	if err := claudeJSONUnmarshal(data, &rawItems); err == nil {
		items := make([]claudeMessageContent, 0, len(rawItems))

		for _, raw := range rawItems {
			content, ok := parseClaudeMessageContent(raw)
			if ok {
				items = append(items, content)
			}
		}

		*c = items

		return nil
	}

	var str string
	if err := claudeJSONUnmarshal(data, &str); err == nil {
		*c = claudeMessageContentList{{Type: "text", Text: str}}

		return nil
	}

	return fmt.Errorf("unsupported message content type")
}

func parseClaudeMessageContent(raw any) (claudeMessageContent, bool) {
	switch value := raw.(type) {
	case string:
		return claudeMessageContent{Type: "text", Text: value}, true
	case map[string]any:
		content := claudeMessageContent{}
		if contentType, ok := value["type"].(string); ok {
			content.Type = contentType
		}

		if text, ok := value["text"].(string); ok {
			content.Text = text
		}

		return content, true
	default:
		return claudeMessageContent{}, false
	}
}

func (l *claudeLogLine) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := claudeJSONUnmarshal(data, &raw); err != nil {
		return err
	}

	claudeDecodeField(raw, "timestamp", &l.Timestamp)
	claudeDecodeField(raw, "sessionId", &l.SessionID)
	claudeDecodeField(raw, "version", &l.Version)
	claudeDecodeField(raw, "toolUseResult", &l.ToolUseResult)
	claudeDecodeField(raw, "usage", &l.Usage)
	claudeDecodeField(raw, "message", &l.Message)
	claudeDecodeField(raw, "isSidechain", &l.IsSideChain)
	claudeDecodeField(raw, "promptId", &l.PromptID)
	claudeDecodeField(raw, "type", &l.Type)
	claudeDecodeField(raw, "cwd", &l.Cwd)

	return nil
}

// claudeDecodeField decodes one optional transcript field independently. Claude
// Code's JSONL schema evolves frequently, so a new shape for one field must not
// discard other usable data from the same line.
func claudeDecodeField(raw map[string]json.RawMessage, key string, value any) {
	data, ok := raw[key]
	if !ok {
		return
	}

	_ = claudeJSONUnmarshal(data, value)
}

func claudeJSONUnmarshal(data []byte, v any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("json decoder panicked: %v", r)
		}
	}()

	return json.Unmarshal(data, v)
}

func (v *contentValue) lineChanges() int {
	if v == nil {
		return 0
	}

	if v.String != nil {
		return countStringLines(*v.String)
	}

	if v.Array == nil {
		return 0
	}

	lineChanges := 0

	for _, item := range *v.Array {
		var str string
		if err := claudeJSONUnmarshal(item, &str); err == nil {
			lineChanges += countStringLines(str)

			continue
		}

		var block struct {
			Text *string `json:"text"`
		}
		if err := claudeJSONUnmarshal(item, &block); err == nil && block.Text != nil {
			lineChanges += countStringLines(*block.Text)
		}
	}

	return lineChanges
}

// Parse parses the Claude JSONL session transcript logs for ai heartbeats.
func (g Claude) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	state := g.loadTranscriptState(ctx)

	transcripts, err := g.transcriptPaths(ctx, state)
	if err != nil {
		return nil, err
	}

	if len(transcripts) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d transcript logs modified after %s for %s", len(transcripts), g.After, g.Name())

	subscriptionPlan := g.subscriptionPlan(ctx)

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		cutoff, exclusive := g.transcriptCutoff(state, transcript)

		parsed, err := g.parseTranscript(ctx, transcript, cutoff, exclusive)
		if err != nil {
			logger.Warnf("failed parsing claude transcript %q: %s", transcript, err)
			continue
		}

		if last, ok := claudeNewestHeartbeatTime(parsed); ok {
			state[transcript] = last
		}

		heartbeats = append(heartbeats, parsed...)
	}

	g.saveTranscriptState(ctx, state)

	return claudeApplySubscriptionPlan(heartbeats, subscriptionPlan), nil
}

// transcriptCutoff returns the timestamp from which a transcript should be
// parsed. Transcripts seen before are resumed right after the newest
// heartbeat already generated from them (exclusive), so token usage logged
// after that heartbeat is picked up by the next one instead of being lost.
// Unseen transcripts fall back to the global cutoff (inclusive).
func (g Claude) transcriptCutoff(state claudeTranscriptState, transcript string) (time.Time, bool) {
	if last, ok := state[transcript]; ok && !last.IsZero() {
		return last, true
	}

	return g.After, false
}

func (g Claude) transcriptPaths(ctx context.Context, state claudeTranscriptState) ([]string, error) {
	configDirs, err := claudeConfigDirs(ctx)
	if err != nil {
		return nil, err
	}

	var transcripts []string

	for _, dir := range configDirs {
		claudeProjectsDir := filepath.Join(dir, "projects")

		info, err := os.Stat(claudeProjectsDir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to read claude projects directory %q: %s", claudeProjectsDir, err)
		}

		if !info.IsDir() {
			continue
		}

		err = filepath.WalkDir(claudeProjectsDir, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if d.IsDir() || filepath.Ext(d.Name()) != ".jsonl" {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return nil
			}

			cutoff, exclusive := g.transcriptCutoff(state, path)
			if !claudeAfterCutoff(info.ModTime(), cutoff, exclusive) {
				return nil
			}

			transcripts = append(transcripts, path)

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed walking claude projects directory %q: %s", claudeProjectsDir, err)
		}
	}

	return transcripts, nil
}

func (g Claude) subscriptionPlan(ctx context.Context) string {
	logger := log.Extract(ctx)

	paths, err := g.configPaths(ctx)
	if err != nil {
		logger.Debugf("failed finding claude config files: %s", err)

		return ""
	}

	for _, path := range paths {
		plan, err := g.subscriptionPlanFromConfig(path)
		if err != nil {
			logger.Debugf("failed reading claude subscription plan from %q: %s", path, err)

			continue
		}

		if plan != "" {
			return plan
		}
	}

	return ""
}

func (Claude) configPaths(ctx context.Context) ([]string, error) {
	configDirs, err := claudeConfigDirs(ctx)
	if err != nil {
		return nil, err
	}

	home, err := ini.UserHomeDir(ctx)
	if err != nil && os.Getenv("CLAUDE_CONFIG_DIR") == "" {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	var defaultConfigDir string
	if home != "" {
		defaultConfigDir = filepath.Join(home, ".claude")
	}

	paths := make([]string, 0, len(configDirs))

	for _, dir := range configDirs {
		if defaultConfigDir != "" && dir == defaultConfigDir {
			paths = append(paths, filepath.Join(home, ".claude.json"))
		} else {
			paths = append(paths, filepath.Join(dir, ".claude.json"))
		}

		backups, err := claudeConfigBackupPaths(dir)
		if err != nil {
			return nil, err
		}

		paths = append(paths, backups...)
	}

	return paths, nil
}

func claudeConfigDirs(ctx context.Context) ([]string, error) {
	if raw := os.Getenv("CLAUDE_CONFIG_DIR"); raw != "" {
		dirs := make([]string, 0, strings.Count(raw, ",")+1)
		seen := make(map[string]bool)

		for _, dir := range strings.Split(raw, ",") {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}

			dir = filepath.Clean(dir)
			if seen[dir] {
				continue
			}

			seen[dir] = true
			dirs = append(dirs, dir)
		}

		if len(dirs) > 0 {
			return dirs, nil
		}
	}

	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	return []string{filepath.Join(home, ".claude")}, nil
}

func claudeConfigBackupPaths(configDir string) ([]string, error) {
	backupDir := filepath.Join(configDir, "backups")

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var paths []string

	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), ".claude.json.backup.") {
			continue
		}

		paths = append(paths, filepath.Join(backupDir, entry.Name()))
	}

	return paths, nil
}

func (Claude) subscriptionPlanFromConfig(path string) (string, error) {
	//nolint:gosec
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	var config claudeConfig
	if err := claudeJSONUnmarshal(data, &config); err != nil {
		return "", err
	}

	if config.OAuthAccount == nil {
		return "", nil
	}

	if plan := claudeNormalizeSubscriptionPlan(config.OAuthAccount.SubscriptionType); plan != "" {
		return plan, nil
	}

	return claudeNormalizeSubscriptionPlan(config.OAuthAccount.OrganizationType), nil
}

func (g Claude) parseTranscript(
	ctx context.Context,
	transcript string,
	cutoff time.Time,
	exclusive bool,
) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open claude transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat claude transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek claude transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read claude transcript %q: %s", transcript, err)
		}
	}

	var (
		heartbeats Heartbeats
		tokens     heartbeat.AITokens
		lastMsg    claudeLastMessage
	)

	claudeVersion := ""
	model := ""
	complexity := ""
	cwd := ""
	cwdFromTranscript := false
	ideSession := false
	sessionID := g.sessionIDFromPath(transcript)
	sessionEntity := appHeartbeatEntity("Claude", transcript)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var logLine claudeLogLine
		if err := claudeJSONUnmarshal(line, &logLine); err != nil {
			logger.Warnf("failed parsing claude transcript line from %q: %s", transcript, err)
			logger.Debugf("failed parsing claude transcript line: %s", line)

			continue
		}

		if logLine.Version != "" {
			claudeVersion = logLine.Version
		}

		if logLine.Message != nil {
			model = firstNonEmptyString(logLine.Message.Model, model)
			complexity = firstNonEmptyString(logLine.Message.Effort, complexity)
		}

		if logLine.SessionID != "" {
			sessionID = logLine.SessionID
		}

		if lineCwd := g.projectPath(logLine, !cwdFromTranscript); lineCwd != "" {
			cwd = lineCwd
			cwdFromTranscript = logLine.Cwd != nil && *logLine.Cwd != ""
		}

		if claudeHasIDEContext(logLine) {
			ideSession = true
		}

		tokens = g.claudeTokenCounts(logLine, tokens, &lastMsg)

		if logLine.Timestamp.IsZero() || !claudeAfterCutoff(logLine.Timestamp, cutoff, exclusive) {
			tokens = g.advanceTokens(tokens)
			continue
		}

		parsed := g.claudeHeartbeats(
			logLine,
			sessionEntity,
			sessionID,
			claudeVersion,
			model,
			complexity,
			cwd,
			cwdFromTranscript,
			ideSession,
			tokens,
		)
		if len(parsed) == 0 {
			if g.shouldAdvanceTokensForNoopToolResult(logLine.ToolUseResult) {
				tokens = g.advanceTokens(tokens)
			}

			continue
		}

		tokens = g.advanceTokens(tokens)

		heartbeats = append(heartbeats, parsed...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading claude transcript %q: %s", transcript, err)
	}

	return heartbeats, nil
}

func claudeNormalizeSubscriptionPlan(value string) string {
	plan := strings.ToLower(strings.TrimSpace(value))
	plan = strings.ReplaceAll(plan, "-", "_")
	plan = strings.TrimPrefix(plan, "claude_")
	plan = strings.TrimPrefix(plan, "claude ")

	switch {
	case plan == "pro" || strings.HasPrefix(plan, "pro_"):
		return "pro"
	case plan == "max" || strings.HasPrefix(plan, "max_"):
		return "max"
	default:
		return ""
	}
}

func claudeApplySubscriptionPlan(heartbeats Heartbeats, subscriptionPlan string) Heartbeats {
	if subscriptionPlan == "" {
		return heartbeats
	}

	for i := range heartbeats {
		heartbeats[i].AISubscriptionPlan = subscriptionPlan
	}

	return heartbeats
}

func (Claude) shouldAdvanceTokensForNoopToolResult(result *toolUseResultValue) bool {
	if result == nil || result.String != nil {
		return false
	}

	return true
}

func (g Claude) claudeTokenCounts(
	line claudeLogLine,
	previous heartbeat.AITokens,
	lastMsg *claudeLastMessage,
) heartbeat.AITokens {
	if line.Message != nil && line.Message.Usage != nil {
		return g.messageTokenCounts(*line.Message, previous, lastMsg)
	}

	if line.Usage == nil {
		return previous
	}

	current := previous

	if line.Usage.InputTokens != nil || line.Usage.CacheCreationInputTokens != nil {
		current.CurrentInput = claudeTokenValue(line.Usage.InputTokens) +
			claudeTokenValue(line.Usage.CacheCreationInputTokens)
	}

	if line.Usage.CacheReadInputTokens != nil {
		current.CurrentCachedInput = claudeTokenValue(line.Usage.CacheReadInputTokens)
	}

	switch {
	case line.Usage.OutputTokens != nil:
		current.CurrentOutput = int64(*line.Usage.OutputTokens)
	case line.Usage.TotalTokens != nil:
		current.CurrentOutput = int64(*line.Usage.TotalTokens)
	}

	return current
}

func (Claude) messageTokenCounts(
	msg claudeMessage,
	previous heartbeat.AITokens,
	lastMsg *claudeLastMessage,
) heartbeat.AITokens {
	usage := msg.Usage
	current := previous

	inputTokens := claudeTokenValue(usage.InputTokens) + claudeTokenValue(usage.CacheCreationInputTokens)
	cachedInputTokens := claudeTokenValue(usage.CacheReadInputTokens)

	outputTokens := int64(0)

	switch {
	case usage.OutputTokens != nil:
		outputTokens = int64(*usage.OutputTokens)
	case usage.TotalTokens != nil:
		outputTokens = int64(*usage.TotalTokens)
	}

	if msg.ID != "" && msg.ID == lastMsg.ID {
		// Same message (streaming update): replace previous contribution with latest.
		current.CurrentInput += inputTokens - lastMsg.InputTokens
		current.CurrentCachedInput += cachedInputTokens - lastMsg.CachedInputTokens
		current.CurrentOutput += outputTokens - lastMsg.OutputTokens
	} else {
		// New message: accumulate onto the running total.
		current.CurrentInput += inputTokens
		current.CurrentCachedInput += cachedInputTokens
		current.CurrentOutput += outputTokens
	}

	lastMsg.ID = msg.ID
	lastMsg.InputTokens = inputTokens
	lastMsg.CachedInputTokens = cachedInputTokens
	lastMsg.OutputTokens = outputTokens

	return current
}

func claudeTokenValue(value *int) int64 {
	if value == nil || *value < 0 {
		return 0
	}

	return int64(*value)
}

func (g Claude) claudeHeartbeats(
	logLine claudeLogLine,
	sessionEntity string,
	sessionID string,
	version string,
	model string,
	complexity string,
	cwd string,
	cwdFromTranscript bool,
	ideSession bool,
	tokens heartbeat.AITokens,
) Heartbeats {
	var heartbeats Heartbeats

	assignTokens := g.hasTokenDelta(tokens)

	appTokens := g.tokensForFirstHeartbeat(assignTokens, tokens)
	if heartbeat := g.claudeAppHeartbeat(
		logLine,
		sessionEntity,
		sessionID,
		version,
		model,
		complexity,
		cwd,
		ideSession,
		appTokens,
	); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
		assignTokens = false
	}

	fileTokens := g.tokensForFirstHeartbeat(assignTokens, tokens)
	if heartbeat := g.claudeFileHeartbeat(
		logLine,
		sessionID,
		version,
		model,
		complexity,
		cwd,
		cwdFromTranscript,
		ideSession,
		fileTokens,
	); heartbeat != nil {
		heartbeats = append(heartbeats, *heartbeat)
	}

	return heartbeats
}

func (g Claude) claudeAppHeartbeat(
	logLine claudeLogLine,
	sessionEntity string,
	sessionID string,
	version string,
	model string,
	complexity string,
	cwd string,
	ideSession bool,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	lineChanges := g.appLineChanges(logLine.ToolUseResult)

	promptLength := claudePromptLength(logLine)
	if lineChanges == 0 && promptLength == 0 {
		return nil
	}

	h := g.newHeartbeat(
		nil,
		sessionID,
		tokens,
		sessionEntity,
		heartbeat.AppType,
		heartbeat.PointerTo(false),
		cwd,
		heartbeatTimestamp(logLine.Timestamp),
		g.userAgent(sessionEntity, version, model, complexity, ideSession),
	)
	h.AIPromptLength = promptLength

	return &h
}

func (g Claude) claudeFileHeartbeat(
	logLine claudeLogLine,
	sessionID string,
	version string,
	model string,
	complexity string,
	cwd string,
	cwdFromTranscript bool,
	ideSession bool,
	tokens *heartbeat.AITokens,
) *heartbeat.Heartbeat {
	if logLine.ToolUseResult == nil || logLine.ToolUseResult.Object == nil {
		return nil
	}

	filePath := g.getFilePath(*logLine.ToolUseResult.Object)
	if filePath == "" {
		return nil
	}

	if isClaudeTaskOutputFilePath(filePath) {
		return nil
	}

	lineChanges := g.lineChanges(*logLine.ToolUseResult.Object)
	isWrite := g.isWrite(*logLine.ToolUseResult.Object)

	projectPathOverride := ""
	if cwdFromTranscript {
		projectPathOverride = cwd
	}

	h := g.newHeartbeat(
		heartbeat.PointerTo(lineChanges),
		sessionID,
		tokens,
		filePath,
		heartbeat.FileType,
		heartbeat.PointerTo(isWrite),
		projectPathOverride,
		heartbeatTimestamp(logLine.Timestamp),
		g.userAgent(filePath, version, model, complexity, ideSession),
	)

	return &h
}

func (g Claude) userAgent(entity string, version string, model string, complexity string, ideSession bool) string {
	modelToken := aiModelUserAgentToken(model, complexity)
	cli := "claude-code/" + unknownIfEmpty(version)

	if !ideSession && len(g.UserAgents) > 0 {
		return strings.TrimSpace(modelToken + " " + cli)
	}

	return aiUserAgentWithModelAndEditor(
		entity, g.UserAgents, g.FallbackUserAgent, model, complexity, cli)
}

func (Claude) tokenDelta(tokens heartbeat.AITokens) (int64, int64) {
	input := tokens.CurrentInput - tokens.LastInput
	if input < 0 {
		input = 0
	}

	output := tokens.CurrentOutput - tokens.LastOutput
	if output < 0 {
		output = 0
	}

	return input, output
}

func (g Claude) hasTokenDelta(tokens heartbeat.AITokens) bool {
	input, output := g.tokenDelta(tokens)
	cachedInput := tokens.CurrentCachedInput - tokens.LastCachedInput

	return input > 0 || cachedInput > 0 || output > 0
}

func (Claude) advanceTokens(tokens heartbeat.AITokens) heartbeat.AITokens {
	tokens.LastInput = tokens.CurrentInput
	tokens.LastCachedInput = tokens.CurrentCachedInput
	tokens.LastOutput = tokens.CurrentOutput

	return tokens
}

func (Claude) tokensForFirstHeartbeat(assign bool, tokens heartbeat.AITokens) *heartbeat.AITokens {
	if !assign {
		return nil
	}

	copy := tokens

	return &copy
}

func (Claude) newHeartbeat(
	aiLineChanges *int,
	aiSession string,
	aiTokens *heartbeat.AITokens,
	entity string,
	entityType heartbeat.EntityType,
	isWrite *bool,
	projectPathOverride string,
	timestamp float64,
	userAgent string,
) heartbeat.Heartbeat {
	if aiTokens != nil {
		return heartbeat.NewWithAITokens(
			aiLineChanges,
			aiSession,
			*aiTokens,
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			entity,
			entityType,
			nil,
			false,
			isWrite,
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			projectPathOverride,
			timestamp,
			userAgent,
		)
	}

	h := heartbeat.New(
		aiLineChanges,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		entity,
		entityType,
		nil,
		false,
		isWrite,
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		projectPathOverride,
		timestamp,
		userAgent,
	)
	h.AISession = aiSession

	return h
}

func (Claude) getFilePath(result toolUseResult) string {
	if result.FilePath != nil {
		return *result.FilePath
	}

	if result.File != nil && result.File.FilePath != nil {
		return *result.File.FilePath
	}

	return ""
}

func (g Claude) projectPath(logLine claudeLogLine, fallbackToFilePath bool) string {
	if logLine.Cwd != nil && *logLine.Cwd != "" {
		return *logLine.Cwd
	}

	if !fallbackToFilePath {
		return ""
	}

	result := logLine.ToolUseResult

	if result == nil || result.Object == nil {
		return ""
	}

	filePath := g.getFilePath(*result.Object)
	if filePath == "" {
		return ""
	}

	if isClaudeTaskOutputFilePath(filePath) {
		return ""
	}

	return filepath.Dir(filePath)
}

// isClaudeTaskOutputFilePath reports whether path points at a Claude Code
// sub-agent task output artifact under /tmp/claude-<uid>/.../tasks/*.output.
func isClaudeTaskOutputFilePath(path string) bool {
	if path == "" {
		return false
	}

	// Normalize backslashes from Windows-style paths so the pattern matches
	// regardless of the host OS the CLI is running on.
	normalized := strings.ReplaceAll(filepath.ToSlash(path), `\`, "/")

	return claudeTaskOutputPathPattern.MatchString(normalized)
}

func (Claude) lineChanges(result toolUseResult) int {
	if result.StructuredPatch != nil && len(*result.StructuredPatch) > 0 {
		lineChanges := 0
		for _, patch := range *result.StructuredPatch {
			lineChanges += patch.NewLines - patch.OldLines
		}

		return lineChanges
	}

	if result.NewString != nil {
		newLines := countStringLines(*result.NewString)
		if result.OldString != nil {
			return newLines - countStringLines(*result.OldString)
		}

		return newLines
	}

	// originalFile with content means this is a read/verify, not a write.
	// An empty string (from creates where no original existed) is not a read.
	if result.OriginalFile != nil && *result.OriginalFile != "" {
		return 0
	}

	// Use top-level Content for writes/creates.
	if lineChanges := result.Content.lineChanges(); lineChanges != 0 {
		return lineChanges
	}

	// File subfield represents read results (Read tool), not writes.
	// Don't count File.Content as line changes.

	return 0
}

func (Claude) isWrite(result toolUseResult) bool {
	if result.StructuredPatch != nil && len(*result.StructuredPatch) > 0 {
		return true
	}

	if result.Type != nil {
		switch *result.Type {
		case "create", "update", "delete":
			return true
		}
	}

	if result.NewString != nil || result.OldString != nil {
		return true
	}

	if result.OriginalFile != nil && *result.OriginalFile != "" {
		return false
	}

	return result.Content.lineChanges() != 0
}

func (g Claude) appLineChanges(result *toolUseResultValue) int {
	if result == nil {
		return 0
	}

	if result.String != nil && strings.TrimSpace(*result.String) != "" {
		return countStringLines(*result.String)
	}

	if result.Array != nil {
		return result.Array.lineChanges()
	}

	if result.Object == nil {
		return 0
	}

	if result.Object.isAgenticOnly() {
		return 0
	}

	if g.getFilePath(*result.Object) != "" {
		return 0
	}

	return result.Object.appLineChanges()
}

func (r toolUseResult) isAgenticOnly() bool {
	if r.Type != nil || r.File != nil || r.FilePath != nil || r.OriginalFile != nil ||
		r.OldString != nil || r.NewString != nil || r.StructuredPatch != nil ||
		r.Stdout != nil || r.Stderr != nil || r.Result != nil || r.CodeText != nil {
		return false
	}

	for _, key := range []string{
		"agentId",
		"agentType",
		"matches",
		"query",
		"results",
		"statusChange",
		"task",
		"taskId",
		"total_deferred_tools",
		"updatedFields",
		"verificationNudgeNeeded",
	} {
		if _, ok := r.Raw[key]; ok {
			return true
		}
	}

	return false
}

func (r toolUseResult) appLineChanges() int {
	lineChanges := 0

	for _, value := range []*contentValue{
		r.Content,
		r.Stdout,
		r.Stderr,
		r.Result,
		r.CodeText,
	} {
		lineChanges += value.lineChanges()
	}

	if r.File != nil {
		lineChanges += r.File.Content.lineChanges()
	}

	return lineChanges
}

func claudePromptLength(logLine claudeLogLine) int {
	if logLine.IsSideChain != nil && *logLine.IsSideChain {
		return 0
	}

	if logLine.Type == nil || *logLine.Type != "user" || logLine.Message == nil {
		return 0
	}

	if !strings.EqualFold(logLine.Message.Role, "user") {
		return 0
	}

	total := 0

	for _, item := range logLine.Message.Content {
		if item.Type != "text" {
			continue
		}

		total += claudePromptTextLength(item.Text)
	}

	return total
}

func claudePromptTextLength(text string) int {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return 0
	}

	if !strings.HasPrefix(trimmed, "<") {
		return promptLength(text)
	}

	for strings.HasPrefix(trimmed, "<") {
		closeOpenTag := strings.Index(trimmed, ">")
		if closeOpenTag < 2 {
			return 0
		}

		tag := strings.TrimSpace(trimmed[1:closeOpenTag])
		if tag == "" || strings.HasPrefix(tag, "/") {
			return 0
		}

		tagName := strings.Fields(tag)[0]

		tagName = strings.TrimSuffix(tagName, "/")
		if tagName == "" {
			return 0
		}

		if strings.HasSuffix(tag, "/") {
			trimmed = strings.TrimSpace(trimmed[closeOpenTag+1:])
			continue
		}

		closeTag := "</" + tagName + ">"

		closeTagIndex := strings.Index(trimmed, closeTag)
		if closeTagIndex == -1 {
			return 0
		}

		trimmed = strings.TrimSpace(trimmed[closeTagIndex+len(closeTag):])
	}

	return promptLength(trimmed)
}

func claudeHasIDEContext(logLine claudeLogLine) bool {
	if logLine.Message == nil {
		return false
	}

	for _, item := range logLine.Message.Content {
		if item.Type != "text" {
			continue
		}

		if strings.Contains(item.Text, "<ide_") {
			return true
		}
	}

	return false
}

func (Claude) sessionIDFromPath(path string) string {
	base := filepath.Base(path)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// claudeTranscriptStateFile stores, per transcript path, the timestamp of the
// newest heartbeat already generated from that transcript. A single global
// cutoff (ai_logs_last_parsed_at) is the newest heartbeat across all
// transcripts, so with several concurrent sessions every line of the slower
// sessions that predates it would be skipped, and its token usage dropped.
const claudeTranscriptStateFile = "ai-claude-transcripts.json"

type claudeTranscriptState map[string]time.Time

// claudeAfterCutoff reports whether timestamp falls inside the parsing
// window. Exclusive cutoffs come from per-transcript state where the cutoff
// line itself was already turned into heartbeats.
func claudeAfterCutoff(timestamp time.Time, cutoff time.Time, exclusive bool) bool {
	if exclusive {
		return timestamp.After(cutoff)
	}

	return timestampAtOrAfterCutoff(timestamp, cutoff)
}

func claudeNewestHeartbeatTime(heartbeats Heartbeats) (time.Time, bool) {
	var newest time.Time

	for _, h := range heartbeats {
		if t := heartbeatTime(h.Time); t.After(newest) {
			newest = t
		}
	}

	return newest, !newest.IsZero()
}

func (Claude) transcriptStatePath(ctx context.Context) (string, error) {
	dir, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, claudeTranscriptStateFile), nil
}

func (g Claude) loadTranscriptState(ctx context.Context) claudeTranscriptState {
	logger := log.Extract(ctx)
	state := claudeTranscriptState{}

	path, err := g.transcriptStatePath(ctx)
	if err != nil {
		logger.Debugf("failed to resolve claude transcript state path: %s", err)
		return state
	}

	data, err := os.ReadFile(path) // nolint:gosec
	if err != nil {
		if !os.IsNotExist(err) {
			logger.Debugf("failed to read claude transcript state %q: %s", path, err)
		}

		return state
	}

	raw := map[string]string{}
	if err := json.Unmarshal(data, &raw); err != nil {
		logger.Warnf("failed to parse claude transcript state %q: %s", path, err)
		return state
	}

	for transcript, value := range raw {
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil {
			continue
		}

		state[transcript] = parsed
	}

	return state
}

func (g Claude) saveTranscriptState(ctx context.Context, state claudeTranscriptState) {
	logger := log.Extract(ctx)

	path, err := g.transcriptStatePath(ctx)
	if err != nil {
		logger.Debugf("failed to resolve claude transcript state path: %s", err)
		return
	}

	raw := make(map[string]string, len(state))

	for transcript, last := range state {
		// drop transcripts that were deleted since the last run
		if _, err := os.Stat(transcript); err != nil {
			continue
		}

		raw[transcript] = last.UTC().Format(time.RFC3339Nano)
	}

	data, err := json.Marshal(raw)
	if err != nil {
		logger.Warnf("failed to encode claude transcript state: %s", err)
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		logger.Warnf("failed to create claude transcript state dir: %s", err)
		return
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		logger.Warnf("failed to write claude transcript state %q: %s", tmp, err)
		return
	}

	if err := os.Rename(tmp, path); err != nil {
		logger.Warnf("failed to replace claude transcript state %q: %s", path, err)
	}
}

// Name returns its name.
func (Claude) Name() string {
	return "Claude"
}
