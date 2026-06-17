//go:build netbsd || dragonfly || (freebsd && !(amd64 || arm64)) || (openbsd && !(amd64 || arm64))

package ai

import (
	"context"
)

// Cursor contains params for detecting heartbeats from Cursor transcripts.
type Cursor ParserConfig

// Parse is a no-op on unsupported BSD targets because the Cursor SQLite reader is not built there.
func (Cursor) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Cursor) Name() string {
	return "Cursor"
}
