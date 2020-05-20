package config

// ErrFileRead handles a custom error while reading wakatime config file.
type ErrFileRead string

// Error implements error interface.
func (e ErrFileRead) Error() string {
	return string(e)
}
