package ai

import (
	"context"
	"encoding/binary"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
)

type wslHomesKey struct{}

func wslHomes(ctx context.Context) []string {
	if homes, ok := ctx.Value(wslHomesKey{}).([]string); ok {
		return homes
	}

	return discoverWSLHomes(ctx)
}

func parseWSLDistros(raw []byte) []string {
	text := string(raw)
	if strings.ContainsRune(text, 0) {
		units := make([]uint16, 0, len(raw)/2)
		for i := 0; i+1 < len(raw); i += 2 {
			units = append(units, binary.LittleEndian.Uint16(raw[i:]))
		}

		text = string(utf16.Decode(units))
	}

	var result []string

	for _, line := range strings.Split(text, "\n") {
		name := strings.TrimSpace(strings.TrimPrefix(line, "\ufeff"))
		if name == "" || strings.HasSuffix(name, ".") || strings.ContainsAny(name, `/\:<>"|?*`+"\x00") ||
			strings.HasPrefix(strings.ToLower(name), "docker-desktop") {
			continue
		}

		valid := true

		for _, r := range name {
			if r < 32 {
				valid = false
			}
		}

		if valid {
			result = append(result, name)
		}
	}

	return result
}

// Entity paths recorded inside WSL are Linux paths, not paths on Windows C:.
func translateWSLHeartbeats(transcript string, heartbeats Heartbeats) Heartbeats {
	normalized := strings.ReplaceAll(transcript, `\`, "/")

	lower := strings.ToLower(normalized)
	if !strings.HasPrefix(lower, "//wsl$/") && !strings.HasPrefix(lower, "//wsl.localhost/") {
		return heartbeats
	}

	parts := strings.SplitN(strings.TrimPrefix(normalized, "//"), "/", 3)
	if len(parts) < 3 {
		return heartbeats
	}

	root := `\\` + parts[0] + `\` + parts[1]
	translate := func(path string) string {
		if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
			return path
		}

		return filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(path, "/")))
	}

	for i := range heartbeats {
		h := &heartbeats[i]
		if h.EntityType == heartbeat.FileType {
			h.Entity = translate(h.Entity)
		}

		h.ProjectPath = translate(h.ProjectPath)
		h.ProjectPathOverride = translate(h.ProjectPathOverride)
	}

	return heartbeats
}
