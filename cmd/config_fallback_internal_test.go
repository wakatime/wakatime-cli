package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfigFiles_XDG(t *testing.T) {
	tests := map[string]struct {
		LegacyFile bool
		Contents   string
		Directory  bool
		Expected   string
	}{
		"xdg config": {
			Expected: "xdg-key",
		},
		"legacy takes precedence": {
			LegacyFile: true,
			Contents:   "[settings]\napi_key = legacy-key\n",
			Expected:   "legacy-key",
		},
		"empty legacy takes precedence": {
			LegacyFile: true,
		},
		"legacy read error": {
			Directory: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			xdgHome := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("WAKATIME_HOME", "")
			t.Setenv("XDG_CONFIG_HOME", xdgHome)

			xdgFile := filepath.Join(xdgHome, "wakatime", "wakatime.cfg")
			require.NoError(t, os.MkdirAll(filepath.Dir(xdgFile), 0700))
			require.NoError(t, os.WriteFile(xdgFile, []byte("[settings]\napi_key = xdg-key\n"), 0600))

			legacy := filepath.Join(home, ".wakatime.cfg")
			if test.LegacyFile {
				require.NoError(t, os.WriteFile(legacy, []byte(test.Contents), 0600))
			}

			if test.Directory {
				require.NoError(t, os.Mkdir(legacy, 0700))
			}

			v := viper.New()

			err := parseConfigFiles(t.Context(), v)
			if test.Directory {
				require.ErrorContains(t, err, "failed to load configuration file")
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.Expected, v.GetString("settings.api_key"))
		})
	}
}
