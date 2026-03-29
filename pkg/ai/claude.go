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

	toolUseResultValue struct {
		Object *toolUseResult
		String *string
	}

	claudeLogLine struct {
		Timestamp     time.Time           `json:"timestamp"`
		Version       string              `json:"version"`
		ToolUseResult *toolUseResultValue `json:"toolUseResult"`
	}
)

func (v *toolUseResultValue) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var result toolUseResult
	if err := json.Unmarshal(data, &result); err == nil {
		v.Object = &result

		return nil
	}

	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		v.String = &str

		return nil
	}

	return fmt.Errorf("unsupported toolUseResult type")
}

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

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open claude transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat claude transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek claude transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read claude transcript %q: %s", transcript, err)
		}
	}

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

		var logLine claudeLogLine
		if err := json.Unmarshal(line, &logLine); err != nil {
			logger.Warnf("failed parsing claude transcript line from %q: %s", transcript, err)
			logger.Debugf("failed parsing claude transcript line: %s", line)

			continue
		}

		if logLine.Timestamp.IsZero() || logLine.Timestamp.Before(g.After) {
			continue
		}

		if logLine.Version != "" {
			claudeVersion = logLine.Version
		}

		if logLine.ToolUseResult == nil || logLine.ToolUseResult.Object == nil {
			continue
		}

		filePath := getClaudeFilePath(*logLine.ToolUseResult.Object)
		if filePath == "" {
			continue
		}

		lineChanges := claudeLineChanges(*logLine.ToolUseResult.Object)

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

func getClaudeFilePath(result toolUseResult) string {
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
		return "ClaudeCode"
	}

	return "ClaudeCode/" + version
}

// ID returns its id.
func (Claude) ID() ParserID {
	return ClaudeParser
}
