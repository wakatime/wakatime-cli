package arguments

import (
	"strconv"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/configs"
)

// GetBooleanOrList Get a boolean or list of regexes from args and configs
func GetBooleanOrList(section string, key string, cfg *configs.ConfigFile) []string {
	arr := []string{}

	values, err := cfg.Get(section, key)
	if err == nil {
		b, err := strconv.ParseBool(*values)
		if err == nil {
			if b {
				arr = append(arr, ".*")
				return arr
			}

			arr = append(arr, "")
			return arr
		}
	}

	// not bool? try parse list
	parts := strings.Split(*values, "\n")

	for _, part := range parts {
		part = strings.TrimSpace(part)

		if len(part) > 0 {
			arr = append(arr, part)
		}
	}

	return arr
}
