package heartbeat

import (
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/utils"
)

// Heartbeat Heartbeat structure
type Heartbeat struct {
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
	ApiUrl                     string
	Timeout                    int
	SyncOfflineActivity        int
	ConfigPath                 string
	UserAgent                  string
	Verbose                    bool
	Skip                       string
}

// ValidateHeartbeat Heartbeat data for sending to API or storing in offline cache.
func ValidateHeartbeat(heartbeat Heartbeat, clone bool) *Heartbeat {
	heartbeat.UserAgent = utils.GetUserAgent(heartbeat.Plugin)

	if !utils.IsValidEntityType(heartbeat.EntityType) {
		heartbeat.EntityType = "file"
	}

	if !utils.IsValidCategory(heartbeat.Category) {
		heartbeat.Category = ""
	}

	if !clone {
		exclude := utils.ShouldExcludeByPattern(heartbeat.Entity, heartbeat.Include, heartbeat.Exclude)
		if exclude != nil {
			heartbeat.Skip = fmt.Sprintf("Skipping because matches exclude pattern: %s", exclude)
			return &heartbeat
		}

		if heartbeat.EntityType == "file" {
			heartbeat.Entity = utils.FormatFilePath(heartbeat.Entity)
			heartbeat.LocalFile = utils.FormatUncPath(heartbeat.Entity)

			if !utils.FileExists(heartbeat.Entity) {
				heartbeat.Skip = "File does not exist; ignoring this heartbeat."
			}

			if utils.IsExcludedByMissingProjectFile(heartbeat.Entity, heartbeat.IncludeOnlyWithProjectFile) {
				heartbeat.Skip = "Skipping because missing .wakatime-project file in parent path."
			}
		}

		if len(strings.TrimSpace(heartbeat.LocalFile)) > 0 && !utils.FileExists(heartbeat.LocalFile) {
			heartbeat.LocalFile = ""
		}

		//todo: continue here
	}

	return nil
}
