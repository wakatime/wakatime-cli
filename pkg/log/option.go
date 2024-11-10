package log

// Option is a functional option for Logger.
type Option func(*Logger)

// WithVerbose sets verbose mode.
func WithVerbose(verbose bool) Option {
	return func(l *Logger) {
		l.SetVerbose(verbose)
	}
}

// WithMetrics sets metrics mode.
func WithMetrics(metrics bool) Option {
	return func(l *Logger) {
		l.metrics = metrics
	}
}

// WithSendDiagsOnErrors sets send diagnostics on errors mode.
func WithSendDiagsOnErrors(sendDiagsOnErrors bool) Option {
	return func(l *Logger) {
		l.sendDiagsOnErrors = sendDiagsOnErrors
	}
}
