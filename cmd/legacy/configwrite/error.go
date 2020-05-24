package configwrite

// ErrFileWrite handles a custom error while writing to wakatime config file.
type ErrFileWrite string

// Error implements error interface.
func (e ErrFileWrite) Error() string {
	return string(e)
}
