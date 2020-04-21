package arguments

import (
	"strconv"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/configs"
)

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
	ObsoleteArgs        ObsoleteArguments
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
	LineNo    int32
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

// ObsoleteArguments ObsoleteArguments
type ObsoleteArguments struct {
	File           string //file
	HideFilenames1 bool   //hide-filenames
	HideFilenames2 bool   //hidefilenames
	LogFile        string //log-file
	APIURL         string //apiurl
}

// NewArguments NewArguments
func NewArguments() *Arguments {
	return &Arguments{
		Entity:       EntityArguments{},
		Editor:       EditorArguments{},
		Project:      ProjectArguments{},
		Obfuscate:    ObfuscateArguments{},
		Exclude:      ExcludeArguments{},
		Include:      IncludeArguments{},
		Config:       ConfigArguments{},
		Proxy:        ProxyArguments{},
		ObsoleteArgs: ObsoleteArguments{},
	}
}

// GetBooleanOrList Get a boolean or list of regexes from args and configs
func GetBooleanOrList(section string, key string, alternativeNames []string, cfg configs.WakaTimeConfig) []string {
	arr := []string{}

	//todo: implement alternative names

	values, err := cfg.Get(section, key)
	if err == nil {
		b, err := strconv.ParseBool(values)
		if err == nil {
			if b {
				arr = append(arr, ".*")
				return arr
			}

			arr = append(arr, "")
			return arr
		}

		// not bool? try parse list
		parts := strings.Split(values, "\n")

		for _, part := range parts {
			part = strings.TrimSpace(part)

			if len(part) > 0 {
				arr = append(arr, part)
			}
		}
	}

	return arr
}
