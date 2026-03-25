package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

const maxClaudeTranscriptLineSize = 10 * 1024 * 1024

// Claude contains params for detecting heartbeats from Claude transcripts.
type Claude struct {
	After time.Time
}

type (
	structuredPatch struct {
		NewLines int `json:"newLines"`
		OldLines int `json:"oldLines"`
	}

	toolUseResultFile struct {
		Content  *string `json:"content"`
		FilePath *string `json:"filePath"`
	}

	toolUseResult struct {
		Type            *string            `json:"type"`
		File            *toolUseResultFile `json:"file"`
		Content         *string            `json:"content"`
		FilePath        *string            `json:"filePath"`
		OriginalFile    *string            `json:"originalFile"`
		StructuredPatch *[]structuredPatch `json:"structuredPatch"`
	}

	logLine struct {
		Timestamp     time.Time      `json:"timestamp"`
		Version       string         `json:"version"`
		ToolUseResult *toolUseResult `json:"toolUseResult"`
	}
)

// Parse parses the Claude JSONL session transcript logs for ai heartbeats.
func (g Claude) Parse(ctx context.Context) (Heartbeats, error) {
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

func (g Claude) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	claudeProjectsDir := filepath.Join(home, ".claude", "projects")

	projects, err := os.ReadDir(claudeProjectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to read .claude projects directory: %s", err)
	}

	var transcripts []string

	for _, project := range projects {
		if !project.IsDir() {
			continue
		}

		files, err := os.ReadDir(filepath.Join(claudeProjectsDir, project.Name()))
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || filepath.Ext(file.Name()) != ".jsonl" {
				continue
			}

			info, err := file.Info()
			if err != nil || info.ModTime().Before(g.After) {
				continue
			}

			transcripts = append(transcripts, filepath.Join(claudeProjectsDir, project.Name(), file.Name()))
		}
	}

	return transcripts, nil
}

func (g Claude) parseTranscript(ctx context.Context, transcript string) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec // Transcript paths are discovered from ~/.claude/projects in transcriptPaths.
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open claude transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxClaudeTranscriptLineSize)

	var heartbeats Heartbeats

	claudeVersion := ""

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var logLine logLine
		if err := json.Unmarshal(line, &logLine); err != nil {
			logger.Warnf("failed parsing claude transcript line from %q: %s", transcript, err)
			logger.Debugf("failed parsing claude transcript line: %s", line)

			continue
		}

		if logLine.Timestamp.IsZero() {
			continue
		}

		if logLine.Version != "" {
			claudeVersion = logLine.Version
		}

		if logLine.Timestamp.Before(g.After) {
			continue
		}

		if logLine.ToolUseResult == nil {
			continue
		}

		filePath := getFilePath(*logLine.ToolUseResult)
		if filePath == "" {
			continue
		}

		lineChanges := claudeLineChanges(*logLine.ToolUseResult)

		isWrite := lineChanges != 0

		heartbeats = append(heartbeats, heartbeat.New(
			heartbeat.PointerTo(lineChanges),
			"",
			heartbeat.AICodingCategory.String(),
			nil,
			filePath,
			heartbeat.FileType,
			nil,
			false,
			heartbeat.PointerTo(isWrite),
			nil,
			"",
			nil,
			nil,
			"",
			"",
			false,
			"",
			"",
			float64(logLine.Timestamp.Unix()),
			heartbeat.UserAgent(ctx, claudePlugin(claudeVersion)),
		))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading claude transcript %q: %s", transcript, err)
	}

	return heartbeats, nil
}

func getFilePath(result toolUseResult) string {
	if result.FilePath != nil {
		return *result.FilePath
	}

	if result.File != nil && result.File.FilePath != nil {
		return *result.File.FilePath
	}

	return ""
}

func claudeLineChanges(result toolUseResult) int {
	if result.StructuredPatch != nil && len(*result.StructuredPatch) > 0 {
		lineChanges := 0
		for _, patch := range *result.StructuredPatch {
			lineChanges += patch.NewLines - patch.OldLines
		}

		return lineChanges
	}

	if result.Content != nil && result.OriginalFile == nil {
		lineChanges := 1

		for _, char := range *result.Content {
			if char == '\n' {
				lineChanges++
			}
		}

		return lineChanges
	}

	return 0
}

func claudePlugin(version string) string {
	if version == "" {
		return "Claude Code"
	}

	return "Claude Code/" + version
}

// ID returns its id.
func (Claude) ID() ParserID {
	return ClaudeParser
}
