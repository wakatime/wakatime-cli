//go:build netbsd || dragonfly || (freebsd && !(amd64 || arm64)) || (openbsd && !(amd64 || arm64))

package ai

import "context"

// Cody contains params for detecting heartbeats from Cody by Sourcegraph transcripts.
type Cody ParserConfig

// Parse is a no-op on unsupported BSD targets because the Cody SQLite reader is not built there.
func (Cody) Parse(context.Context) (Heartbeats, error) {
	return Heartbeats{}, nil
}

// Name returns its name.
func (Cody) Name() string {
	return "Cody"
}
