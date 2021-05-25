package offlinesync

import (
	"errors"
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyapi"
	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyparams"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/spf13/viper"
)

// Run executes the sync-offline-activity command.
func Run(v *viper.Viper) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		log.Fatalf("failed to load offline queue filepath: %s", err)
	}

	err = SyncOfflineActivity(v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			log.Errorf(
				"failed to sync offline activity: %s. Find your api key from wakatime.com/settings/api-key",
				errauth,
			)
			os.Exit(exitcode.ErrAuth)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			log.Errorf("failed to sync offline activity: %s", err)
			os.Exit(exitcode.ErrAPI)
		}

		log.Fatalf("failed to sync offline activity: %s", err)
	}

	log.Debugln("successfully synced offline activity")
	os.Exit(exitcode.Success)
}

// SyncOfflineActivity syncs offline activity by sending heartbeats
// from the offline queue to the WakaTime API.
func SyncOfflineActivity(v *viper.Viper, queueFilepath string) error {
	params, err := legacyparams.Load(v)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	if params.OfflineDisabled {
		return fmt.Errorf("sync offline is disabled. cannot sync offline activity: %w", err)
	}

	offlineHandleOpt, err := offline.WithQueue(queueFilepath, params.OfflineSyncMax)
	if err != nil {
		return fmt.Errorf("failed to initialize offline queue handle option: %w", err)
	}

	apiClient, err := legacyapi.NewClient(params.API)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	handle := heartbeat.NewHandle(apiClient, []heartbeat.HandleOption{offlineHandleOpt}...)

	_, err = handle(nil)
	if err != nil {
		return fmt.Errorf("failed to sync offline activity via api client: %w", err)
	}

	return nil
}
