package apikey

import (
	"context"
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

type (
	// Config contains apikey project detection configurations.
	Config struct {
		// DefaultAPIKey contains the default api key.
		DefaultAPIKey string
		// DefaultAPIURL contains the default api url.
		DefaultAPIURL string
		// Patterns contains the overridden api key per path.
		MapPatterns []MapPattern
		// URLPatterns contains the overridden api url and key per path.
		URLPatterns []URLPattern
	}

	// MapPattern contains [project_api_key] data.
	MapPattern struct {
		// APIKey is the project related api key.
		APIKey string
		// Regex is the regular expression for a specific path.
		Regex regex.Regex
	}

	// URLPattern contains [api_urls] data mapping paths to API URL and key combinations.
	URLPattern struct {
		// APIURL is the API URL for matching paths.
		APIURL string
		// APIKey is the API key for matching paths.
		APIKey string
		// Regex is the regular expression for a specific path.
		Regex regex.Regex
	}

	// destination represents a unique API URL and key combination.
	destination struct {
		url string
		key string
	}
)

// WithReplacing initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to replace default api key
// and api url for a heartbeat following the provided configurations.
//
// The behavior is:
//  1. Always send to api_url (default WakaTime) with api_key (or project_api_key if matched)
//  2. Also send to ALL matching api_urls patterns
//  3. Use a hash map of api_url|api_key to deduplicate destinations
func WithReplacing(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			// logger.Debugln("execute api key replacing")
			var result []heartbeat.Heartbeat

			for _, h := range hh {
				// Use a map to deduplicate destinations by url|key
				destinations := make(map[string]destination)

				// 1. Add default destination with api_key (or project_api_key if matched)
				apiKey := config.DefaultAPIKey

				// Check if a project_api_key pattern matches (overrides default key)
				if matchedKey, ok := MatchPattern(ctx, h.Entity, config.MapPatterns); ok {
					apiKey = matchedKey
				}

				key := fmt.Sprintf("%s|%s", config.DefaultAPIURL, apiKey)
				destinations[key] = destination{url: config.DefaultAPIURL, key: apiKey}

				// 2. Add all matching url patterns (not just the first one)
				matchedURLPatterns := MatchAllURLPatterns(ctx, h.Entity, config.URLPatterns)
				for _, pattern := range matchedURLPatterns {
					key := fmt.Sprintf("%s|%s", pattern.APIURL, pattern.APIKey)
					destinations[key] = destination{url: pattern.APIURL, key: pattern.APIKey}
				}

				// 3. Create a heartbeat copy for each unique destination
				for _, dest := range destinations {
					copy := h
					copy.APIURL = dest.url
					copy.APIKey = dest.key
					result = append(result, copy)
				}
			}

			return next(ctx, result)
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

// MatchAllURLPatterns matches regex against entity's path and returns ALL matching url patterns.
func MatchAllURLPatterns(ctx context.Context, fp string, patterns []URLPattern) []URLPattern {
	logger := log.Extract(ctx)

	var matched []URLPattern

	for _, pattern := range patterns {
		if pattern.Regex.MatchString(ctx, fp) {
			logger.Debugf("api url pattern %q matched path %q", pattern.Regex.String(), fp)
			matched = append(matched, pattern)
		} else {
			logger.Debugf("api url pattern %q did not match path %q", pattern.Regex.String(), fp)
		}
	}

	return matched
}
