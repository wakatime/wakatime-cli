package wakaerror

// Error is a custom error interface.
type Error interface {
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
