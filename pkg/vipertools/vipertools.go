package vipertools

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	viperini "github.com/go-viper/encoding/ini"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	iniv1 "gopkg.in/ini.v1"
)

// New creates a new viper instance with the ini codec registered.
func New() (*viper.Viper, error) {
	multilineOption := iniv1.LoadOptions{AllowPythonMultilineValues: true}
	iniCodec := viperini.Codec{LoadOptions: multilineOption}

	codecRegistry := viper.NewCodecRegistry()
	if err := codecRegistry.RegisterCodec("ini", iniCodec); err != nil {
		return nil, fmt.Errorf("failed to register ini codec: %w", err)
	}

	return viper.NewWithOptions(viper.WithCodecRegistry(codecRegistry)), nil
}

// MustNew creates a new viper instance with the ini codec registered and panics if it fails.
// This is useful for testing.
func MustNew() *viper.Viper {
	v, err := New()
	if err != nil {
		panic(fmt.Sprintf("failed to create viper instance: %s", err))
	}

	return v
}

// FirstNonEmptyBool accepts multiple keys and returns the first non-empty bool value
// from viper.Viper via these keys. Non-empty meaning key not set will not be accepted.
// Will return false as second parameter, if non-empty bool value could not be retrieved.
func FirstNonEmptyBool(v *viper.Viper, keys ...string) bool {
	if v == nil {
		return false
	}

	for _, key := range keys {
		if !v.IsSet(key) {
			continue
		}

		value := v.Get(key)

		parsed, err := cast.ToBoolE(value)
		if err != nil {
			continue
		}

		return parsed
	}

	return false
}

// FirstNonEmptyInt accepts multiple keys and returns the first non-empty int value
// from viper.Viper via these keys. Non-empty meaning key not set will not be accepted.
// Will return false as second parameter, if non-empty int value could not be retrieved.
func FirstNonEmptyInt(v *viper.Viper, keys ...string) (int, bool) {
	if v == nil {
		return 0, false
	}

	for _, key := range keys {
		if !v.IsSet(key) {
			continue
		}

		// Zero means a valid value when set, so it needs to use generic function and later cast it to int
		value := v.Get(key)

		// If the value is not an int, it will continue to find the next non-empty key
		parsed, err := cast.ToIntE(value)
		if err != nil {
			continue
		}

		return parsed, true
	}

	return 0, false
}

// FirstNonEmptyString accepts multiple keys and returns the first non-empty string value
// from viper.Viper via these keys. Returns empty string by default if a value couldn't be found.
func FirstNonEmptyString(v *viper.Viper, keys ...string) string {
	if v == nil {
		return ""
	}

	for _, key := range keys {
		if !v.IsSet(key) {
			continue
		}

		value := v.Get(key)

		parsed, err := cast.ToStringE(value)
		if err != nil {
			continue
		}

		return strings.Trim(parsed, `"'`)
	}

	return ""
}

// GetString gets a parameter/setting by key and strips any quotes.
func GetString(v *viper.Viper, key string) string {
	return strings.Trim(v.GetString(key), `"'`)
}

// GetStringMapString gets a parameter/setting by key prefix and strips any quotes.
func GetStringMapString(v *viper.Viper, prefix string) map[string]string {
	m := map[string]string{}

	for _, k := range v.AllKeys() {
		if !strings.HasPrefix(k, prefix+".") {
			continue
		}

		m[strings.TrimPrefix(k, prefix+".")] = GetString(v, k)
	}

	return m
}

// SafeTimeParse turns a string into a time.Time.
func SafeTimeParse(format, s string) (parsed time.Time, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panicked: failed to time.Parse: %v. Stack: %s", r, string(debug.Stack()))
		}
	}()

	parsed, err = time.Parse(format, s)

	return parsed, err
}
