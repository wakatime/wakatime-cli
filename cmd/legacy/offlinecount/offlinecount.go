package offlinecount

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/cmd/legacy/legacyparams"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/spf13/viper"
)

// Run executes the offline-count command.
func Run(v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath()
	if err != nil {
		return exitcode.ErrDefault, fmt.Errorf(
			"failed to load offline queue filepath: %s",
			err,
		)
	}

	params, err := legacyparams.Load(v)
	if err != nil {
		return exitcode.ErrDefault, fmt.Errorf("failed to load command parameters: %w", err)
	}

	if params.OfflineQueueFile != "" {
		queueFilepath = params.OfflineQueueFile
	}

	count, err := offline.CountHeartbeats(queueFilepath)
	if err != nil {
		fmt.Println(err)
		return exitcode.ErrDefault, fmt.Errorf("failed to count offline heartbeats: %w", err)
	}

	fmt.Println(count)

	return exitcode.Success, nil
}
