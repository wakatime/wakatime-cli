package apikey

import (
	"github.com/slongfield/pyfmt"
	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// Config contains apikey project detection configurations.
type Config struct {
	// DefaultApiKey is the key used when no pattern matches a heartbeat's entity
	DefaultApiKey string
	// Patterns contains the overridden api key per path.
	Patterns []ProjectPattern
}

// ProjectPattern contains [project_api_key] data.
type ProjectPattern struct {
	// ApiKey is the project related api key.
	ApiKey string
	// Regex is the regular expression to match an entity path.
	Regex regex.Regex
}

// WithApiKey populates the ApiKey field on a Heartbeat.
func WithApiKey(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("assign api key to heartbeats")

			for n, h := range hh {
				result, ok := MatchPattern(h.Entity, config.Patterns)
				if !ok {
					hh[n].ApiKey = config.DefaultApiKey
					continue
				}

				hh[n].ApiKey = result
			}

			return next(hh)
		}
	}
}

// MatchPattern matches regex against entity's path to find alternate api key.
func MatchPattern(fp string, patterns []ProjectPattern) (string, bool) {
	for _, pattern := range patterns {
		if pattern.Regex.MatchString(fp) {
			matches := pattern.Regex.FindStringSubmatch(fp)
			if len(matches) > 0 {
				log.Debugf("project api key pattern matched: %s", pattern.Regex.String())

				params := make([]interface{}, len(matches[1:]))
				for i, v := range matches[1:] {
					params[i] = v
				}

				result, err := pyfmt.Fmt(pattern.ApiKey, params...)
				if err != nil {
					log.Errorf("error formatting %q: %s", pattern.ApiKey, err)
					continue
				}

				return result, true
			}
		}
	}

	return "", false
}
