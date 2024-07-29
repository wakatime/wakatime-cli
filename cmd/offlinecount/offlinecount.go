package offlinecount

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"

	"github.com/spf13/viper"
)

// Run executes the offline-count command.
func Run(v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath(v)
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"failed to load offline queue filepath: %s",
			err,
		)
	}

	count, err := offline.CountHeartbeats(queueFilepath)
	if err != nil {
		fmt.Println(err)
		return exitcode.ErrGeneric, fmt.Errorf("failed to count offline heartbeats: %w", err)
	}

	fmt.Println(count)

	return exitcode.Success, nil
}
