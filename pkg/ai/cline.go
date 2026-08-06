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

// Cline contains params for detecting heartbeats from Cline task logs.
type Cline ParserConfig

type (
	clineUIMessage struct {
		Timestamp int64   `json:"ts"`
		Type      string  `json:"type"`
		Ask       *string `json:"ask"`
		Say       *string `json:"say"`
		Text      *string `json:"text"`
	}

	clineAPIRequest struct {
		Request     string `json:"request"`
		TokensIn    int64  `json:"tokensIn"`
		TokensOut   int64  `json:"tokensOut"`
		CacheWrites int64  `json:"cacheWrites"`
		CacheReads  int64  `json:"cacheReads"`
	}

	clineToolMessage struct {
		Tool    string `json:"tool"`
		Path    string `json:"path"`
		Diff    string `json:"diff"`
		Content string `json:"content"`
	}

	clineToolHeartbeat struct {
		filePath    string
		lineChanges int
		isWrite     bool
		ok          bool
	}
)

// Parse parses the Cline task logs for ai heartbeats.
func (g Cline) Parse(ctx context.Context) (Heartbeats, error) {
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

func (g Cline) taskDirs(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	codeStorage := func(parts ...string) string {
		return filepath.Join(append([]string{home}, parts...)...)
	}

	roots := []string{
		codeStorage(
			"Library", "Application Support", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			"Library", "Application Support", "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			"Library", "Application Support", "Windsurf", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			"AppData", "Roaming", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			"AppData", "Roaming", "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			"AppData", "Roaming", "Windsurf", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			".config", "Cursor", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			".config", "Windsurf", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			".vscode-server", "data", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
		codeStorage(
			".cursor-server", "data", "User", "globalStorage", "saoudrizwan.claude-dev", "tasks",
		),
	}

	var taskDirs []string

	for _, root := range roots {
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to read Cline tasks directory %q: %s", root, err)
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

func (g Cline) taskDirModifiedAfter(taskDir string) bool {
	if g.After.IsZero() {
		return true
	}

	paths := []string{
		filepath.Join(taskDir, "ui_messages.json"),
		filepath.Join(taskDir, "api_conversation_history.json"),
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err == nil && timestampAtOrAfterCutoff(info.ModTime(), g.After) {
			return true
		}
	}

	return false
}

func (g Cline) parseTaskDir(taskDir string) (Heartbeats, error) {
	// nolint:gosec
	contents, err := os.ReadFile(filepath.Clean(filepath.Join(taskDir, "ui_messages.json")))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read Cline ui messages from %q: %s", taskDir, err)
	}

	var messages []clineUIMessage
	if err := json.Unmarshal(contents, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse Cline ui messages from %q: %s", taskDir, err)
	}

	sessionID := filepath.Base(taskDir)
	sessionEntity := appHeartbeatEntity(g.Name(), sessionID)
	cwd := ""

	var heartbeats Heartbeats

	for _, message := range messages {
		timestamp := time.UnixMilli(message.Timestamp)
		if timestamp.IsZero() || !timestampAtOrAfterCutoff(timestamp, g.After) {
			continue
		}

		if message.Say != nil && *message.Say == "api_req_started" && message.Text != nil {
			request := clineAPIRequest{}
			if err := json.Unmarshal([]byte(*message.Text), &request); err == nil {
				if parsedCwd := clineCurrentWorkingDirectory(request.Request); parsedCwd != "" {
					cwd = parsedCwd
				}

				heartbeats = append(heartbeats, g.appHeartbeat(
					timestamp,
					sessionEntity,
					sessionID,
					cwd,
					clineTaskText(request.Request),
					heartbeat.AITokens{
						CurrentInput:       max(request.TokensIn, 0) + max(request.CacheWrites, 0),
						CurrentCachedInput: max(request.CacheReads, 0),
						CurrentOutput:      max(request.TokensOut, 0),
					},
				))
			}
		}

		if message.Text == nil {
			continue
		}

		if message.Say != nil && *message.Say == "tool" {
			var tool clineToolMessage
			if err := json.Unmarshal([]byte(*message.Text), &tool); err != nil {
				continue
			}

			fileHB := g.toolHeartbeat(tool, cwd)
			if !fileHB.ok {
				continue
			}

			heartbeats = append(
				heartbeats,
				g.fileHeartbeat(
					timestamp,
					sessionID,
					fileHB.filePath,
					fileHB.lineChanges,
					fileHB.isWrite,
				),
			)
		}
	}

	return heartbeats, nil
}

func (g Cline) appHeartbeat(
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
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, "", ""),
	)
	h.AIPromptLength = promptLength(taskText)

	return h
}

func (g Cline) fileHeartbeat(
	timestamp time.Time,
	sessionID string,
	filePath string,
	lineChanges int,
	isWrite bool,
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
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, "", ""),
	)
}

func (Cline) toolHeartbeat(tool clineToolMessage, cwd string) clineToolHeartbeat {
	if tool.Path == "" {
		return clineToolHeartbeat{}
	}

	filePath := tool.Path
	if !clineIsAbsPath(filePath) && cwd != "" {
		filePath = clineJoinPath(cwd, filePath)
	}

	switch tool.Tool {
	case "editedExistingFile":
		if strings.TrimSpace(tool.Diff) != "" {
			return clineToolHeartbeat{
				filePath:    filePath,
				lineChanges: clineAppliedDiffLineChanges(tool.Diff),
				isWrite:     true,
				ok:          true,
			}
		}

		if strings.TrimSpace(tool.Content) == "" {
			return clineToolHeartbeat{}
		}

		return clineToolHeartbeat{
			filePath:    filePath,
			lineChanges: countStringLines(tool.Content),
			isWrite:     true,
			ok:          true,
		}
	case "newFileCreated":
		if strings.TrimSpace(tool.Content) == "" {
			return clineToolHeartbeat{}
		}

		return clineToolHeartbeat{
			filePath:    filePath,
			lineChanges: countStringLines(tool.Content),
			isWrite:     true,
			ok:          true,
		}
	case "readFile":
		return clineToolHeartbeat{
			filePath: filePath,
			ok:       true,
		}
	case "fileDeleted":
		if strings.TrimSpace(tool.Content) == "" {
			return clineToolHeartbeat{}
		}

		return clineToolHeartbeat{
			filePath:    filePath,
			lineChanges: -countStringLines(tool.Content),
			isWrite:     true,
			ok:          true,
		}
	default:
		return clineToolHeartbeat{}
	}
}

func clineTaskText(request string) string {
	task, ok := clineBetween(request, "<task>", "</task>")
	if !ok {
		return ""
	}

	return strings.TrimSpace(task)
}

func clineCurrentWorkingDirectory(request string) string {
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

func clineIsAbsPath(filePath string) bool {
	if filepath.IsAbs(filePath) || strings.HasPrefix(filePath, "/") {
		return true
	}

	return clineUsesPosixPaths(filePath)
}

func clineJoinPath(base string, filePath string) string {
	if clineUsesPosixPaths(base) || strings.HasPrefix(filePath, "/") {
		return pathpkg.Join(base, filePath)
	}

	return filepath.Join(base, filePath)
}

func clineUsesPosixPaths(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.Contains(path, `\`)
}

func clineAppliedDiffLineChanges(diff string) int {
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

func clineBetween(value string, start string, end string) (string, bool) {
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
func (Cline) Name() string {
	return "Cline"
}
