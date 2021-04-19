package log

import (
	"io"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/version"

	l "github.com/sirupsen/logrus"
)

// nolint
var logEntry = new()

func new() *l.Entry {
	entry := l.NewEntry(&l.Logger{
		Out: os.Stdout,
		Formatter: &l.JSONFormatter{
			FieldMap: l.FieldMap{
				l.FieldKeyTime: "now",
				l.FieldKeyFile: "caller",
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
func SetOutput(w io.Writer) {
	logEntry.Logger.Out = w
}

// SetVerbose sets log level to debug if enabled.
func SetVerbose(verbose bool) {
	if verbose {
		logEntry.Logger.SetLevel(l.DebugLevel)
	} else {
		logEntry.Logger.SetLevel(l.InfoLevel)
	}
}

// WithField adds a single field to the Entry.
func WithField(key string, value interface{}) {
	logEntry.WithField(key, value)
}

// WithFields adds a map of fields to the Entry.
func WithFields(fields map[string]interface{}) {
	logEntry.WithFields(fields)
}

// Debugf logs a message at level Debug.
func Debugf(format string, args ...interface{}) {
	logEntry.Debugf(format, args...)
}

// Infof logs a message at level Info.
func Infof(format string, args ...interface{}) {
	logEntry.Infof(format, args...)
}

// Warnf logs a message at level Warn.
func Warnf(format string, args ...interface{}) {
	logEntry.Warnf(format, args...)
}

// Errorf logs a message at level Error.
func Errorf(format string, args ...interface{}) {
	logEntry.Errorf(format, args...)
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func Fatalf(format string, args ...interface{}) {
	logEntry.Fatalf(format, args...)
}

// Debugln logs a message at level Debug.
func Debugln(args ...interface{}) {
	logEntry.Debugln(args...)
}

// Infoln logs a message at level Info.
func Infoln(args ...interface{}) {
	logEntry.Infoln(args...)
}

// Warnln logs a message at level Warn.
func Warnln(args ...interface{}) {
	logEntry.Warnln(args...)
}

// Errorln logs a message at level Error.
func Errorln(args ...interface{}) {
	logEntry.Errorln(args...)
}

// Fatalln logs a message at level Fatal then the process will exit with status set to 1.
func Fatalln(args ...interface{}) {
	logEntry.Fatalln(args...)
}
