package configwrite

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/spf13/viper"
)

// Params contains config write parameters.
type Params struct {
	Section  string
	KeyValue map[string]string
}

// Run loads wakatime config file and call Write().
func Run(v *viper.Viper) {
	w, err := config.NewIniWriter(v, config.FilePath)
	if err != nil {
		var cfperr config.ErrFileParse
		if errors.As(err, &cfperr) {
			log.Errorf(err.Error())
			os.Exit(exitcode.ErrConfigFileParse)
		}

		log.Fatalln(err)
	}

	if err := Write(v, w); err != nil {
		log.Errorf("failed to write on config file: %s", err)

		var cfwerr config.ErrFileWrite
		if errors.As(err, &cfwerr) {
			os.Exit(exitcode.ErrConfigFileWrite)
		}

		os.Exit(exitcode.ErrDefault)
	}

	os.Exit(exitcode.Success)
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
		return Params{},
			config.ErrFileWrite("neither section nor key/value can be empty")
	}

	return Params{
		Section:  section,
		KeyValue: kv,
	}, nil
}
