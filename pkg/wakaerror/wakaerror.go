package wakaerror

type (
	// Error is a custom error interface.
	Error interface {
		// ExitCode returns the exit code for the error.
		ExitCode() int
		// Message returns the error message.
		Message() string
		// SendDiagsOnErrors returns true when diagnostics should be sent on error.
		SendDiagsOnErrors() bool
		// ShouldLogError returns true when error should be logged.
		ShouldLogError() bool
		error
	}

	// LogLevel is a custom log level interface to return log level for error.
	LogLevel interface {
		// LogLevel returns the log level for the error.
		LogLevel() int8
	}
)
