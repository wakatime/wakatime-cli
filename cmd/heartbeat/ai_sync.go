package heartbeat

import (
	"context"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// RunAISyncActivity parses AI transcripts and sends resulting AI heartbeats
// without requiring a CLI heartbeat entity.
func RunAISyncActivity(ctx context.Context, v *viper.Viper) (int, error) {
	logger := log.Extract(ctx)

	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		logger.Warnf("failed to load offline queue filepath: %s", err)
	}

	aiParams, err := params.LoadAIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return exitcode.ErrAuth, fmt.Errorf("failed to load ai parameters: %w", err)
	}

	apiParams, err := params.LoadAPIParams(ctx, v, params.FlagReadOrderFlagPrecedence)
	if err != nil {
		return exitcode.ErrAuth, fmt.Errorf("failed to load API parameters: %w", err)
	}

	loadedParams := params.Params{
		AI:      aiParams,
		API:     apiParams,
		Offline: params.LoadOfflineParams(ctx, v, params.FlagReadOrderFlagPrecedence),
	}

	heartbeats, err := applyAIParsing(ctx, v, loadedParams, []heartbeat.Heartbeat{})
	if err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("sending ai activity failed: %w", errwaka)
		}

		return exitcode.ErrGeneric, fmt.Errorf("sending ai activity failed: %w", err)
	}

	if len(heartbeats) == 0 {
		logger.Debugln("no ai activity found to sync")

		return exitcode.Success, nil
	}

	if err := sendPreparedHeartbeats(ctx, v, loadedParams, queueFilepath, heartbeats, false); err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("sending ai activity failed: %w", errwaka)
		}

		return exitcode.ErrGeneric, fmt.Errorf("sending ai activity failed: %w", err)
	}

	logger.Debugln("successfully synced ai activity")

	return exitcode.Success, nil
}
