package offlinesync

import (
	"context"
	"fmt"
	"os"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	cmdheartbeat "github.com/wakatime/wakatime-cli/cmd/heartbeat"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// RunWithoutRateLimiting executes the sync-offline-activity command without rate limiting.
func RunWithoutRateLimiting(ctx context.Context, v *viper.Viper) (int, error) {
	return run(ctx, v)
}

// RunWithRateLimiting executes sync-offline-activity command with rate limiting enabled.
func RunWithRateLimiting(ctx context.Context, v *viper.Viper) (int, error) {
	paramOffline := params.LoadOfflineParams(ctx, v)

	logger := log.Extract(ctx)

	if cmdheartbeat.RateLimited(cmdheartbeat.RateLimitParams{
		Disabled:   paramOffline.Disabled,
		LastSentAt: paramOffline.LastSentAt,
		Timeout:    paramOffline.RateLimit,
	}) {
		logger.Debugln("skip syncing offline activity to respect rate limit")
		return exitcode.Success, nil
	}

	return run(ctx, v)
}

func run(ctx context.Context, v *viper.Viper) (int, error) {
	paramOffline := params.LoadOfflineParams(ctx, v)
	if paramOffline.Disabled {
		return exitcode.Success, nil
	}

	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: failed to load offline queue filepath: %s",
			err,
		)
	}

	logger := log.Extract(ctx)

	queueFilepathLegacy, err := offline.QueueFilepathLegacy(ctx, v)
	if err != nil {
		logger.Warnf("legacy offline sync failed: failed to load offline queue filepath: %s", err)
	}

	if err = syncOfflineActivityLegacy(ctx, v, queueFilepathLegacy); err != nil {
		logger.Warnf("legacy offline sync failed: %s", err)
	}

	if err = SyncOfflineActivity(ctx, v, queueFilepath); err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("offline sync failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: %s",
			err,
		)
	}

	logger.Debugln("successfully synced offline activity")

	return exitcode.Success, nil
}

// syncOfflineActivityLegacy syncs the old offline activity by sending heartbeats
// from the legacy offline queue to the WakaTime API.
func syncOfflineActivityLegacy(ctx context.Context, v *viper.Viper, queueFilepath string) error {
	if queueFilepath == "" {
		return nil
	}

	if !fileExists(queueFilepath) {
		return nil
	}

	paramOffline := params.LoadOfflineParams(ctx, v)

	paramAPI, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(ctx, paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, paramOffline.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: paramAPI.Key,
			MapPatterns:   paramAPI.KeyPatterns,
		}),
	)

	_, err = handle(ctx, nil)
	if err != nil {
		return err
	}

	logger := log.Extract(ctx)

	if err := os.Remove(queueFilepath); err != nil {
		logger.Warnf("failed to delete legacy offline file: %s", err)
	}

	return nil
}

// SyncOfflineActivity syncs offline activity by sending heartbeats
// from the offline queue to the WakaTime API.
func SyncOfflineActivity(ctx context.Context, v *viper.Viper, queueFilepath string) error {
	paramAPI, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(ctx, paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	paramOffline := params.LoadOfflineParams(ctx, v)

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, paramOffline.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: paramAPI.Key,
			MapPatterns:   paramAPI.KeyPatterns,
		}),
	)

	_, err = handle(ctx, nil)
	if err != nil {
		return err
	}

	logger := log.Extract(ctx)

	if err := cmdheartbeat.ResetRateLimit(ctx, v); err != nil {
		logger.Errorf("failed to reset rate limit: %s", err)
	}

	return nil
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
