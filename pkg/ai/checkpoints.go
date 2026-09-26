package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
	goini "gopkg.in/ini.v1"
)

// Checkpoints are read and replaced while holding AcquireSyncLock. Keep them
// beside the internal config so profiles do not share progress. A single atomic
// replacement commits all parser/session cutoffs together.
type syncCheckpoints struct {
	Version        int                          `json:"version"`
	PreviousGlobal time.Time                    `json:"previous_global"`
	Global         time.Time                    `json:"global"`
	Parsers        map[string]*parserCheckpoint `json:"parsers"`
	path           string
	original       []byte
}

type parserCheckpoint struct {
	Cutoff      time.Time                     `json:"cutoff"`
	Sessions    map[string]*sessionCheckpoint `json:"sessions"`
	Sources     map[string]checkpointSource   `json:"sources,omitempty"`
	Cursors     map[string]json.RawMessage    `json:"cursors,omitempty"`
	incremental map[string]int
	// global is this run's global cutoff (ai_logs_last_parsed_at). It is the
	// cutoff a session gets the first time it is seen, and is never persisted.
	global time.Time
	// legacy holds pre-checkpoint cursor files that were read as a fallback.
	// They are removed once their state is safely saved in Cursors, so a lost
	// or reset checkpoint cannot resurrect their stale progress.
	legacy map[string]struct{}
}

// clone deep-copies p so callers can roll back to it after a failed parser
// run without paying for a JSON marshal/unmarshal round-trip on every run.
func (p *parserCheckpoint) clone() *parserCheckpoint {
	clone := &parserCheckpoint{Cutoff: p.Cutoff, global: p.global}

	if p.Sessions != nil {
		clone.Sessions = make(map[string]*sessionCheckpoint, len(p.Sessions))

		for id, session := range p.Sessions {
			copied := *session
			if session.Boundary != nil {
				copied.Boundary = make(map[string][]checkpointTokens, len(session.Boundary))
				for key, tokens := range session.Boundary {
					copied.Boundary[key] = append([]checkpointTokens(nil), tokens...)
				}
			}

			clone.Sessions[id] = &copied
		}
	}

	if p.Sources != nil {
		clone.Sources = make(map[string]checkpointSource, len(p.Sources))
		maps.Copy(clone.Sources, p.Sources)
	}

	if p.Cursors != nil {
		clone.Cursors = make(map[string]json.RawMessage, len(p.Cursors))
		maps.Copy(clone.Cursors, p.Cursors)
	}

	if p.incremental != nil {
		clone.incremental = make(map[string]int, len(p.incremental))
		maps.Copy(clone.incremental, p.incremental)
	}

	return clone
}

type sessionCheckpoint struct {
	Cutoff   time.Time                     `json:"cutoff"`
	Boundary map[string][]checkpointTokens `json:"boundary,omitempty"`
}

func loadSyncCheckpoints(ctx context.Context, v *viper.Viper, after time.Time) (*syncCheckpoints, error) {
	path, err := ini.InternalFilePath(ctx, v)
	if err != nil {
		return nil, err
	}

	path = strings.TrimSuffix(path, filepath.Ext(path)) + "-ai-parsing.json"
	state := &syncCheckpoints{Version: 1, Global: after, Parsers: make(map[string]*parserCheckpoint), path: path}

	raw, err := os.ReadFile(filepath.Clean(path))
	if os.IsNotExist(err) {
		return state, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed reading ai checkpoints: %w", err)
	}

	if err := json.Unmarshal(raw, state); err != nil {
		// An unreadable file (for example zero-length after a crash) can never
		// heal itself: start over from the global cutoff instead of leaving ai
		// sync disabled until someone deletes the file by hand.
		log.Extract(ctx).Warnf("discarding unreadable ai checkpoints %q: %s", path, err)

		return &syncCheckpoints{Version: 1, Global: after, Parsers: make(map[string]*parserCheckpoint), path: path}, nil
	}

	if state.Version != 1 {
		return nil, fmt.Errorf("unsupported ai checkpoint version: %d", state.Version)
	}
	// An explicitly rewound global cutoff requests a backfill.
	if after.Before(state.Global) && !after.Equal(state.PreviousGlobal) {
		state.Parsers = make(map[string]*parserCheckpoint)
	}

	if state.Parsers == nil {
		state.Parsers = make(map[string]*parserCheckpoint)
	}

	// Drop any corrupted entry instead of failing the whole file: one bad
	// parser or session must not permanently disable ai sync for the rest.
	for name, parser := range state.Parsers {
		if parser == nil {
			delete(state.Parsers, name)
			continue
		}

		for id, session := range parser.Sessions {
			if session == nil {
				delete(parser.Sessions, id)
			}
		}
	}

	state.Global = after
	state.PreviousGlobal = after
	state.original = raw

	return state, nil
}

func (s *syncCheckpoints) parser(name string, after time.Time) *parserCheckpoint {
	state := s.Parsers[name]
	if state == nil {
		state = &parserCheckpoint{Cutoff: after}
		s.Parsers[name] = state
	}

	state.global = after

	if state.Sessions == nil {
		state.Sessions = make(map[string]*sessionCheckpoint)
	}

	return state
}

func (p *parserCheckpoint) discoveryAfter() time.Time {
	// With no recorded session cutoff this is just the global cutoff.
	after := p.global
	for _, session := range p.Sessions {
		if session.Cutoff.Before(after) {
			after = session.Cutoff
		}
	}

	return checkpointReadAfter(after)
}

// Include the boundary; identities below suppress previously emitted records,
// while allowing records appended later at exactly the same timestamp. The
// small overlap also compensates for float64 heartbeat timestamp rounding.
func checkpointReadAfter(after time.Time) time.Time {
	if after.IsZero() {
		return after
	}

	return after.Add(-time.Microsecond)
}

func (c ParserConfig) sessionAfter(id string) time.Time {
	if c.checkpoint == nil {
		return c.After
	}

	p := c.checkpoint

	// The first time a session is seen it starts at the global cutoff, and is
	// recorded there so later runs keep that cutoff even after other parsers
	// advance the global one.
	session := p.Sessions[id]
	if session == nil {
		session = &sessionCheckpoint{Cutoff: p.global}
		p.Sessions[id] = session
	}

	return checkpointReadAfter(session.Cutoff)
}

type checkpointTokens struct {
	Input  int64 `json:"input,omitempty"`
	Cached int64 `json:"cached,omitempty"`
	Output int64 `json:"output,omitempty"`
}

func (p *parserCheckpoint) filter(hh Heartbeats) Heartbeats {
	// The first time a session is seen it starts at the global cutoff.
	fallback := p.global
	result := make(Heartbeats, 0, len(hh))
	next := make(map[string]*sessionCheckpoint)
	occurrences := make(map[string]map[string]int)

	for _, h := range hh {
		previous := p.Sessions[h.AISession]
		if previous == nil {
			previous = &sessionCheckpoint{Cutoff: fallback}
			p.Sessions[h.AISession] = previous
		}

		timestamp := heartbeatTime(h.Time)

		incremental := p.takeIncremental(h)
		if timestamp.Before(previous.Cutoff) && !incremental {
			continue
		}

		key := checkpointHeartbeatKey(h)
		if occurrences[h.AISession] == nil {
			occurrences[h.AISession] = make(map[string]int)
		}

		index := occurrences[h.AISession][key]
		occurrences[h.AISession][key]++
		counts := checkpointTokens{h.AIInputTokens, h.AICachedInputTokens, h.AIOutputTokens}

		if !incremental && timestamp.Equal(previous.Cutoff) && index < len(previous.Boundary[key]) {
			old := previous.Boundary[key][index]
			counts.Input = max(counts.Input, old.Input)
			counts.Cached = max(counts.Cached, old.Cached)
			counts.Output = max(counts.Output, old.Output)
			h.AIInputTokens = counts.Input - old.Input
			h.AICachedInputTokens = counts.Cached - old.Cached

			h.AIOutputTokens = counts.Output - old.Output
			if h.AIInputTokens > 0 || h.AICachedInputTokens > 0 || h.AIOutputTokens > 0 {
				// Some stores update usage in place without changing the activity's time.
				// Emit only growth, at observation time, and retain the source checkpoint.
				h.Time = heartbeatTimestamp(time.Now())
				h.AILineChanges = nil
				h.HumanLineChanges = nil
				h.AIPromptLength = 0
				result = append(result, h)
			}
		} else {
			result = append(result, h)
		}

		if timestamp.Before(previous.Cutoff) {
			continue // A durable SQLite row cursor may deliver older rows in a later batch.
		}

		current := next[h.AISession]
		if current == nil || timestamp.After(current.Cutoff) {
			current = &sessionCheckpoint{Cutoff: timestamp, Boundary: make(map[string][]checkpointTokens)}
			next[h.AISession] = current
		}

		if timestamp.Equal(current.Cutoff) {
			current.Boundary[key] = append(current.Boundary[key], counts)
		}

		if timestamp.After(p.Cutoff) {
			p.Cutoff = timestamp
		}
	}

	for id, current := range next {
		previous := p.Sessions[id]
		if current.Cutoff.Equal(previous.Cutoff) {
			for key, counts := range previous.Boundary {
				if len(counts) > len(current.Boundary[key]) {
					current.Boundary[key] = append(current.Boundary[key], counts[len(current.Boundary[key]):]...)
				}
			}
		}

		p.Sessions[id] = current
	}

	return result
}

func checkpointHeartbeatKey(h heartbeat.Heartbeat) string {
	// Token deltas at the boundary may be zero after older usage establishes the
	// baseline. They are not part of the source activity's identity.
	raw, _ := json.Marshal(struct {
		Time     float64
		Entity   string
		Type     heartbeat.EntityType
		Category string
		Write    *bool
		Prompt   int
		Lines    *int
	}{h.Time, h.Entity, h.EntityType, h.Category, h.IsWrite, h.AIPromptLength, h.AILineChanges})

	return fmt.Sprintf("%x", sha256.Sum256(raw))
}

func (s *syncCheckpoints) save() error {
	// Commit before updating the legacy global cutoff. PreviousGlobal allows
	// a retry if that separate INI update fails or the process exits first.
	raw, err := json.Marshal(s)
	if err != nil {
		return err
	}

	if string(raw) == string(s.original) {
		return nil
	}

	if err := atomicWriteFile(filepath.Dir(s.path), ".ai-parsing-*", s.path, raw); err != nil {
		return err
	}

	s.removeMigratedLegacyCursors()

	return nil
}

func (s *syncCheckpoints) removeMigratedLegacyCursors() {
	for _, p := range s.Parsers {
		for path := range p.legacy {
			if _, ok := p.Cursors[path]; ok {
				_ = os.Remove(path)
			}
		}

		p.legacy = nil
	}
}

// atomicWriteFile replaces path's contents by writing to a temporary file in
// dir and renaming it into place, so a crash or concurrent read never
// observes a partially written file.
func atomicWriteFile(dir, pattern, path string, contents []byte) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	file, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return err
	}
	defer os.Remove(file.Name()) // nolint:errcheck

	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return os.Rename(file.Name(), path)
}

func newCheckpointOpenCode(config ParserConfig) OpenCode {
	return OpenCode{After: config.After, FallbackUserAgent: config.FallbackUserAgent, UserAgents: config.UserAgents,
		ProjectInfo: config.ProjectInfo, checkpoint: config.checkpoint}
}

// getSyncLastParsedAt reads the freshest on-disk value of ai_logs_last_parsed_at,
// since a prior sync pass in this same process (for example while replaying
// offline heartbeats) may have written it without updating v in memory. This
// is a read, so it deliberately avoids ini.NewWriter's config mutex: it must
// not add a blocking or failing dependency to the per-heartbeat ai sync path.
func getSyncLastParsedAt(ctx context.Context, v *viper.Viper) (time.Time, error) {
	path, err := ini.InternalFilePath(ctx, v)
	if err != nil {
		return time.Time{}, err
	}

	file, fileErr := goini.LoadSources(goini.LoadOptions{
		AllowPythonMultilineValues: true,
		SkipUnrecognizableLines:    true,
	}, path)
	if fileErr == nil {
		if value := file.Section("internal").Key("ai_logs_last_parsed_at").String(); value != "" {
			fresh := viper.New()
			fresh.Set("internal.ai_logs_last_parsed_at", value)

			return getLastParsedAt(ctx, fresh)
		}
	}

	return getLastParsedAt(ctx, v)
}

type checkpointSource struct {
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// Cache only independent sources. Parsers that reconcile records across files
// (for example Codex forks) must still visit those files together.
func parseCheckpointSource(config ParserConfig, path string, parse func() (Heartbeats, error)) (Heartbeats, error) {
	if config.checkpoint == nil {
		return parse()
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	source := checkpointSource{Size: info.Size(), Modified: info.ModTime()}

	state := config.checkpoint
	if previous, ok := state.Sources[path]; ok && previous.Size == source.Size && previous.Modified.Equal(source.Modified) {
		return nil, nil
	}

	hh, err := parse()
	if err != nil {
		return nil, err
	}

	if state.Sources == nil {
		state.Sources = make(map[string]checkpointSource)
	}
	// Snapshot before reading: a concurrent append causes a retry next run.
	state.Sources[path] = source

	return hh, nil
}

// SQLite cursors already identify new and changed records, including older rows
// reached in a later budgeted batch. Do not discard these at a time watermark.
func (p *parserCheckpoint) markIncremental(hh Heartbeats) {
	if p.incremental == nil {
		p.incremental = make(map[string]int)
	}

	for _, h := range hh {
		p.incremental[incrementalHeartbeatKey(h)]++
	}
}

func (p *parserCheckpoint) takeIncremental(h heartbeat.Heartbeat) bool {
	if len(p.incremental) == 0 {
		return false
	}

	key := incrementalHeartbeatKey(h)
	if p.incremental[key] == 0 {
		return false
	}

	p.incremental[key]--

	return true
}

func incrementalHeartbeatKey(h heartbeat.Heartbeat) string {
	raw, _ := json.Marshal(h)
	return fmt.Sprintf("%x", sha256.Sum256(raw))
}
