//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	// Register the pure-Go SQLite driver used to read VS Code state.vscdb files.
	_ "modernc.org/sqlite"
)

// Cody contains params for detecting heartbeats from Cody by Sourcegraph transcripts.
type Cody ParserConfig

type (
	codyStorage struct {
		ChatHistory json.RawMessage `json:"cody-local-chatHistory-v2"`
	}

	codyUserLocalHistory struct {
		Chat codyChatHistory `json:"chat"`
	}

	codyChatHistory map[string]codyTranscript

	codyTranscript struct {
		ID                       string            `json:"id"`
		LastInteractionTimestamp time.Time         `json:"lastInteractionTimestamp"`
		Interactions             []codyInteraction `json:"interactions"`
	}

	codyInteraction struct {
		HumanMessage     codyMessage  `json:"humanMessage"`
		AssistantMessage *codyMessage `json:"assistantMessage"`
	}

	codyMessage struct {
		Text         string            `json:"text"`
		Model        string            `json:"model"`
		TokenUsage   codyTokenUsage    `json:"tokenUsage"`
		ContextFiles []codyContextItem `json:"contextFiles"`
		Processes    []codyProcessStep `json:"processes"`
		SubMessages  []codySubMessage  `json:"subMessages"`
	}

	codyTokenUsage struct {
		CompletionTokens int64 `json:"completionTokens"`
		PromptTokens     int64 `json:"promptTokens"`
	}

	codyProcessStep struct {
		Items []codyContextItem `json:"items"`
	}

	codySubMessage struct {
		ContextFiles []codyContextItem `json:"contextFiles"`
	}

	codyContextItem struct {
		Type     string          `json:"type"`
		ToolName string          `json:"toolName"`
		URI      codyURI         `json:"uri"`
		Metadata json.RawMessage `json:"metadata"`
	}

	codyURI struct {
		Scheme string `json:"scheme"`
		Path   string `json:"path"`
		FSPath string `json:"fsPath"`
	}
)

const codyStorageKey = "sourcegraph.cody-ai"

// Parse parses Cody chat history from VS Code global storage.
func (g Cody) Parse(ctx context.Context) (Heartbeats, error) {
	dbPaths, err := g.stateDBPaths(ctx)
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, dbPath := range dbPaths {
		if !g.stateDBModifiedAfter(dbPath, g.After) {
			continue
		}

		storageValue, err := g.queryStorage(ctx, dbPath)
		if err != nil {
			return nil, err
		}

		parsed := g.heartbeatsFromStorage(storageValue)
		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (Cody) stateDBPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	candidates := []string{
		filepath.Join(home, "Library", "Application Support", "Code", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "Library", "Application Support", "Code - Insiders", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "Library", "Application Support", "VSCodium", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Code", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "Code - Insiders", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, "AppData", "Roaming", "VSCodium", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Code", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "Code - Insiders", "User", "globalStorage", "state.vscdb"),
		filepath.Join(home, ".config", "VSCodium", "User", "globalStorage", "state.vscdb"),
	}

	var paths []string

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			paths = append(paths, candidate)
		}
	}

	return paths, nil
}

func (Cody) stateDBModifiedAfter(dbPath string, after time.Time) bool {
	if after.IsZero() {
		return true
	}

	info, err := os.Stat(dbPath)
	if err != nil {
		return false
	}

	return timestampAtOrAfterCutoff(info.ModTime(), after)
}

func (Cody) queryStorage(ctx context.Context, dbPath string) (string, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return "", fmt.Errorf("failed opening cody sqlite db %q: %s", dbPath, err)
	}
	defer db.Close() // nolint:errcheck

	var value string

	err = db.QueryRowContext(ctx, `
SELECT CAST(value AS TEXT)
FROM ItemTable
WHERE key = ?
LIMIT 1;
`, codyStorageKey).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", fmt.Errorf("failed querying cody sqlite db %q: %s", dbPath, err)
	}

	return strings.TrimSpace(value), nil
}

func (g Cody) heartbeatsFromStorage(value string) Heartbeats {
	if value == "" {
		return nil
	}

	history := g.chatHistory(value)
	if len(history) == 0 {
		return nil
	}

	var heartbeats Heartbeats

	for _, userHistory := range history {
		for chatID, transcript := range userHistory.Chat {
			transcriptID := firstNonEmptyString(transcript.ID, chatID)
			if transcript.LastInteractionTimestamp.IsZero() ||
				!timestampAtOrAfterCutoff(transcript.LastInteractionTimestamp, g.After) {
				continue
			}

			interaction, ok := codyLastInteraction(transcript.Interactions)
			if !ok {
				continue
			}

			heartbeats = append(heartbeats, g.interactionHeartbeats(
				transcript.LastInteractionTimestamp,
				transcriptID,
				interaction,
			)...)
		}
	}

	return heartbeats
}

func (Cody) chatHistory(value string) map[string]codyUserLocalHistory {
	var storage codyStorage
	if err := json.Unmarshal([]byte(value), &storage); err != nil || len(storage.ChatHistory) == 0 {
		return nil
	}

	var encoded string
	if err := json.Unmarshal(storage.ChatHistory, &encoded); err == nil {
		storage.ChatHistory = json.RawMessage(encoded)
	}

	var history map[string]codyUserLocalHistory
	if err := json.Unmarshal(storage.ChatHistory, &history); err != nil {
		return nil
	}

	return history
}

func codyLastInteraction(interactions []codyInteraction) (codyInteraction, bool) {
	for i := len(interactions) - 1; i >= 0; i-- {
		if strings.TrimSpace(interactions[i].HumanMessage.Text) != "" || interactions[i].AssistantMessage != nil {
			return interactions[i], true
		}
	}

	return codyInteraction{}, false
}

func (g Cody) interactionHeartbeats(
	timestamp time.Time,
	sessionID string,
	interaction codyInteraction,
) Heartbeats {
	var heartbeats Heartbeats

	model := ""

	var tokens *heartbeat.AITokens

	if interaction.AssistantMessage != nil {
		model = interaction.AssistantMessage.Model
		tokens = g.tokens(interaction.AssistantMessage.TokenUsage)
	}

	entity := appHeartbeatEntity(g.Name(), sessionID)
	appHeartbeat := g.newHeartbeat(
		nil,
		sessionID,
		tokens,
		entity,
		heartbeat.AppType,
		heartbeat.PointerTo(false),
		"",
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(entity, g.UserAgents, g.FallbackUserAgent, model, ""),
	)
	appHeartbeat.AIPromptLength = promptLength(interaction.HumanMessage.Text)
	heartbeats = append(heartbeats, appHeartbeat)

	for _, item := range codyContextItems(interaction) {
		if heartbeat := g.fileHeartbeat(timestamp, sessionID, model, item); heartbeat != nil {
			heartbeats = append(heartbeats, *heartbeat)
		}
	}

	return heartbeats
}

func (Cody) tokens(usage codyTokenUsage) *heartbeat.AITokens {
	if usage.PromptTokens == 0 && usage.CompletionTokens == 0 {
		return nil
	}

	return &heartbeat.AITokens{
		CurrentInput:  usage.PromptTokens,
		CurrentOutput: usage.CompletionTokens,
	}
}

func codyContextItems(interaction codyInteraction) []codyContextItem {
	var items []codyContextItem

	items = append(items, interaction.HumanMessage.ContextFiles...)

	if interaction.AssistantMessage == nil {
		return items
	}

	items = append(items, interaction.AssistantMessage.ContextFiles...)

	for _, process := range interaction.AssistantMessage.Processes {
		items = append(items, process.Items...)
	}

	for _, subMessage := range interaction.AssistantMessage.SubMessages {
		items = append(items, subMessage.ContextFiles...)
	}

	return items
}

func (g Cody) fileHeartbeat(
	timestamp time.Time,
	sessionID string,
	model string,
	item codyContextItem,
) *heartbeat.Heartbeat {
	filePath := item.filePath()
	if filePath == "" {
		return nil
	}

	isWrite := false
	lineChanges := 0

	if strings.EqualFold(item.ToolName, "text_editor") {
		isWrite = true

		changes, ok := item.lineChanges()
		if !ok {
			return nil
		}

		lineChanges = changes
	}

	h := g.newHeartbeat(
		heartbeat.PointerTo(lineChanges),
		sessionID,
		nil,
		filePath,
		heartbeat.FileType,
		heartbeat.PointerTo(isWrite),
		"",
		heartbeatTimestamp(timestamp),
		aiUserAgentWithModel(filePath, g.UserAgents, g.FallbackUserAgent, model, ""),
	)

	return &h
}

func (item codyContextItem) filePath() string {
	rawPath := ""

	switch {
	case item.URI.FSPath != "":
		rawPath = item.URI.FSPath
	case item.URI.Scheme == "file" && item.URI.Path != "":
		rawPath = item.URI.Path
	case strings.HasPrefix(strings.ToLower(item.URI.Path), "file://"):
		rawPath = item.URI.Path
	}

	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(rawPath), "file://") {
		stripped := rawPath[len("file://"):]

		if len(stripped) >= 2 && stripped[1] == ':' {
			return filepath.FromSlash(stripped)
		}

		u, err := url.Parse(item.URI.Path)
		if err == nil && u.Path != "" {
			rawPath = u.Path
		} else {
			rawPath = stripped
		}
	}

	if len(rawPath) >= 3 && (rawPath[0] == '/' || rawPath[0] == '\\') && rawPath[2] == ':' {
		rawPath = rawPath[1:]
	}

	return filepath.FromSlash(rawPath)
}

func (item codyContextItem) lineChanges() (int, bool) {
	var contentPair []string
	if err := json.Unmarshal(item.Metadata, &contentPair); err != nil || len(contentPair) < 2 {
		return 0, false
	}

	return codyDiffLineChanges(contentPair[0], contentPair[1]), true
}

func codyDiffLineChanges(oldContent string, newContent string) int {
	oldLines := codyContentLines(oldContent)
	newLines := codyContentLines(newContent)

	if len(oldLines) == 0 {
		return len(newLines)
	}

	if len(newLines) == 0 {
		return len(oldLines)
	}

	const maxDiffCells = 4_000_000
	if len(oldLines)*len(newLines) > maxDiffCells {
		diff := len(newLines) - len(oldLines)
		if diff < 0 {
			return -diff
		}

		return diff
	}

	lcs := codyLCSLineCount(oldLines, newLines)

	return len(oldLines) + len(newLines) - 2*lcs
}

func codyContentLines(content string) []string {
	if content == "" {
		return nil
	}

	return strings.Split(strings.TrimSuffix(content, "\n"), "\n")
}

func codyLCSLineCount(oldLines []string, newLines []string) int {
	previous := make([]int, len(newLines)+1)
	current := make([]int, len(newLines)+1)

	for i := range oldLines {
		for j := range newLines {
			if oldLines[i] == newLines[j] {
				current[j+1] = previous[j] + 1
			} else {
				current[j+1] = max(current[j], previous[j+1])
			}
		}

		previous, current = current, previous
		clear(current)
	}

	return previous[len(newLines)]
}

func (Cody) newHeartbeat(
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
func (Cody) Name() string {
	return "Cody"
}
