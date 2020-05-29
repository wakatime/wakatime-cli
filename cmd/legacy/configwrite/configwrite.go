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
		jww.FATAL.Fatalln(err)

		var cfperr config.ErrFileParse
		if errors.As(err, &cfperr) {
			os.Exit(exitcode.ErrConfigFileParse)
		}

		os.Exit(exitcode.ErrDefault)
	}

	if err := Write(v, w); err != nil {
		jww.FATAL.Printf("failed to write on config file: %s", err)

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
	section := strings.TrimSpace(v.GetString("config-section"))
	kv := v.GetStringMapString("config-write")

	jww.DEBUG.Println("section:", section)
	jww.DEBUG.Println("key/value:", flattenKeyValue(kv))

	if section == "" || len(kv) == 0 {
		return Params{},
			config.ErrFileWrite(
				fmt.Sprintf("neither section nor key/value can be empty"),
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
