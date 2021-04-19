package logfile2

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/version"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	l "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// nolint
var LogEntry = l.NewEntry(new())

const defaultFile = ".wakatime.log"

// Params contains log file parameters.
type Params struct {
	File string
}

func new() *l.Logger {
	return &l.Logger{
		Out: os.Stdout,
		Formatter: &l.JSONFormatter{
			FieldMap: l.FieldMap{
				l.FieldKeyTime: "now",
				l.FieldKeyFile: "caller",
				l.FieldKeyFunc: "caller_func",
				l.FieldKeyMsg:  "message",
			},
			DisableHTMLEscape: true,
		},
		Level:        l.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: true,
	}
}

// SetOutput defines output log to a file.
func SetOutput(v *viper.Viper) {
	params, err := LoadParams(v)
	if err != nil {
		l.Fatalf("error loading log file params: %s", err)
	}

	f, err := os.OpenFile(params.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		l.Fatalf("error opening log file: %s", err)
	}

	LogEntry.Logger.Out = f
	LogEntry.Data["version"] = version.Version
}

// SetVerbose sets log level to debug if enabled.
func SetVerbose(v *viper.Viper) {
	var debug bool
	if b := v.GetBool("settings.debug"); v.IsSet("settings.debug") {
		debug = b
	}

	if v.GetBool("verbose") || debug {
		LogEntry.Logger.SetLevel(l.DebugLevel)
	}
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	logFile, _ := vipertools.FirstNonEmptyString(v, "log-file", "logfile", "settings.log_file")

	if logFile != "" {
		p, err := homedir.Expand(logFile)
		if err != nil {
			return Params{},
				ErrLogFile2(fmt.Sprintf("failed expading log file: %s", err))
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
				ErrLogFile2(fmt.Sprintf("failed parsing WAKATIME_HOME environment variable: %s", err))
		}
	} else {
		home, err = os.UserHomeDir()
		if err != nil {
			return Params{},
				ErrLogFile2(fmt.Sprintf("failed getting user's home directory: %s", err))
		}
	}

	return Params{
		File: filepath.Join(home, defaultFile),
	}, nil
}
