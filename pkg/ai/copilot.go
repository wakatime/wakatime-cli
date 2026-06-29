package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Copilot contains params for detecting heartbeats from GitHub Copilot Chat sessions.
type Copilot ParserConfig

type (
	copilotSession struct {
		Version         int              `json:"version"`
		CreationDate    int64            `json:"creationDate"`
		LastMessageDate int64            `json:"lastMessageDate"`
		SessionID       string           `json:"sessionId"`
		Requests        []copilotRequest `json:"requests"`
	}

	copilotRequest struct {
		RequestID    string                `json:"requestId"`
		Timestamp    int64                 `json:"timestamp"`
		Agent        *copilotAgent         `json:"agent"`
		Message      *copilotMessage       `json:"message"`
		Usage        json.RawMessage       `json:"usage"`
		Result       json.RawMessage       `json:"result"`
		VariableData *copilotVariableData  `json:"variableData"`
		Response     []copilotResponseItem `json:"response"`
		ModelState   *copilotModelState    `json:"modelState"`
	}

	copilotAgent struct {
		ExtensionVersion string `json:"extensionVersion"`
	}

	copilotModelState struct {
		Value       int   `json:"value"`
		CompletedAt int64 `json:"completedAt"`
	}

	copilotMessage struct {
		Text string `json:"text"`
	}

	copilotVariableData struct {
		Variables []copilotVariable `json:"variables"`
	}

	copilotVariable struct {
		Kind  string            `json:"kind"`
		ID    string            `json:"id"`
		Value *copilotPathValue `json:"value"`
	}

	copilotPathValue struct {
		FSPath string          `json:"fsPath"`
		URI    *copilotFileURI `json:"uri"`
	}

	copilotFileURI struct {
		FSPath string `json:"fsPath"`
	}

	copilotResponseItem struct {
		Kind              string                  `json:"kind"`
		InvocationMessage *copilotMessageWithURIs `json:"invocationMessage"`
		PastTenseMessage  *copilotMessageWithURIs `json:"pastTenseMessage"`
		InlineReference   *copilotInlineReference `json:"inlineReference"`
	}

	copilotMessageWithURIs struct {
		Value string                         `json:"value"`
		URIs  map[string]copilotReferenceURI `json:"uris"`
	}

	copilotInlineReference struct {
		FSPath string `json:"fsPath"`
		URI    *struct {
			FSPath string `json:"fsPath"`
		} `json:"uri"`
	}

	copilotReferenceURI struct {
		Path   string `json:"path"`
		FSPath string `json:"fsPath"`
	}

	copilotJSONLPatch struct {
		Kind int               `json:"kind"`
		K    []json.RawMessage `json:"k"`
		V    json.RawMessage   `json:"v"`
		I    int               `json:"i"`
	}

	copilotEditState struct {
		Timeline *struct {
			FileBaselines [][]json.RawMessage    `json:"fileBaselines"`
			Operations    []copilotTextOperation `json:"operations"`
		} `json:"timeline"`
	}

	copilotTextOperation struct {
		Type      string               `json:"type"`
		RequestID string               `json:"requestId"`
		URI       *copilotOperationURI `json:"uri"`
		Epoch     int                  `json:"epoch"`
		Edits     []copilotTextEdit    `json:"edits"`
	}

	copilotOperationURI struct {
		FSPath string `json:"fsPath"`
	}

	copilotTextEdit struct {
		Text  string            `json:"text"`
		Range *copilotTextRange `json:"range"`
	}

	copilotTextRange struct {
		StartLineNumber int `json:"startLineNumber"`
		StartColumn     int `json:"startColumn"`
		EndLineNumber   int `json:"endLineNumber"`
		EndColumn       int `json:"endColumn"`
	}

	copilotWorkspaceState struct {
		workspaceDir string
		sessions     map[string]*copilotSession
	}

	copilotRequestMeta struct {
		timestamp time.Time
		plugin    string
		sessionID string
	}

	copilotTimedHeartbeat struct {
		timestamp time.Time
		heartbeat heartbeat.Heartbeat
	}
)

func (v *copilotVariable) UnmarshalJSON(data []byte) error {
	type alias struct {
		Kind  string          `json:"kind"`
		ID    string          `json:"id"`
		Value json.RawMessage `json:"value"`
	}

	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	v.Kind = decoded.Kind
	v.ID = decoded.ID

	if len(decoded.Value) == 0 || string(decoded.Value) == "null" {
		return nil
	}

	var pathValue copilotPathValue
	if err := json.Unmarshal(decoded.Value, &pathValue); err == nil {
		v.Value = &pathValue
	}

	return nil
}

func (m *copilotMessageWithURIs) UnmarshalJSON(data []byte) error {
	if data == nil || string(data) == "null" {
		return nil
	}

	var text string
	if err := json.Unmarshal(data, &text); err == nil {
		m.Value = text
		return nil
	}

	type alias copilotMessageWithURIs

	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	m.Value = decoded.Value
	m.URIs = decoded.URIs

	return nil
}

// Parse parses VS Code GitHub Copilot Chat sessions and Copilot CLI event.jsonl transcripts for ai heartbeats.
func (g Copilot) Parse(ctx context.Context) (Heartbeats, error) {
	workspaces, err := g.workspaceStates(ctx)
	if err != nil {
		return nil, err
	}

	logger := log.Extract(ctx)

	var timed []copilotTimedHeartbeat

	if len(workspaces) > 0 {
		logger.Debugf("Found %d Copilot workspace storage directories for %s", len(workspaces), g.Name())
	}

	for _, ws := range workspaces {
		sessionHeartbeats, requestMeta := g.sessionHeartbeats(ws)
		timed = append(timed, sessionHeartbeats...)

		editHeartbeats, err := g.editHeartbeats(ws.workspaceDir, requestMeta)
		if err != nil {
			return nil, err
		}

		timed = append(timed, editHeartbeats...)
	}

	cliHeartbeats, err := g.cliHeartbeats(ctx)
	if err != nil {
		return nil, err
	}

	timed = append(timed, cliHeartbeats...)

	if len(timed) == 0 {
		return Heartbeats{}, nil
	}

	sort.SliceStable(timed, func(i, j int) bool {
		return timed[i].timestamp.Before(timed[j].timestamp)
	})

	heartbeats := make(Heartbeats, 0, len(timed))
	for _, item := range timed {
		if item.timestamp.Before(g.After) {
			continue
		}

		heartbeats = append(heartbeats, item.heartbeat)
	}

	return heartbeats, nil
}

func (g Copilot) workspaceStates(ctx context.Context) ([]copilotWorkspaceState, error) {
	dirs, err := workspaceStorageDirs(ctx)
	if err != nil {
		return nil, err
	}

	logger := log.Extract(ctx)

	var workspaces []copilotWorkspaceState

	for _, dir := range dirs {
		sessions, err := g.loadWorkspaceSessions(logger, dir)
		if err != nil {
			return nil, err
		}

		if len(sessions) == 0 {
			continue
		}

		workspaces = append(workspaces, copilotWorkspaceState{
			workspaceDir: dir,
			sessions:     sessions,
		})
	}

	return workspaces, nil
}

func workspaceStorageDirs(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Code", "User", "workspaceStorage"),
		filepath.Join(home, "AppData", "Roaming", "Code", "User", "workspaceStorage"),
		filepath.Join(home, ".config", "Code", "User", "workspaceStorage"),
		filepath.Join(home, ".config", "code", "User", "workspaceStorage"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			entries, err := os.ReadDir(candidate)
			if err != nil {
				return nil, fmt.Errorf("failed reading copilot workspace storage %q: %s", candidate, err)
			}

			dirs := make([]string, 0, len(entries))
			for _, entry := range entries {
				if entry.IsDir() {
					dirs = append(dirs, filepath.Join(candidate, entry.Name()))
				}
			}

			return dirs, nil
		}
	}

	return nil, nil
}

func (g Copilot) loadWorkspaceSessions(logger *log.Logger, workspaceDir string) (map[string]*copilotSession, error) {
	sessionDir := filepath.Join(workspaceDir, "chatSessions")

	entries, err := os.ReadDir(sessionDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed reading copilot chat sessions %q: %s", sessionDir, err)
	}

	sessions := make(map[string]*copilotSession)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".json" && ext != ".jsonl" {
			continue
		}

		info, err := entry.Info()
		if err != nil || info.ModTime().Before(g.After) {
			continue
		}

		path := filepath.Join(sessionDir, entry.Name())

		var session *copilotSession
		if ext == ".json" {
			session, err = g.parseJSONSession(path)
		} else {
			session, err = g.parseJSONLSession(path)
		}

		if err != nil {
			logger.Warnf("failed parsing copilot session %q: %s", path, err)
			continue
		}

		if session == nil {
			continue
		}

		if session.SessionID == "" {
			session.SessionID = strings.TrimSuffix(entry.Name(), ext)
		}

		sessions[session.SessionID] = session
	}

	return sessions, nil
}

func (Copilot) parseJSONSession(path string) (*copilotSession, error) {
	//nolint:gosec
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed reading copilot session %q: %s", path, err)
	}

	var session copilotSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed parsing copilot session %q: %s", path, err)
	}

	return &session, nil
}

func (g Copilot) parseJSONLSession(path string) (*copilotSession, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("failed opening copilot session %q: %s", path, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	var session *copilotSession

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var patch copilotJSONLPatch
		if err := json.Unmarshal(line, &patch); err != nil {
			continue
		}

		switch patch.Kind {
		case 0:
			var root struct {
				V copilotSession `json:"v"`
			}
			if err := json.Unmarshal(line, &root); err != nil {
				continue
			}

			copy := root.V
			session = &copy
		case 1:
			if session == nil {
				continue
			}

			g.applyJSONLSet(session, patch)
		case 2:
			if session == nil {
				continue
			}

			g.applyJSONLInsert(session, patch)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading copilot jsonl session %q: %s", path, err)
	}

	return session, nil
}

func (g Copilot) applyJSONLSet(session *copilotSession, patch copilotJSONLPatch) {
	if len(patch.K) == 0 {
		return
	}

	head, ok := g.parseJSONString(patch.K[0])
	if !ok || head != "requests" {
		return
	}

	if len(patch.K) < 3 {
		return
	}

	idx, ok := g.parseJSONInt(patch.K[1])
	if !ok || idx < 0 || idx >= len(session.Requests) {
		return
	}

	field, ok := g.parseJSONString(patch.K[2])
	if !ok {
		return
	}

	switch field {
	case "modelState":
		var state copilotModelState
		if err := json.Unmarshal(patch.V, &state); err == nil {
			session.Requests[idx].ModelState = &state
		}
	case "result":
		session.Requests[idx].Result = append(session.Requests[idx].Result[:0], patch.V...)
	case "response":
		var items []copilotResponseItem
		if err := json.Unmarshal(patch.V, &items); err == nil {
			session.Requests[idx].Response = items
		}
	case "message":
		var message copilotMessage
		if err := json.Unmarshal(patch.V, &message); err == nil {
			session.Requests[idx].Message = &message
		}
	case "variableData":
		var variableData copilotVariableData
		if err := json.Unmarshal(patch.V, &variableData); err == nil {
			session.Requests[idx].VariableData = &variableData
		}
	}
}

func (g Copilot) applyJSONLInsert(session *copilotSession, patch copilotJSONLPatch) {
	if len(patch.K) == 0 {
		return
	}

	head, ok := g.parseJSONString(patch.K[0])
	if !ok || head != "requests" {
		return
	}

	if len(patch.K) == 1 {
		var requests []copilotRequest
		if err := json.Unmarshal(patch.V, &requests); err == nil {
			session.Requests = g.insertRequests(
				session.Requests,
				patch.I,
				requests,
			)
		}

		return
	}

	if len(patch.K) == 3 {
		idx, ok := g.parseJSONInt(patch.K[1])
		if !ok || idx < 0 || idx >= len(session.Requests) {
			return
		}

		field, ok := g.parseJSONString(patch.K[2])
		if !ok || field != "response" {
			return
		}

		var items []copilotResponseItem
		if err := json.Unmarshal(patch.V, &items); err == nil {
			session.Requests[idx].Response = g.insertResponses(
				session.Requests[idx].Response,
				patch.I,
				items,
			)
		}
	}
}

func (Copilot) insertRequests(existing []copilotRequest, index int, incoming []copilotRequest) []copilotRequest {
	if index < 0 || index > len(existing) {
		index = len(existing)
	}

	result := make([]copilotRequest, 0, len(existing)+len(incoming))
	result = append(result, existing[:index]...)
	result = append(result, incoming...)
	result = append(result, existing[index:]...)

	return result
}

func (Copilot) insertResponses(
	existing []copilotResponseItem,
	index int,
	incoming []copilotResponseItem,
) []copilotResponseItem {
	if index < 0 || index > len(existing) {
		index = len(existing)
	}

	result := make([]copilotResponseItem, 0, len(existing)+len(incoming))
	result = append(result, existing[:index]...)
	result = append(result, incoming...)
	result = append(result, existing[index:]...)

	return result
}

func (Copilot) parseJSONString(raw json.RawMessage) (string, bool) {
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", false
	}

	return value, true
}

func (Copilot) parseJSONInt(raw json.RawMessage) (int, bool) {
	var value int
	if err := json.Unmarshal(raw, &value); err != nil {
		return 0, false
	}

	return value, true
}

func (g Copilot) sessionHeartbeats(
	ws copilotWorkspaceState,
) ([]copilotTimedHeartbeat, map[string]copilotRequestMeta) {
	var timed []copilotTimedHeartbeat

	requestMeta := make(map[string]copilotRequestMeta)

	sessionIDs := make([]string, 0, len(ws.sessions))
	for sessionID := range ws.sessions {
		sessionIDs = append(sessionIDs, sessionID)
	}

	sort.Strings(sessionIDs)

	for _, sessionID := range sessionIDs {
		session := ws.sessions[sessionID]
		sessionEntity := appHeartbeatEntity("Copilot", session.SessionID)

		var tokens heartbeat.AITokens

		for _, request := range session.Requests {
			requestTime := g.millisTime(request.Timestamp)
			if requestTime.IsZero() {
				continue
			}

			plugin := "github-copilot/" + unknownIfEmpty(g.version(request.Agent))
			requestMeta[request.RequestID] = copilotRequestMeta{
				timestamp: requestTime,
				plugin:    plugin,
				sessionID: session.SessionID,
			}
			projectPath := g.requestProjectPath(request)

			var requestTimed []copilotTimedHeartbeat

			assignTokens := false

			tokens = g.copilotTokenCounts(request, tokens)
			assignTokens = g.hasTokenDelta(tokens)

			if strings.TrimSpace(g.messageText(request.Message)) != "" {
				requestTimed = append(requestTimed, copilotTimedHeartbeat{
					timestamp: requestTime,
					heartbeat: g.appHeartbeat(
						sessionEntity,
						session.SessionID,
						g.tokensForFirstHeartbeat(assignTokens, tokens),
						promptLength(g.messageText(request.Message)),
						requestTime,
						projectPath,
						g.UserAgents,
						g.FallbackUserAgent,
						plugin,
					),
				})
				assignTokens = false
			}

			readFiles := g.readFiles(request)
			for i, path := range readFiles {
				readTime := requestTime.Add(time.Duration(i+1) * time.Millisecond)
				requestTimed = append(requestTimed, copilotTimedHeartbeat{
					timestamp: readTime,
					heartbeat: g.fileHeartbeat(
						path,
						session.SessionID,
						g.tokensForFirstHeartbeat(assignTokens, tokens),
						readTime,
						g.UserAgents,
						g.FallbackUserAgent,
						plugin,
						false,
						nil,
					),
				})
				assignTokens = false
			}

			if completedAt := g.requestCompletedAt(request); !completedAt.IsZero() {
				requestTimed = append(requestTimed, copilotTimedHeartbeat{
					timestamp: completedAt,
					heartbeat: g.appHeartbeat(
						sessionEntity,
						session.SessionID,
						g.tokensForFirstHeartbeat(assignTokens, tokens),
						0,
						completedAt,
						projectPath,
						g.UserAgents,
						g.FallbackUserAgent,
						plugin,
					),
				})
				assignTokens = false
			}

			if len(requestTimed) == 0 {
				continue
			}

			tokens = g.advanceTokens(tokens)

			timed = append(timed, requestTimed...)
		}
	}

	return timed, requestMeta
}

func (g Copilot) copilotTokenCounts(request copilotRequest, previous heartbeat.AITokens) heartbeat.AITokens {
	input, output, cumulative := g.usageCounts(request)
	if input == nil && output == nil {
		return previous
	}

	current := previous

	if cumulative {
		if input != nil {
			current.CurrentInput = int64(*input)
		}

		if output != nil {
			current.CurrentOutput = int64(*output)
		}

		return current
	}

	if input != nil {
		current.CurrentInput = previous.LastInput + int64(*input)
	}

	if output != nil {
		current.CurrentOutput = previous.LastOutput + int64(*output)
	}

	return current
}

func (Copilot) tokenDelta(tokens heartbeat.AITokens) (int64, int64) {
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

func (g Copilot) hasTokenDelta(tokens heartbeat.AITokens) bool {
	input, output := g.tokenDelta(tokens)
	return input > 0 || output > 0
}

func (Copilot) advanceTokens(tokens heartbeat.AITokens) heartbeat.AITokens {
	tokens.LastInput = tokens.CurrentInput
	tokens.LastOutput = tokens.CurrentOutput

	return tokens
}

func (Copilot) tokensForFirstHeartbeat(assign bool, tokens heartbeat.AITokens) *heartbeat.AITokens {
	if !assign {
		return nil
	}

	copy := tokens

	return &copy
}

func (g Copilot) usageCounts(request copilotRequest) (*int, *int, bool) {
	if input, output := g.parseTokenCounts(request.Usage); input != nil || output != nil {
		return input, output, true
	}

	if len(request.Result) == 0 || string(request.Result) == "null" {
		return nil, nil, false
	}

	var result struct {
		Metadata json.RawMessage `json:"metadata"`
	}
	if err := json.Unmarshal(request.Result, &result); err != nil || len(result.Metadata) == 0 {
		return nil, nil, false
	}

	if input, output := g.parseResultMetadataTokenCounts(result.Metadata); input != nil || output != nil {
		return input, output, false
	}

	return nil, nil, false
}

func (Copilot) parseTokenCounts(raw json.RawMessage) (*int, *int) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	var usage struct {
		InputTokens      *int `json:"inputTokens"`
		OutputTokens     *int `json:"outputTokens"`
		TotalTokens      *int `json:"totalTokens"`
		PromptTokens     *int `json:"promptTokens"`
		CompletionTokens *int `json:"completionTokens"`
	}
	if err := json.Unmarshal(raw, &usage); err != nil {
		return nil, nil
	}

	input := usage.PromptTokens
	if input == nil {
		input = usage.InputTokens // fallback in case they rename the attribute
	}

	output := usage.OutputTokens
	if output == nil {
		output = usage.CompletionTokens
	}

	if output == nil {
		output = usage.TotalTokens
	}

	return input, output
}

func (g Copilot) parseResultMetadataTokenCounts(raw json.RawMessage) (*int, *int) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}

	input, output := g.parseTokenCounts(raw)
	if input != nil || output != nil {
		return input, output
	}

	var metadata struct {
		Usage json.RawMessage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return nil, nil
	}

	return g.parseTokenCounts(metadata.Usage)
}

func (g Copilot) editHeartbeats(
	workspaceDir string,
	requestMeta map[string]copilotRequestMeta,
) ([]copilotTimedHeartbeat, error) {
	editSessionsDir := filepath.Join(workspaceDir, "chatEditingSessions")

	info, err := os.Stat(editSessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed reading copilot edit sessions %q: %s", editSessionsDir, err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("copilot edit sessions %q is not a directory", editSessionsDir)
	}

	entries, err := os.ReadDir(editSessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed reading copilot edit sessions %q: %s", editSessionsDir, err)
	}

	var timed []copilotTimedHeartbeat

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		statePath := filepath.Join(editSessionsDir, entry.Name(), "state.json")
		if !g.fileModifiedAfter(statePath, g.After) {
			continue
		}

		sessionTimed, err := g.parseEditState(statePath, requestMeta)
		if err != nil {
			return nil, err
		}

		timed = append(timed, sessionTimed...)
	}

	return timed, nil
}

func (Copilot) fileModifiedAfter(path string, after time.Time) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.ModTime().Before(after)
}

func (g Copilot) parseEditState(
	statePath string,
	requestMeta map[string]copilotRequestMeta,
) ([]copilotTimedHeartbeat, error) {
	//nolint:gosec
	data, err := os.ReadFile(filepath.Clean(statePath))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed reading copilot edit state %q: %s", statePath, err)
	}

	var state copilotEditState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed parsing copilot edit state %q: %s", statePath, err)
	}

	if state.Timeline == nil {
		return nil, nil
	}

	baselines := make(map[string]string)

	for _, baseline := range state.Timeline.FileBaselines {
		if len(baseline) < 2 {
			continue
		}

		var (
			key   string
			value struct {
				Content string `json:"content"`
			}
		)

		if err := json.Unmarshal(baseline[0], &key); err != nil {
			continue
		}

		if err := json.Unmarshal(baseline[1], &value); err != nil {
			continue
		}

		baselines[key] = value.Content
	}

	currentContent := make(map[string]string)

	var timed []copilotTimedHeartbeat

	for opIndex, op := range state.Timeline.Operations {
		if op.Type != "textEdit" || op.URI == nil || op.URI.FSPath == "" || len(op.Edits) == 0 {
			continue
		}

		meta, found := requestMeta[op.RequestID]
		if !found || meta.timestamp.IsZero() {
			continue
		}

		key := op.URI.FSPath + "::" + op.RequestID

		content, found := currentContent[key]
		if !found {
			if baseline, ok := baselines["file://"+op.URI.FSPath+"::"+op.RequestID]; ok {
				content = baseline
			} else if baseline, ok := baselines[g.pathToFileURI(op.URI.FSPath)+"::"+op.RequestID]; ok {
				content = baseline
			} else {
				content = ""
			}
		}

		oldLineCount := g.lineCount(content)

		updatedContent := content
		for _, edit := range op.Edits {
			updatedContent = g.applyTextEdit(updatedContent, edit)
		}

		currentContent[key] = updatedContent

		delta := g.lineCount(updatedContent) - oldLineCount
		timestamp := meta.timestamp.Add(100*time.Millisecond + time.Duration(opIndex)*time.Millisecond)
		lineChanges := heartbeat.PointerTo(delta)

		timed = append(timed, copilotTimedHeartbeat{
			timestamp: timestamp,
			heartbeat: g.fileHeartbeat(
				op.URI.FSPath,
				meta.sessionID,
				nil,
				timestamp,
				g.UserAgents,
				g.FallbackUserAgent,
				meta.plugin,
				true,
				lineChanges,
			),
		})
	}

	return timed, nil
}

func (Copilot) pathToFileURI(path string) string {
	if path == "" {
		return ""
	}

	u := url.URL{Scheme: "file", Path: path}

	return u.String()
}

func (Copilot) messageText(message *copilotMessage) string {
	if message == nil {
		return ""
	}

	return message.Text
}

func (Copilot) version(agent *copilotAgent) string {
	if agent == nil {
		return ""
	}

	return agent.ExtensionVersion
}

func (Copilot) millisTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}

	return time.UnixMilli(value)
}

func (g Copilot) requestCompletedAt(request copilotRequest) time.Time {
	if request.ModelState != nil && request.ModelState.CompletedAt > 0 {
		return g.millisTime(request.ModelState.CompletedAt)
	}

	if g.responseHasActivity(request.Response) {
		return g.millisTime(request.Timestamp).Add(time.Second)
	}

	return time.Time{}
}

func (Copilot) responseHasActivity(response []copilotResponseItem) bool {
	for _, item := range response {
		switch item.Kind {
		case "thinking", "markdown", "toolInvocationSerialized", "prepareToolInvocation", "textEditGroup", "inlineReference":
			return true
		}
	}

	return false
}

func (g Copilot) requestProjectPath(request copilotRequest) string {
	for _, path := range g.readFiles(request) {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}

		if path != "" {
			return filepath.Dir(path)
		}
	}

	if request.VariableData != nil {
		for _, variable := range request.VariableData.Variables {
			if path := g.variablePath(variable); path != "" {
				return filepath.Dir(path)
			}
		}
	}

	return ""
}

func (g Copilot) readFiles(request copilotRequest) []string {
	seen := make(map[string]struct{})

	var files []string

	if request.VariableData != nil {
		for _, variable := range request.VariableData.Variables {
			if path := g.variablePath(variable); path != "" && g.shouldTrackReadPath(path) {
				if _, found := seen[path]; !found {
					seen[path] = struct{}{}
					files = append(files, path)
				}
			}
		}
	}

	for _, item := range request.Response {
		for _, path := range g.responsePaths(item) {
			if !g.shouldTrackReadPath(path) {
				continue
			}

			if _, found := seen[path]; !found {
				seen[path] = struct{}{}
				files = append(files, path)
			}
		}
	}

	return files
}

func (g Copilot) variablePath(variable copilotVariable) string {
	if variable.Value != nil {
		if variable.Value.URI != nil && variable.Value.URI.FSPath != "" {
			return variable.Value.URI.FSPath
		}

		if variable.Value.FSPath != "" {
			return variable.Value.FSPath
		}
	}

	if strings.HasPrefix(variable.ID, "file://") {
		if path, err := g.fileURIToPath(variable.ID); err == nil {
			return path
		}
	}

	return ""
}

func (g Copilot) responsePaths(item copilotResponseItem) []string {
	var files []string

	collect := func(message *copilotMessageWithURIs) {
		if message == nil {
			return
		}

		for key, ref := range message.URIs {
			path := ref.FSPath
			if path == "" {
				path = ref.Path
			}

			if path == "" && strings.HasPrefix(key, "file://") {
				if parsed, err := g.fileURIToPath(key); err == nil {
					path = parsed
				}
			}

			if path != "" {
				files = append(files, path)
			}
		}
	}

	collect(item.InvocationMessage)
	collect(item.PastTenseMessage)

	if item.InlineReference != nil {
		if item.InlineReference.FSPath != "" {
			files = append(files, item.InlineReference.FSPath)
		} else if item.InlineReference.URI != nil && item.InlineReference.URI.FSPath != "" {
			files = append(files, item.InlineReference.URI.FSPath)
		}
	}

	return files
}

func (Copilot) fileURIToPath(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}

	return parsed.Path, nil
}

func (Copilot) shouldTrackReadPath(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	return !info.IsDir()
}

func (Copilot) appHeartbeat(
	entity string,
	sessionID string,
	aiTokens *heartbeat.AITokens,
	promptLength int,
	timestamp time.Time,
	projectPath string,
	userAgents map[string]string,
	fallbackUserAgent string,
	plugin string,
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
			aiUserAgent(entity, userAgents, fallbackUserAgent, plugin),
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
		aiUserAgent(entity, userAgents, fallbackUserAgent, plugin),
	)
	h.AISession = sessionID
	h.AIPromptLength = promptLength

	return h
}

func (Copilot) fileHeartbeat(
	path string,
	sessionID string,
	aiTokens *heartbeat.AITokens,
	timestamp time.Time,
	userAgents map[string]string,
	fallbackUserAgent string,
	plugin string,
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
			aiUserAgent(path, userAgents, fallbackUserAgent, plugin),
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
		aiUserAgent(path, userAgents, fallbackUserAgent, plugin),
	)
	h.AISession = sessionID

	return h
}

func (Copilot) lineCount(content string) int {
	if content == "" {
		return 0
	}

	return countStringLines(content)
}

func (g Copilot) applyTextEdit(content string, edit copilotTextEdit) string {
	if edit.Range == nil {
		return edit.Text
	}

	start := g.offset(content, edit.Range.StartLineNumber, edit.Range.StartColumn)

	end := g.offset(content, edit.Range.EndLineNumber, edit.Range.EndColumn)
	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}

	if end > len(content) {
		end = len(content)
	}

	return content[:start] + edit.Text + content[end:]
}

func (Copilot) offset(content string, line int, column int) int {
	if line <= 1 && column <= 1 {
		return 0
	}

	currentLine := 1

	currentColumn := 1
	for offset, r := range content {
		if currentLine == line && currentColumn == column {
			return offset
		}

		if r == '\n' {
			currentLine++
			currentColumn = 1
		} else {
			currentColumn++
		}
	}

	return len(content)
}

// Name returns its name.
func (Copilot) Name() string {
	return "Copilot"
}
