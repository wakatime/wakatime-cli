package vipertools

import (
	"github.com/spf13/viper"
)

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
		if value := v.GetString(key); value != "" {
			return value, true
		}
	}

	return "", false
}
