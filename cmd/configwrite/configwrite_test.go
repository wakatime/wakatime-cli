package configwrite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/configwrite"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	configFile := filepath.Join(t.TempDir(), ".wakatime.cfg")

	v := viper.New()
	v.Set("config", configFile)
	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "true"})

	code, err := configwrite.Run(t.Context(), v)

	require.NoError(t, err)
	assert.Equal(t, exitcode.Success, code)

	contents, err := os.ReadFile(configFile)
	require.NoError(t, err)
	assert.Contains(t, string(contents), "debug")
	assert.Contains(t, string(contents), "true")
}

func TestRunErr(t *testing.T) {
	v := viper.New()
	v.Set("config", filepath.Join(t.TempDir(), ".wakatime.cfg"))
	v.Set("config-section", "settings")

	code, err := configwrite.Run(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrGeneric, code)
	assert.Contains(t, err.Error(), "failed to write to config file")
}

func TestRunNewWriterErr(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(parent, nil, 0600))

	v := viper.New()
	v.Set("config", filepath.Join(parent, ".wakatime.cfg"))
	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "true"})

	code, err := configwrite.Run(t.Context(), v)

	require.Error(t, err)
	assert.Equal(t, exitcode.ErrConfigFileParse, code)
	assert.Contains(t, err.Error(), "failed to write to config file")
}

func TestRun_ConfigPathError(t *testing.T) {
	v := viper.New()
	v.Set("config", "~missing-user/.wakatime.cfg")
	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "true"})

	code, err := configwrite.Run(t.Context(), v)

	require.ErrorContains(t, err, "failed to expand config param")
	assert.Equal(t, exitcode.ErrConfigFileParse, code)
}

func TestRun_XDG(t *testing.T) {
	tests := map[string]struct {
		CustomHome bool
		Existing   bool
		Legacy     bool
	}{
		"create default config": {},
		"create custom config":  {CustomHome: true},
		"update xdg config":     {CustomHome: true, Existing: true},
		"update legacy config":  {CustomHome: true, Existing: true, Legacy: true},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			t.Setenv("USERPROFILE", home)
			t.Setenv("WAKATIME_HOME", "")
			t.Setenv("XDG_CONFIG_HOME", "")

			xdgHome := filepath.Join(home, ".config")
			if test.CustomHome {
				xdgHome = filepath.Join(home, "custom-config")
				t.Setenv("XDG_CONFIG_HOME", xdgHome)
			}

			xdgFile := filepath.Join(xdgHome, "wakatime", "wakatime.cfg")
			original := []byte("[settings]\napi_key = keep-key\ndebug = false\n")

			if test.Existing {
				require.NoError(t, os.MkdirAll(filepath.Dir(xdgFile), 0750))
				require.NoError(t, os.Chmod(filepath.Dir(xdgFile), 0750))
				require.NoError(t, os.WriteFile(xdgFile, original, 0600))
			}

			expected := xdgFile

			legacy := filepath.Join(home, ".wakatime.cfg")
			if test.Legacy {
				require.NoError(t, os.WriteFile(legacy, original, 0600))
				expected = legacy
			}

			v := viper.New()
			v.Set("config-section", "settings")
			v.Set("config-write", map[string]string{"debug": "true"})

			code, err := configwrite.Run(t.Context(), v)
			require.NoError(t, err)
			assert.Equal(t, exitcode.Success, code)

			resolved, err := ini.FilePath(t.Context(), v)
			require.NoError(t, err)
			assert.Equal(t, expected, resolved)
			require.NoError(t, ini.ReadInConfig(v, resolved))
			assert.Equal(t, "true", v.GetString("settings.debug"))

			if test.Existing {
				assert.Equal(t, "keep-key", v.GetString("settings.api_key"))
			}

			if test.Legacy {
				contents, err := os.ReadFile(xdgFile)
				require.NoError(t, err)
				assert.Equal(t, original, contents)
			} else {
				assert.NoFileExists(t, legacy)
			}

			if runtime.GOOS != "windows" && !test.Legacy {
				dirInfo, err := os.Stat(filepath.Dir(xdgFile))
				require.NoError(t, err)

				expectedMode := os.FileMode(0700)
				if test.Existing {
					expectedMode = 0750
				}

				assert.Equal(t, expectedMode, dirInfo.Mode().Perm())

				fileInfo, err := os.Stat(xdgFile)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
			}
		})
	}
}

func TestLoadParams(t *testing.T) {
	tests := map[string]struct {
		Value   map[string]string
		Section string
	}{
		"single_keyvalue":          {map[string]string{"yo": "hi"}, "settings"},
		"double_value":             {map[string]string{"yo": "hi", "oh": "hi=there"}, "git"},
		"empty_value":              {map[string]string{"yo": ""}, "subversion"},
		"empty_value_double_value": {map[string]string{"yo": "", "oh": "hi=there"}, "default"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("config-section", test.Section)
			v.Set("config-write", test.Value)

			params, err := configwrite.LoadParams(v)
			require.NoError(t, err)

			assert.Equal(t, test.Section, params.Section)
			assert.Equal(t, test.Value, params.KeyValue)
		})
	}
}

func TestLoadParamsErr(t *testing.T) {
	tests := map[string]struct {
		Value   map[string]string
		Section string
	}{
		"section_missing": {
			Value: map[string]string{},
		},
		"key_missing": {
			Section: "settings",
		},
		"all_missing": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("config-section", test.Section)
			v.Set("config-write", test.Value)

			_, err := configwrite.LoadParams(v)
			assert.Error(t, err)
		})
	}
}

func TestWrite(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := t.Context()

	v := viper.New()
	ini, err := ini.NewWriter(ctx, v, func(_ context.Context, vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "false"})

	err = configwrite.Write(ctx, v, ini)
	require.NoError(t, err)

	err = ini.File.Reload()
	require.NoError(t, err)

	assert.Equal(t, "false", ini.File.Section("settings").Key("debug").String())
}

func TestWriteErr(t *testing.T) {
	tests := map[string]struct {
		Value   map[string]string
		Section string
	}{
		"empty_value": {
			Section: "settings",
			Value:   map[string]string{},
		},
		"section_missing": {
			Value: map[string]string{"debug": "false"},
		},
		"key_missing": {
			Section: "settings",
		},
		"all_missing": {},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			w := &ini.WriterConfig{}

			v.Set("config-section", test.Section)
			v.Set("config-write", test.Value)

			err := configwrite.Write(t.Context(), v, w)
			require.Error(t, err)

			assert.Equal(
				t,
				"failed to load command parameters: neither section nor key/value can be empty",
				err.Error(),
				fmt.Sprintf("error %q differs from the string set", err),
			)
		})
	}
}

func TestWriteSaveErr(t *testing.T) {
	v := viper.New()
	w := &mockWriter{
		WriteFn: func(_ context.Context, section string, keyValue map[string]string) error {
			assert.Equal(t, "settings", section)
			assert.Equal(t, map[string]string{"debug": "false"}, keyValue)

			return errors.New("error")
		},
	}

	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "false"})

	err := configwrite.Write(t.Context(), v, w)
	assert.Error(t, err)
}

type mockWriter struct {
	WriteFn func(ctx context.Context, section string, keyValue map[string]string) error
}

func (m *mockWriter) Write(ctx context.Context, section string, keyValue map[string]string) error {
	return m.WriteFn(ctx, section, keyValue)
}
