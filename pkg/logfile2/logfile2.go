package logfile2

import (
	"os"

	"github.com/wakatime/wakatime-cli/pkg/version"

	l "github.com/sirupsen/logrus"
)

// nolint
var LogEntry = new()

func new() *l.Entry {
	entry := l.NewEntry(&l.Logger{
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
	})
	entry.Data["version"] = version.Version

	return entry
}

// SetOutput defines output log to a file.
func SetOutput(filepath string) {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		l.Fatalf("error opening log file: %s", err)
	}

	LogEntry.Logger.Out = f
}

// SetVerbose sets log level to debug if enabled.
func SetVerbose(verbose bool) {
	if verbose {
		LogEntry.Logger.SetLevel(l.DebugLevel)
	}
}
