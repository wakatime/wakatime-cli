package vipertools

import (
	"strings"

	"github.com/spf13/viper"
)

// FirstNonEmptyBool accepts multiple keys and returns the first non empty bool value
// from viper.Viper via these keys. Non-empty meaning false value will not be accepted.
func FirstNonEmptyBool(v *viper.Viper, keys ...string) bool {
	if v == nil {
		return false
	}

	for _, key := range keys {
		if value := v.GetBool(key); value {
			return value
		}
	}

	return false
}

// FirstNonEmptyInt accepts multiple keys and returns the first non empty int value
// from viper.Viper via these keys. Non-empty meaning 0 value will not be accepted.
// Will return false as second parameter, if non-empty int value could not be retrieved.
func FirstNonEmptyInt(v *viper.Viper, keys ...string) (int, bool) {
	if v == nil {
		return 0, false
	}

	for _, key := range keys {
		if value := v.GetInt(key); value != 0 {
			return value, true
		}
	}

	return 0, false
}

// FirstNonEmptyString accepts multiple keys and returns the first non empty string value
// from viper.Viper via these keys. Non-empty meaning "" value will not be accepted.
// Will return false as second parameter, if non-empty string value could not be retrieved.
func FirstNonEmptyString(v *viper.Viper, keys ...string) (string, bool) {
	if v == nil {
		return "", false
	}

	for _, key := range keys {
		if value := GetString(v, key); value != "" {
			return strings.Trim(value, `"'`), true
		}
	}

	return "", false
}

// GetString gets a parameter/setting by key and strips any quotes.
func GetString(v *viper.Viper, key string) string {
	return strings.Trim(v.GetString(key), `"'`)
}
