package wakaerror

// Error is a custom error interface.
type Error interface {
	ExitCode() int
	Message() string
	error
}
