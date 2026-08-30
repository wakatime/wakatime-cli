package ai

import (
	"context"
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

// CodeBuddyCode contains parameters for detecting heartbeats from Tencent CodeBuddy Code transcripts.
type CodeBuddyCode ParserConfig

type (
	codeBuddyCodeTranscript struct {
		path     string
		subagent bool
		parentID string
	}

	codeBuddyCodeToolCall struct {
		arguments map[string]any
		cwd       string
		model     string
		name      string
		sessionID string
	}
)

var codeBuddyCodeIntegerPattern = regexp.MustCompile(`\d+`)

// Parse parses Tencent CodeBuddy Code JSONL session transcripts for AI heartbeats.
func (g CodeBuddyCode) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	configDir := envOrDefault("CODEBUDDY_CONFIG_DIR", filepath.Join(home, ".codebuddy"))

	transcripts, err := codeBuddyCodeTranscriptPaths(configDir, g.After)
	if err != nil {
		return nil, err
	}

	logger := log.Extract(ctx)

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		values, err := genericAIJSONLines(ctx, g.Name(), transcript.path)
		if err != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return nil, ctxErr
			}

			logger.Warnf("failed parsing %s transcript %q: %s", g.Name(), transcript.path, err)

			continue
		}

		heartbeats = append(heartbeats, g.parseTranscript(ctx, transcript, values)...)
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
	}

	return heartbeats, nil
}

func codeBuddyCodeTranscriptPaths(configDir string, after time.Time) ([]codeBuddyCodeTranscript, error) {
	projectsDir := filepath.Join(configDir, "projects")

	projects, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read CodeBuddy Code projects directory: %s", err)
	}

	var transcripts []codeBuddyCodeTranscript

	for _, project := range projects {
		if !project.IsDir() {
			continue
		}

		projectDir := filepath.Join(projectsDir, project.Name())

		entries, err := os.ReadDir(projectDir)
		if err != nil {
			return nil, fmt.Errorf("failed to read CodeBuddy Code project directory %q: %s", projectDir, err)
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jsonl") {
				continue
			}

			path := filepath.Join(projectDir, entry.Name())
			if codeBuddyCodeTranscriptIsRecent(entry, after) {
				transcripts = append(transcripts, codeBuddyCodeTranscript{path: path})
			}
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			parentID := entry.Name()
			subagentsDir := filepath.Join(projectDir, parentID, "subagents")

			subagents, err := os.ReadDir(subagentsDir)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}

				return nil, fmt.Errorf("failed to read CodeBuddy Code subagents directory %q: %s", subagentsDir, err)
			}

			for _, subagent := range subagents {
				if subagent.IsDir() || !strings.EqualFold(filepath.Ext(subagent.Name()), ".jsonl") ||
					!codeBuddyCodeTranscriptIsRecent(subagent, after) {
					continue
				}

				transcripts = append(transcripts, codeBuddyCodeTranscript{
					path:     filepath.Join(subagentsDir, subagent.Name()),
					subagent: true,
					parentID: parentID,
				})
			}
		}
	}

	sort.Slice(transcripts, func(i, j int) bool {
		if transcripts[i].subagent != transcripts[j].subagent {
			return !transcripts[i].subagent
		}

		return transcripts[i].path < transcripts[j].path
	})

	return transcripts, nil
}

func codeBuddyCodeTranscriptIsRecent(entry os.DirEntry, after time.Time) bool {
	info, err := entry.Info()

	return err == nil && timestampAtOrAfterCutoff(info.ModTime(), after)
}

func (g CodeBuddyCode) parseTranscript(
	ctx context.Context,
	transcript codeBuddyCodeTranscript,
	values []any,
) Heartbeats {
	defaultSessionID := genericAISessionID(transcript.path)
	if transcript.subagent && transcript.parentID != "" {
		defaultSessionID = transcript.parentID + "/" + defaultSessionID
	}

	pendingTools := make(map[string]codeBuddyCodeToolCall)
	model := ""
	cwd := ""

	var heartbeats Heartbeats

	for i, value := range values {
		if ctx.Err() != nil {
			break
		}

		event := genericAIEventFromValue(value)
		event.sessionID = codeBuddyCodeSessionID(value, defaultSessionID)
		event.cwd = codeBuddyCodeTopLevelString(value, "cwd")
		event.model = codeBuddyCodeModel(value)

		if event.cwd == "" {
			event.cwd = cwd
		} else {
			cwd = event.cwd
		}

		if event.model != "" {
			model = event.model
		} else {
			event.model = model
		}

		if event.prompt != "" {
			if !codeBuddyCodeIsExternalPrompt(value) {
				event.prompt = ""
			} else if nextModel := codeBuddyCodeFollowingModel(values, i); nextModel != "" {
				event.model = nextModel
			}
		}

		codeBuddyCodeNormalizeUsage(value, &event)

		recordType := strings.ToLower(codeBuddyCodeTopLevelString(value, "type"))
		switch recordType {
		case "function_call":
			if !transcript.subagent {
				heartbeats = append(heartbeats, g.nonToolHeartbeats(event)...)
			}

			callID := codeBuddyCodeCallID(value)
			if callID != "" {
				pendingTools[callID] = codeBuddyCodeToolCall{
					arguments: codeBuddyCodeArguments(value),
					cwd:       event.cwd,
					model:     event.model,
					name:      codeBuddyCodeTopLevelString(value, "name"),
					sessionID: event.sessionID,
				}
			}

			continue
		case "function_call_output", "function_call_result":
			if !transcript.subagent {
				heartbeats = append(heartbeats, g.nonToolHeartbeats(event)...)
			}

			callID := codeBuddyCodeCallID(value)

			call, ok := pendingTools[callID]
			if !ok {
				continue
			}

			delete(pendingTools, callID)

			if !codeBuddyCodeToolResultSucceeded(value) {
				continue
			}

			filePath, isWrite, lineChanges := codeBuddyCodeToolHeartbeatInfo(call, value)
			if filePath == "" || codeBuddyCodeDirectoryResult(call, value, filePath) {
				continue
			}

			toolEvent := genericAIEvent{
				timestamp:   event.timestamp,
				sessionID:   call.sessionID,
				model:       call.model,
				cwd:         call.cwd,
				filePath:    filePath,
				toolName:    call.name,
				lineChanges: lineChanges,
				isWrite:     isWrite,
			}
			if toolEvent.sessionID == "" {
				toolEvent.sessionID = defaultSessionID
			}

			if toolEvent.timestamp.IsZero() || !timestampAtOrAfterCutoff(toolEvent.timestamp, g.After) {
				continue
			}

			if transcript.subagent {
				heartbeats = append(heartbeats, g.fileHeartbeat(toolEvent))
			} else {
				heartbeats = append(heartbeats, genericAIEventHeartbeats(g, ParserConfig(g), toolEvent)...)
			}

			continue
		}

		if transcript.subagent || event.timestamp.IsZero() ||
			!timestampAtOrAfterCutoff(event.timestamp, g.After) || !event.hasActivity() {
			continue
		}

		heartbeats = append(heartbeats, genericAIEventHeartbeats(g, ParserConfig(g), event)...)
	}

	return heartbeats
}

func (g CodeBuddyCode) nonToolHeartbeats(event genericAIEvent) Heartbeats {
	event.filePath = ""
	event.toolName = ""
	event.prompt = ""
	event.lineChanges = nil
	event.isWrite = false

	if event.timestamp.IsZero() || !timestampAtOrAfterCutoff(event.timestamp, g.After) || !event.hasActivity() {
		return nil
	}

	return genericAIEventHeartbeats(g, ParserConfig(g), event)
}

func (g CodeBuddyCode) fileHeartbeat(event genericAIEvent) heartbeat.Heartbeat {
	filePath := strings.ReplaceAll(filepath.Clean(event.filePath), `\`, "/")
	h := heartbeat.New(
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
		aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, event.model, ""),
	)
	h.AISession = event.sessionID

	return h
}

func codeBuddyCodeNormalizeUsage(value any, event *genericAIEvent) {
	cacheCreation, _ := genericAINumericField(value, genericAICacheCreationKeys()...)
	rawInput := max(event.input-cacheCreation, 0)
	event.input = max(rawInput-event.cachedInput, 0)
}

func codeBuddyCodeFollowingModel(values []any, promptIndex int) string {
	for i := promptIndex + 1; i < len(values); i++ {
		if genericAIPrompt(values[i]) != "" && codeBuddyCodeIsExternalPrompt(values[i]) {
			return ""
		}

		if model := codeBuddyCodeModel(values[i]); model != "" {
			return model
		}
	}

	return ""
}

func codeBuddyCodeIsExternalPrompt(value any) bool {
	role := strings.ToLower(codeBuddyCodeTopLevelString(value, "role"))
	if role != "user" && role != "human" && role != "user_message" {
		return false
	}

	providerData := codeBuddyCodeTopLevelObject(value, "providerData")
	if startsNewRequest, ok := codeBuddyCodeBoolField(providerData, "startsNewUserRequest"); ok {
		return startsNewRequest
	}

	for _, key := range []string{
		"skipRun", "isMeta", "isCompactInternal", "isSummary", "isAdditionalData", "isSystemTeammate",
		"systemTeammate",
	} {
		if flagged, _ := codeBuddyCodeBoolField(providerData, key); flagged {
			return false
		}
	}

	agent := strings.ToLower(codeBuddyCodeObjectString(providerData, "agent"))

	return agent != "compact" && agent != "teammate"
}

func codeBuddyCodeToolHeartbeatInfo(
	call codeBuddyCodeToolCall,
	result any,
) (string, bool, *int) {
	filePath := codeBuddyCodeToolPath(call.arguments, call.cwd)
	if filePath == "" {
		return "", false, nil
	}

	switch strings.ToLower(strings.TrimSpace(call.name)) {
	case "read", "read_image":
		return filePath, false, heartbeat.PointerTo(0)
	case "write":
		if !codeBuddyCodeWriteCreatedFile(result) {
			return filePath, true, nil
		}

		return filePath, true, heartbeat.PointerTo(codeBuddyCodeTextLineCount(
			codeBuddyCodeObjectString(call.arguments, "content"),
		))
	case "edit":
		lineChanges, ok := codeBuddyCodeEditLineChanges(call.arguments, result)
		if !ok {
			return filePath, true, nil
		}

		return filePath, true, heartbeat.PointerTo(lineChanges)
	case "multiedit", "multi_edit":
		lineChanges, ok := codeBuddyCodeMultiEditLineChanges(call.arguments)
		if !ok {
			return filePath, true, nil
		}

		return filePath, true, heartbeat.PointerTo(lineChanges)
	default:
		return "", false, nil
	}
}

func codeBuddyCodeEditLineChanges(arguments map[string]any, result any) (int, bool) {
	oldText := codeBuddyCodeObjectString(arguments, "old_string")
	newText := codeBuddyCodeObjectString(arguments, "new_string")
	repeat := 1

	if replaceAll, _ := codeBuddyCodeBoolField(arguments, "replace_all"); replaceAll {
		var ok bool

		repeat, ok = codeBuddyCodeReplacementCount(result)
		if !ok {
			return 0, false
		}
	}

	return repeat * (codeBuddyCodeTextLineCount(newText) - codeBuddyCodeTextLineCount(oldText)), true
}

func codeBuddyCodeMultiEditLineChanges(arguments map[string]any) (int, bool) {
	edits, ok := arguments["edits"].([]any)
	if !ok {
		return 0, false
	}

	lineChanges := 0

	for _, value := range edits {
		edit, ok := value.(map[string]any)
		if !ok {
			return 0, false
		}

		if replaceAll, _ := codeBuddyCodeBoolField(edit, "replace_all"); replaceAll {
			return 0, false
		}

		lineChanges += codeBuddyCodeTextLineCount(codeBuddyCodeObjectString(edit, "new_string")) -
			codeBuddyCodeTextLineCount(codeBuddyCodeObjectString(edit, "old_string"))
	}

	return lineChanges, true
}

func codeBuddyCodeReplacementCount(result any) (int, bool) {
	providerData := codeBuddyCodeTopLevelObject(result, "providerData")
	toolResult := codeBuddyCodeTopLevelObject(providerData, "toolResult")
	title := codeBuddyCodeObjectString(toolResult, "title")

	lowerTitle := strings.ToLower(title)
	if !strings.Contains(lowerTitle, "made") || !strings.Contains(lowerTitle, "replacement") {
		return 0, false
	}

	count, err := strconv.Atoi(codeBuddyCodeIntegerPattern.FindString(title))

	return count, err == nil
}

func codeBuddyCodeWriteCreatedFile(result any) bool {
	text := strings.ToLower(codeBuddyCodeResultText(result))

	return strings.Contains(text, "successfully created and wrote to new file")
}

func codeBuddyCodeDirectoryResult(call codeBuddyCodeToolCall, result any, filePath string) bool {
	if !strings.EqualFold(strings.TrimSpace(call.name), "read") {
		return false
	}

	if info, err := os.Stat(filepath.FromSlash(filePath)); err == nil && info.IsDir() {
		return true
	}

	providerData := codeBuddyCodeTopLevelObject(result, "providerData")
	toolResult := codeBuddyCodeTopLevelObject(providerData, "toolResult")
	renderer := codeBuddyCodeTopLevelObject(toolResult, "renderer")

	return strings.EqualFold(codeBuddyCodeObjectString(renderer, "type"), "list")
}

func codeBuddyCodeToolResultSucceeded(value any) bool {
	status := strings.ToLower(codeBuddyCodeTopLevelString(value, "status"))
	switch status {
	case "incomplete", "failed", "error", "cancelled", "canceled", "rejected":
		return false
	}

	providerData := codeBuddyCodeTopLevelObject(value, "providerData")
	if failed, _ := codeBuddyCodeBoolField(providerData, "error"); failed {
		return false
	}

	toolResult := codeBuddyCodeTopLevelObject(providerData, "toolResult")
	if codeBuddyCodeDirectError(value) || codeBuddyCodeDirectError(providerData) ||
		codeBuddyCodeDirectError(toolResult) ||
		codeBuddyCodeDirectError(codeBuddyCodeTopLevelValue(value, "output")) {
		return false
	}

	return true
}

func codeBuddyCodeDirectError(value any) bool {
	object, ok := genericAIDecodedValue(value).(map[string]any)
	if !ok {
		return false
	}

	if failed, _ := codeBuddyCodeBoolField(object, "isError"); failed {
		return true
	}

	return codeBuddyCodeNonEmptyField(object, "error")
}

func codeBuddyCodeResultText(value any) string {
	var parts []string

	for _, key := range []string{"output", "content"} {
		if text := genericAIContentText(codeBuddyCodeTopLevelValue(value, key)); text != "" {
			parts = append(parts, text)
		}
	}

	providerData := codeBuddyCodeTopLevelObject(value, "providerData")
	if text := genericAIContentText(codeBuddyCodeTopLevelValue(providerData, "toolResult")); text != "" {
		parts = append(parts, text)
	}

	return strings.Join(parts, "\n")
}

func codeBuddyCodeToolPath(arguments map[string]any, cwd string) string {
	path := codeBuddyCodeObjectString(arguments, "file_path")
	if path == "" {
		return ""
	}

	if !genericAIPathIsAbs(path) && cwd != "" {
		path = filepath.Join(cwd, filepath.FromSlash(path))
	}

	return filepath.ToSlash(filepath.Clean(path))
}

func codeBuddyCodeArguments(value any) map[string]any {
	arguments := genericAIDecodedValue(codeBuddyCodeTopLevelValue(value, "arguments"))
	object, _ := arguments.(map[string]any)

	return object
}

func codeBuddyCodeCallID(value any) string {
	if callID := codeBuddyCodeTopLevelString(value, "callId"); callID != "" {
		return callID
	}

	return codeBuddyCodeTopLevelString(value, "id")
}

func codeBuddyCodeSessionID(value any, fallback string) string {
	for _, key := range genericAISessionKeys() {
		if sessionID := codeBuddyCodeTopLevelString(value, key); sessionID != "" {
			return sessionID
		}
	}

	return fallback
}

func codeBuddyCodeModel(value any) string {
	for _, key := range genericAIModelKeys() {
		if model := codeBuddyCodeTopLevelString(value, key); model != "" {
			return model
		}
	}

	providerData := codeBuddyCodeTopLevelObject(value, "providerData")
	for _, key := range genericAIModelKeys() {
		if model := codeBuddyCodeTopLevelString(providerData, key); model != "" {
			return model
		}
	}

	return ""
}

func codeBuddyCodeTopLevelObject(value any, key string) map[string]any {
	object, _ := codeBuddyCodeTopLevelValue(value, key).(map[string]any)

	return object
}

func codeBuddyCodeTopLevelString(value any, key string) string {
	return genericAIString(codeBuddyCodeTopLevelValue(value, key))
}

func codeBuddyCodeObjectString(value map[string]any, key string) string {
	return genericAIString(codeBuddyCodeTopLevelValue(value, key))
}

func codeBuddyCodeTopLevelValue(value any, key string) any {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	wanted := genericAIKey(key)
	for childKey, child := range object {
		if genericAIKey(childKey) == wanted {
			return genericAIDecodedValue(child)
		}
	}

	return nil
}

func codeBuddyCodeBoolField(value map[string]any, key string) (bool, bool) {
	return codeBuddyCodeBoolValue(codeBuddyCodeTopLevelValue(value, key))
}

func codeBuddyCodeBoolValue(value any) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(strings.TrimSpace(typed))
		return parsed, err == nil
	case float64:
		return typed != 0, true
	case int:
		return typed != 0, true
	case int64:
		return typed != 0, true
	default:
		return false, false
	}
}

func codeBuddyCodeNonEmptyField(value map[string]any, key string) bool {
	field := codeBuddyCodeTopLevelValue(value, key)

	return field != nil && !genericAIEmpty(field)
}

func codeBuddyCodeTextLineCount(text string) int {
	if text == "" {
		return 0
	}

	return countStringLines(text)
}

// Name returns the Tencent CodeBuddy Code parser name.
func (CodeBuddyCode) Name() string { return "CodeBuddy Code" }
