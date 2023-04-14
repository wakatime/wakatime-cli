package ini

import (
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
	// WakaHomeTypeUnknown is unknown WakaTime home type.
	WakaHomeTypeUnknown WakaHomeType = iota
	// WakaHomeTypeEnvVar is WakaTime home type from environment variable.
	WakaHomeTypeEnvVar
	// WakaHomeTypeOSDir is WakaTime home type from OS directory.
	WakaHomeTypeOSDir
)

const (
	// defaultFile is the name of the default wakatime config file.
	defaultFile = ".wakatime.cfg"
	// defaultInternalFile is the name of the default wakatime internal config file.
	defaultInternalFile = "wakatime-internal.cfg"
	// DateFormat is the default format for date in config file.
	DateFormat = time.RFC3339
	// defaultTimeout is the default timeout for acquiring a lock.
	defaultTimeout = time.Second * 5
)

// Writer defines the methods to write to config file.
type Writer interface {
	Write(section string, keyValue map[string]string) error
}

// WriterConfig stores the configuration necessary to write to config file.
type WriterConfig struct {
	ConfigFilepath string
	File           *ini.File
}

// NewWriter creates a new writer instance.
func NewWriter(v *viper.Viper, filepathFn func(v *viper.Viper) (string, error)) (*WriterConfig, error) {
	configFilepath, err := filepathFn(v)
	if err != nil {
		return nil, fmt.Errorf("error getting filepath: %s", err)
	}

	// check if file exists
	if !fileExists(configFilepath) {
		log.Debugf("it will create missing config file %q", configFilepath)

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
func (w *WriterConfig) Write(section string, keyValue map[string]string) error {
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
		log.Debugf("failed to acquire mutex: %s", err)
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
		return fmt.Errorf("error parsing config file: %s", err)
	}

	return nil
}

// FilePath returns the path for wakatime config file.
func FilePath(v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed expanding config param: %s", err)
		}

		return p, nil
	}

	home, _, err := WakaHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return filepath.Join(home, defaultFile), nil
}

// ImportFilePath returns the path for import wakatime config file.
func ImportFilePath(v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "settings.import_cfg")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed expanding settings.import_cfg param: %s", err)
		}

		return p, nil
	}

	return "", nil
}

// InternalFilePath returns the path for the wakatime internal config file.
func InternalFilePath(v *viper.Viper) (string, error) {
	configFilepath := vipertools.GetString(v, "internal-config")
	if configFilepath != "" {
		p, err := homedir.Expand(configFilepath)
		if err != nil {
			return "", fmt.Errorf("failed expanding internal-config param: %s", err)
		}

		return p, nil
	}

	folder, err := WakaResourcesDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
	}

	return filepath.Join(folder, defaultInternalFile), nil
}

// WakaHomeDir returns the current user's home directory.
func WakaHomeDir() (string, WakaHomeType, error) {
	home, exists := os.LookupEnv("WAKATIME_HOME")
	if exists && home != "" {
		home, err := homedir.Expand(home)
		if err == nil {
			return home, WakaHomeTypeEnvVar, nil
		}

		log.Warnf("failed to expand WAKATIME_HOME filepath: %s", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Warnf("failed to get user home dir: %s", err)
	}

	if home != "" {
		return home, WakaHomeTypeOSDir, nil
	}

	u, err := user.LookupId(strconv.Itoa(os.Getuid()))
	if err != nil {
		log.Warnf("failed to user info by userid: %s", err)
	}

	if u.HomeDir != "" {
		return u.HomeDir, WakaHomeTypeOSDir, nil
	}

	return "", WakaHomeTypeUnknown, fmt.Errorf("could not determine wakatime home dir")
}

// WakaResourcesDir returns the ~/.wakatime/ folder.
func WakaResourcesDir() (string, error) {
	home, hometype, err := WakaHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed getting user's home directory: %s", err)
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
