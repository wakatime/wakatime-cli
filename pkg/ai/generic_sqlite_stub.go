//go:build (freebsd && !amd64 && !arm64) || (openbsd && !amd64 && !arm64) || netbsd || dragonfly

package ai

import "context"

func aiSQLiteBudgetContext(ctx context.Context) context.Context { return ctx }

func parseGenericAISQLite(context.Context, genericAIProvider) (Heartbeats, error) {
	return Heartbeats{}, nil
}
