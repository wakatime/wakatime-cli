package offlinesync

import (
	"context"
	"fmt"
	"os"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/params"
	"github.com/wakatime/wakatime-cli/pkg/ratelimit"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// RunWithoutRateLimiting executes the sync-offline-activity command without rate limiting.
func RunWithoutRateLimiting(ctx context.Context, v *viper.Viper) (int, error) {
	return run(ctx, v)
}

// RunWithRateLimiting executes sync-offline-activity command with rate limiting enabled.
func RunWithRateLimiting(ctx context.Context, v *viper.Viper) (int, error) {
	offlineParams := params.LoadOfflineParams(ctx, v)

	logger := log.Extract(ctx)

	if ratelimit.IsRateLimited(ratelimit.Params{
		Disabled:   offlineParams.Disabled,
		LastSentAt: offlineParams.LastSentAt,
		Timeout:    offlineParams.RateLimit,
	}) {
		logger.Debugln("skip syncing offline activity to respect rate limit")

		return exitcode.Success, nil
	}

	return run(ctx, v)
}

func run(ctx context.Context, v *viper.Viper) (int, error) {
	offlineParams := params.LoadOfflineParams(ctx, v)
	if offlineParams.Disabled {
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

	defer func() {
		if err := os.Remove(queueFilepath); err != nil {
			logger := log.Extract(ctx)
			logger.Warnf("failed to delete legacy offline file: %s", err)
		}
	}()

	offlineParams := params.LoadOfflineParams(ctx, v)

	apiParams, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(ctx, apiParams)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, offlineParams.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: apiParams.Key,
			MapPatterns:   apiParams.KeyPatterns,
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
	offlineParams := params.LoadOfflineParams(ctx, v)

	apiParams, err := params.LoadAPIParams(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(ctx, apiParams)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, offlineParams.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: apiParams.Key,
			MapPatterns:   apiParams.KeyPatterns,
		}),
	)

	_, err = handle(ctx, nil)
	if err != nil {
		return err
	}

	if err := ratelimit.Reset(ctx, v); err != nil {
		logger := log.Extract(ctx)
		logger.Errorf("failed to reset rate limit: %s", err)
	}

	return nil
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)

	return err == nil || os.IsExist(err)
}
