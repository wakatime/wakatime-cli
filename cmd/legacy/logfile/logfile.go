package logfile

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

const defaultFile = ".wakatime.log"

// Params contains log file parameters.
type Params struct {
	File string
}

// Set define output log to a file.
func Set(v *viper.Viper) {
	params, err := LoadParams(v)
	if err != nil {
		jww.WARN.Printf("error loading log file params: %s", err)
	}

	f, err := os.OpenFile(params.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		jww.WARN.Printf("error opening log file: %s", err)
	}

	jww.SetStdoutOutput(f)
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	logFile := firstNonEmptyString(v, "log-file", "logfile", "settings.log_file")

	if logFile != "" {
		p, err := homedir.Expand(logFile)
		if err != nil {
			return Params{},
				ErrLogFile(fmt.Sprintf("failed expading log file: %s", err))
		}

		return Params{
			File: p,
		}, nil
	}

	var (
		home string
		err  error
	)

	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		home, err = homedir.Expand(home)
		if err != nil {
			return Params{},
				ErrLogFile(fmt.Sprintf("failed parsing WAKATIME_HOME environment variable: %s", err))
		}
	} else {
		home, err = os.UserHomeDir()
		if err != nil {
			return Params{},
				ErrLogFile(fmt.Sprintf("failed getting user's home directory: %s", err))
		}
	}

	return Params{
		File: filepath.Join(home, defaultFile),
	}, nil
}

// firstNonEmptyString accepts multiple keys and returns the first non empty string value
// it is able to retrieve from viper.Viper via these keys.
func firstNonEmptyString(v *viper.Viper, keys ...string) string {
	for _, key := range keys {
		if value := v.GetString(key); value != "" {
			return value
		}
	}

	return ""
}
