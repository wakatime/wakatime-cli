package setup

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/cmd/logfile"
	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logging uses the --log-file param to configure logging to file or stdout.
// It returns a logger with the configured settings or the default settings if it's not set.
func Logging(ctx context.Context, v *viper.Viper) (*log.Logger, error) {
	params, err := logfile.LoadParams(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("failed to load log params: %s", err)
	}

	var destOutput io.Writer = os.Stdout

	if !params.ToStdout {
		dir := filepath.Dir(params.File)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0750)
			if err != nil {
				return nil, fmt.Errorf("failed to create log file directory %q: %s", dir, err)
			}
		}

		// rotate log files
		destOutput = &lumberjack.Logger{
			Filename:   params.File,
			MaxSize:    log.MaxLogFileSize,
			MaxBackups: log.MaxNumberOfBackups,
		}
	}

	logger := log.New(
		destOutput,
		log.WithVerbose(params.Verbose),
		log.WithSendDiagsOnErrors(params.SendDiagsOnErrors),
		log.WithMetrics(params.Metrics),
	)

	return logger, nil
}
