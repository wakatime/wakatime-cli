package offlinesync

import (
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
func RunWithoutRateLimiting(v *viper.Viper) (int, error) {
	return run(v)
}

// RunWithRateLimiting executes sync-offline-activity command with rate limiting enabled.
func RunWithRateLimiting(v *viper.Viper) (int, error) {
	paramOffline := params.LoadOfflineParams(v)

	if cmdheartbeat.RateLimited(cmdheartbeat.RateLimitParams{
		Disabled:   paramOffline.Disabled,
		LastSentAt: paramOffline.LastSentAt,
		Timeout:    paramOffline.RateLimit,
	}) {
		log.Debugln("skip syncing offline activity to respect rate limit")
		return exitcode.Success, nil
	}

	return run(v)
}

func run(v *viper.Viper) (int, error) {
	paramOffline := params.LoadOfflineParams(v)
	if paramOffline.Disabled {
		return exitcode.Success, nil
	}

	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: failed to load offline queue filepath: %s",
			err,
		)
	}

	queueFilepathLegacy, err := offline.QueueFilepathLegacy()
	if err != nil {
		log.Warnf("legacy offline sync failed: failed to load offline queue filepath: %s", err)
	}

	if err = syncOfflineActivityLegacy(v, queueFilepathLegacy); err != nil {
		log.Warnf("legacy offline sync failed: %s", err)
	}

	if err = SyncOfflineActivity(v, queueFilepath); err != nil {
		if errwaka, ok := err.(wakaerror.Error); ok {
			return errwaka.ExitCode(), fmt.Errorf("offline sync failed: %s", errwaka.Message())
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: %s",
			err,
		)
	}

	log.Debugln("successfully synced offline activity")

	return exitcode.Success, nil
}

// syncOfflineActivityLegacy syncs the old offline activity by sending heartbeats
// from the legacy offline queue to the WakaTime API.
func syncOfflineActivityLegacy(v *viper.Viper, queueFilepath string) error {
	if queueFilepath == "" {
		return nil
	}

	paramOffline := params.LoadOfflineParams(v)

	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	if paramOffline.QueueFileLegacy != "" {
		queueFilepath = paramOffline.QueueFileLegacy
	}

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, paramOffline.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: paramAPI.Key,
			MapPatterns:   paramAPI.KeyPatterns,
		}),
	)

	_, err = handle(nil)
	if err != nil {
		return err
	}

	if err := os.Remove(queueFilepath); err != nil {
		log.Warnf("failed to delete legacy offline file: %s", err)
	}

	return nil
}

// SyncOfflineActivity syncs offline activity by sending heartbeats
// from the offline queue to the WakaTime API.
func SyncOfflineActivity(v *viper.Viper, queueFilepath string) error {
	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	paramOffline := params.LoadOfflineParams(v)

	if paramOffline.QueueFile != "" {
		queueFilepath = paramOffline.QueueFile
	}

	handle := heartbeat.NewHandle(apiClient,
		offline.WithSync(queueFilepath, paramOffline.SyncMax),
		apikey.WithReplacing(apikey.Config{
			DefaultAPIKey: paramAPI.Key,
			MapPatterns:   paramAPI.KeyPatterns,
		}),
	)

	_, err = handle(nil)
	if err != nil {
		return err
	}

	if err := cmdheartbeat.ResetRateLimit(v); err != nil {
		log.Errorf("failed to reset rate limit: %s", err)
	}

	return nil
}
