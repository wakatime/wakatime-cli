package logfile

import (
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const defaultFile = ".wakatime.log"

// Params contains log file parameters.
type Params struct {
	File     string
	ToStdout bool
	Verbose  bool
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	var debug bool
	if b := v.GetBool("settings.debug"); v.IsSet("settings.debug") {
		debug = b
	}

	params := Params{
		ToStdout: v.GetBool("log-to-stdout"),
		Verbose:  v.GetBool("verbose") || debug,
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

	home, err := ini.WakaHomeDir()
	if err != nil {
		return Params{}, fmt.Errorf("failed getting user's home directory: %s", err)
	}

	params.File = filepath.Join(home, defaultFile)

	return params, nil
}
