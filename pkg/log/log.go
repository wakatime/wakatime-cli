package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/version"

	l "github.com/sirupsen/logrus"
	jww "github.com/spf13/jwalterweatherman"
)

// nolint:gochecknoglobals
var (
	logEntry = new()
	// Debugf logs a message at level Debug.
	Debugf = logEntry.Debugf
	// Infof logs a message at level Info.
	Infof = logEntry.Infof
	// Warnf logs a message at level Warn.
	Warnf = logEntry.Warnf
	// Errorf logs a message at level Error.
	Errorf = logEntry.Errorf
	// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
	Fatalf = logEntry.Fatalf
	// Debugln logs a message at level Debug.
	Debugln = logEntry.Debugln
	// Infoln logs a message at level Info.
	Infoln = logEntry.Infoln
	// Warnln logs a message at level Warn.
	Warnln = logEntry.Warnln
	// Errorln logs a message at level Error.
	Errorln = logEntry.Errorln
	// Fatalln logs a message at level Fatal then the process will exit with status set to 1.
	Fatalln = logEntry.Fatalln
)

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
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				// Simplifies function description by removing package name from it.
				lastSlash := strings.LastIndexByte(f.Function, '/')
				if lastSlash < 0 {
					lastSlash = 0
				}
				lastDot := strings.LastIndexByte(f.Function[lastSlash:], '.') + lastSlash

				return f.Function[lastDot+1:], fmt.Sprintf("%s:%d", f.File, f.Line)
			},
		},
		Level:        l.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: true,
	})
	entry.Data["version"] = version.Version

	return entry
}

// Output returns the current log output.
func Output() io.Writer {
	return logEntry.Logger.Out
}

// SetOutput defines sets the log output to io.Writer.
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

// SetJww sets jww log when debug enabled.
func SetJww(verbose bool, w io.Writer) {
	if verbose {
		jww.SetLogThreshold(jww.LevelDebug)
		jww.SetLogOutput(w)
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
