package offline

import (
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/ini"
)

// dbLegacyFilename is the legacy bolt db filename.
const dbLegacyFilename = ".wakatime.bdb"

// QueueFilepathLegacy returns the path for offline queue db file. If
// the user's $HOME folder cannot be detected, it defaults to the
// current directory.
// This is used to support the old db file name and will be removed in the future.
func QueueFilepathLegacy() (string, error) {
	home, _, err := ini.WakaHomeDir()
	if err != nil {
		return dbFilename, fmt.Errorf("failed getting user's home directory, defaulting to current directory: %s", err)
	}

	return filepath.Join(home, dbLegacyFilename), nil
}
