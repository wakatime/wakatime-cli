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
)

// Continue contains params for detecting heartbeats from Continue transcripts.
type Continue ParserConfig

type (
	continueEventKind int

	continueEvent struct {
		Timestamp     time.Time
		Kind          continueEventKind
		SessionID     string
		UserAgent     string
		Model         string
		Provider      string
		Prompt        string
		FilePath      string
		WorkspacePath string
		LineChange    *int
		Tokens        heartbeat.AITokens
	}

	continueChatInteraction struct {
		Timestamp     time.Time `json:"timestamp"`
		UserAgent     string    `json:"userAgent"`
		EventName     string    `json:"eventName"`
		Prompt        string    `json:"prompt"`
		ModelName     string    `json:"modelName"`
		ModelProvider string    `json:"modelProvider"`
		SessionID     string    `json:"sessionId"`
	}

	continueToolUsage struct {
		Timestamp    time.Time       `json:"timestamp"`
		UserAgent    string          `json:"userAgent"`
		EventName    string          `json:"eventName"`
		FunctionName string          `json:"functionName"`
		ToolCallArgs json.RawMessage `json:"toolCallArgs"`
		Accepted     bool            `json:"accepted"`
		Succeeded    bool            `json:"succeeded"`
	}

	continueEditOutcome struct {
		Timestamp     time.Time `json:"timestamp"`
		UserAgent     string    `json:"userAgent"`
		EventName     string    `json:"eventName"`
		Prompt        string    `json:"prompt"`
		ModelName     string    `json:"modelName"`
		ModelProvider string    `json:"modelProvider"`
		Accepted      bool      `json:"accepted"`
		LineChange    *int      `json:"lineChange"`
		FilePath      string    `json:"filepath"`
	}

	continueTokensGenerated struct {
		Timestamp       time.Time `json:"timestamp"`
		EventName       string    `json:"eventName"`
		Model           string    `json:"model"`
		Provider        string    `json:"provider"`
		PromptTokens    int64     `json:"promptTokens"`
		GeneratedTokens int64     `json:"generatedTokens"`
	}

	continueSession struct {
		SessionID          string `json:"sessionId"`
		WorkspaceDirectory string `json:"workspaceDirectory"`
	}
)

const (
	continueEventChat continueEventKind = iota
	continueEventRead
	continueEventEdit
	continueEventTokens
)

// Parse parses Continue local dev_data jsonl logs for ai heartbeats.
func (g Continue) Parse(ctx context.Context) (Heartbeats, error) {
	root, err := g.root(ctx)
	if err != nil {
		return nil, err
	}

	if root == "" {
		return Heartbeats{}, nil
	}

	workspaces := g.sessionWorkspaces(filepath.Join(root, "sessions", "sessions.json"))

	events, err := g.events(filepath.Join(root, "dev_data"), workspaces)
	if err != nil {
		return nil, err
	}

	sort.SliceStable(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})

	return g.heartbeats(events), nil
}

func (Continue) root(ctx context.Context) (string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to find user home dir: %s", err)
	}

	root := filepath.Join(home, ".continue")
	if _, err := os.Stat(root); err == nil {
		return root, nil
	}

	return "", nil
}

func (g Continue) events(devDataDir string, workspaces map[string]string) ([]continueEvent, error) {
	var events []continueEvent

	entries, err := os.ReadDir(devDataDir)
	if os.IsNotExist(err) {
		return events, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed reading continue dev data dir %q: %s", devDataDir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		schemaDir := filepath.Join(devDataDir, entry.Name())
		for _, filename := range []string{
			"tokensGenerated.jsonl",
			"chatInteraction.jsonl",
			"toolUsage.jsonl",
			"editOutcome.jsonl",
		} {
			parsed, err := g.eventsFromFile(filepath.Join(schemaDir, filename), workspaces)
			if err != nil {
				return nil, err
			}

			events = append(events, parsed...)
		}
	}

	return events, nil
}

func (g Continue) eventsFromFile(path string, workspaces map[string]string) ([]continueEvent, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed opening continue log %q: %s", path, err)
	}

	defer fh.Close() // nolint:errcheck

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	var events []continueEvent

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		event, ok := g.eventFromLine([]byte(line), workspaces)
		if ok && !event.Timestamp.IsZero() {
			events = append(events, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading continue log %q: %s", path, err)
	}

	return events, nil
}

func (g Continue) eventFromLine(line []byte, workspaces map[string]string) (continueEvent, bool) {
	var header struct {
		EventName string `json:"eventName"`
	}
	if err := json.Unmarshal(line, &header); err != nil {
		return continueEvent{}, false
	}

	switch header.EventName {
	case "chatInteraction":
		var event continueChatInteraction
		if err := json.Unmarshal(line, &event); err != nil {
			return continueEvent{}, false
		}

		return continueEvent{
			Timestamp:     event.Timestamp,
			Kind:          continueEventChat,
			SessionID:     event.SessionID,
			UserAgent:     event.UserAgent,
			Model:         event.ModelName,
			Provider:      event.ModelProvider,
			Prompt:        event.Prompt,
			WorkspacePath: workspaces[event.SessionID],
		}, true
	case "toolUsage":
		var event continueToolUsage
		if err := json.Unmarshal(line, &event); err != nil {
			return continueEvent{}, false
		}

		if !event.Accepted || !event.Succeeded || event.FunctionName != "read_file" {
			return continueEvent{}, false
		}

		filePath := g.toolFilePath(event.ToolCallArgs)
		if filePath == "" {
			return continueEvent{}, false
		}

		return continueEvent{
			Timestamp: event.Timestamp,
			Kind:      continueEventRead,
			UserAgent: event.UserAgent,
			FilePath:  filePath,
		}, true
	case "editOutcome":
		var event continueEditOutcome
		if err := json.Unmarshal(line, &event); err != nil {
			return continueEvent{}, false
		}

		filePath := g.filePath(event.FilePath)
		if !event.Accepted || filePath == "" {
			return continueEvent{}, false
		}

		return continueEvent{
			Timestamp:  event.Timestamp,
			Kind:       continueEventEdit,
			UserAgent:  event.UserAgent,
			Model:      event.ModelName,
			Provider:   event.ModelProvider,
			Prompt:     event.Prompt,
			FilePath:   filePath,
			LineChange: event.LineChange,
		}, true
	case "tokensGenerated":
		var event continueTokensGenerated
		if err := json.Unmarshal(line, &event); err != nil {
			return continueEvent{}, false
		}

		return continueEvent{
			Timestamp: event.Timestamp,
			Kind:      continueEventTokens,
			Model:     event.Model,
			Provider:  event.Provider,
			Tokens: heartbeat.AITokens{
				CurrentInput:  event.PromptTokens,
				CurrentOutput: event.GeneratedTokens,
			},
		}, true
	default:
		return continueEvent{}, false
	}
}

func (g Continue) heartbeats(events []continueEvent) Heartbeats {
	var heartbeats Heartbeats

	var (
		currentSessionID string
		currentWorkspace string
		currentTokens    *continueEvent
	)

	for _, event := range events {
		switch event.Kind {
		case continueEventTokens:
			copy := event
			currentTokens = &copy
		case continueEventChat:
			if event.SessionID != "" {
				currentSessionID = event.SessionID
			}

			if event.WorkspacePath != "" {
				currentWorkspace = event.WorkspacePath
			}

			if event.Timestamp.Before(g.After) {
				continue
			}

			tokens := g.tokensForEvent(event, currentTokens)
			if tokens != nil {
				currentTokens = nil
			}

			heartbeats = append(heartbeats, g.newHeartbeat(
				nil,
				currentSessionID,
				tokens,
				appHeartbeatEntity(g.Name(), currentSessionID),
				heartbeat.AppType,
				heartbeat.PointerTo(false),
				currentWorkspace,
				float64(event.Timestamp.Unix()),
				g.userAgent(event, appHeartbeatEntity(g.Name(), currentSessionID)),
			))
			heartbeats[len(heartbeats)-1].AIPromptLength = g.promptLength(event.Prompt)
		case continueEventRead:
			if event.Timestamp.Before(g.After) {
				continue
			}

			filePath := g.resolvePath(event.FilePath, currentWorkspace)
			heartbeats = append(heartbeats, g.newHeartbeat(
				heartbeat.PointerTo(0),
				currentSessionID,
				nil,
				filePath,
				heartbeat.FileType,
				heartbeat.PointerTo(false),
				"",
				float64(event.Timestamp.Unix()),
				g.userAgent(event, filePath),
			))
		case continueEventEdit:
			if event.Timestamp.Before(g.After) {
				continue
			}

			lineChange := event.LineChange
			if lineChange == nil {
				lineChange = heartbeat.PointerTo(0)
			}

			filePath := g.resolvePath(event.FilePath, currentWorkspace)
			heartbeats = append(heartbeats, g.newHeartbeat(
				lineChange,
				currentSessionID,
				nil,
				filePath,
				heartbeat.FileType,
				heartbeat.PointerTo(true),
				"",
				float64(event.Timestamp.Unix()),
				g.userAgent(event, filePath),
			))
		}
	}

	return heartbeats
}

func (Continue) tokensForEvent(event continueEvent, tokens *continueEvent) *heartbeat.AITokens {
	if tokens == nil {
		return nil
	}

	if event.Model != "" && tokens.Model != "" && !strings.EqualFold(event.Model, tokens.Model) {
		return nil
	}

	if event.Provider != "" && tokens.Provider != "" && !strings.EqualFold(event.Provider, tokens.Provider) {
		return nil
	}

	diff := event.Timestamp.Sub(tokens.Timestamp)
	if diff < 0 {
		diff = -diff
	}

	if diff > 5*time.Second {
		return nil
	}

	copy := tokens.Tokens

	return &copy
}

func (g Continue) userAgent(event continueEvent, entity string) string {
	parser := aiPlugin(g, event.Model)

	fallback := event.UserAgent
	if fallback == "" {
		fallback = g.FallbackUserAgent
	}

	return aiUserAgent(entity, g.UserAgents, fallback, parser)
}

func (Continue) promptLength(prompt string) int {
	idx := strings.LastIndex(prompt, "<user>")
	if idx != -1 {
		prompt = prompt[idx+len("<user>"):]
	}

	if idx := strings.Index(prompt, "<"); idx != -1 {
		prompt = prompt[:idx]
	}

	return promptLength(strings.TrimSpace(prompt))
}

func (g Continue) toolFilePath(raw json.RawMessage) string {
	var args struct {
		FilePath string `json:"filepath"`
	}

	var encoded string
	if err := json.Unmarshal(raw, &encoded); err == nil {
		_ = json.Unmarshal([]byte(encoded), &args)
	} else {
		_ = json.Unmarshal(raw, &args)
	}

	return g.filePath(args.FilePath)
}

func (Continue) filePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(path), "file://") {
		stripped := path[len("file://"):]

		if len(stripped) >= 2 && stripped[1] == ':' {
			return filepath.FromSlash(stripped)
		}

		u, err := url.Parse(path)
		if err == nil && u.Path != "" {
			path = u.Path
		} else {
			path = stripped
		}
	}

	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}

	return filepath.FromSlash(path)
}

func (Continue) resolvePath(path string, workspace string) string {
	if filepath.IsAbs(path) || workspace == "" {
		return path
	}

	return filepath.Join(workspace, path)
}

func (g Continue) sessionWorkspaces(path string) map[string]string {
	workspaces := make(map[string]string)

	//nolint:gosec
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return workspaces
	}

	var sessions []continueSession
	if err := json.Unmarshal(data, &sessions); err != nil {
		return workspaces
	}

	for _, session := range sessions {
		workspace := g.filePath(session.WorkspaceDirectory)
		if session.SessionID != "" && workspace != "" {
			workspaces[session.SessionID] = workspace
		}
	}

	return workspaces
}

func (Continue) newHeartbeat(
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

// Name returns its name.
func (Continue) Name() string {
	return "Continue"
}
