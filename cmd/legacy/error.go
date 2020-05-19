package legacy

// ErrConfigFileRead handles a custom error while reading wakatime config file.
type ErrConfigFileRead string

func (e ErrConfigFileRead) Error() string {
	return string(e)
}

// ErrConfigFileParse handles a custom error while parsing wakatime config file.
type ErrConfigFileParse string

func (e ErrConfigFileParse) Error() string {
	return string(e)
}
