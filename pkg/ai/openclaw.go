package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// OpenClaw contains parameters for detecting heartbeats from OpenClaw logs.
type OpenClaw ParserConfig

// Parse parses OpenClaw logs for AI heartbeats.
func (g OpenClaw) Parse(ctx context.Context) (Heartbeats, error) {
	home, err := aiUserHome(ctx)
	if err != nil {
		return nil, err
	}

	paths := genericAIPaths{
		roots: []string{
			filepath.Join(home, ".openclaw", "agents"),
			filepath.Join(home, ".clawdbot", "agents"),
			filepath.Join(home, ".moldbot", "agents"),
		},
		extensions: []string{".jsonl"},
	}

	provider := genericAIProvider{parser: g, config: ParserConfig(g), roots: paths.roots, extensions: stringSet(".jsonl")}

	transcripts, err := provider.transcriptPaths()
	if err != nil {
		return nil, err
	}

	var result Heartbeats

	for _, path := range transcripts {
		values, err := genericAIJSONLines(ctx, g.Name(), path)
		if err != nil {
			return nil, err
		}

		clock, err := newTranscriptClock(ctx, path)
		if err != nil {
			return nil, err
		}

		for _, value := range values {
			if object, ok := value.(map[string]any); ok {
				raw, _ := json.Marshal(object)
				object["timestamp"] = clock.resolve(fmt.Sprintf("%x", sha256.Sum256(raw)),
					genericAITime(object["timestamp"])).Format(time.RFC3339Nano)
			}
		}

		parsed, err := genericAIHeartbeats(ctx, provider, path, values)
		if err != nil {
			return nil, err
		}

		if err := clock.save(); err != nil {
			return nil, err
		}

		result = append(result, parsed...)
	}

	return result, nil
}

// Name returns the OpenClaw parser name.
func (OpenClaw) Name() string { return "OpenClaw" }
