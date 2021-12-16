package offlinesync

import (
	"errors"
	"fmt"

	apicmd "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/api"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/spf13/viper"
)

// Run executes the sync-offline-activity command.
func Run(v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: failed to load offline queue filepath: %s",
			err,
		)
	}

	err = SyncOfflineActivity(v, queueFilepath)
	if err != nil {
		var errauth api.ErrAuth
		if errors.As(err, &errauth) {
			return exitcode.ErrAuth, fmt.Errorf(
				"offline sync failed: invalid api key... find yours at wakatime.com/api-key. %s",
				errauth,
			)
		}

		var errbadRequest api.ErrBadRequest
		if errors.As(err, &errbadRequest) {
			return exitcode.ErrGeneric, fmt.Errorf(
				"offline sync failed: bad request: %s",
				err,
			)
		}

		var errapi api.Err
		if errors.As(err, &errapi) {
			return exitcode.ErrAPI, fmt.Errorf(
				"offline sync failed: api error: %s",
				err,
			)
		}

		return exitcode.ErrGeneric, fmt.Errorf(
			"offline sync failed: %s",
			err,
		)
	}

	log.Debugln("successfully synced offline activity")

	return exitcode.Success, nil
}

// SyncOfflineActivity syncs offline activity by sending heartbeats
// from the offline queue to the WakaTime API.
func SyncOfflineActivity(v *viper.Viper, queueFilepath string) error {
	p, err := params.Load(v, true)
	if err != nil {
		return fmt.Errorf("failed to load command parameters: %w", err)
	}

	apiClient, err := apicmd.NewClient(p.API)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

	if p.OfflineQueueFile != "" {
		queueFilepath = p.OfflineQueueFile
	}

	syncFn := offline.Sync(queueFilepath, p.OfflineSyncMax)

	return syncFn(apiClient.SendHeartbeats)
}
