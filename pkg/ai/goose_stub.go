//go:build freebsd || openbsd || netbsd || dragonfly

package ai

import "context"

// Goose contains params for detecting heartbeats from Goose session logs.
type Goose ParserConfig

// Parse is a no-op on BSD because the Goose SQLite reader is not built there.
func (Goose) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its id.
func (Goose) Name() string {
	return "Goose"
}
