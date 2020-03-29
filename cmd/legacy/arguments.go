package legacy

// Arguments Legacy structure
type Arguments struct {
	Entity                     string
	Key                        string
	Write                      bool
	Plugin                     string
	Time                       int64
	Lineo                      int32
	CursorPos                  int32
	EntityType                 string
	Category                   string
	Proxy                      string
	NoSslVerify                bool
	SslCertsFile               string
	Project                    string
	AlternateProject           string
	AlternateLanguage          string
	Language                   string
	LocalFile                  string
	Hostname                   string
	DisableOffline             bool
	HideFileNames              bool
	HiddenFileNames            []string
	HideProjectNames           bool
	HiddenProjectNames         []string
	Branch                     string
	HideBranchNames            bool
	HiddenBranchNames          []string
	Exclude                    []string
	ExcludeUnknownProject      bool
	Include                    []string
	IncludeOnlyWithProjectFile bool
	Ignore                     []string
	ExtraHeartbeats            bool
	LogFile                    string
	APIURL                     string
	Timeout                    int
	SyncOfflineActivity        int
	Today                      bool
	ConfigPath                 string
	ConfigSection              string
	ConfigRead                 string
	ConfigWrite                map[string]string
	Verbose                    bool
	Version                    bool
}
