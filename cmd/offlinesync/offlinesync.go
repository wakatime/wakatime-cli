package offlinesync

import (
	"fmt"

	cmdapi "github.com/wakatime/wakatime-cli/cmd/api"
	"github.com/wakatime/wakatime-cli/cmd/params"
	"github.com/wakatime/wakatime-cli/pkg/apikey"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

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

// SyncOfflineActivity syncs offline activity by sending heartbeats
// from the offline queue to the WakaTime API.
func SyncOfflineActivity(v *viper.Viper, queueFilepath string) error {
	paramOffline := params.LoadOfflineParams(v)

	paramAPI, err := params.LoadAPIParams(v)
	if err != nil {
		return fmt.Errorf("failed to load API parameters: %w", err)
	}

	apiClient, err := cmdapi.NewClientWithoutAuth(paramAPI)
	if err != nil {
		return fmt.Errorf("failed to initialize api client: %w", err)
	}

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

	return nil
}
