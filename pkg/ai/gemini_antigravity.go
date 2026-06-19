package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
)

type (
	antigravityProduct struct {
		appDir           string
		entityName       string
		userAgentProduct string
	}

	antigravityTranscript struct {
		path             string
		entityName       string
		userAgentProduct string
		sessionID        string
	}

	antigravityLogLine struct {
		Source    string                   `json:"source"`
		Type      string                   `json:"type"`
		Status    string                   `json:"status"`
		CreatedAt string                   `json:"created_at"`
		Content   string                   `json:"content"`
		Model     string                   `json:"model"`
		ToolCalls []antigravityLogToolCall `json:"tool_calls"`
	}

	antigravityLogToolCall struct {
		Args map[string]json.RawMessage `json:"args"`
	}

	antigravityHistoryLine struct {
		ConversationID string `json:"conversationId"`
		Workspace      string `json:"workspace"`
	}
)

func antigravityProductDirs() []antigravityProduct {
	return []antigravityProduct{
		{appDir: "antigravity", entityName: "Antigravity Desktop", userAgentProduct: "antigravity-desktop"},
		{appDir: "antigravity-ide", entityName: "Antigravity IDE", userAgentProduct: "antigravity-ide"},
		{appDir: "antigravity-cli", entityName: "Antigravity CLI", userAgentProduct: "antigravity-cli"},
	}
}

func antigravityProductVersion(
	home string,
	appRoot string,
	product antigravityProduct,
) string {
	switch product.userAgentProduct {
	case "antigravity-desktop":
		for _, path := range antigravityDesktopPlistPaths(home, appRoot) {
			if version := antigravityPlistVersion(path); version != "" {
				return version
			}
		}
	case "antigravity-ide":
		for _, path := range antigravityIDEProductPaths(home, appRoot) {
			if version := antigravityJSONVersion(path); version != "" {
				return version
			}
		}
	case "antigravity-cli":
		return "unknown"
	}

	return "unknown"
}

func antigravityDesktopPlistPaths(home string, appRoot string) []string {
	var paths []string

	if executable := antigravityAgentAPIExecutable(appRoot); executable != "" {
		slashed := filepath.ToSlash(executable)
		if end := strings.Index(slashed, ".app/"); end != -1 {
			bundle := filepath.FromSlash(slashed[:end+len(".app")])
			paths = append(paths, filepath.Join(bundle, "Contents", "Info.plist"))
		}
	}

	for _, parent := range []string{"/Applications", filepath.Join(home, "Applications")} {
		paths = append(paths, filepath.Join(parent, "Antigravity.app", "Contents", "Info.plist"))
	}

	return paths
}

func antigravityIDEProductPaths(home string, appRoot string) []string {
	var paths []string

	if executable := antigravityAgentAPIExecutable(appRoot); executable != "" {
		slashed := filepath.ToSlash(executable)
		if end := strings.Index(slashed, "/extensions/"); end != -1 {
			paths = append(paths, filepath.Join(filepath.FromSlash(slashed[:end]), "product.json"))
		}
	}

	paths = append(paths,
		filepath.Join("/Applications", "Antigravity IDE.app", "Contents", "Resources", "app", "product.json"),
		filepath.Join(home, "Applications", "Antigravity IDE.app", "Contents", "Resources", "app", "product.json"),
		filepath.Join(home, "AppData", "Local", "Programs", "Antigravity IDE", "resources", "app", "product.json"),
		filepath.Join("/usr", "share", "antigravity-ide", "resources", "app", "product.json"),
		filepath.Join("/opt", "Antigravity IDE", "resources", "app", "product.json"),
	)

	return paths
}

func antigravityAgentAPIExecutable(appRoot string) string {
	contents, err := os.ReadFile(filepath.Clean(filepath.Join(appRoot, "bin", "agentapi")))
	if err != nil {
		return ""
	}

	_, remainder, found := strings.Cut(string(contents), `exec "`)
	if !found {
		return ""
	}

	executable, _, found := strings.Cut(remainder, `"`)
	if !found {
		return ""
	}

	return antigravityFilePath(executable)
}

func antigravityJSONVersion(path string) string {
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}

	var metadata struct {
		Version string `json:"version"`
	}
	if json.Unmarshal(contents, &metadata) != nil {
		return ""
	}

	return antigravityVersionString(metadata.Version)
}

func antigravityPlistVersion(path string) string {
	contents, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return ""
	}

	const key = "<key>CFBundleShortVersionString</key>"

	_, remainder, found := strings.Cut(string(contents), key)
	if !found {
		return ""
	}

	_, remainder, found = strings.Cut(remainder, "<string>")
	if !found {
		return ""
	}

	value, _, found := strings.Cut(remainder, "</string>")
	if !found {
		return ""
	}

	return antigravityVersionString(value)
}

func antigravityVersionString(value string) string {
	for _, field := range strings.Fields(value) {
		field = strings.Trim(strings.TrimSpace(field), "vV,;()[]{}")
		if field == "" || !strings.ContainsAny(field, "0123456789") {
			continue
		}

		valid := true

		for _, char := range field {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || strings.ContainsRune(".-_+", char) {
				continue
			}

			valid = false

			break
		}

		if valid {
			return strings.ToLower(field)
		}
	}

	return ""
}

func (g Gemini) parseAntigravity(ctx context.Context) (Heartbeats, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	transcripts, err := g.antigravityTranscriptPaths(home)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		projectPath := antigravityProjectPath(transcript)

		parsed, err := g.parseAntigravityTranscript(transcript, projectPath)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Gemini) antigravityTranscriptPaths(home string) ([]antigravityTranscript, error) {
	var transcripts []antigravityTranscript

	for _, product := range antigravityProductDirs() {
		appRoot := filepath.Join(home, ".gemini", product.appDir)
		brainDir := filepath.Join(appRoot, "brain")

		if _, err := os.Stat(brainDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to stat %s brain directory: %s", product.entityName, err)
		}

		userAgentProduct := product.userAgentProduct + "/" + antigravityProductVersion(home, appRoot, product)

		err := filepath.WalkDir(brainDir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() || entry.Name() != "transcript.jsonl" {
				return nil
			}

			logsDir := filepath.Dir(path)
			if filepath.Base(logsDir) != "logs" || filepath.Base(filepath.Dir(logsDir)) != ".system_generated" {
				return nil
			}

			info, err := entry.Info()
			if err != nil || info.ModTime().Before(g.After) {
				return nil
			}

			sessionDir := filepath.Dir(filepath.Dir(logsDir))

			sessionID := filepath.Base(sessionDir)
			if sessionID == "." || sessionID == string(filepath.Separator) || sessionID == "" {
				return nil
			}

			transcripts = append(transcripts, antigravityTranscript{
				path:             path,
				entityName:       product.entityName,
				userAgentProduct: userAgentProduct,
				sessionID:        sessionID,
			})

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk %s brain directory: %s", product.entityName, err)
		}
	}

	return transcripts, nil
}

func (g Gemini) parseAntigravityTranscript(
	transcript antigravityTranscript,
	projectPath string,
) (Heartbeats, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript.path))
	if err != nil {
		return nil, fmt.Errorf("failed opening Antigravity transcript %q: %s", transcript.path, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	var lines []antigravityLogLine

	for scanner.Scan() {
		var line antigravityLogLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			continue
		}

		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading Antigravity transcript %q: %s", transcript.path, err)
	}

	if projectPath == "" {
		projectPath = antigravityTranscriptProjectPath(lines)
	}

	entity := appHeartbeatEntity(transcript.entityName, transcript.sessionID)

	var (
		heartbeats Heartbeats
		model      string
	)

	for _, line := range lines {
		if selectedModel := antigravitySelectedModel(line.Content); selectedModel != "" {
			model = selectedModel
		}

		if line.Model != "" {
			model = line.Model
		} else {
			line.Model = model
		}

		if line.Status != "" && line.Status != "DONE" {
			continue
		}

		timestamp := parseGeminiTime(line.CreatedAt)
		if timestamp.IsZero() || timestamp.Before(g.After) {
			continue
		}

		if h := g.antigravityAppHeartbeat(
			timestamp,
			entity,
			transcript,
			projectPath,
			line,
		); h != nil {
			heartbeats = append(heartbeats, *h)
		}

		if h := g.antigravityFileHeartbeat(timestamp, transcript, line); h != nil {
			heartbeats = append(heartbeats, *h)
		}
	}

	return heartbeats, nil
}

func antigravityTranscriptProjectPath(lines []antigravityLogLine) string {
	for _, key := range []string{"Cwd", "DirectoryPath"} {
		for _, line := range lines {
			for _, toolCall := range line.ToolCalls {
				raw, found := toolCall.Args[key]
				if !found {
					continue
				}

				var value string
				if json.Unmarshal(raw, &value) != nil {
					continue
				}

				var decoded string
				if json.Unmarshal([]byte(value), &decoded) == nil {
					value = decoded
				}

				value = antigravityFilePath(strings.Trim(value, `"`))
				if value != "" {
					return value
				}
			}
		}
	}

	return ""
}

func (g Gemini) antigravityAppHeartbeat(
	timestamp time.Time,
	entity string,
	transcript antigravityTranscript,
	projectPath string,
	line antigravityLogLine,
) *heartbeat.Heartbeat {
	content := strings.TrimSpace(line.Content)
	if content == "" {
		return nil
	}

	isPrompt := line.Source == "USER_EXPLICIT" && line.Type == "USER_INPUT"

	isResponse := line.Source == "MODEL" && line.Type == "PLANNER_RESPONSE"
	if !isPrompt && !isResponse {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		nil,
		transcript.sessionID,
		heartbeat.AITokens{},
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
		float64(timestamp.Unix()),
		aiUserAgentWithAgentPrefix(
			entity,
			g.UserAgents,
			g.FallbackUserAgent,
			transcript.userAgentProduct,
			line.Model,
		),
	)
	if isPrompt {
		h.AIPromptLength = promptLength(antigravityUserRequest(content))
	}

	return &h
}

func (g Gemini) antigravityFileHeartbeat(
	timestamp time.Time,
	transcript antigravityTranscript,
	line antigravityLogLine,
) *heartbeat.Heartbeat {
	if line.Source != "MODEL" || line.Type != "CODE_ACTION" {
		return nil
	}

	filePath := antigravityCodeActionPath(line.Content)

	lineChanges, ok := antigravityCodeActionLineChanges(line.Content)
	if filePath == "" || !ok {
		return nil
	}

	h := heartbeat.NewWithAITokens(
		heartbeat.PointerTo(lineChanges),
		transcript.sessionID,
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
		aiUserAgentWithAgentPrefix(
			filePath,
			g.UserAgents,
			g.FallbackUserAgent,
			transcript.userAgentProduct,
			line.Model,
		),
	)

	return &h
}

func antigravityUserRequest(content string) string {
	const (
		startTag = "<USER_REQUEST>"
		endTag   = "</USER_REQUEST>"
	)

	start := strings.Index(content, startTag)
	if start == -1 {
		return content
	}

	request := content[start+len(startTag):]
	if end := strings.Index(request, endTag); end != -1 {
		request = request[:end]
	}

	return strings.TrimSpace(request)
}

func antigravitySelectedModel(content string) string {
	const setting = "`Model Selection`"

	start := strings.Index(content, setting)
	if start == -1 {
		return ""
	}

	remainder := content[start+len(setting):]

	_, model, found := strings.Cut(remainder, " to ")
	if !found {
		return ""
	}

	end := len(model)
	for _, delimiter := range []string{". No need", "</USER_SETTINGS_CHANGE>", "\n"} {
		if index := strings.Index(model, delimiter); index != -1 && index < end {
			end = index
		}
	}

	model = model[:end]

	model = strings.Trim(strings.TrimSpace(model), ".`")

	fields := strings.Fields(model)
	if len(fields) < 2 {
		return strings.ToLower(model)
	}

	product := strings.ToLower(strings.Trim(fields[0], "()"))
	version := strings.ToLower(strings.Join(fields[1:], "-"))
	version = strings.NewReplacer("(", "", ")", "").Replace(version)

	version = strings.Trim(version, "-_/.")
	if product == "" || version == "" {
		return ""
	}

	return product + "/" + version
}

func antigravityCodeActionPath(content string) string {
	const marker = "The following changes were made by the "

	for _, line := range strings.Split(content, "\n") {
		if !strings.HasPrefix(line, marker) {
			continue
		}

		_, remainder, found := strings.Cut(line, " to: ")
		if !found {
			continue
		}

		if path, _, found := strings.Cut(remainder, ". If relevant"); found {
			remainder = path
		}

		return antigravityFilePath(strings.Trim(strings.TrimSpace(remainder), "`"))
	}

	return ""
}

func antigravityCodeActionLineChanges(content string) (int, bool) {
	const (
		startMarker = "[diff_block_start]"
		endMarker   = "[diff_block_end]"
	)

	start := strings.Index(content, startMarker)
	if start == -1 {
		return 0, false
	}

	diff := content[start+len(startMarker):]
	if end := strings.Index(diff, endMarker); end != -1 {
		diff = diff[:end]
	}

	var (
		added   int
		removed int
	)

	for _, line := range strings.Split(diff, "\n") {
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++ "):
			added++
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "--- "):
			removed++
		}
	}

	return added - removed, true
}

func antigravityProjectPath(transcript antigravityTranscript) string {
	appRoot := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(transcript.path)))))

	contents, err := os.ReadFile(filepath.Clean(filepath.Join(appRoot, "cache", "last_conversations.json")))
	if err == nil {
		var conversations map[string]string
		if json.Unmarshal(contents, &conversations) == nil {
			for workspace, sessionID := range conversations {
				if sessionID == transcript.sessionID {
					return antigravityFilePath(workspace)
				}
			}
		}
	}

	// The CLI records the workspace alongside the conversation ID when a
	// conversation is resumed or exited.
	historyPath := filepath.Join(appRoot, "history.jsonl")
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(historyPath))
	if err == nil {
		defer fh.Close() // nolint:errcheck,gosec

		scanner := bufio.NewScanner(fh)
		scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

		for scanner.Scan() {
			var line antigravityHistoryLine
			if json.Unmarshal(scanner.Bytes(), &line) == nil &&
				line.ConversationID == transcript.sessionID && line.Workspace != "" {
				return antigravityFilePath(line.Workspace)
			}
		}
	}

	return ""
}

func antigravityFilePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(path), "file://") {
		stripped := path[len("file://"):]
		if len(stripped) >= 2 && stripped[1] == ':' {
			return filepath.FromSlash(stripped)
		}

		uri, err := url.Parse(path)
		if err == nil && uri.Path != "" {
			path = uri.Path
		} else {
			path = stripped
		}
	}

	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}

	return filepath.FromSlash(path)
}
