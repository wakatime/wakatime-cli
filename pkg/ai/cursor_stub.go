//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import (
	"context"
)

// Cursor contains params for detecting heartbeats from Cursor transcripts.
type Cursor ParserConfig

// Parse is a no-op on BSD because the Cursor SQLite reader is not built there.
func (Cursor) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Cursor) Name() string {
	return "Cursor"
}
