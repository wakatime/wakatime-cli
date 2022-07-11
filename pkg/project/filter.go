package project

import (
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
)

// FilterConfig contains project filtering configurations.
type FilterConfig struct {
	// ExcludeUnknownProject determines if heartbeat should be skipped when the project cannot be detected.
	ExcludeUnknownProject bool
}

// WithFiltering initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to filter heartbeats following
// the provided configurations.
func WithFiltering(config FilterConfig) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute project filtering")

			var filtered []heartbeat.Heartbeat

			for _, h := range hh {
				err := Filter(h, config)
				if err != nil {
					log.Debugln(err.Error())

					if h.LocalFileNeedsCleanup {
						err = os.Remove(h.LocalFile)
						if err != nil {
							log.Warnf("unable to delete tmp file: %s", err)
						}
					}

					continue
				}

				filtered = append(filtered, h)
			}

			return next(filtered)
		}
	}
}

// Filter determines, following the passed in configurations, if a heartbeat
// should be skipped.
func Filter(h heartbeat.Heartbeat, config FilterConfig) error {
	// exclude unknown project
	if config.ExcludeUnknownProject && (h.Project == nil || *h.Project == "") {
		return fmt.Errorf("skipping because of unknown project")
	}

	return nil
}
