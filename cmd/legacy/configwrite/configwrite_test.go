package configwrite_test

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/ini.v1"
)

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

			var fwerr configwrite.ErrFileWrite
			assert.True(t, errors.As(err, &fwerr))
		})
	}
}

func TestWrite(t *testing.T) {
	v := viper.New()
	cfg := ini.Empty()
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v.SetConfigFile(tmpFile.Name())

	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "false"})

	err = configwrite.Write(v, cfg)
	require.NoError(t, err)
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
			cfg := ini.Empty()
			v.Set("config-section", test.Section)
			v.Set("config-write", test.Value)

			err := configwrite.Write(v, cfg)

			var fwerr configwrite.ErrFileWrite

			assert.True(t, errors.As(err, &fwerr))
		})
	}
}
