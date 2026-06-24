package offlinecount

import (
	"context"
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/offline"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"

	"github.com/spf13/viper"
)

// Run executes the offline-count command.
func Run(ctx context.Context, v *viper.Viper) (int, error) {
	queueFilepath, err := offline.QueueFilepath(ctx, v)
	if err != nil {
		return exitcode.ErrGeneric, fmt.Errorf(
			"failed to load offline queue filepath: %s",
			err,
		)
	}

	count, err := offline.CountHeartbeats(ctx, queueFilepath)
	if err != nil {
		fmt.Println(err)

		var errwaka wakaerror.Error
		if errors.As(err, &errwaka) {
			return errwaka.ExitCode(), fmt.Errorf("failed to count offline heartbeats: %w", err)
		}

		return exitcode.ErrGeneric, fmt.Errorf("failed to count offline heartbeats: %w", err)
	}

	fmt.Println(count)

	return exitcode.Success, nil
}
