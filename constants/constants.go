package constants

const (
	// Success Exit code used when a heartbeat was sent successfully.
	Success int = 0
	// ApiError Exit code used when the WakaTime API returned an error.
	ApiError int = 102
	// ConfigFileParseError Exit code used when the ~/.wakatime.cfg config file could not be parsed.
	ConfigFileParseError int = 103
	// AuthError Exit code used when our api key is invalid.
	AuthError int = 104
	// UnknownError Exit code used when there was an unhandled exception.
	UnknownError int = 105
	//ConnectionError Exit code used when there was proxy or other problem connecting to the WakaTime API servers.
	ConnectionError int = 107

	// MaxFileSizeSupported Files larger than this in bytes will not have a line count stat for performance. Default is 2MB.
	MaxFileSizeSupported int = 2000000
	// DefaultSyncOfflineActivity Default limit of number of offline heartbeats to sync before exiting.
	DefaultSyncOfflineActivity int = 100
	// HeartbeatsPerRequest Even when sending more heartbeats, this is the number of heartbeats sent per individual https request to the WakaTime API.
	HeartbeatsPerRequest int = 25
	// DefaultTimeout Default timeout
	DefaultTimeout int = 60
)
