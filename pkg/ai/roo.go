package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	pathpkg "path"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

// RooCode contains params for detecting heartbeats from Roo Code task logs.
type RooCode ParserConfig

type (
	rooUIMessage struct {
		Timestamp int64   `json:"ts"`
		Type      string  `json:"type"`
		Ask       *string `json:"ask"`
		Say       *string `json:"say"`
		Text      *string `json:"text"`
	}

	rooAPIRequest struct {
		Request   string `json:"request"`
		TokensIn  int64  `json:"tokensIn"`
		TokensOut int64  `json:"tokensOut"`
	}

	rooToolAsk struct {
		Tool    string `json:"tool"`
		Path    string `json:"path"`
		Diff    string `json:"diff"`
		Content string `json:"content"`
	}
)

// Parse parses the Roo Code task logs for ai heartbeats.
func (g RooCode) Parse(ctx context.Context) (Heartbeats, error) {
	taskDirs, err := g.taskDirs(ctx)
	if err != nil {
		return nil, err
	}

	if len(taskDirs) == 0 {
		return Heartbeats{}, nil
	}

	var heartbeats Heartbeats

	for _, taskDir := range taskDirs {
		parsed, err := g.parseTaskDir(taskDir)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g RooCode) taskDirs(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	codeStorage := func(parts ...string) string {
		return filepath.Join(append([]string{home}, parts...)...)
	}

	roots := []string{
		codeStorage(
			"Library", "Application Support", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks",
		),
		codeStorage(
			"Library", "Application Support", "Cursor", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks",
		),
		codeStorage("AppData", "Roaming", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage("AppData", "Roaming", "Cursor", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage(".config", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage(".config", "Cursor", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage(".vscode-server", "data", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage(".cursor-server", "data", "User", "globalStorage", "RooVeterinaryInc.roo-cline", "tasks"),
		codeStorage(
			"Library", "Application Support", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-code-nightly", "tasks",
		),
		codeStorage("AppData", "Roaming", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-code-nightly", "tasks"),
		codeStorage(".config", "Code", "User", "globalStorage", "RooVeterinaryInc.roo-code-nightly", "tasks"),
		codeStorage(".vscode-server", "data", "User", "globalStorage", "RooVeterinaryInc.roo-code-nightly", "tasks"),
	}

	var taskDirs []string

	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to read Roo Code tasks directory %q: %s", root, err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			taskDir := filepath.Join(root, entry.Name())
			if !g.taskDirModifiedAfter(taskDir) {
				continue
			}

			taskDirs = append(taskDirs, taskDir)
		}
	}

	return taskDirs, nil
}

func (g RooCode) taskDirModifiedAfter(taskDir string) bool {
	if g.After.IsZero() {
		return true
	}

	paths := []string{
		filepath.Join(taskDir, "ui_messages.json"),
		filepath.Join(taskDir, "api_conversation_history.json"),
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && !info.ModTime().Before(g.After) {
			return true
		}
	}

	return false
}

func (g RooCode) parseTaskDir(taskDir string) (Heartbeats, error) {
	// nolint:gosec // Reads a task log from a previously discovered Roo Code task directory.
	contents, err := os.ReadFile(filepath.Clean(filepath.Join(taskDir, "ui_messages.json")))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read Roo Code ui messages from %q: %s", taskDir, err)
	}

	var messages []rooUIMessage
	if err := json.Unmarshal(contents, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse Roo Code ui messages from %q: %s", taskDir, err)
	}

	sessionID := filepath.Base(taskDir)
	sessionEntity := appHeartbeatEntity(g.Name(), sessionID)
	cwd := ""

	var heartbeats Heartbeats

	for _, message := range messages {
		timestamp := time.UnixMilli(message.Timestamp)
		if timestamp.IsZero() || timestamp.Before(g.After) {
			continue
		}

		if message.Say != nil && *message.Say == "api_req_started" && message.Text != nil {
			request := rooAPIRequest{}
			if err := json.Unmarshal([]byte(*message.Text), &request); err == nil {
				if parsedCwd := rooCurrentWorkingDirectory(request.Request); parsedCwd != "" {
					cwd = parsedCwd
				}

				heartbeats = append(heartbeats, g.appHeartbeat(
					timestamp,
					sessionEntity,
					sessionID,
					cwd,
					rooTaskText(request.Request),
					heartbeat.AITokens{
						CurrentInput:  request.TokensIn,
						CurrentOutput: request.TokensOut,
					},
				))
			}
		}

		if message.Ask != nil && *message.Ask == "tool" && message.Text != nil {
			tool := rooToolAsk{}
			if err := json.Unmarshal([]byte(*message.Text), &tool); err != nil {
				continue
			}

			filePath, lineChanges, ok := rooToolHeartbeat(tool, cwd)
			if !ok {
				continue
			}

			heartbeats = append(heartbeats, g.fileHeartbeat(timestamp, sessionID, filePath, lineChanges))
		}
	}

	return heartbeats, nil
}

func (g RooCode) appHeartbeat(
	timestamp time.Time,
	entity string,
	sessionID string,
	cwd string,
	taskText string,
	tokens heartbeat.AITokens,
) heartbeat.Heartbeat {
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
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, "", ""),
	)
	h.AIPromptLength = promptLength(taskText)

	return h
}

func (g RooCode) fileHeartbeat(
	timestamp time.Time,
	sessionID string,
	filePath string,
	lineChanges int,
) heartbeat.Heartbeat {
	return heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		sessionID,
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
		float64(timestamp.Unix()),
		aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, "", ""),
	)
}

func rooTaskText(request string) string {
	task, ok := rooBetween(request, "<task>", "</task>")
	if !ok {
		return ""
	}

	return strings.TrimSpace(task)
}

func rooCurrentWorkingDirectory(request string) string {
	const prefix = "# Current Working Directory ("

	idx := strings.Index(request, prefix)
	if idx == -1 {
		return ""
	}

	remaining := request[idx+len(prefix):]

	end := strings.Index(remaining, ")")
	if end == -1 {
		return ""
	}

	return strings.TrimSpace(remaining[:end])
}

func rooToolHeartbeat(tool rooToolAsk, cwd string) (string, int, bool) {
	if tool.Path == "" {
		return "", 0, false
	}

	filePath := tool.Path
	if !rooIsAbsPath(filePath) && cwd != "" {
		filePath = rooJoinPath(cwd, filePath)
	}

	switch tool.Tool {
	case "appliedDiff":
		return filePath, rooAppliedDiffLineChanges(tool.Diff), true
	case "writeToFile":
		if strings.TrimSpace(tool.Content) == "" {
			return "", 0, false
		}

		return filePath, countStringLines(tool.Content), true
	default:
		return "", 0, false
	}
}

func rooIsAbsPath(filePath string) bool {
	if filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "/") {
		return true
	}

	return rooUsesPosixPaths(filePath)
}

func rooJoinPath(base string, filePath string) string {
	if rooUsesPosixPaths(base) || strings.HasPrefix(filePath, "/") {
		return pathpkg.Join(base, filePath)
	}

	return filepath.Join(base, filePath)
}

func rooUsesPosixPaths(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.Contains(path, `\`)
}

func rooAppliedDiffLineChanges(diff string) int {
	lines := strings.Split(diff, "\n")
	oldLines := 0
	newLines := 0
	total := 0
	state := ""

	flush := func() {
		if state == "" {
			return
		}

		total += newLines - oldLines
		oldLines = 0
		newLines = 0
		state = ""
	}

	for _, line := range lines {
		switch line {
		case "<<<<<<< SEARCH":
			flush()

			state = "old"
		case "=======":
			if state == "old" {
				state = "new"
			}
		case ">>>>>>> REPLACE":
			flush()
		default:
			switch state {
			case "old":
				oldLines++
			case "new":
				newLines++
			}
		}
	}

	flush()

	return total
}

func rooBetween(value string, start string, end string) (string, bool) {
	after, ok := strings.CutPrefix(value, start)
	if !ok {
		startIdx := strings.Index(value, start)
		if startIdx == -1 {
			return "", false
		}

		after = value[startIdx+len(start):]
	}

	beforeEnd, _, ok := strings.Cut(after, end)
	if !ok {
		return "", false
	}

	return beforeEnd, true
}

// Name returns its id.
func (RooCode) Name() string {
	return "Roo Code"
}
