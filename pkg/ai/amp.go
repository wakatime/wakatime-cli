package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Amp contains params for detecting heartbeats from Amp CLI thread logs.
type Amp ParserConfig

type (
	ampToolArgs struct {
		PatchText string `json:"patchText"`
		Workdir   string `json:"workdir"`
	}

	ampToolData struct {
		ToolCallID string          `json:"toolCallId"`
		ToolName   string          `json:"toolName"`
		Args       json.RawMessage `json:"args"`
	}

	ampLogLine struct {
		Timestamp       time.Time    `json:"@timestamp"`
		Message         string       `json:"message"`
		ThreadID        string       `json:"threadId"`
		Type            string       `json:"type"`
		ToolCallID      string       `json:"toolCallId"`
		RunStatus       string       `json:"runStatus"`
		HasRunError     bool         `json:"hasRunError"`
		ReasoningEffort string       `json:"reasoningEffort"`
		Data            *ampToolData `json:"data"`
		PID             int          `json:"pid"`
	}

	ampCLILogLine struct {
		Timestamp         time.Time `json:"@timestamp"`
		Message           string    `json:"message"`
		Version           string    `json:"version"`
		CurrentVersion    string    `json:"currentVersion"`
		WorkspaceRoot     string    `json:"workspaceRoot"`
		WorkspaceRootPath string    `json:"workspaceRootPath"`
		PID               int       `json:"pid"`
	}

	ampProcessMetadata struct {
		cwd       string
		startedAt time.Time
		version   string
	}

	ampSessionState struct {
		cwd     string
		id      string
		version string
	}

	ampPendingPatch struct {
		cwd             string
		patchText       string
		reasoningEffort string
		sessionID       string
		version         string
	}

	ampParseState struct {
		cwd             string
		heartbeats      Heartbeats
		pendingPatches  map[string]ampPendingPatch
		reasoningEffort string
		sessionID       string
		version         string
	}

	ampSessionMetadata map[int][]ampProcessMetadata
)

// Parse parses Amp CLI JSONL thread logs for ai heartbeats.
func (g Amp) Parse(ctx context.Context) (Heartbeats, error) {
	logger := log.Extract(ctx)

	transcripts, err := g.transcriptPaths(ctx)
	if err != nil {
		return nil, err
	}

	if len(transcripts) == 0 {
		return Heartbeats{}, nil
	}

	logger.Debugf("Found %d transcript logs modified after %s for %s", len(transcripts), g.After, g.Name())
	metadata := g.readCLIMetadata(logger, transcripts)

	var heartbeats Heartbeats

	for _, transcript := range transcripts {
		parsed, err := g.parseTranscript(ctx, transcript, metadata)
		if err != nil {
			logger.Warnf("failed parsing amp transcript %q: %s", transcript, err)
			continue
		}

		heartbeats = append(heartbeats, parsed...)
	}

	return heartbeats, nil
}

func (g Amp) transcriptPaths(ctx context.Context) ([]string, error) {
	home, err := ini.UserHomeDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find user home dir: %s", err)
	}

	transcriptDirs := []string{
		filepath.Join(home, ".cache", "amp", "logs", "threads"),
	}

	if cacheDir, err := os.UserCacheDir(); err == nil {
		transcriptDirs = append(transcriptDirs, filepath.Join(cacheDir, "amp", "logs", "threads"))
	}

	seenDirs := make(map[string]struct{}, len(transcriptDirs))
	seenTranscripts := make(map[string]struct{})

	var transcripts []string

	for _, transcriptDir := range transcriptDirs {
		transcriptDir = filepath.Clean(transcriptDir)
		if _, found := seenDirs[transcriptDir]; found {
			continue
		}

		seenDirs[transcriptDir] = struct{}{}

		if _, err := os.Stat(transcriptDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, fmt.Errorf("failed to stat Amp thread logs directory %q: %s", transcriptDir, err)
		}

		err := filepath.WalkDir(transcriptDir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}

			if entry.IsDir() || filepath.Ext(entry.Name()) != ".log" {
				return nil
			}

			info, err := entry.Info()
			if err != nil || info.ModTime().Before(g.After) {
				return nil
			}

			if _, found := seenTranscripts[path]; !found {
				seenTranscripts[path] = struct{}{}
				transcripts = append(transcripts, path)
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to walk Amp thread logs directory %q: %s", transcriptDir, err)
		}
	}

	sort.Strings(transcripts)

	return transcripts, nil
}

func (Amp) readCLIMetadata(logger *log.Logger, transcripts []string) ampSessionMetadata {
	metadata := make(ampSessionMetadata)
	seen := make(map[string]struct{})

	for _, transcript := range transcripts {
		logsDir := filepath.Dir(filepath.Dir(transcript))

		cliLog := filepath.Join(logsDir, "cli.log")
		if _, found := seen[cliLog]; found {
			continue
		}

		seen[cliLog] = struct{}{}

		parsed, err := readAmpCLIMetadata(cliLog)
		if err != nil {
			if !os.IsNotExist(err) {
				logger.Warnf("failed parsing Amp CLI metadata from %q: %s", cliLog, err)
			}

			continue
		}

		for pid, records := range parsed {
			metadata[pid] = append(metadata[pid], records...)
		}
	}

	for pid := range metadata {
		sort.Slice(metadata[pid], func(i int, j int) bool {
			return metadata[pid][i].startedAt.Before(metadata[pid][j].startedAt)
		})
	}

	return metadata
}

func readAmpCLIMetadata(path string) (ampSessionMetadata, error) {
	//nolint:gosec
	fh, err := os.Open(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	defer fh.Close() // nolint:errcheck,gosec

	metadata := make(ampSessionMetadata)
	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	for scanner.Scan() {
		var line ampCLILogLine
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil || line.PID == 0 {
			continue
		}

		records := metadata[line.PID]
		if line.Message == "Starting Amp CLI." {
			records = append(records, ampProcessMetadata{
				startedAt: line.Timestamp,
				version:   line.Version,
			})
		} else if len(records) == 0 {
			records = append(records, ampProcessMetadata{})
		}

		record := &records[len(records)-1]
		if line.Version != "" {
			record.version = line.Version
		} else if line.CurrentVersion != "" {
			record.version = line.CurrentVersion
		}

		workspace := firstNonEmptyString(line.WorkspaceRootPath, line.WorkspaceRoot)
		if workspace != "" {
			record.cwd = ampWorkspacePath(workspace)
		}

		metadata[line.PID] = records
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading Amp CLI log %q: %s", path, err)
	}

	return metadata, nil
}

func (m ampSessionMetadata) process(pid int, timestamp time.Time) ampProcessMetadata {
	records := m[pid]
	if len(records) == 0 {
		return ampProcessMetadata{}
	}

	var selected ampProcessMetadata

	for _, record := range records {
		if !timestamp.IsZero() && !record.startedAt.IsZero() && record.startedAt.After(timestamp) {
			break
		}

		selected = record
	}

	return selected
}

func ampWorkspacePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(path), "file://") {
		stripped := path[len("file://"):]
		if len(stripped) >= 2 && stripped[1] == ':' {
			return filepath.FromSlash(stripped)
		}

		uri, err := url.Parse(path)
		if err == nil && uri.Path != "" {
			path = uri.Path
		} else {
			path = stripped
		}
	}

	if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
		path = path[1:]
	}

	return filepath.FromSlash(path)
}

func (g Amp) parseTranscript(
	ctx context.Context,
	transcript string,
	metadata ampSessionMetadata,
) (Heartbeats, error) {
	logger := log.Extract(ctx)

	//nolint:gosec
	fh, err := os.Open(filepath.Clean(transcript))
	if err != nil {
		return nil, fmt.Errorf("failed to open amp transcript %q: %s", transcript, err)
	}
	defer fh.Close() // nolint:errcheck,gosec

	session, err := g.readSessionState(logger, fh, transcript, metadata)
	if err != nil {
		return nil, err
	}

	scanner, err := ampScanner(fh, transcript)
	if err != nil {
		return nil, err
	}

	state := ampParseState{
		cwd:            session.cwd,
		pendingPatches: make(map[string]ampPendingPatch),
		sessionID:      session.id,
		version:        session.version,
	}

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		g.handleTranscriptLine(logger, transcript, scanner.Bytes(), &state)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading amp transcript %q: %s", transcript, err)
	}

	return state.heartbeats, nil
}

func (Amp) readSessionState(
	logger *log.Logger,
	fh *os.File,
	transcript string,
	metadata ampSessionMetadata,
) (ampSessionState, error) {
	reader := bufio.NewReader(fh)

	firstLine, err := reader.ReadBytes('\n')
	if err != nil && err != io.EOF {
		return ampSessionState{}, fmt.Errorf("failed to read amp transcript %q: %s", transcript, err)
	}

	state := ampSessionState{
		id: strings.TrimSuffix(filepath.Base(transcript), filepath.Ext(transcript)),
	}

	if len(firstLine) > 0 {
		var line ampLogLine
		if err := json.Unmarshal(firstLine, &line); err != nil {
			logger.Debugf("failed parsing amp session metadata from %q: %s", transcript, err)
		} else {
			if line.ThreadID != "" {
				state.id = line.ThreadID
			}

			process := metadata.process(line.PID, line.Timestamp)
			state.cwd = process.cwd
			state.version = process.version
		}
	}

	if _, err := fh.Seek(0, 0); err != nil {
		return ampSessionState{}, fmt.Errorf("failed to rewind amp transcript %q: %s", transcript, err)
	}

	return state, nil
}

func ampScanner(fh *os.File, transcript string) (*bufio.Scanner, error) {
	info, err := fh.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat amp transcript %q: %s", transcript, err)
	}

	skipFirstLine := false

	if info.Size() > maxTranscriptLineSize {
		if _, err := fh.Seek(info.Size()-maxTranscriptLineSize, 0); err != nil {
			return nil, fmt.Errorf("failed to seek amp transcript %q: %s", transcript, err)
		}

		skipFirstLine = true
	}

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), maxTranscriptLineSize)

	if skipFirstLine {
		scanner.Scan()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read amp transcript %q: %s", transcript, err)
		}
	}

	return scanner, nil
}

func (g Amp) handleTranscriptLine(
	logger *log.Logger,
	transcript string,
	line []byte,
	state *ampParseState,
) {
	if len(line) == 0 {
		return
	}

	var logLine ampLogLine
	if err := json.Unmarshal(line, &logLine); err != nil {
		logger.Warnf("failed parsing amp transcript line from %q: %s", transcript, err)
		logger.Debugf("failed parsing amp transcript line: %s", line)

		return
	}

	if logLine.ThreadID != "" {
		state.sessionID = logLine.ThreadID
	}

	if reasoningEffort := strings.TrimSpace(logLine.ReasoningEffort); reasoningEffort != "" {
		state.reasoningEffort = reasoningEffort
	}

	if logLine.Message == "onToolLease" && logLine.Data != nil {
		g.handleToolLease(logger, transcript, *logLine.Data, state)

		return
	}

	if logLine.Type != "executor_tool_result" || logLine.ToolCallID == "" {
		return
	}

	pending, found := state.pendingPatches[logLine.ToolCallID]
	if !found {
		return
	}

	delete(state.pendingPatches, logLine.ToolCallID)

	if logLine.RunStatus != "done" || logLine.HasRunError ||
		logLine.Timestamp.IsZero() || logLine.Timestamp.Before(g.After) {
		return
	}

	state.heartbeats = append(state.heartbeats, g.patchHeartbeats(
		logLine.Timestamp,
		pending.cwd,
		pending.patchText,
		pending.reasoningEffort,
		pending.sessionID,
		pending.version,
	)...)
}

func (Amp) handleToolLease(
	logger *log.Logger,
	transcript string,
	data ampToolData,
	state *ampParseState,
) {
	var args ampToolArgs
	if err := json.Unmarshal(data.Args, &args); err != nil {
		logger.Debugf("failed parsing amp tool arguments from %q: %s", transcript, err)

		return
	}

	workdir := state.cwd
	if args.Workdir != "" {
		workdir = args.Workdir
		if state.cwd == "" {
			state.cwd = workdir
		}
	}

	if !strings.EqualFold(data.ToolName, "apply_patch") || data.ToolCallID == "" || args.PatchText == "" {
		return
	}

	state.pendingPatches[data.ToolCallID] = ampPendingPatch{
		cwd:             workdir,
		patchText:       args.PatchText,
		reasoningEffort: state.reasoningEffort,
		sessionID:       state.sessionID,
		version:         state.version,
	}
}

func (g Amp) patchHeartbeats(
	timestamp time.Time,
	cwd string,
	input string,
	reasoningEffort string,
	sessionID string,
	version string,
) Heartbeats {
	var heartbeats Heartbeats

	var (
		currentFile string
		additions   int
		deletions   int
	)

	lines := strings.Split(input, "\n")
	for i := range lines {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		if moveFile := codexMoveFilePath(cwd, line); moveFile != "" {
			if currentFile != "" {
				currentFile = moveFile
			}

			continue
		}

		if filePath := codexFilePath(cwd, line); filePath != "" {
			if currentFile != "" {
				heartbeats = append(heartbeats, g.heartbeat(
					currentFile,
					reasoningEffort,
					sessionID,
					timestamp,
					version,
					additions,
					deletions,
				))
			}

			currentFile = filePath
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
		heartbeats = append(heartbeats, g.heartbeat(
			currentFile,
			reasoningEffort,
			sessionID,
			timestamp,
			version,
			additions,
			deletions,
		))
	}

	return heartbeats
}

func (g Amp) heartbeat(
	currentFile string,
	reasoningEffort string,
	sessionID string,
	timestamp time.Time,
	version string,
	additions int,
	deletions int,
) heartbeat.Heartbeat {
	return heartbeat.NewWithAITokens(
		heartbeat.PointerTo(additions-deletions),
		sessionID,
		heartbeat.AITokens{},
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
		float64(timestamp.UnixMilli())/1000,
		aiUserAgent(currentFile, g.UserAgents, g.FallbackUserAgent, ampAgentVersion(version, reasoningEffort)),
	)
}

func ampAgentVersion(version string, reasoningEffort string) string {
	version = unknownIfEmpty(version)

	reasoningEffort = strings.TrimSpace(reasoningEffort)
	if reasoningEffort == "" {
		return "amp/" + version
	}

	return "amp/" + version + "-" + reasoningEffort
}

// Name returns its id.
func (Amp) Name() string {
	return "Amp"
}
