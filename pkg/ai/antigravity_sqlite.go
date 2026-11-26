//go:build (!freebsd && !openbsd && !netbsd && !dragonfly) || (freebsd && (amd64 || arm64)) || (openbsd && (amd64 || arm64))

package ai

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

type antigravityProtoField struct {
	number uint64
	value  uint64
	bytes  []byte
}

// Decode just the protobuf wire format; unknown fields remain safely skippable.
func antigravityProto(data []byte) []antigravityProtoField {
	var fields []antigravityProtoField

	for len(data) > 0 {
		tag, n := binary.Uvarint(data)
		if n <= 0 || tag>>3 == 0 {
			return nil
		}

		data = data[n:]

		field := antigravityProtoField{number: tag >> 3}
		switch tag & 7 {
		case 0:
			v, n := binary.Uvarint(data)
			if n <= 0 {
				return nil
			}

			field.value = v
			data = data[n:]
		case 1:
			if len(data) < 8 {
				return nil
			}

			data = data[8:]
		case 2:
			length, n := binary.Uvarint(data)
			if n <= 0 {
				return nil
			}

			data = data[n:]
			if length > uint64(len(data)) {
				return nil
			}

			field.bytes = data[:length]
			data = data[length:]
		case 5:
			if len(data) < 4 {
				return nil
			}

			data = data[4:]
		default:
			return nil
		}

		fields = append(fields, field)
	}

	return fields
}

func antigravityField(fields []antigravityProtoField, number uint64) antigravityProtoField {
	for _, field := range fields {
		if field.number == number {
			return field
		}
	}

	return antigravityProtoField{}
}

func antigravityProtoTime(field antigravityProtoField) time.Time {
	if stamp := genericAITime(string(field.bytes)); !stamp.IsZero() {
		return stamp
	}

	if len(field.bytes) > 0 {
		nested := antigravityProto(field.bytes)

		seconds := antigravityField(nested, 1).value

		nanos := antigravityField(nested, 2).value
		if seconds > 0 && seconds < 1<<62 && nanos < 1e9 {
			return time.Unix(int64(seconds), int64(nanos)).UTC()
		}
	}

	if field.value > 0 && field.value < 1<<62 {
		return genericAITime(int64(field.value))
	}

	return time.Time{}
}

func (g Gemini) antigravitySQLite(ctx context.Context, home string) (Heartbeats, error) {
	var result Heartbeats

	for _, product := range antigravityProductDirs() {
		for _, directory := range []string{"conversations", "implicit"} {
			paths, err := filepath.Glob(filepath.Join(home, ".gemini", product.appDir, directory, "*.db"))
			if err != nil {
				return nil, err
			}

			for _, path := range paths {
				sessionID := strings.TrimSuffix(filepath.Base(path), ".db")
				if !aiSQLiteModifiedAfter(path, ParserConfig(g).sessionAfter(sessionID)) {
					continue
				}

				transcript := antigravityTranscript{path: path, sessionID: sessionID,
					entityName: product.entityName, userAgentProduct: product.userAgentProduct + "/unknown"}

				parsed, err := g.parseAntigravityDB(ctx, transcript)
				if err != nil {
					if isAntigravitySkippableSQLiteError(err) {
						log.Extract(ctx).Warnf("skipping Antigravity sqlite database %q: %s", path, err)
						continue
					}

					return result, err
				}

				result = append(result, parsed...)
			}
		}
	}

	return result, nil
}

func (g Gemini) parseAntigravityDB(ctx context.Context, transcript antigravityTranscript) (Heartbeats, error) {
	ctx, cancel := aiSQLiteContext(ctx)
	defer cancel()

	db, err := openAISQLiteDB(ctx, transcript.path)
	if err != nil {
		return nil, err
	}
	defer db.Close() // nolint:errcheck

	tables, err := genericAISQLiteTables(ctx, db)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(tables, "gen_metadata") {
		return nil, nil
	}

	steps := make(map[uint64][]byte)

	if slices.Contains(tables, "steps") {
		rows, err := db.QueryContext(ctx, `SELECT idx, metadata FROM steps`)
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var (
				idx      uint64
				metadata []byte
			)

			if err := rows.Scan(&idx, &metadata); err != nil {
				_ = rows.Close()
				return nil, err
			}

			if err := aiSQLiteRead(ctx, string(metadata)); err != nil {
				_ = rows.Close()
				return nil, err
			}

			steps[idx] = metadata
		}

		err = rows.Err()
		_ = rows.Close()

		if err != nil {
			return nil, err
		}
	}

	rows, err := db.QueryContext(ctx, `SELECT idx,data FROM gen_metadata ORDER BY idx`)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint:errcheck

	clock, err := newTranscriptClock(ctx, transcript.path)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)

	var result Heartbeats

	for rows.Next() {
		var (
			idx int64
			raw []byte
		)

		if err := rows.Scan(&idx, &raw); err != nil {
			return nil, err
		}

		if err := aiSQLiteRead(ctx, string(raw)); err != nil {
			return nil, err
		}

		result = append(result, g.antigravityRow(transcript, idx, raw, steps, clock, seen)...)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := clock.save(); err != nil {
		return nil, err
	}

	return result, nil
}

func (g Gemini) antigravityRow(transcript antigravityTranscript, idx int64, raw []byte,
	steps map[uint64][]byte, clock *transcriptClock, seen map[string]bool) Heartbeats {
	root := antigravityProto(raw)
	chat := antigravityProto(antigravityField(root, 1).bytes)

	usage := antigravityProto(antigravityField(chat, 4).bytes)
	if len(usage) == 0 {
		return nil
	}

	id := string(antigravityField(usage, 11).bytes)
	if id == "" {
		id = fmt.Sprint(idx)
	}

	if seen[id] {
		return nil
	}

	seen[id] = true
	indices := antigravityStepIndices(root)
	metadata := antigravityProto(antigravityField(chat, 9).bytes)

	parsedStamp := antigravityProtoTime(antigravityField(metadata, 4))
	if parsedStamp.IsZero() {
		parsedStamp = antigravityStepsTime(indices, steps)
	}

	stamp := clock.resolve(id, parsedStamp)
	if !timestampAtOrAfterCutoff(stamp, ParserConfig(g).sessionAfter(transcript.sessionID)) {
		return nil
	}

	model := string(antigravityField(chat, 19).bytes)
	if model == "" {
		model = string(antigravityField(chat, 21).bytes)
	}

	if model == "" {
		for _, field := range chat {
			if field.number == 20 {
				pair := antigravityProto(field.bytes)
				if string(antigravityField(pair, 1).bytes) == "model_enum" {
					model = string(antigravityField(pair, 2).bytes)
				}
			}
		}
	}

	input := antigravityField(usage, 2).value
	if input == 0 {
		input = antigravityField(usage, 1).value
	}

	output := antigravityField(usage, 3).value
	if output == 0 {
		output = antigravityField(usage, 9).value + antigravityField(usage, 10).value
	}

	if input > 1<<62 || output > 1<<62 {
		return nil
	}

	event := genericAIEvent{sessionID: transcript.sessionID, model: model, timestamp: stamp, input: int64(input),
		output: int64(output), tokensFound: true}

	emitted := genericAIEventHeartbeats(g, ParserConfig(g), event)
	for i := range emitted {
		emitted[i].Entity = transcript.entityName
		emitted[i].UserAgent = aiUserAgentWithModelAndEditor(emitted[i].Entity, g.UserAgents, g.FallbackUserAgent,
			model, "", transcript.userAgentProduct)
	}

	result := emitted

	return append(result, g.antigravityTools(transcript, indices, steps, event, seen)...)
}

func antigravityStepIndices(root []antigravityProtoField) []uint64 {
	var indices []uint64

	for _, field := range root {
		if field.number != 2 {
			continue
		}

		if field.bytes == nil {
			indices = append(indices, field.value)
			continue
		}

		data := field.bytes
		for len(data) > 0 {
			index, n := binary.Uvarint(data)
			if n <= 0 {
				break
			}

			indices = append(indices, index)
			data = data[n:]
		}
	}

	return indices
}

func antigravityStepTime(metadata []byte) time.Time {
	if len(metadata) == 0 {
		return time.Time{}
	}

	fields := antigravityProto(metadata)

	// Field 1 is the step created_at timestamp; fields 7, 8, and 6 carry completion timestamps.
	for _, fieldNumber := range []uint64{1, 7, 8, 6} {
		stamp := antigravityProtoTime(antigravityField(fields, fieldNumber))
		if !stamp.IsZero() && stamp.Year() >= 2020 {
			return stamp
		}
	}

	return time.Time{}
}

func antigravityStepsTime(indices []uint64, steps map[uint64][]byte) time.Time {
	for _, index := range indices {
		if stamp := antigravityStepTime(steps[index]); !stamp.IsZero() {
			return stamp
		}
	}

	return time.Time{}
}

func isAntigravitySkippableSQLiteError(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	code := sqliteErr.Code()
	primary := code & 0xff

	return primary == sqlite3.SQLITE_BUSY ||
		primary == sqlite3.SQLITE_LOCKED ||
		code == sqlite3.SQLITE_READONLY_RECOVERY
}

func antigravityToolLineChanges(name string, args map[string]any) *int {
	switch name {
	case "write_to_file":
		if content, ok := args["CodeContent"].(string); ok {
			count := countStringLines(content)
			return &count
		}
	case "write_file":
		if content, ok := args["Content"].(string); ok {
			count := countStringLines(content)
			return &count
		}
	case "edit_file":
		target, okTarget := args["Target"].(string)

		replacement, okReplacement := args["Replacement"].(string)
		if okTarget && okReplacement {
			count := countStringLines(replacement) - countStringLines(target)
			return &count
		}
	}

	return nil
}

func (g Gemini) antigravityTools(transcript antigravityTranscript, indices []uint64,
	steps map[uint64][]byte, event genericAIEvent, seen map[string]bool) Heartbeats {
	var result Heartbeats

	model := event.model

	for _, index := range indices {
		key := fmt.Sprintf("step:%d", index)
		if seen[key] {
			continue
		}

		seen[key] = true

		stamp := antigravityStepTime(steps[index])
		if stamp.IsZero() {
			stamp = event.timestamp
		}

		for _, field := range antigravityProto(steps[index]) {
			if field.number != 4 {
				continue
			}

			tool := antigravityProto(field.bytes)
			name := string(antigravityField(tool, 2).bytes)

			var args map[string]any
			if json.Unmarshal(antigravityField(tool, 3).bytes, &args) != nil {
				continue
			}
			// A reference from committed generation metadata establishes completion.
			file := firstNonEmptyString(
				genericAIString(args["TargetFile"]),
				genericAIString(args["AbsolutePath"]),
				genericAIString(args["Path"]),
				genericAIString(args["file_path"]),
			)
			if file == "" {
				continue
			}

			write := name == "write_to_file" ||
				name == "replace_file_content" ||
				name == "multi_replace_file_content" ||
				name == "write_file" ||
				name == "edit_file"

			activity := genericAIEvent{
				sessionID:   transcript.sessionID,
				model:       model,
				timestamp:   stamp,
				filePath:    file,
				toolName:    name,
				isWrite:     write,
				lineChanges: antigravityToolLineChanges(name, args),
			}

			emitted := genericAIEventHeartbeats(g, ParserConfig(g), activity)
			for i := range emitted {
				emitted[i].UserAgent = aiUserAgentWithModelAndEditor(emitted[i].Entity, g.UserAgents,
					g.FallbackUserAgent, model, "", transcript.userAgentProduct)
			}

			result = append(result, emitted...)
		}
	}

	return result
}
