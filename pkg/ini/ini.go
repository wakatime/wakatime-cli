package ini

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/vipertools"

	"github.com/juju/mutex"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

// WakaHomeType is WakaTime home type.
type WakaHomeType int

const (
	defaultFolder = ".wakatime"
	// defaultFile is the name of the default wakatime config file.
	defaultFile = ".wakatime.cfg"
	// defaultInternalFile is the name of the default wakatime internal config file.
	defaultInternalFile = "wakatime-internal.cfg"
	// DateFormat is the default format for date in config file.
	DateFormat = time.RFC3339
	// defaultTimeout is the default timeout for acquiring a lock.
	defaultTimeout = time.Second * 5

	// WakaHomeTypeUnknown is unknown WakaTime home type.
	WakaHomeTypeUnknown WakaHomeType = iota
	// WakaHomeTypeEnvVar is WakaTime home type from environment variable.
	WakaHomeTypeEnvVar
	// WakaHomeTypeOSDir is WakaTime home type from OS directory.
	WakaHomeTypeOSDir
)

// Writer defines the methods to write to config file.
type Writer interface {
	Write(ctx context.Context, section string, keyValue map[string]string) error
}

// WriterConfig stores the configuration necessary to write to config file.
type WriterConfig struct {
	ConfigFilepath string
	File           *ini.File
}

// NewWriter creates a new writer instance.
func NewWriter(
	ctx context.Context,
	v *viper.Viper,
	filepathFn func(ctx context.Context, v *viper.Viper) (string, error),
) (*WriterConfig, error) {
	configFilepath, err := filepathFn(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("error getting filepath: %s", err)
	}

	logger := log.Extract(ctx)

	// check if file exists
	if !fileExists(configFilepath) {
		logger.Debugf("it will create missing config file %q", configFilepath)

		f, err := os.Create(configFilepath) // nolint:gosec
		if err != nil {
			return nil, fmt.Errorf("failed creating file: %s", err)
		}

		if err = f.Close(); err != nil {
			return nil, fmt.Errorf("failed to close file: %s", err)
		}
	}

	ini, err := ini.LoadSources(ini.LoadOptions{
		AllowPythonMultilineValues: true,
		SkipUnrecognizableLines:    true,
	}, configFilepath)
	if err != nil {
		return nil, fmt.Errorf("error loading config file: %s", err)
	}

	return &WriterConfig{
		ConfigFilepath: configFilepath,
		File:           ini,
	}, nil
}

// Write persists key(s) and value(s) on disk.
func (w *WriterConfig) Write(ctx context.Context, section string, keyValue map[string]string) error {
	logger := log.Extract(ctx)

	if w.File == nil || w.ConfigFilepath == "" {
		return errors.New("got undefined wakatime config file instance")
	}

	for key, value := range keyValue {
		// prevent writing null characters
		key = strings.ReplaceAll(key, "\x00", "")
		value = strings.ReplaceAll(value, "\x00", "")

		w.File.Section(section).Key(key).SetValue(value)
	}

	releaser, err := mutex.Acquire(mutex.Spec{
		Name:    "wakatime-cli-config-mutex",
		Delay:   time.Millisecond,
		Timeout: defaultTimeout,
		Clock:   &mutexClock{delay: time.Millisecond},
	})
	if err != nil {
		logger.Debugf("failed to acquire mutex: %s", err)
	}

	defer func() {
		if releaser != nil {
			releaser.Release()
		}
	}()

	if err := w.File.SaveTo(w.ConfigFilepath); err != nil {
		return fmt.Errorf("error saving wakatime config: %s", err)
	}

	return nil
}

// ReadInConfig reads wakatime config file in memory.
func ReadInConfig(v *viper.Viper, configFilePath string) error {
	v.SetConfigType("ini")
	v.SetConfigFile(configFilePath)

	if err := v.MergeInConfig(); err != nil {
		return fmt.Errorf("failed to merge config file: %s", err)
	}

	return nil
}

// FilePath returns the path for wakatime config file.
func FilePath(ctx context.Context, v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed to expand config param: %s", err)
		}

		return p, nil
	}

	home, _, err := WakaHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %s", err)
	}

	return filepath.Join(home, defaultFile), nil
}

// ImportFilePath returns the path for custom wakatime config file.
// It's used to keep the api key out ofthe home folder, and usually it's to avoid backing up sensitive wakatime config file.
// https://github.com/wakatime/wakatime-cli/issues/464
func ImportFilePath(_ context.Context, v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "settings.import_cfg")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed to expand settings.import_cfg param: %s", err)
		}

		return p, nil
	}

	return "", nil
}

// InternalFilePath returns the path for the wakatime internal config file which contains
// last heartbeat timestamp and backoff time.
func InternalFilePath(ctx context.Context, v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "internal-config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed to expand internal-config param: %s", err)
		}

		return p, nil
	}

	folder, err := WakaResourcesDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %s", err)
	}

	return filepath.Join(folder, defaultInternalFile), nil
}

// WakaHomeDir returns the current user's home directory.
func WakaHomeDir(ctx context.Context) (string, WakaHomeType, error) {
	logger := log.Extract(ctx)

	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		home, err := homedir.Expand(home)
		if err == nil {
			return home, WakaHomeTypeEnvVar, nil
		}

		logger.Warnf("failed to expand WAKATIME_HOME filepath: %s. It will try to get user home dir.", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		logger.Warnf("failed to get user home dir: %s", err)
	}

	if home != "" {
		return home, WakaHomeTypeOSDir, nil
	}

	u, err := user.LookupId(strconv.Itoa(os.Getuid()))
	if err != nil {
		logger.Warnf("failed to user info by userid: %s", err)
	}

	if u.HomeDir != "" {
		return u.HomeDir, WakaHomeTypeOSDir, nil
	}

	return "", WakaHomeTypeUnknown, fmt.Errorf("could not determine wakatime home dir")
}

// WakaResourcesDir returns the ~/.wakatime/ folder.
func WakaResourcesDir(ctx context.Context) (string, error) {
	home, hometype, err := WakaHomeDir(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get user's home directory: %s", err)
	}

	switch hometype {
	case WakaHomeTypeEnvVar:
		return home, nil
	default:
		return filepath.Join(home, defaultFolder), nil
	}
}

// mutexClock is used to implement mutex.Clock interface.
type mutexClock struct {
	delay time.Duration
}

func (mc *mutexClock) After(time.Duration) <-chan time.Time {
	return time.After(mc.delay)
}

func (*mutexClock) Now() time.Time {
	return time.Now()
}

// fileExists checks if a file or directory exist.
func fileExists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}
