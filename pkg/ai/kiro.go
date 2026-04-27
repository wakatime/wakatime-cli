package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	pathutil "path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

// Kiro contains params for detecting heartbeats from Kiro local activity.
type Kiro ParserConfig

type (
	kiroSession struct {
		SessionID          string             `json:"sessionId"`
		WorkspacePath      string             `json:"workspacePath"`
		WorkspaceDirectory string             `json:"workspaceDirectory"`
		History            []kiroHistoryEntry `json:"history"`
	}

	kiroHistoryEntry struct {
		ExecutionID string      `json:"executionId"`
		Message     kiroMessage `json:"message"`
	}

	kiroMessage struct {
		Role    string          `json:"role"`
		Content json.RawMessage `json:"content"`
	}

	kiroContentBlock struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}

	kiroSessionInfo struct {
		SessionID string
		Workspace string
	}

	kiroPrompt struct {
		SessionID string
		Workspace string
		Length    int
	}

	kiroExecution struct {
		ExecutionID   string       `json:"executionId"`
		ChatSessionID string       `json:"chatSessionId"`
		StartTimeMS   int64        `json:"startTime"`
		Input         kiroInput    `json:"input"`
		Actions       []kiroAction `json:"actions"`
	}

	kiroInput struct {
		Data kiroInputData `json:"data"`
	}

	kiroInputData struct {
		ChatSessionID      string `json:"chatSessionId"`
		WorkspacePath      string `json:"workspacePath"`
		WorkspaceDirectory string `json:"workspaceDirectory"`
		WorkspaceRoot      string `json:"workspaceRoot"`
	}

	kiroAction struct {
		ActionType  string          `json:"actionType"`
		ActionState string          `json:"actionState"`
		EmittedAtMS int64           `json:"emittedAt"`
		Input       kiroActionInput `json:"input"`
	}

	kiroActionInput struct {
		File            string           `json:"file"`
		Path            string           `json:"path"`
		Local           string           `json:"local"`
		OriginalContent string           `json:"originalContent"`
		ModifiedContent string           `json:"modifiedContent"`
		Files           []kiroActionFile `json:"files"`
	}

	kiroActionFile struct {
		Path string `json:"path"`
	}
)

// Parse parses Kiro's local JSON session and execution logs for ai heartbeats.
func (g Kiro) Parse(ctx context.Context) (Heartbeats, error) {
	root, err := g.storageRoot(ctx)
	if err != nil {
		return nil, err
	}

	if root == "" {
		return Heartbeats{}, nil
	}

	sessions, prompts, err := kiroSessions(root)
	if err != nil {
		return nil, err
	}

	executions, err := kiroExecutions(root)
	if err != nil {
		return nil, err
	}

	sort.Slice(executions, func(i, j int) bool {
		return executions[i].StartTimeMS < executions[j].StartTimeMS
	})

	var heartbeats Heartbeats

	for _, execution := range executions {
		session := kiroExecutionSession(execution, sessions)
		if session.SessionID == "" || execution.StartTimeMS == 0 {
			continue
		}

		if prompt, ok := prompts[execution.ExecutionID]; ok {
			timestamp := time.UnixMilli(execution.StartTimeMS)
			if !timestamp.Before(g.After) {
				heartbeats = append(heartbeats, g.promptHeartbeat(prompt, timestamp))
			}
		}

		heartbeats = append(heartbeats, g.actionHeartbeats(execution, session)...)
	}

	return heartbeats, nil
}

func (Kiro) storageRoot(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Kiro", "User", "globalStorage", "kiro.kiroagent"),
		filepath.Join(home, "AppData", "Roaming", "Kiro", "User", "globalStorage", "kiro.kiroagent"),
		filepath.Join(home, ".config", "Kiro", "User", "globalStorage", "kiro.kiroagent"),
		filepath.Join(home, ".config", "kiro", "User", "globalStorage", "kiro.kiroagent"),
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	return "", nil
}

func kiroSessions(root string) (map[string]kiroSessionInfo, map[string]kiroPrompt, error) {
	sessionPaths, err := filepath.Glob(filepath.Join(root, "workspace-sessions", "*", "*.json"))
	if err != nil {
		return nil, nil, fmt.Errorf("failed globbing kiro sessions: %s", err)
	}

	sessions := make(map[string]kiroSessionInfo)
	prompts := make(map[string]kiroPrompt)

	for _, sessionPath := range sessionPaths {
		if filepath.Base(sessionPath) == "sessions.json" {
			continue
		}

		var session kiroSession
		if err := kiroReadJSON(root, sessionPath, &session); err != nil {
			continue
		}

		if session.SessionID == "" {
			continue
		}

		workspace := firstNonEmptyString(session.WorkspacePath, session.WorkspaceDirectory)
		sessions[session.SessionID] = kiroSessionInfo{
			SessionID: session.SessionID,
			Workspace: workspace,
		}

		var pendingPrompt string

		for _, entry := range session.History {
			switch entry.Message.Role {
			case "user":
				pendingPrompt = kiroMessageText(entry.Message.Content)
			case "assistant":
				if pendingPrompt == "" || entry.ExecutionID == "" {
					continue
				}

				prompts[entry.ExecutionID] = kiroPrompt{
					SessionID: session.SessionID,
					Workspace: workspace,
					Length:    promptLength(pendingPrompt),
				}
				pendingPrompt = ""
			}
		}
	}

	return sessions, prompts, nil
}

func kiroExecutions(root string) ([]kiroExecution, error) {
	var executions []kiroExecution

	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if entry.IsDir() {
			if path != root && filepath.Base(path) == "workspace-sessions" {
				return filepath.SkipDir
			}

			if kiroPathDepth(root, path) >= 3 {
				return filepath.SkipDir
			}

			return nil
		}

		if kiroPathDepth(root, path) > 3 {
			return nil
		}

		var execution kiroExecution
		if err := kiroReadJSON(root, path, &execution); err != nil {
			return nil
		}

		if execution.ExecutionID == "" || execution.StartTimeMS == 0 {
			return nil
		}

		executions = append(executions, execution)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed walking kiro storage root: %s", err)
	}

	return executions, nil
}

func kiroPathDepth(root string, filePath string) int {
	rel, err := filepath.Rel(root, filePath)
	if err != nil || rel == "." {
		return 0
	}

	return len(strings.Split(rel, string(os.PathSeparator)))
}

func kiroReadJSON(root string, filePath string, target any) error {
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	cleanPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}

	rel, err := filepath.Rel(cleanRoot, cleanPath)
	if err != nil {
		return err
	}

	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("kiro path outside storage root: %s", filePath)
	}

	data, err := os.ReadFile(cleanPath) // #nosec G304 -- path is discovered and verified under Kiro storage root.
	if err != nil {
		return err
	}

	if len(data) == 0 || data[0] != '{' {
		return fmt.Errorf("not a kiro json object: %s", filePath)
	}

	return json.Unmarshal(data, target)
}

func kiroMessageText(raw json.RawMessage) string {
	var blocks []kiroContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var parts []string

		for _, block := range blocks {
			if block.Type == "text" && strings.TrimSpace(block.Text) != "" {
				parts = append(parts, block.Text)
			}
		}

		return strings.TrimSpace(strings.Join(parts, "\n"))
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		return strings.TrimSpace(text)
	}

	return ""
}

func kiroExecutionSession(execution kiroExecution, sessions map[string]kiroSessionInfo) kiroSessionInfo {
	sessionID := firstNonEmptyString(execution.ChatSessionID, execution.Input.Data.ChatSessionID)
	workspace := firstNonEmptyString(
		execution.Input.Data.WorkspacePath,
		execution.Input.Data.WorkspaceDirectory,
		execution.Input.Data.WorkspaceRoot,
	)

	if fromSession, ok := sessions[sessionID]; ok {
		if workspace == "" {
			workspace = fromSession.Workspace
		}

		return kiroSessionInfo{SessionID: sessionID, Workspace: workspace}
	}

	return kiroSessionInfo{SessionID: sessionID, Workspace: workspace}
}

func (g Kiro) promptHeartbeat(prompt kiroPrompt, timestamp time.Time) heartbeat.Heartbeat {
	entity := appHeartbeatEntity("Kiro", prompt.SessionID)

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
		prompt.Workspace,
		float64(timestamp.UnixMilli())/1000,
		aiUserAgent(entity, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
	)
	h.AISession = prompt.SessionID
	h.AIPromptLength = prompt.Length

	return h
}

func (g Kiro) actionHeartbeats(execution kiroExecution, session kiroSessionInfo) Heartbeats {
	var heartbeats Heartbeats

	for _, action := range execution.Actions {
		if action.ActionState != "Accepted" {
			continue
		}

		timestampMS := action.EmittedAtMS
		if timestampMS == 0 {
			timestampMS = execution.StartTimeMS
		}

		timestamp := time.UnixMilli(timestampMS)
		if timestamp.IsZero() || timestamp.Before(g.After) {
			continue
		}

		switch action.ActionType {
		case "readFiles":
			for _, file := range action.Input.Files {
				if heartbeat := g.fileHeartbeat(file.Path, false, nil, session, timestamp); heartbeat != nil {
					heartbeats = append(heartbeats, *heartbeat)
				}
			}
		case "replace":
			lineChanges := kiroLineChanges(action.Input.OriginalContent, action.Input.ModifiedContent)

			filePath := firstNonEmptyString(action.Input.Local, action.Input.File, action.Input.Path)
			if heartbeat := g.fileHeartbeat(filePath, true, &lineChanges, session, timestamp); heartbeat != nil {
				heartbeats = append(heartbeats, *heartbeat)
			}
		}
	}

	return heartbeats
}

func (g Kiro) fileHeartbeat(
	rawPath string,
	isWrite bool,
	lineChanges *int,
	session kiroSessionInfo,
	timestamp time.Time,
) *heartbeat.Heartbeat {
	filePath := kiroFilePath(rawPath, session.Workspace)
	if filePath == "" {
		return nil
	}

	h := heartbeat.New(
		lineChanges,
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		filePath,
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
		aiUserAgent(filePath, g.UserAgents, g.FallbackUserAgent, aiPlugin(g, "")),
	)
	h.AISession = session.SessionID

	return &h
}

func kiroFilePath(rawPath string, workspace string) string {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return ""
	}

	if strings.HasPrefix(rawPath, "file://") {
		uri, err := url.Parse(rawPath)
		if err == nil && uri.Path != "" {
			rawPath = uri.Path
		}
	}

	if len(rawPath) >= 3 && rawPath[0] == '/' && rawPath[2] == ':' {
		return filepath.FromSlash(rawPath[1:])
	}

	if strings.HasPrefix(rawPath, "/") {
		return rawPath
	}

	if filepath.IsAbs(rawPath) {
		return filepath.FromSlash(rawPath)
	}

	if workspace == "" {
		return filepath.FromSlash(rawPath)
	}

	if strings.HasPrefix(workspace, "/") {
		return pathutil.Join(workspace, rawPath)
	}

	return filepath.Join(filepath.FromSlash(workspace), filepath.FromSlash(rawPath))
}

func kiroLineChanges(original string, modified string) int {
	if strings.TrimSpace(original) == "" {
		return countStringLines(modified)
	}

	if strings.TrimSpace(modified) == "" {
		return -countStringLines(original)
	}

	return countStringLines(modified) - countStringLines(original)
}

// Name returns its name.
func (Kiro) Name() string {
	return "Kiro"
}
