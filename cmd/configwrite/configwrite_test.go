package configwrite_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/wakatime/wakatime-cli/cmd/configwrite"
	"github.com/wakatime/wakatime-cli/pkg/ini"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Error(t, err)
		})
	}
}

func TestWrite(t *testing.T) {
	tmpFile, err := os.CreateTemp(t.TempDir(), "wakatime")
	require.NoError(t, err)

	defer tmpFile.Close()

	ctx := context.Background()

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

			err := configwrite.Write(context.Background(), v, w)
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

	err := configwrite.Write(context.Background(), v, w)
	assert.Error(t, err)
}

type mockWriter struct {
	WriteFn func(ctx context.Context, section string, keyValue map[string]string) error
}

func (m *mockWriter) Write(ctx context.Context, section string, keyValue map[string]string) error {
	return m.WriteFn(ctx, section, keyValue)
}
