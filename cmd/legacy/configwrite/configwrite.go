package configwrite

import (
	"errors"
	"fmt"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

// Params contains config write parameters.
type Params struct {
	Section  string
	KeyValue map[string]string
}

// Run loads wakatime config file and call Write().
func Run(v *viper.Viper) (int, error) {
	w, err := config.NewIniWriter(v, config.FilePath)
	if err != nil {
		return exitcode.ErrConfigFileParse, fmt.Errorf(
			"failed to parse config file: %s",
			err,
		)
	}

	if err := Write(v, w); err != nil {
		return exitcode.ErrDefault, fmt.Errorf(
			"failed to write to config file: %s",
			err,
		)
	}

	return exitcode.Success, nil
}

// Write writes value(s) to given config key(s) and persist on disk.
func Write(v *viper.Viper, w config.Writer) error {
	params, err := LoadParams(v)
	if err != nil {
		return fmt.Errorf("failed loading params: %w", err)
	}

	if err := w.Write(params.Section, params.KeyValue); err != nil {
		return err
	}

	return nil
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	section := strings.TrimSpace(vipertools.GetString(v, "config-section"))
	kv := v.GetStringMapString("config-write")

	if section == "" || len(kv) == 0 {
		return Params{}, errors.New(
			"neither section nor key/value can be empty",
		)
	}

	return Params{
		Section:  section,
		KeyValue: kv,
	}, nil
}
