package log

import (
	"fmt"
	"io"

	jww "github.com/spf13/jwalterweatherman"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/wakatime/wakatime-cli/pkg/version"
)

const (
	// MaxLogFileSize is the maximum size of the log file.
	MaxLogFileSize = 25 // 25MB
	// MaxNumberOfBackups is the maximum number of log file backups.
	MaxNumberOfBackups = 4
)

// Logger is the log entry.
type Logger struct {
	entry              *zap.Logger
	atomicLevel        zap.AtomicLevel
	currentOutput      io.Writer
	dynamicWriteSyncer *DynamicWriteSyncer
	metrics            bool
	sendDiagsOnErrors  bool
	verbose            bool
}

// New creates a new Logger that writes to dest.
func New(dest io.Writer, opts ...Option) *Logger {
	atom := zap.NewAtomicLevel()
	dynamicWriteSyncer := NewDynamicWriteSyncer(zapcore.AddSync(dest))

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "now"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.MessageKey = "message"
	encoderCfg.FunctionKey = "func"

	l := zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		dynamicWriteSyncer,
		atom,
	),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.FatalLevel),
	)

	l = l.With(
		zap.String("version", version.Version),
		zap.String("os/arch", fmt.Sprintf("%s/%s", version.OS, version.Arch)),
	)

	logger := &Logger{
		entry:              l,
		atomicLevel:        atom,
		currentOutput:      dest,
		dynamicWriteSyncer: dynamicWriteSyncer,
	}

	for _, option := range opts {
		option(logger)
	}

	return logger
}

// IsMetricsEnabled returns true if it should collect metrics.
func (l *Logger) IsMetricsEnabled() bool {
	return l.metrics
}

// IsVerboseEnabled returns true if debug is enabled.
func (l *Logger) IsVerboseEnabled() bool {
	return l.verbose
}

// Output returns the current log output.
func (l *Logger) Output() io.Writer {
	return l.currentOutput
}

// SendDiagsOnErrors returns true if diagnostics should be sent on errors.
func (l *Logger) SendDiagsOnErrors() bool {
	return l.sendDiagsOnErrors
}

// SetOutput defines sets the log output to io.Writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.currentOutput = w
	l.dynamicWriteSyncer.SetWriter(zapcore.AddSync(w))
}

// SetVerbose sets log level to debug if enabled.
func (l *Logger) SetVerbose(verbose bool) {
	l.verbose = verbose

	if verbose {
		l.atomicLevel.SetLevel(zap.DebugLevel)
	} else {
		l.atomicLevel.SetLevel(zap.InfoLevel)
	}
}

// Flush flushes the log output and closes the file.
func (l *Logger) Flush() {
	if err := l.entry.Sync(); err != nil {
		l.Debugf("failed to flush log file: %s", err)
	}

	if closer, ok := l.currentOutput.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			l.Debugf("failed to close log file: %s", err)
		}
	}
}

// SetJww sets jww log when debug enabled.
func SetJww(verbose bool, w io.Writer) {
	if verbose {
		jww.SetLogThreshold(jww.LevelDebug)
		jww.SetStdoutThreshold(jww.LevelDebug)

		jww.SetLogOutput(w)
		jww.SetStdoutOutput(w)
	}
}

// Debugf logs a message at level Debug.
func (l *Logger) Debugf(format string, args ...any) {
	l.entry.Log(zapcore.DebugLevel, fmt.Sprintf(format, args...))
}

// Infof logs a message at level Info.
func (l *Logger) Infof(format string, args ...any) {
	l.entry.Log(zapcore.InfoLevel, fmt.Sprintf(format, args...))
}

// Warnf logs a message at level Warn.
func (l *Logger) Warnf(format string, args ...any) {
	l.entry.Log(zapcore.WarnLevel, fmt.Sprintf(format, args...))
}

// Errorf logs a message at level Error.
func (l *Logger) Errorf(format string, args ...any) {
	l.entry.Log(zapcore.ErrorLevel, fmt.Sprintf(format, args...))
}

// Fatalf logs a message at level Fatal then the process will exit with status set to 1.
func (l *Logger) Fatalf(format string, args ...any) {
	l.entry.Log(zapcore.FatalLevel, fmt.Sprintf(format, args...))
}

// Debugln logs a message at level Debug.
func (l *Logger) Debugln(msg string) {
	l.entry.Log(zapcore.DebugLevel, msg)
}

// Infoln logs a message at level Info.
func (l *Logger) Infoln(msg string) {
	l.entry.Log(zapcore.InfoLevel, msg)
}

// Warnln logs a message at level Warn.
func (l *Logger) Warnln(msg string) {
	l.entry.Log(zapcore.WarnLevel, msg)
}

// Errorln logs a message at level Error.
func (l *Logger) Errorln(msg string) {
	l.entry.Log(zapcore.ErrorLevel, msg)
}

// Fatalln logs a message at level Fatal then the process will exit with status set to 1.
func (l *Logger) Fatalln(msg string) {
	l.entry.Log(zapcore.FatalLevel, msg)
}

// WithField adds a single field to the Logger.
func (l *Logger) WithField(key string, value any) {
	l.entry = l.entry.With(zap.Any(key, value))
}
