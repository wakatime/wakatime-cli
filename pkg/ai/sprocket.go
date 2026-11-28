package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

// Sprocket contains parameters for detecting heartbeats from Sprocket transcripts.
type Sprocket ParserConfig

type sprocketRun struct {
	WorkspacePath string `json:"workspacePath"`
	Version       string `json:"version"`
	model         string
	calls         map[string]sprocketCall
}

type sprocketCall struct {
	sprocketItem
	cwd    string
	model  string
	status string
}

type sprocketPart struct {
	Kind      string `json:"kind"`
	RunID     string `json:"runId"`
	CreatedAt int64  `json:"createdAt"`
	Prompt    *struct {
		Text   string `json:"text"`
		Hidden string `json:"hidden"`
	} `json:"prompt"`
	Completion *struct {
		Items []sprocketItem `json:"items"`
	} `json:"completion"`
	Tool *sprocketTool `json:"tool"`
}

type sprocketItem struct {
	Type             string          `json:"type"`
	Name             string          `json:"name"`
	CallID           string          `json:"callId"`
	ToolInvocationID string          `json:"toolInvocationId"`
	Input            json.RawMessage `json:"input"`
	StartedAt        int64           `json:"startedAt"`
	CompletedAt      int64           `json:"completedAt"`
	ProviderMetadata struct {
		Origin struct {
			Model string `json:"model"`
		} `json:"origin"`
	} `json:"providerMetadata"`
}

type sprocketTool struct {
	ToolInvocationID string          `json:"toolInvocationId"`
	CallID           string          `json:"callId"`
	Name             string          `json:"name"`
	Status           string          `json:"status"`
	Input            json.RawMessage `json:"input"`
	Output           json.RawMessage `json:"output"`
}

// Parse reads locally opted-in runs. Downloaded transcripts have no run sidecar.
func (g Sprocket) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := envOrDefault("SPROCKET_DATA_DIR", filepath.Join(home, ".sprocket"))

	paths, err := filepath.Glob(filepath.Join(root, "transcripts", "*", "*", "wakatime"))
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, path := range paths {
		parsed, err := g.parseThread(ctx, filepath.Dir(path))
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Sprocket) parseThread(ctx context.Context, thread string) (Heartbeats, error) {
	paths, err := filepath.Glob(filepath.Join(thread, "wakatime", "*.json"))
	if err != nil {
		return nil, err
	}

	runs := make(map[string]*sprocketRun)

	for _, path := range paths {
		var run sprocketRun

		raw, err := os.ReadFile(filepath.Clean(path))
		if err != nil || json.Unmarshal(raw, &run) != nil || !filepath.IsAbs(run.WorkspacePath) {
			continue
		}

		run.calls = make(map[string]sprocketCall)
		runs[strings.TrimSuffix(filepath.Base(path), ".json")] = &run
	}

	if len(runs) == 0 {
		return nil, nil
	}

	paths, err = filepath.Glob(filepath.Join(thread, "parts", "*.jsonl"))
	if err != nil {
		return nil, err
	}

	var heartbeats Heartbeats

	for _, path := range paths {
		parsed, err := g.parseChunk(ctx, path, filepath.Base(thread), runs)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Sprocket) parseChunk(
	ctx context.Context, path, session string, runs map[string]*sprocketRun,
) (Heartbeats, error) {
	file, err := os.Open(filepath.Clean(path))
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer file.Close() // nolint:errcheck

	var heartbeats Heartbeats

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var part sprocketPart
		if json.Unmarshal(scanner.Bytes(), &part) != nil {
			continue
		}

		run := runs[part.RunID]
		if run == nil {
			continue
		}

		heartbeats = append(heartbeats, g.partHeartbeats(part, session, run)...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading Sprocket chunk %q: %w", path, err)
	}

	return heartbeats, nil
}

func (g Sprocket) partHeartbeats(part sprocketPart, session string, run *sprocketRun) Heartbeats {
	event := genericAIEvent{sessionID: session, cwd: run.WorkspacePath, model: run.model}
	if part.CreatedAt > 0 {
		event.timestamp = time.UnixMilli(part.CreatedAt)
	}

	switch {
	case part.Kind == "prompt" && part.Prompt != nil && part.Prompt.Hidden == "":
		event.prompt = part.Prompt.Text
		if promptLength(event.prompt) > 0 {
			return g.eventHeartbeats(event, run.Version)
		}
	case part.Kind == "completion" && part.Completion != nil:
		return g.completionHeartbeats(part, event, run)
	case part.Kind == "tool" && part.Tool != nil:
		return g.toolHeartbeats(*part.Tool, event, run)
	}

	return nil
}

func (g Sprocket) completionHeartbeats(part sprocketPart, event genericAIEvent, run *sprocketRun) Heartbeats {
	var heartbeats Heartbeats

	for _, item := range part.Completion.Items {
		if model := item.ProviderMetadata.Origin.Model; model != "" {
			run.model = model
		}

		if item.Type == "tool-call" {
			key := sprocketCallKey(item.ToolInvocationID, item.CallID)
			run.calls[key] = sprocketCall{
				sprocketItem: item, cwd: run.WorkspacePath, model: run.model, status: run.calls[key].status,
			}
		}

		if item.Type != "text" && item.Type != "reasoning" && item.Type != "tool-call" {
			continue
		}

		itemEvent := event
		itemEvent.model = run.model

		stamp := item.CompletedAt
		if stamp <= 0 {
			stamp = item.StartedAt
		}

		if stamp > 0 {
			itemEvent.timestamp = time.UnixMilli(stamp)
		}

		heartbeats = append(heartbeats, g.eventHeartbeats(itemEvent, run.Version)...)
	}

	return heartbeats
}

func sprocketCallKey(invocation, call string) string {
	if invocation != "" {
		return "invocation:" + invocation
	}

	return "call:" + call
}

func (g Sprocket) toolHeartbeats(tool sprocketTool, event genericAIEvent, run *sprocketRun) Heartbeats {
	key := sprocketCallKey(tool.ToolInvocationID, tool.CallID)

	call := run.calls[key]
	if call.status == "completed" || call.status == "failed" || call.status == "cancelled" {
		return nil
	}

	if call.cwd != "" {
		event.cwd = call.cwd
		event.model = call.model
	}

	if tool.Name == "" {
		tool.Name = call.Name
	}

	if len(tool.Input) == 0 || string(tool.Input) == "null" {
		tool.Input = call.Input
	}

	if tool.Status == "started" {
		run.calls[key] = sprocketCall{
			sprocketItem: sprocketItem{Name: tool.Name, Input: tool.Input},
			cwd:          event.cwd, model: event.model, status: tool.Status,
		}

		return nil
	}

	run.calls[key] = sprocketCall{status: tool.Status}
	if tool.Status != "completed" {
		return nil
	}

	var output struct {
		CWD     string `json:"cwd"`
		Success *bool  `json:"success"`
		Changes []struct {
			Path      string `json:"path"`
			Operation string `json:"operation"`
		} `json:"changes"`
	}

	_ = json.Unmarshal(tool.Output, &output)

	heartbeats := g.eventHeartbeats(event, run.Version)
	if output.Success != nil && !*output.Success {
		return heartbeats
	}

	if tool.Name == "set_cwd" && filepath.IsAbs(output.CWD) {
		run.WorkspacePath = output.CWD
		return heartbeats
	}

	var input struct {
		Path  string `json:"path"`
		Patch string `json:"patch"`
	}
	if json.Unmarshal(tool.Input, &input) != nil {
		return heartbeats
	}

	if tool.Name == "parse_file" && input.Path != "" {
		event.filePath = input.Path
		if !filepath.IsAbs(event.filePath) {
			event.filePath = filepath.Join(event.cwd, event.filePath)
		}

		return g.eventHeartbeats(event, run.Version)
	}

	if tool.Name != "apply_patch" {
		return heartbeats
	}

	counts := sprocketPatchCounts(input.Patch)

	for i, change := range output.Changes {
		if !filepath.IsAbs(change.Path) {
			continue
		}

		event.filePath = change.Path
		event.isWrite = true
		event.lineChanges = nil

		switch change.Operation {
		case "created", "updated", "renamed":
			if i < len(counts) {
				event.lineChanges = heartbeat.PointerTo(counts[i])
			}
		}

		file := g.eventHeartbeats(event, run.Version)
		if len(file) > 1 {
			file[1].IsUnsavedEntity = change.Operation == "deleted"
			heartbeats = append(heartbeats, file[1])
		}
	}

	return heartbeats
}

func sprocketPatchCounts(patch string) []int {
	var counts []int

	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "*** Add File: ") || strings.HasPrefix(line, "*** Update File: ") ||
			strings.HasPrefix(line, "*** Delete File: ") || strings.HasPrefix(line, "*** Copy File: ") {
			counts = append(counts, 0)
			continue
		}

		if len(counts) == 0 {
			continue
		}

		if strings.HasPrefix(line, "+") {
			counts[len(counts)-1]++
		} else if strings.HasPrefix(line, "-") {
			counts[len(counts)-1]--
		}
	}

	return counts
}

func (g Sprocket) eventHeartbeats(event genericAIEvent, version string) Heartbeats {
	if event.timestamp.IsZero() || !event.timestamp.After(ParserConfig(g).sessionAfter(event.sessionID)) {
		return nil
	}

	heartbeats := genericAIEventHeartbeats(g, ParserConfig(g), event)
	for i := range heartbeats {
		heartbeats[i].UserAgent = aiUserAgentWithModelAndEditor(
			heartbeats[i].Entity, g.UserAgents, g.FallbackUserAgent, event.model, "", aiPlugin(g, version),
		)
		heartbeats[i].ProjectPathOverride = event.cwd
	}

	return heartbeats
}

// Name returns the Sprocket parser name.
func (Sprocket) Name() string { return "Sprocket" }
