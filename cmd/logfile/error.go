package logfile

// ErrLogFile handles a custom error for setting log file.
type ErrLogFile string

// Error implements error interface.
func (e ErrLogFile) Error() string {
	return string(e)
}
