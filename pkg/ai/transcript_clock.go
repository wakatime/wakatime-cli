package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/ini"
)

// Persist only fallback timestamps. Appending a transcript must not move old
// unstamped records to the new mtime and replay them on every sync.
type transcriptClock struct {
	path               string
	fallback, previous time.Time
	stamps             map[string]time.Time
	dirty              bool
}

func newTranscriptClock(ctx context.Context, path string) (*transcriptClock, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	dir, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return nil, err
	}

	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	key := sha256.Sum256([]byte(absolute))
	clock := &transcriptClock{path: filepath.Join(dir, "ai-timestamps", fmt.Sprintf("%x.json", key)),
		fallback: info.ModTime(), stamps: make(map[string]time.Time)}

	raw, err := os.ReadFile(clock.path)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &clock.stamps); err != nil {
			return nil, err
		}
	}

	return clock, nil
}

func (c *transcriptClock) resolve(key string, timestamp time.Time) time.Time {
	if !timestamp.IsZero() {
		c.previous = timestamp
		return timestamp
	}

	if previous := c.stamps[key]; !previous.IsZero() {
		c.previous = previous
		return previous
	}

	timestamp = c.previous
	if timestamp.IsZero() {
		timestamp = c.fallback
	}

	c.stamps[key] = timestamp
	c.dirty = true

	return timestamp
}

func (c *transcriptClock) normalize(line []byte) []byte {
	var value map[string]any
	if json.Unmarshal(line, &value) != nil {
		return line
	}

	stamp := genericAITime(value["timestamp"])
	key := fmt.Sprintf("%x", sha256.Sum256(line))
	value["timestamp"] = c.resolve(key, stamp).Format(time.RFC3339Nano)

	normalized, err := json.Marshal(value)
	if err != nil {
		return line
	}

	return normalized
}

func (c *transcriptClock) save() error {
	if !c.dirty {
		return nil
	}

	raw, err := json.Marshal(c.stamps)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(c.path), 0o700); err != nil {
		return err
	}

	file, err := os.CreateTemp(filepath.Dir(c.path), ".timestamps-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name()) // nolint:errcheck

	if _, err := file.Write(raw); err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), c.path)
}
