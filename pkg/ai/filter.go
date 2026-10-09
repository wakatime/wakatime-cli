package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// FilterConfig contains per-project AI tracking configurations.
type FilterConfig struct {
	// SyncDisabled determines if heartbeats with category "ai coding" should be skipped.
	// It can be set globally or per-project via a .wakatime file containing
	// sync_ai_disabled = true. Normal coding activity is still logged.
	SyncDisabled bool
	// ExcludeProjects contains project name patterns excluded from AI tracking.
	// Heartbeats with category "ai coding" matching any of these patterns are
	// skipped, while normal coding activity is still logged. POSIX regex syntax.
	ExcludeProjects []regex.Regex
}

// WithFiltering initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to filter ai coding heartbeats
// following the provided configurations. It must run after project detection,
// so the project name is already resolved, and before sanitization, so project
// names are not yet obfuscated.
func WithFiltering(config FilterConfig) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(ctx context.Context, hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			logger := log.Extract(ctx)

			var filtered []heartbeat.Heartbeat

			for _, h := range hh {
				err := Filter(ctx, h, config)
				if err != nil {
					logger.Debugln(err.Error())

					if h.LocalFileNeedsCleanup {
						err = os.Remove(h.LocalFile)
						if err != nil {
							logger.Warnf("unable to delete tmp file: %s", err)
						}
					}

					continue
				}

				filtered = append(filtered, h)
			}

			return next(ctx, filtered)
		}
	}
}

// Filter determines, following the passed in configurations, if an ai coding
// heartbeat should be skipped. Heartbeats of any other category are never skipped.
func Filter(ctx context.Context, h heartbeat.Heartbeat, config FilterConfig) error {
	if h.Category != heartbeat.AICodingCategory.String() {
		return nil
	}

	if config.SyncDisabled {
		return fmt.Errorf("skipping ai coding heartbeat because ai tracking is disabled for this project")
	}

	var projectName string

	if h.Project != nil {
		projectName = *h.Project
	}

	if projectName == "" {
		return nil
	}

	for _, pattern := range config.ExcludeProjects {
		if pattern.MatchString(ctx, projectName) {
			return fmt.Errorf(
				"skipping ai coding heartbeat because project %q matches ai exclude projects pattern %q",
				projectName,
				pattern.String(),
			)
		}
	}

	return nil
}
