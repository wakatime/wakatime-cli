package config

// ErrFileParse handles a custom error while parsing wakatime config file.
type ErrFileParse string

// Error implements error interface.
func (e ErrFileParse) Error() string {
	return string(e)
}
