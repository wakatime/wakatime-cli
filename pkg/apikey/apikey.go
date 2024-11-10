package apikey

import (
	"context"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// Config contains apikey project detection configurations.
type Config struct {
	// DefaultAPIKey contains the default api key.
	DefaultAPIKey string
	// Patterns contains the overridden api key per path.
	MapPatterns []MapPattern
}

// MapPattern contains [project_api_key] data.
type MapPattern struct {
	// APIKey is the project related api key.
	APIKey string
	// Regex is the regular expression for a specific path.
	Regex regex.Regex
}

// WithReplacing initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to replace default api key
// for a heartbeat following the provided configurations.
func WithReplacing(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute api key replacing")

			for n, h := range hh {
				result, ok := MatchPattern(ctx, h.Entity, config.MapPatterns)
				if ok {
					hh[n].APIKey = result

					continue
				}

				hh[n].APIKey = config.DefaultAPIKey
			}

			return next(ctx, hh)
		}
	}
}

// MatchPattern matches regex against entity's path to find alternate api key.
func MatchPattern(ctx context.Context, fp string, patterns []MapPattern) (string, bool) {
	logger := log.Extract(ctx)

	for _, pattern := range patterns {
		if pattern.Regex.MatchString(ctx, fp) {
			logger.Debugf("api key pattern %q matched path %q", pattern.Regex.String(), fp)
			return pattern.APIKey, true
		}

		logger.Debugf("api key pattern %q did not match path %q", pattern.Regex.String(), fp)
	}

	return "", false
}
