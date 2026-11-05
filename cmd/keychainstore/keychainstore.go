package keychainstore

import (
	"context"
	"errors"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/ini"
	"github.com/wakatime/wakatime-cli/pkg/keychain"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/params"

	"github.com/spf13/viper"
)

// Run stores the api key in the OS keychain, then switches the config file over to it.
func Run(ctx context.Context, v *viper.Viper) (int, error) {
	apiKey, err := params.LoadConfigAPIKey(v)
	if err != nil {
		return exitcode.ErrAuth, fmt.Errorf("failed to load api key: %s", err)
	}

	if err := Store(apiKey); err != nil {
		return exitcode.ErrGeneric, fmt.Errorf("failed to store api key in keychain: %s", err)
	}

	w, err := ini.NewWriter(ctx, v, ini.FilePath)
	if err != nil {
		return exitcode.ErrConfigFileParse, fmt.Errorf("failed to write to config file: %s", err)
	}

	if err := UpdateConfig(ctx, v, w); err != nil {
		return exitcode.ErrConfigFileWrite, fmt.Errorf("failed to write to config file: %s", err)
	}

	log.Extract(ctx).Debugln("stored api key in keychain")

	return exitcode.Success, nil
}

// Store writes the api key to the OS keychain and reads it back to make sure
// it was saved, before anything is removed from the config file.
func Store(apiKey string) error {
	if err := keychain.WriteAPIKey(apiKey); err != nil {
		return err
	}

	stored, err := keychain.ReadAPIKey()
	if err != nil {
		return fmt.Errorf("failed to read back api key: %s", err)
	}

	if stored != apiKey {
		return errors.New("api key read back from keychain does not match")
	}

	return nil
}

// UpdateConfig sets api_key_storage to keychain and blanks any api key left in the
// [settings] section of the config file.
func UpdateConfig(ctx context.Context, v *viper.Viper, w ini.Writer) error {
	keyValue := map[string]string{
		"api_key_storage": keychain.StorageKeychain,
	}

	for _, key := range []string{"api_key", "apikey"} {
		if v.IsSet("settings." + key) {
			keyValue[key] = ""
		}
	}

	return w.Write(ctx, "settings", keyValue)
}
