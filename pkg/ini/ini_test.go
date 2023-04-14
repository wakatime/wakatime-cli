package ini_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"
	iniv1 "gopkg.in/ini.v1"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadInConfig(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/wakatime.cfg")

	filePath, err := ini.FilePath(v)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, filePath)
	require.NoError(t, err)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", vipertools.GetString(v, "settings.api_key"))
	assert.Equal(t, "true", vipertools.GetString(v, "settings.debug"))
	assert.Equal(t, "us", vipertools.GetString(v, "other.country"))
	assert.Equal(t, "project-y", vipertools.GetString(v, "projectmap./some/path"))
	assert.Equal(t, "project-x", vipertools.GetString(v, "project_api_key./some/path"))
	assert.Equal(t, "project-2", vipertools.GetString(v, "project_api_key./other/tmp/path"))
}

func TestReadInConfig_Multiline(t *testing.T) {
	multilineOption := viper.IniLoadOptions(iniv1.LoadOptions{AllowPythonMultilineValues: true})
	v := viper.NewWithOptions(multilineOption)

	v.Set("config", "testdata/wakatime-multiline.cfg")

	filePath, err := ini.FilePath(v)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, filePath)
	require.NoError(t, err)

	ignoreConfig := strings.ReplaceAll(vipertools.GetString(v, "settings.ignore"), "\r", "")
	assert.Equal(t, "\n  COMMIT_EDITMSG$\n  PULLREQ_EDITMSG$\n  MERGE_MSG$\n  TAG_EDITMSG$", ignoreConfig)

	gitConfig := strings.ReplaceAll(vipertools.GetString(v, "git.submodules_disabled"), "\r", "")
	assert.Equal(t, "\n  .*secret.*\n  fix.*", gitConfig)
}

func TestReadInConfig_Multiple(t *testing.T) {
	v := viper.New()

	v.Set("config", "testdata/wakatime.cfg")
	v.Set("internal-config", "testdata/wakatime-internal.cfg")

	filePath, err := ini.FilePath(v)
	require.NoError(t, err)

	internalFilePath, err := ini.InternalFilePath(v)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, filePath)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, internalFilePath)
	require.NoError(t, err)

	assert.Equal(t, "b9485572-74bf-419a-916b-22056ca3a24c", vipertools.GetString(v, "settings.api_key"))
	assert.Equal(t, "true", vipertools.GetString(v, "settings.debug"))
	assert.Equal(t, "us", vipertools.GetString(v, "other.country"))
	assert.Equal(t, "2021-11-25T12:17:21-07:00", vipertools.GetString(v, "internal.backoff_at"))
	assert.Equal(t, "3", vipertools.GetString(v, "internal.backoff_retries"))
}

func TestReadInConfig_Corrupted(t *testing.T) {
	iniOption := viper.IniLoadOptions(iniv1.LoadOptions{SkipUnrecognizableLines: true})
	v := viper.NewWithOptions(iniOption)

	v.Set("config", "testdata/corrupted.cfg")

	filePath, err := ini.FilePath(v)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, filePath)
	require.NoError(t, err)

	assert.Empty(t, vipertools.GetString(v, "internal.backoff_at"))
	assert.Equal(t, "0", vipertools.GetString(v, "internal.backoff_retries"))
	assert.Equal(t, "v1.45.3", vipertools.GetString(v, "internal.cli_version"))
	assert.Equal(t, "Mon, 16 May 2022 21:32:42 GMT", vipertools.GetString(v, "internal.cli_version_last_modified"))
}

func TestReadInConfig_Missing(t *testing.T) {
	v := viper.New()

	err := ini.ReadInConfig(v, "not-exists")

	assert.Error(t, err, "error parsing config file: open not-exists: no such file or directory")
}

func TestReadInConfig_Malformed(t *testing.T) {
	v := viper.New()
	v.Set("config", "testdata/malformed.cfg")

	filePath, err := ini.FilePath(v)
	require.NoError(t, err)

	err = ini.ReadInConfig(v, filePath)
	require.Error(t, err)
}

func TestFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperValue string
		EnvVar     string
		Expected   string
	}{
		"default": {
			Expected: filepath.Join(home, ".wakatime.cfg"),
		},
		"viper": {
			ViperValue: "~/path/.wakatime.cfg",
			Expected:   filepath.Join(home, "path", ".wakatime.cfg"),
		},
		"surrounding double quotes": {
			ViperValue: `"~/path/.wakatime.cfg"`,
			Expected:   filepath.Join(home, "path", ".wakatime.cfg"),
		},
		"env trailing slash": {
			EnvVar:   "~/path2/",
			Expected: filepath.Join(home, "path2", ".wakatime.cfg"),
		},
		"env without trailing slash": {
			EnvVar:   "~/path2",
			Expected: filepath.Join(home, "path2", ".wakatime.cfg"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("config", test.ViperValue)
			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			defer os.Unsetenv("WAKATIME_HOME")

			configFilepath, err := ini.FilePath(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, configFilepath)
		})
	}
}

func TestInternalFilePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := map[string]struct {
		ViperValue string
		EnvVar     string
		Expected   string
	}{
		"default": {
			Expected: filepath.Join(home, ".wakatime", "wakatime-internal.cfg"),
		},
		"env trailing slash": {
			EnvVar:   "~/path2/",
			Expected: filepath.Join(home, "path2", "wakatime-internal.cfg"),
		},
		"env without trailing slash": {
			EnvVar:   "~/path2",
			Expected: filepath.Join(home, "path2", "wakatime-internal.cfg"),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			v := viper.New()
			v.Set("internal-config", test.ViperValue)

			err := os.Setenv("WAKATIME_HOME", test.EnvVar)
			require.NoError(t, err)

			defer os.Unsetenv("WAKATIME_HOME")

			configFilepath, err := ini.InternalFilePath(v)
			require.NoError(t, err)

			assert.Equal(t, test.Expected, configFilepath)
		})
	}
}

func TestNewWriter(t *testing.T) {
	v := viper.New()

	w, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "testdata/wakatime.cfg", nil
	})
	require.NoError(t, err)

	assert.Equal(t, "testdata/wakatime.cfg", w.ConfigFilepath)
	assert.NotNil(t, w.File)
}

func TestNewWriterErr(t *testing.T) {
	v := viper.New()

	_, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "", errors.New("error")
	})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error getting filepath")
}

func TestNewWriter_MissingFile(t *testing.T) {
	v := viper.New()

	tmpDir := t.TempDir()

	w, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return filepath.Join(tmpDir, "missing.cfg"), nil
	})
	require.NoError(t, err)

	assert.FileExists(t, w.ConfigFilepath)

	assert.Equal(t, filepath.Join(tmpDir, "missing.cfg"), w.ConfigFilepath)
	assert.NotNil(t, w.File)
}

func TestNewWriter_CorruptedFile(t *testing.T) {
	v := viper.New()

	w, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return "testdata/corrupted.cfg", nil
	})
	require.NoError(t, err)

	assert.Equal(t, "testdata/corrupted.cfg", w.ConfigFilepath)
	assert.NotNil(t, w.File)
}

func TestWrite(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	tests := map[string]struct {
		Value   map[string]string
		Section string
	}{
		"single_value": {
			Value: map[string]string{
				"debug": "true",
			},
			Section: "settings",
		},
		"double_value": {
			Value: map[string]string{
				"debug":   "true",
				"api_key": "b9485572-74bf-419a-916b-22056ca3a24c",
			},
			Section: "settings",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := ini.WriterConfig{
				File:           iniv1.Empty(),
				ConfigFilepath: tmpFile.Name(),
			}

			err := w.Write(test.Section, test.Value)

			require.NoError(t, err)
		})
	}
}

func TestWrite_NoMultilineSideEffects(t *testing.T) {
	multilineOption := viper.IniLoadOptions(iniv1.LoadOptions{AllowPythonMultilineValues: true})
	v := viper.NewWithOptions(multilineOption)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("config", tmpFile.Name())

	copyFile(t, "testdata/wakatime-multiline.cfg", tmpFile.Name())

	w, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	err = w.Write("settings", map[string]string{"debug": "true"})
	require.NoError(t, err)

	actual, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/wakatime-multiline-expected.cfg")
	require.NoError(t, err)

	assert.Equal(t,
		strings.ReplaceAll(string(expected), "\r", ""),
		strings.ReplaceAll(string(actual), "\r", ""))
}

func TestWrite_NullsRemoved(t *testing.T) {
	multilineOption := viper.IniLoadOptions(iniv1.LoadOptions{AllowPythonMultilineValues: true})
	v := viper.NewWithOptions(multilineOption)

	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	v.Set("config", tmpFile.Name())

	copyFile(t, "testdata/wakatime-nulls.cfg", tmpFile.Name())

	w, err := ini.NewWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	err = w.Write("settings", map[string]string{"debug": "true"})
	require.NoError(t, err)

	actual, err := os.ReadFile(tmpFile.Name())
	require.NoError(t, err)

	expected, err := os.ReadFile("testdata/wakatime-nulls-expected.cfg")
	require.NoError(t, err)

	assert.Equal(t,
		strings.ReplaceAll(string(expected), "\r", ""),
		strings.ReplaceAll(string(actual), "\r", ""))
}

func TestWriteErr(t *testing.T) {
	w := ini.WriterConfig{}

	err := w.Write("settings", map[string]string{"debug": "true"})
	require.Error(t, err)

	assert.Equal(t, "got undefined wakatime config file instance", err.Error())
}

func copyFile(t *testing.T, source, destination string) {
	input, err := os.ReadFile(source)
	require.NoError(t, err)

	err = os.WriteFile(destination, input, 0600)
	require.NoError(t, err)
}
