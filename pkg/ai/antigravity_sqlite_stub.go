//go:build netbsd || dragonfly || (freebsd && !(amd64 || arm64)) || (openbsd && !(amd64 || arm64))

package ai

import "context"

func (Gemini) antigravitySQLite(context.Context, string) (Heartbeats, error) { return nil, nil }
