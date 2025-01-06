package logfile

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const defaultFile = "wakatime.log"

// Params contains log file parameters.
type Params struct {
	File              string
	Metrics           bool
	SendDiagsOnErrors bool
	ToStdout          bool
	Verbose           bool
}

// LoadParams loads needed data from the configuration file.
func LoadParams(ctx context.Context, v *viper.Viper) (Params, error) {
	params := Params{
		Metrics: vipertools.FirstNonEmptyBool(
			v,
			"metrics",
			"settings.metrics",
		),
		SendDiagsOnErrors: vipertools.FirstNonEmptyBool(
			v,
			"send-diagnostics-on-errors",
			"settings.send_diagnostics_on_errors",
		),
		ToStdout: v.GetBool("log-to-stdout"),
		Verbose: vipertools.FirstNonEmptyBool(
			v,
			"verbose",
			"settings.debug",
		),
	}

	logFile := vipertools.FirstNonEmptyString(v, "log-file", "logfile", "settings.log_file")
	if logFile != "" {
		p, err := homedir.Expand(logFile)
		if err != nil {
			return Params{}, fmt.Errorf("failed to expand log file: %s", err)
		}

		params.File = p

		return params, nil
	}

	folder, err := ini.WakaResourcesDir(ctx)
	if err != nil {
		return Params{}, fmt.Errorf("failed to get resource directory: %s", err)
	}

	params.File = filepath.Join(folder, defaultFile)

	return params, nil
}
