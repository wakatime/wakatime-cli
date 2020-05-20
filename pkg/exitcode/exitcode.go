package exitcode

const (
	// Success is used when a heartbeat was sent successfully.
	Success = 0
	// ErrDefault is used for general errors.
	ErrDefault = 1
	// ErrConfigFileParse is used when the ~/.wakatime.cfg config file could not be parsed.
	ErrConfigFileParse = 103
	// ErrConfigFileRead is used for errors of config read command.
	ErrConfigFileRead = 110
)
