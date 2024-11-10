package offline

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

// dbLegacyFilename is the legacy bolt db filename.
const dbLegacyFilename = ".wakatime.bdb"

// QueueFilepathLegacy returns the path for offline queue db file. If
// the user's $HOME folder cannot be detected, it defaults to the
// current directory.
// This is used to support the old db file name and will be removed in the future.
func QueueFilepathLegacy(ctx context.Context, v *viper.Viper) (string, error) {
	paramFile := vipertools.GetString(v, "offline-queue-file-legacy")
	if paramFile != "" {
		p, err := homedir.Expand(paramFile)
		if err != nil {
			return "", fmt.Errorf("failed expanding offline-queue-file-legacy param: %s", err)
		}

		return p, nil
	}

	home, _, err := ini.WakaHomeDir(ctx)
	if err != nil {
		return dbFilename, fmt.Errorf("failed getting user's home directory, defaulting to current directory: %s", err)
	}

	return filepath.Join(home, dbLegacyFilename), nil
}
