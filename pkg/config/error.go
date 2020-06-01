package config

// ErrFileParse handles a custom error while parsing wakatime config file.
type ErrFileParse string

// Error implements error interface.
func (e ErrFileParse) Error() string {
	return string(e)
}

// ErrFileWrite handles a custom error while writing to wakatime config file.
type ErrFileWrite string

// Error implements error interface.
func (e ErrFileWrite) Error() string {
	return string(e)
}
