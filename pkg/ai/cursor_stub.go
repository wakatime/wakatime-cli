//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import (
	"context"
	"time"
)

// Cursor contains params for detecting heartbeats from Cursor transcripts.
type Cursor struct {
	After             time.Time
	FallbackUserAgent string
	UserAgents        map[string]string
}

// Parse is a no-op on BSD because the Cursor SQLite reader is not built there.
func (Cursor) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// ID returns its id.
func (Cursor) ID() ParserID {
	return CursorParser
}
