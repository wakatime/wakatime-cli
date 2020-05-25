package configwrite_test

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/legacy/configwrite"
	"github.com/wakatime/wakatime-cli/pkg/config"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type writerMock struct {
	WriteFn func(section string, keyValue map[string]string) error
}

func (m *writerMock) Write(section string, keyValue map[string]string) error {
	return m.WriteFn(section, keyValue)
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

			var fwerr config.ErrFileWrite
			assert.True(t, errors.As(err, &fwerr))
		})
	}
}

func TestWrite(t *testing.T) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "wakatime")
	require.NoError(t, err)

	defer os.Remove(tmpFile.Name())

	v := viper.New()
	ini, err := config.NewIniWriter(v, func(vp *viper.Viper) (string, error) {
		assert.Equal(t, v, vp)
		return tmpFile.Name(), nil
	})
	require.NoError(t, err)

	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "false"})

	err = configwrite.Write(v, ini)
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
			w := &config.IniWriter{}

			v.Set("config-section", test.Section)
			v.Set("config-write", test.Value)

			err := configwrite.Write(v, w)

			var fwerr config.ErrFileWrite
			errMsg := fmt.Sprintf("error %q differs from the string set", err)

			assert.True(t, errors.As(err, &fwerr))
			assert.Equal(
				t,
				err.Error(),
				"failed to write on wakatime config file. neither section nor key/value can be empty",
				errMsg,
			)
		})
	}
}

func TestWriteSaveErr(t *testing.T) {
	v := viper.New()
	w := &writerMock{
		WriteFn: func(section string, keyValue map[string]string) error {
			assert.Equal(t, "settings", section)
			assert.Equal(t, map[string]string{"debug": "false"}, keyValue)
			return config.ErrFileWrite("error")
		},
	}

	v.Set("config-section", "settings")
	v.Set("config-write", map[string]string{"debug": "false"})

	err := configwrite.Write(v, w)

	var fwerr config.ErrFileWrite

	assert.True(t, errors.As(err, &fwerr))
}
