package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

type genericAIProvider struct {
	parser             Parser
	config             ParserConfig
	roots              []string
	fileNames          map[string]bool
	extensions         map[string]bool
	sqliteRoots        []string
	containerKey       string
	preferMessagesFile bool
	tokenCounterMode   genericAICounterMode
	lineCounterMode    genericAICounterMode
}

type genericAIEvent struct {
	timestamp   time.Time
	sessionID   string
	model       string
	cwd         string
	prompt      string
	filePath    string
	toolName    string
	input       int64
	cachedInput int64
	output      int64
	lineChanges *int
	isWrite     bool
	tokensFound bool
}

type genericAISessionState struct {
	event       genericAIEvent
	input       int64
	cachedInput int64
	output      int64
}

var genericAICWDPattern = regexp.MustCompile(
	`(?i)(?:current working directory|working directory|cwd)\s*(?:is|:|=)\s*["']?([^\r\n"'<]+)`,
)

func parseGenericAIProvider(ctx context.Context, provider genericAIProvider) (Heartbeats, error) {
	paths, err := provider.transcriptPaths()
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	logger := log.Extract(ctx)

	for _, path := range paths {
		parsed, err := parseGenericAITranscript(ctx, provider, path)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}

			logger.Warnf("failed parsing %s transcript %q: %s", provider.parser.Name(), path, err)

			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	sqliteHeartbeats, err := parseGenericAISQLite(ctx, provider.parser, provider.config, provider.sqliteRoots)
	if err != nil {
		return nil, err
	}

	return append(heartbeats, sqliteHeartbeats...), nil
}

func (g genericAIProvider) transcriptPaths() ([]string, error) {
	seen := make(map[string]bool)

	var paths []string

	for _, root := range g.roots {
		if root == "" {
			continue
		}

		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to stat %s data path %q: %s", g.parser.Name(), root, err)
		}

		if !info.IsDir() {
			if g.accepts(root) && timestampAtOrAfterCutoff(info.ModTime(), g.config.After) {
				paths = append(paths, root)
			}

			continue
		}

		err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() || !g.accepts(path) || seen[path] {
				return nil
			}

			info, err := entry.Info()
			if err != nil || !timestampAtOrAfterCutoff(info.ModTime(), g.config.After) {
				return nil
			}

			seen[path] = true
			paths = append(paths, path)

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk %s data path %q: %s", g.parser.Name(), root, err)
		}
	}

	return paths, nil
}

func (g genericAIProvider) accepts(path string) bool {
	name := filepath.Base(path)
	if g.preferMessagesFile && strings.HasSuffix(name, ".json") {
		if strings.HasSuffix(name, ".messages.json") {
			return true
		}

		messagesPath := strings.TrimSuffix(path, ".json") + ".messages.json"
		if _, err := os.Stat(messagesPath); err == nil {
			return !genericAITranscriptHasTokens(messagesPath)
		}
	}

	if g.fileNames[name] {
		return true
	}

	return g.extensions[strings.ToLower(filepath.Ext(name))]
}

func genericAITranscriptHasTokens(path string) bool {
	//nolint:gosec // The path is a sibling under a discovered provider directory.
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return false
	}

	for _, value := range genericAIJSONValues(contents) {
		if object, ok := value.(map[string]any); ok {
			for _, child := range genericAIContainerValues(object, "messages") {
				if genericAIInt64(genericAIFind(child, genericAIInputKeys()...)) > 0 ||
					genericAIInt64(genericAIFind(child, genericAIOutputKeys()...)) > 0 {
					return true
				}
			}
		}
	}

	return false
}

func parseGenericAITranscript(ctx context.Context, provider genericAIProvider, path string) (Heartbeats, error) {
	if strings.EqualFold(filepath.Ext(path), ".jsonl") {
		values, err := genericAIJSONLines(ctx, provider.parser.Name(), path)
		if err != nil {
			return nil, err
		}

		return genericAIHeartbeats(ctx, provider, path, values)
	}

	//nolint:gosec // The path was discovered under the provider's data directory.
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed reading transcript: %s", err)
	}

	values := genericAIJSONValues(contents)
	if len(values) == 1 && provider.containerKey != "" {
		if object, ok := values[0].(map[string]any); ok {
			values = genericAIContainerValues(object, provider.containerKey)
		}
	}

	if len(values) == 0 {
		return genericAIPlaintextHeartbeats(provider.parser, provider.config, path, contents), nil
	}

	return genericAIHeartbeats(ctx, provider, path, values)
}

func genericAIJSONLines(ctx context.Context, providerName string, path string) ([]any, error) {
	//nolint:gosec // The path was discovered under the provider's data directory.
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed opening transcript: %s", err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed stating transcript: %s", err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed seeking transcript: %s", err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed reading transcript: %s", err)
		}
	}

	logger := log.Extract(ctx)

	var values []any

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var value any
		if err := json.Unmarshal(line, &value); err != nil {
			logger.Warnf("failed parsing %s transcript line from %q: %s", providerName, path, err)
			logger.Debugf("failed parsing transcript line: %s", line)

			continue
		}

		values = append(values, value)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading transcript: %s", err)
	}

	return values, nil
}

func genericAIJSONValues(contents []byte) []any {
	var document any
	if json.Unmarshal(contents, &document) == nil {
		if values, ok := document.([]any); ok {
			return values
		}

		return []any{document}
	}

	var values []any

	scanner := bufio.NewScanner(bytes.NewReader(contents))
	scanner.Buffer(make([]byte, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		var value any
		if json.Unmarshal(scanner.Bytes(), &value) == nil {
			values = append(values, value)
		}
	}

	return values
}

func genericAIContainerValues(object map[string]any, key string) []any {
	values, ok := object[key].([]any)
	if !ok {
		return []any{object}
	}

	sessionID := genericAIString(genericAIFind(object, genericAISessionKeys()...))
	cwd := genericAIString(genericAIFind(object, genericAICWDKeys()...))

	for i, value := range values {
		child, ok := value.(map[string]any)
		if !ok {
			continue
		}

		if sessionID != "" && genericAIString(genericAIFind(child, genericAISessionKeys()...)) == "" {
			child["session_id"] = sessionID
		}

		if cwd != "" && genericAIString(genericAIFind(child, genericAICWDKeys()...)) == "" {
			child["cwd"] = cwd
		}

		values[i] = child
	}

	return values
}

func genericAIPlaintextHeartbeats(
	parser Parser,
	config ParserConfig,
	path string,
	contents []byte,
) Heartbeats {
	if strings.ToLower(filepath.Ext(path)) != ".txt" || len(bytes.TrimSpace(contents)) == 0 {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil || !timestampAtOrAfterCutoff(info.ModTime(), config.After) {
		return nil
	}

	text := string(contents)

	prompt := text
	if start := strings.Index(text, "<user_query>"); start >= 0 {
		start += len("<user_query>")
		if end := strings.Index(text[start:], "</user_query>"); end >= 0 {
			prompt = text[start : start+end]
		}
	}

	event := genericAIEvent{
		timestamp: info.ModTime(),
		sessionID: genericAISessionID(path),
		prompt:    strings.TrimSpace(prompt),
		input:     int64(len([]rune(prompt)) / 4),
		output:    int64(len([]rune(text)) / 4),
	}

	return genericAIEventHeartbeats(parser, config, event)
}

func genericAIHeartbeats(
	ctx context.Context,
	provider genericAIProvider,
	path string,
	values []any,
) (Heartbeats, error) {
	defaultSessionID := genericAISessionID(path)
	states := make(map[string]genericAISessionState)
	lineCounters := make(map[string]int)
	fallbackTimestamp := time.Time{}

	if len(values) == 1 {
		if info, err := os.Stat(path); err == nil {
			fallbackTimestamp = info.ModTime()
		}
	}

	var heartbeats Heartbeats

	for _, value := range values {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		event := genericAIEventFromValue(value)
		if event.timestamp.IsZero() {
			event.timestamp = fallbackTimestamp
		}

		if event.sessionID == "" {
			event.sessionID = defaultSessionID
		}

		state := states[event.sessionID]

		previous := state.event
		if event.cwd == "" {
			event.cwd = previous.cwd
		}

		if event.model == "" {
			event.model = previous.model
		}

		if event.filePath != "" && event.cwd != "" && !genericAIPathIsAbs(event.filePath) {
			event.filePath = filepath.Join(event.cwd, filepath.FromSlash(event.filePath))
		}

		insideCutoff := !event.timestamp.IsZero() &&
			timestampAtOrAfterCutoff(event.timestamp, provider.config.After)

		if provider.tokenCounterMode == genericAICumulativeCounters && event.hasTokens() {
			currentInput := event.input
			currentCachedInput := event.cachedInput
			currentOutput := event.output

			if insideCutoff {
				event.input = nonNegativeDelta(currentInput, state.input)
				event.cachedInput = nonNegativeDelta(currentCachedInput, state.cachedInput)
				event.output = nonNegativeDelta(currentOutput, state.output)
			}

			state.input = currentInput
			state.cachedInput = currentCachedInput
			state.output = currentOutput
		}

		if provider.lineCounterMode == genericAICumulativeCounters && event.lineChanges != nil {
			lineCounterKey := event.sessionID + "\x00" + event.filePath
			current := max(*event.lineChanges, 0)
			previousLines := lineCounters[lineCounterKey]
			lineCounters[lineCounterKey] = current

			if insideCutoff {
				delta := max(current-previousLines, 0)
				event.lineChanges = &delta
			}
		}

		state.event = event

		states[event.sessionID] = state
		if !insideCutoff || !event.hasActivity() {
			continue
		}

		heartbeats = append(heartbeats, genericAIEventHeartbeats(provider.parser, provider.config, event)...)
	}

	return heartbeats, nil
}

func nonNegativeDelta(current int64, previous int64) int64 {
	return max(current-previous, 0)
}

func genericAIEventFromValue(value any) genericAIEvent {
	event := genericAIEvent{
		timestamp: genericAITime(genericAIFind(value, genericAITimestampKeys()...)),
		sessionID: genericAIString(genericAIFind(value, genericAISessionKeys()...)),
		model:     genericAIString(genericAIFind(value, genericAIModelKeys()...)),
		cwd:       genericAIString(genericAIFind(value, genericAICWDKeys()...)),
		toolName:  genericAIToolName(value),
		filePath:  genericAIString(genericAIFind(value, genericAIFileKeys()...)),
	}

	var inputFound, cachedInputFound, outputFound, cacheCreationFound bool

	event.input, inputFound = genericAINumericField(value, genericAIInputKeys()...)
	event.cachedInput, cachedInputFound = genericAINumericField(value, genericAICachedInputKeys()...)
	event.output, outputFound = genericAINumericField(value, genericAIOutputKeys()...)

	cacheCreation, cacheCreationFound := genericAINumericField(value, genericAICacheCreationKeys()...)
	event.input += cacheCreation

	event.tokensFound = inputFound || cachedInputFound || outputFound || cacheCreationFound
	if event.sessionID == "" && event.input > 0 {
		event.sessionID = genericAITopLevelString(value, "id")
	}

	if event.filePath == "" && event.toolName != "" {
		event.filePath = genericAIString(genericAIFind(value, "path"))
	}

	event.prompt = genericAIPrompt(value)
	event.isWrite = genericAIToolIsWrite(event.toolName)
	event.lineChanges = genericAILineChanges(value, event.isWrite)

	if event.cwd == "" {
		event.cwd = genericAICWDFromText(event.prompt)
	}

	return event
}

func genericAIPathIsAbs(path string) bool {
	return filepath.IsAbs(path) || strings.HasPrefix(path, "/") || strings.HasPrefix(path, `\`)
}

func genericAITopLevelString(value any, key string) string {
	object, ok := value.(map[string]any)
	if !ok {
		return ""
	}

	for objectKey, child := range object {
		if genericAIKey(objectKey) == genericAIKey(key) {
			return genericAIString(child)
		}
	}

	return ""
}

func genericAIToolName(value any) string {
	if name := genericAIString(genericAIFind(value, genericAIToolKeys()...)); name != "" {
		return name
	}

	return genericAINestedToolName(value, 0)
}

func genericAINestedToolName(value any, depth int) string {
	if depth > 10 {
		return ""
	}

	switch typed := value.(type) {
	case map[string]any:
		kind := strings.ToLower(genericAITopLevelString(typed, "type"))
		if strings.Contains(kind, "tool") || strings.Contains(kind, "function") {
			if name := genericAITopLevelString(typed, "name"); name != "" {
				return name
			}
		}

		for key, child := range typed {
			if genericAIKey(key) == "function" {
				if name := genericAITopLevelString(genericAIDecodedValue(child), "name"); name != "" {
					return name
				}
			}
		}

		for _, child := range typed {
			if name := genericAINestedToolName(genericAIDecodedValue(child), depth+1); name != "" {
				return name
			}
		}
	case []any:
		for _, child := range typed {
			if name := genericAINestedToolName(child, depth+1); name != "" {
				return name
			}
		}
	}

	return ""
}

func (e genericAIEvent) hasActivity() bool {
	return e.input > 0 || e.cachedInput > 0 || e.output > 0 || e.prompt != "" || e.toolName != "" || e.filePath != ""
}

func (e genericAIEvent) hasTokens() bool {
	return e.tokensFound
}

func genericAIEventHeartbeats(parser Parser, config ParserConfig, event genericAIEvent) Heartbeats {
	entity := appHeartbeatEntity(parser.Name(), event.sessionID)
	tokens := heartbeat.AITokens{
		CurrentInput:       max(event.input, 0),
		CurrentCachedInput: max(event.cachedInput, 0),
		CurrentOutput:      max(event.output, 0),
	}
	userAgent := aiUserAgent(entity, config.UserAgents, config.FallbackUserAgent, aiPlugin(parser, ""))
	userAgent = userAgentWithPrependedAgent(userAgent, aiModelUserAgentToken(event.model, ""))

	app := heartbeat.NewWithAITokens(
		nil,
		event.sessionID,
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
		event.cwd,
		heartbeatTimestamp(event.timestamp),
		userAgent,
	)
	app.AIPromptLength = promptLength(event.prompt)

	heartbeats := Heartbeats{app}
	if event.filePath == "" {
		return heartbeats
	}

	filePath := strings.ReplaceAll(filepath.Clean(event.filePath), `\`, "/")
	fileHeartbeat := heartbeat.New(
		event.lineChanges,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		filePath,
		heartbeat.FileType,
		nil,
		false,
		heartbeat.PointerTo(event.isWrite),
		nil,
		"",
		nil,
		nil,
		"",
		"",
		false,
		"",
		"",
		heartbeatTimestamp(event.timestamp),
		aiUserAgent(filePath, config.UserAgents, config.FallbackUserAgent, aiPlugin(parser, "")),
	)
	fileHeartbeat.AISession = event.sessionID
	heartbeats = append(heartbeats, fileHeartbeat)

	return heartbeats
}

func genericAISessionID(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	switch base {
	case "token_ledger":
		return filepath.Base(filepath.Dir(filepath.Dir(path)))
	case "chat", "events", "messages", "wire", "ui_messages":
		return filepath.Base(filepath.Dir(path))
	default:
		return strings.TrimSuffix(base, ".messages")
	}
}

func genericAIFind(value any, keys ...string) any {
	for _, key := range keys {
		if found := genericAIFindDepth(value, genericAIKey(key), 0); found != nil {
			return found
		}
	}

	return nil
}

func genericAIFindDepth(value any, key string, depth int) any {
	if depth > 10 {
		return nil
	}

	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			if key == genericAIKey(childKey) && !genericAIEmpty(child) {
				return child
			}
		}

		childKeys := make([]string, 0, len(typed))
		for childKey := range typed {
			childKeys = append(childKeys, childKey)
		}

		sort.Strings(childKeys)

		for _, childKey := range childKeys {
			if found := genericAIFindDepth(genericAIDecodedValue(typed[childKey]), key, depth+1); found != nil {
				return found
			}
		}
	case []any:
		for _, child := range typed {
			if found := genericAIFindDepth(child, key, depth+1); found != nil {
				return found
			}
		}
	}

	return nil
}

func genericAIDecodedValue(value any) any {
	text, ok := value.(string)
	if !ok {
		return value
	}

	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "{") && !strings.HasPrefix(text, "[") {
		return value
	}

	var decoded any
	if json.Unmarshal([]byte(text), &decoded) != nil {
		return value
	}

	return decoded
}

func genericAIKey(key string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' {
			return r + ('a' - 'A')
		}

		if r == '_' || r == '-' || r == ' ' {
			return -1
		}

		return r
	}, key)
}

func genericAIEmpty(value any) bool {
	switch typed := value.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(typed) == ""
	case float64:
		return typed == 0
	default:
		return false
	}
}

func genericAIString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	default:
		return ""
	}
}

func genericAIInt64(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	case int64:
		return typed
	case int:
		return int64(typed)
	case json.Number:
		parsed, _ := typed.Int64()
		return parsed
	case string:
		parsed, _ := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed
	default:
		return 0
	}
}

func genericAINumericField(value any, keys ...string) (int64, bool) {
	for _, key := range keys {
		if number, ok := genericAINumericFieldDepth(value, genericAIKey(key), 0); ok {
			return number, true
		}
	}

	return 0, false
}

func genericAINumericFieldDepth(value any, key string, depth int) (int64, bool) {
	if depth > 10 {
		return 0, false
	}

	switch typed := value.(type) {
	case map[string]any:
		for childKey, child := range typed {
			if key != genericAIKey(childKey) {
				continue
			}

			if number, ok := genericAIInt64Value(child); ok {
				return number, true
			}
		}

		childKeys := make([]string, 0, len(typed))
		for childKey := range typed {
			childKeys = append(childKeys, childKey)
		}

		sort.Strings(childKeys)

		for _, childKey := range childKeys {
			if number, ok := genericAINumericFieldDepth(
				genericAIDecodedValue(typed[childKey]), key, depth+1,
			); ok {
				return number, true
			}
		}
	case []any:
		for _, child := range typed {
			if number, ok := genericAINumericFieldDepth(child, key, depth+1); ok {
				return number, true
			}
		}
	}

	return 0, false
}

func genericAIInt64Value(value any) (int64, bool) {
	switch typed := value.(type) {
	case float64:
		return int64(typed), true
	case int64:
		return typed, true
	case int:
		return int64(typed), true
	case json.Number:
		parsed, err := typed.Int64()
		return parsed, err == nil
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(typed), 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func genericAIBool(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed
	case float64:
		return typed != 0
	case int64:
		return typed != 0
	default:
		return false
	}
}

func genericAITime(value any) time.Time {
	switch typed := value.(type) {
	case time.Time:
		return typed
	case string:
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05.999999999-07:00",
			"2006-01-02 15:04:05",
		}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, typed); err == nil {
				return parsed
			}
		}

		if number, err := strconv.ParseFloat(typed, 64); err == nil {
			return genericAINumericTime(number)
		}
	case float64:
		return genericAINumericTime(typed)
	case int64:
		return genericAINumericTime(float64(typed))
	case int:
		return genericAINumericTime(float64(typed))
	}

	return time.Time{}
}

func genericAINumericTime(value float64) time.Time {
	abs := value
	if abs < 0 {
		abs = -abs
	}

	switch {
	case abs >= 1e17:
		return time.Unix(0, int64(value)).UTC()
	case abs >= 1e14:
		return time.UnixMicro(int64(value)).UTC()
	case abs >= 1e11:
		return time.UnixMilli(int64(value)).UTC()
	default:
		seconds := int64(value)
		nanos := int64((value - float64(seconds)) * float64(time.Second))

		return time.Unix(seconds, nanos).UTC()
	}
}

func genericAIPrompt(value any) string {
	role := strings.ToLower(genericAIString(genericAIFind(value, "role", "type")))

	prompt := genericAIString(genericAIFind(
		value,
		"user_message", "userMessage", "user_input", "userInput", "prompt", "request", "title",
	))
	if prompt != "" {
		return prompt
	}

	if role == "turn.prompt" {
		return genericAIString(genericAIFind(value, "text"))
	}

	if genericAIBool(genericAIFind(value, "is_user_input", "isUserInput")) {
		return genericAIString(genericAIFind(value, "message", "content", "text"))
	}

	if role != "user" && role != "human" && role != "user_message" {
		return ""
	}

	return genericAIString(genericAIFind(value, "content", "text", "message"))
}

func genericAICWDFromText(text string) string {
	if cwd := clineCurrentWorkingDirectory(text); cwd != "" {
		return cwd
	}

	match := genericAICWDPattern.FindStringSubmatch(text)
	if len(match) < 2 {
		return ""
	}

	return strings.TrimSpace(match[1])
}

func genericAIToolIsWrite(tool string) bool {
	tool = strings.ToLower(tool)

	return strings.Contains(tool, "write") || strings.Contains(tool, "edit") || strings.Contains(tool, "patch") ||
		strings.Contains(tool, "create") || strings.Contains(tool, "delete") || strings.Contains(tool, "replace")
}

func genericAILineChanges(value any, isWrite bool) *int {
	added, addedFound := genericAINumericField(
		value, "lines_added", "linesAdded", "added_lines", "model_added_lines",
	)

	removed, removedFound := genericAINumericField(
		value, "lines_removed", "linesRemoved", "removed_lines", "model_removed_lines",
	)
	if addedFound || removedFound {
		lineChanges := int(max(added, 0) + max(removed, 0))
		return &lineChanges
	}

	diff := genericAIString(genericAIFind(value, "diff", "patch", "unified_diff"))
	if diff != "" {
		lineChanges := 0

		for _, line := range strings.Split(diff, "\n") {
			if (strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++")) ||
				(strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---")) {
				lineChanges++
			}
		}

		return &lineChanges
	}

	if !isWrite {
		return nil
	}

	content := genericAIString(genericAIFind(value, "content", "new_content", "newContent"))
	if content == "" {
		return nil
	}

	lineChanges := countStringLines(content)

	return &lineChanges
}

func genericAITimestampKeys() []string {
	return []string{
		"timestamp", "created_at", "createdAt", "updated_at", "updatedAt",
		"completed_at", "completedAt", "ended_at", "endedAt", "started_at", "startedAt",
		"end_time", "endTime", "start_time", "startTime", "ts", "time",
	}
}

func genericAISessionKeys() []string {
	return []string{"session_id", "sessionId", "conversation_id", "conversationId", "thread_id", "threadId"}
}

func genericAIModelKeys() []string {
	return []string{"model", "model_id", "modelId", "model_name", "modelName", "active_model", "activeModel"}
}

func genericAICWDKeys() []string {
	return []string{
		"cwd", "working_dir", "workingDirectory", "working_directory", "project_path", "projectPath",
		"workspace_path", "workspacePath", "workspace",
	}
}

func genericAIToolKeys() []string {
	return []string{"tool_name", "toolName", "function_name", "functionName", "tool"}
}

func genericAIFileKeys() []string {
	return []string{"file_path", "filePath", "file_name", "fileName", "target_file", "targetFile"}
}

func genericAIInputKeys() []string {
	return []string{
		"input_tokens", "inputTokens", "tokens_in", "tokensIn", "prompt_tokens", "promptTokens",
		"promptTokenCount", "total_input_tokens", "total_tokens", "session_prompt_tokens", "input_other",
		"inputOther", "input",
	}
}

func genericAIOutputKeys() []string {
	return []string{
		"output_tokens", "outputTokens", "tokens_out", "tokensOut", "completion_tokens", "completionTokens",
		"candidatesTokenCount", "total_output_tokens", "session_completion_tokens", "output",
	}
}

func genericAICachedInputKeys() []string {
	return []string{
		"cached_input_tokens", "cachedInputTokens", "cache_read_input_tokens", "cacheReadInputTokens",
		"cache_read_tokens", "cacheReadTokens", "cacheReads", "cachedContentTokenCount", "input_cache_read",
		"inputCacheRead", "cached",
	}
}

func genericAICacheCreationKeys() []string {
	return []string{
		"cache_creation_input_tokens", "cacheCreationInputTokens", "cache_creation_tokens", "cacheCreationTokens",
		"input_cache_creation", "inputCacheCreation", "cacheWrites",
	}
}
