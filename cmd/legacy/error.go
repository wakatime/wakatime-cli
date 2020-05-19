package legacy

// ErrConfigFileParse handles a custom error while parsing wakatime config file.
type ErrConfigFileParse string

// Error implements error interface.
func (e ErrConfigFileParse) Error() string {
	return string(e)
}
