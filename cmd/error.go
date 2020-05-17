package cmd

// ErrConfigFileParse handles a custom error while parsing wakatime config file.
type ErrConfigFileParse string

func (e ErrConfigFileParse) Error() string {
	return string(e)
}
