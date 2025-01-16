package vipertools

import (
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// FirstNonEmptyBool accepts multiple keys and returns the first non-empty bool value
// from viper.Viper via these keys. Non-empty meaning key not set will not be accepted.
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
		//	if value := GetString(v, key); value != "" {
		//		return value
		//	}
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
