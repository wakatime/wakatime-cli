package logfile

import (
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	defaultFile   = "wakatime.log"
	defaultFolder = ".wakatime"
)

// Params contains log file parameters.
type Params struct {
	File              string
	SendDiagsOnErrors bool
	ToStdout          bool
	Verbose           bool
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	params := Params{
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

	logFile, ok := vipertools.FirstNonEmptyString(v, "log-file", "logfile", "settings.log_file")
	if ok {
		p, err := homedir.Expand(logFile)
		if err != nil {
			return Params{}, fmt.Errorf("failed expanding log file: %s", err)
		}

		params.File = p

		return params, nil
	}

	folder, err := ini.WakaResourcesDir()
	if err != nil {
		return Params{}, fmt.Errorf("failed getting user's home directory: %s", err)
	}

	params.File = filepath.Join(folder, defaultFile)

	return params, nil
}
