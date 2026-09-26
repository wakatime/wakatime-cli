package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (g Kiro) parseModernSessions(ctx context.Context, home string) (Heartbeats, error) {
	root := filepath.Join(envOrDefault("KIRO_HOME", filepath.Join(home, ".kiro")), "sessions")
	provider := genericAIProvider{parser: g, config: ParserConfig(g), roots: []string{root}, extensions: stringSet(".jsonl")}

	paths, err := provider.transcriptPaths()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)

	var result Heartbeats

	for _, path := range paths {
		cli := filepath.Base(filepath.Dir(path)) == "cli"
		if !cli && filepath.Base(path) != "messages.jsonl" {
			continue
		}

		metaPath := filepath.Join(filepath.Dir(path), "session.json")
		if cli {
			metaPath = strings.TrimSuffix(path, ".jsonl") + ".json"
		}

		raw, err := os.ReadFile(filepath.Clean(metaPath))
		if err != nil {
			continue
		}

		var meta map[string]any
		if json.Unmarshal(raw, &meta) != nil {
			continue
		}

		session := firstNonEmptyString(genericAIString(meta["session_id"]), genericAIString(meta["id"]),
			filepath.Base(filepath.Dir(path)))

		cwd := genericAIString(meta["cwd"])
		if paths, ok := meta["workspacePaths"].([]any); ok && len(paths) > 0 {
			cwd = genericAIString(paths[0])
		}

		if filepath.Clean(cwd) == filepath.Clean(home) {
			cwd = ""
		}

		model := genericAIString(genericAIFind(meta, "modelId", "model_id"))

		values, err := genericAIJSONLines(ctx, g.Name(), path)
		if err != nil {
			return nil, err
		}

		if cli {
			result = append(result, g.kiroCLIEntries(values, session, cwd, model, meta, seen)...)
			continue
		}

		for _, value := range values {
			object, ok := value.(map[string]any)
			if !ok {
				continue
			}

			payload, ok := object["payload"].(map[string]any)
			if !ok {
				continue
			}

			kind := genericAIString(payload["type"])
			if kind != "user" && kind != "assistant" && kind != "tool_result" {
				continue
			}

			timestamp := genericAITime(object["timestamp"])
			if timestamp.IsZero() || !timestampAtOrAfterCutoff(timestamp, ParserConfig(g).sessionAfter(session)) {
				continue
			}

			event := genericAIEvent{sessionID: session, cwd: cwd, model: model, timestamp: timestamp}
			if kind == "user" {
				event.prompt = kiroModernText(payload["content"])
			}

			if kind == "tool_result" {
				// Only explicit successful file metadata can establish a completed edit.
				if payload["success"] != true {
					continue
				}

				tool := genericAIEventFromValue(payload)
				event.filePath, event.toolName, event.isWrite, event.lineChanges = tool.filePath, tool.toolName,
					tool.isWrite, tool.lineChanges
			}

			identity := session + ":" + firstNonEmptyString(genericAIString(object["id"]),
				genericAIString(payload["id"]), timestamp.String()+":"+kind+":"+event.prompt)
			if seen[identity] {
				continue
			}

			seen[identity] = true

			result = append(result, genericAIEventHeartbeats(g, ParserConfig(g), event)...)
		}
	}

	return result, nil
}

func kiroModernText(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case []any:
		var text strings.Builder

		for _, part := range value {
			if object, ok := part.(map[string]any); ok {
				if genericAIString(object["kind"]) == "text" {
					text.WriteString(genericAIString(object["data"]))
				}

				if genericAIString(object["type"]) == "text" {
					text.WriteString(genericAIString(object["text"]))
				}
			}
		}

		return text.String()
	}

	return ""
}

func (g Kiro) kiroCLIEntries(values []any, session, cwd, model string, meta map[string]any,
	seen map[string]bool) Heartbeats {
	var result Heartbeats

	turn := -1
	prompt := ""

	var timestamp time.Time

	turns, _ := genericAIFind(meta, "user_turn_metadatas").([]any)
	flush := func() {
		if turn < 0 {
			return
		}

		stamp := timestamp

		if turn < len(turns) {
			if object, ok := turns[turn].(map[string]any); ok {
				if end := genericAITime(object["end_timestamp"]); !end.IsZero() {
					stamp = end
				}
			}
		}

		if stamp.IsZero() {
			stamp = genericAITime(meta["created_at"])
		}

		key := fmt.Sprintf("kiro-cli:%s:%d", session, turn)
		if stamp.IsZero() || !timestampAtOrAfterCutoff(stamp, ParserConfig(g).sessionAfter(session)) || seen[key] {
			return
		}

		seen[key] = true

		result = append(result, genericAIEventHeartbeats(g, ParserConfig(g), genericAIEvent{sessionID: session,
			cwd: cwd, model: model, timestamp: stamp, prompt: prompt})...)
	}

	for _, value := range values {
		object, ok := value.(map[string]any)
		if !ok {
			continue
		}

		if genericAIString(object["kind"]) != "Prompt" {
			continue
		}

		flush()

		turn++
		data, _ := object["data"].(map[string]any)
		prompt = kiroModernText(data["content"])
		details, _ := data["meta"].(map[string]any)
		timestamp = genericAITime(details["timestamp"])
	}

	flush()

	return result
}
