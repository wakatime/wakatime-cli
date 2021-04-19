package logfile2

// ErrLogFile2 handles a custom error for setting log file.
type ErrLogFile2 string

// Error implements error interface.
func (e ErrLogFile2) Error() string {
	return string(e)
}
