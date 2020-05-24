package configwrite

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/config"
	"github.com/wakatime/wakatime-cli/pkg/exitcode"

	jww "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// Params contains config write parameters.
type Params struct {
	Section  string
	KeyValue map[string]string
}

// Run loads wakatime config file and call Write().
func Run(v *viper.Viper) {
	cfg, err := config.LoadIni(v, config.FilePath)
	if err != nil {
		jww.FATAL.Fatalln(err)

		var cfperr config.ErrFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	if err := Write(v, cfg); err != nil {
		jww.ERROR.Println(err)

		var cfwerr ErrFileWrite
		if errors.As(err, &cfwerr) {
			os.Exit(exitcode.ErrConfigFileWrite)
		}

		os.Exit(exitcode.ErrDefault)
	}

	os.Exit(exitcode.Success)
}

// Write writes value(s) to given config key(s) and persist on disk.
func Write(v *viper.Viper, cfg *ini.File) error {
	params, err := LoadParams(v)
	if err != nil {
		return err
	}

	params.SetValue(cfg)

	if err := cfg.SaveTo(v.ConfigFileUsed()); err != nil {
		return ErrFileWrite(fmt.Errorf("error saving wakatime config: %s", err).Error())
	}

	return nil
}

// LoadParams loads needed data from the configuration file.
func LoadParams(v *viper.Viper) (Params, error) {
	section := strings.TrimSpace(v.GetString("config-section"))
	kv := v.GetStringMapString("config-write")

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key/value:", flattenKeyValue(kv))

	if section == "" || len(kv) == 0 {
		return Params{},
			ErrFileWrite(
				fmt.Errorf("failed to write on wakatime config file. neither section nor key/value can be empty").Error(),
			)
	}

	return Params{
		Section:  section,
		KeyValue: kv,
	}, nil
}

func flattenKeyValue(kv map[string]string) string {
	var ret []string
	for k, v := range kv {
		ret = append(ret, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.Join(ret, ",")
}

// SetValue sets the value for the key in the override register.
func (c *Params) SetValue(cfg *ini.File) {
	for key, value := range c.KeyValue {
		cfg.Section(c.Section).Key(key).SetValue(value)
	}
}
