//go:build netbsd || dragonfly || (freebsd && !(amd64 || arm64)) || (openbsd && !(amd64 || arm64))

package ai

import "context"

func (Copilot) reconcileSQLiteUsage(_ context.Context, timed []copilotTimedHeartbeat) ([]copilotTimedHeartbeat, error) {
	return timed, nil
}
func (Copilot) otelHeartbeats(context.Context, []copilotTimedHeartbeat) ([]copilotTimedHeartbeat, error) {
	return nil, nil
}
