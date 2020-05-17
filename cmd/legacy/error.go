package legacy

// ErrConfigFileRead handles a custom error while reading wakatime config file.
type ErrConfigFileRead string

func (e ErrConfigFileRead) Error() string {
	return string(e)
}
