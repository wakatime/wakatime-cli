package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Codex contains params for detecting heartbeats from Codex session transcripts.
type Codex struct {
	After time.Time
}

type (
	codexSessionMeta struct {
		Type    string `json:"type"`
		Payload struct {
			Cwd     string `json:"cwd"`
			Version string `json:"cli_version"`
		} `json:"payload"`
	}

	codexPayload struct {
		Type   string  `json:"type"`
		Name   *string `json:"name"`
		Input  *string `json:"input"`
		Role   *string `json:"role"`
		Status *string `json:"status"`
	}

	codexLogLine struct {
		Timestamp time.Time     `json:"timestamp"`
		Type      string        `json:"type"`
		Payload   *codexPayload `json:"payload"`
	}
)

// Parse parses the Codex JSONL session transcript logs for ai heartbeats.
func (g Codex) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	transcripts, err := g.transcriptPaths(ctx)
	if err != nil {
		return nil, err
	}

	if len(transcripts) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d transcript logs modified after %s for %s", len(transcripts), g.After, g.ID())

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		parsed, err := g.parseTranscript(ctx, transcript)
		if err != nil {
			return nil, err
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Codex) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	now := time.Now()
	sessionsDir := filepath.Join(home, ".codex", "sessions", now.Format("2006"), now.Format("01"), now.Format("02"))

	sessions, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read .codex sessions directory: %s", err)
	}

	var transcripts []string

	for _, file := range sessions {
		if file.IsDir() || filepath.Ext(file.Name()) != ".jsonl" {
			continue
		}

		info, err := file.Info()
		if err != nil || info.ModTime().Before(g.After) {
			continue
		}

		transcripts = append(transcripts, filepath.Join(sessionsDir, file.Name()))
	}

	return transcripts, nil
}

func (g Codex) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open codex transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
	}

	cwd := ""
	version := ""

	if len(firstLine) > 0 {
		var sessionMeta codexSessionMeta
		if err := json.Unmarshal(firstLine, &sessionMeta); err != nil {
			logger.Debugf("failed parsing codex session metadata from %q: %s", transcript, err)
		} else {
			cwd = sessionMeta.Payload.Cwd
			version = sessionMeta.Payload.Version
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to rewind codex transcript %q: %s", transcript, err)
	}

	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat codex transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek codex transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read codex transcript %q: %s", transcript, err)
		}
	}

	var heartbeats Heartbeats

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var logLine codexLogLine
		if err := json.Unmarshal(line, &logLine); err != nil {
			logger.Warnf("failed parsing codex transcript line from %q: %s", transcript, err)
			logger.Debugf("failed parsing codex transcript line: %s", line)

			continue
		}

		if logLine.Timestamp.IsZero() {
			continue
		}

		if logLine.Timestamp.Before(g.After) {
			continue
		}

		if logLine.Payload == nil {
			continue
		}

		entities := getCodexEntities(ctx, logLine.Timestamp, version, cwd, *logLine.Payload)
		if len(entities) == 0 {
			continue
		}

		heartbeats = append(heartbeats, entities...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading codex transcript %q: %s", transcript, err)
	}

	return heartbeats, nil
}

func getCodexEntities(
	ctx context.Context,
	timestamp time.Time,
	version string,
	cwd string,
	payload codexPayload,
) Heartbeats {
	if payload.Name == nil || *payload.Name != "apply_patch" {
		return nil
	}

	var heartbeats Heartbeats

	var (
		currentFile string
		additions   int
		deletions   int
	)

	lines := strings.Split(*payload.Input, "\n")
	for i := range lines {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "*** ") {
			if currentFile != "" {
				heartbeats = append(heartbeats, codexHeartbeat(
					ctx,
					currentFile,
					timestamp,
					version,
					additions,
					deletions,
				))
			}

			currentFile = codexFilePath(cwd, line)
			additions = 0
			deletions = 0
		} else if currentFile != "" {
			if strings.HasPrefix(line, "+") {
				additions++
			} else if strings.HasPrefix(line, "-") {
				deletions++
			}
		}
	}

	if currentFile != "" {
		heartbeats = append(heartbeats, codexHeartbeat(
			ctx,
			currentFile,
			timestamp,
			version,
			additions,
			deletions,
		))
	}

	return heartbeats
}

func codexFilePath(cwd string, line string) string {
	prefixes := []string{
		"*** Update File: ",
		"*** Add File: ",
		"*** Delete File: ",
	}
	for _, prefix := range prefixes {
		if file, ok := strings.CutPrefix(line, prefix); ok {
			if file == "" || filepath.IsAbs(file) || strings.HasPrefix(file, "/") {
				return file
			}

			return filepath.Join(cwd, file)
		}
	}

	return ""
}

func codexHeartbeat(
	ctx context.Context,
	currentFile string,
	timestamp time.Time,
	version string,
	additions int,
	deletions int,
) heartbeat.Heartbeat {
	return heartbeat.New(
		heartbeat.PointerTo(additions-deletions),
		"",
		heartbeat.AICodingCategory.String(),
		nil,
		currentFile,
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
		heartbeat.UserAgent(ctx, codexPlugin(version)),
	)
}

func codexPlugin(version string) string {
	if version == "" {
		return "Codex"
	}

	return "Codex/" + version
}

// ID returns its id.
func (Codex) ID() ParserID {
	return CodexParser
}
