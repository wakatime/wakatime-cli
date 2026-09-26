package ai

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// GrokBot reads the local message mirror of the Grok Bot desktop app.
type GrokBot ParserConfig

// Name returns the desktop provider name (distinct from Grok Build).
func (GrokBot) Name() string { return "Grok Bot" }

// Parse records measured activity and human prompt length. The mirror has no
// model or token usage; never turn text-length estimates into measured tokens.
func (g GrokBot) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	root := filepath.Join(home, ".config", "Grok Bot")

	switch runtime.GOOS {
	case "darwin":
		root = filepath.Join(home, "Library", "Application Support", "Grok Bot")
	case "windows":
		root = filepath.Join(envOrDefault("APPDATA", filepath.Join(home, "AppData", "Roaming")), "Grok Bot")
	}

	return g.parseMirror(ctx, filepath.Join(root, "sand-client-persistence"))
}

func (g GrokBot) parseMirror(ctx context.Context, root string) (Heartbeats, error) {
	files, err := filepath.Glob(filepath.Join(root, "*.blob"))
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)

	var result Heartbeats

	for _, path := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		name := strings.ToUpper(strings.TrimSuffix(filepath.Base(path), ".blob"))

		key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(name)
		if err != nil {
			continue
		}

		account, ok := strings.CutPrefix(string(key), "sand.client.slice.account.")
		if !ok {
			continue
		}

		account, agent, ok := strings.Cut(account, ".transcript.replicas.")
		if !ok || agent == "" {
			continue
		}

		info, err := os.Stat(path)
		if err != nil || info.Size() > maxTranscriptLineSize || !timestampAtOrAfterCutoff(info.ModTime(), g.After) {
			continue
		}

		raw, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, err
		}

		var blob struct {
			Value struct {
				Entries []map[string]any `json:"entries"`
			} `json:"value"`
		}
		if json.Unmarshal(raw, &blob) != nil {
			continue
		}

		for _, entry := range blob.Value.Entries {
			kind := genericAIString(entry["kind"])
			if kind != "message" && kind != "send-message" {
				continue
			}

			timestamp := genericAITime(entry["timestampMs"])
			if timestamp.IsZero() {
				continue
			}

			id := firstNonEmptyString(genericAIString(entry["id"]),
				genericAIString(entry["requestId"])+":"+kind+":"+timestamp.String())

			identity := account + ":" + agent + ":" + id
			if seen[identity] {
				continue
			}

			seen[identity] = true

			event := genericAIEvent{sessionID: "grokbot:" + account + ":" + agent, timestamp: timestamp}
			if !timestampAtOrAfterCutoff(timestamp, ParserConfig(g).sessionAfter(event.sessionID)) {
				continue
			}

			if kind == "message" && genericAIString(entry["role"]) == "user" && entry["fromAgent"] == nil {
				event.prompt = genericAIString(entry["content"])
				if text, ok := entry["content"].(map[string]any); ok {
					event.prompt = genericAIString(text["content"])
				}
			}

			result = append(result, genericAIEventHeartbeats(g, ParserConfig(g), event)...)
		}
	}

	return result, nil
}
