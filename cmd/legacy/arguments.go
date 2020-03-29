package legacy

// Arguments Legacy structure
type Arguments struct {
	Key                 string
	Time                int64
	Category            string
	AlternateLanguage   string
	Language            string
	Hostname            string
	DisableOffline      bool
	ExtraHeartbeats     bool
	LogFile             string
	APIURL              string
	Timeout             int
	SyncOfflineActivity int
	Today               bool
	Verbose             bool
	Version             bool
	Entity              EntityArguments
	Editor              EditorArguments
	Project             ProjectArguments
	Obfuscate           ObfuscateArguments
	Exclude             ExcludeArguments
	Include             IncludeArguments
	Config              ConfigArguments
	Proxy               ProxyArguments
}

// EntityArguments EntityArguments
type EntityArguments struct {
	Entity    string
	LocalFile string
	IsWrite   bool
	Type      string
}

// ExcludeArguments ExcludeArguments
type ExcludeArguments struct {
	Exclude               []string
	ExcludeUnknownProject bool
	Ignore                []string
}

// IncludeArguments IncludeArguments
type IncludeArguments struct {
	Include                    []string
	IncludeOnlyWithProjectFile bool
}

// EditorArguments EditorArguments
type EditorArguments struct {
	Plugin    string
	Lineo     int32
	CursorPos int32
}

// ProjectArguments ProjectArguments
type ProjectArguments struct {
	Name             string
	AlternateProject string
	Branch           string
}

// ProxyArguments ProxyArguments
type ProxyArguments struct {
	Address      string
	NoSslVerify  bool
	SslCertsFile string
}

// ObfuscateArguments ObfuscateArguments
type ObfuscateArguments struct {
	HideFileNames      bool
	HiddenFileNames    []string
	HideProjectNames   bool
	HiddenProjectNames []string
	HideBranchNames    bool
	HiddenBranchNames  []string
}

// ConfigArguments ConfigArguments
type ConfigArguments struct {
	Path    string
	Section string
	Read    string
	Write   map[string]string
}

// NewArguments NewArguments
func NewArguments() *Arguments {
	return &Arguments{
		Config:    ConfigArguments{},
		Obfuscate: ObfuscateArguments{},
	}
}
