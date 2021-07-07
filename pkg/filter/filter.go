package filter

import (
	"errors"
	"fmt"
	"os"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/project"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// Config contains filtering configurations.
type Config struct {
	Exclude                    []regex.Regex
	ExcludeUnknownProject      bool
	Include                    []regex.Regex
	IncludeOnlyWithProjectFile bool
}

// WithFiltering initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to filter heartbeats following
// the provided configurations.
func WithFiltering(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute heartbeat filtering")

			var filtered []heartbeat.Heartbeat

			for _, h := range hh {
				err := Filter(h, config)
				if err != nil {
					var errv Err
					if errors.As(err, &errv) {
						log.Debugln(errv.Error())
						continue
					}

					return nil, fmt.Errorf("error filtering heartbeat: %w", err)
				}

				filtered = append(filtered, h)
			}

			if len(filtered) == 0 {
				log.Debugln("no heartbeat left after filtering. abort heartbeat handling.")
				return []heartbeat.Result{}, nil
			}

			return next(filtered)
		}
	}
}

// Filter determines, following the passed in configurations, if a heartbeat
// should be skipped.
// Returns Err to signal to the caller to skip the heartbeat.
func Filter(h heartbeat.Heartbeat, config Config) error {
	// unknown project
	if config.ExcludeUnknownProject && (h.Project == nil || *h.Project == "") {
		return Err("skipping because of unknown project")
	}

	// filter by pattern
	if err := filterByPattern(h.Entity, config.Include, config.Exclude); err != nil {
		return fmt.Errorf("filter by pattern: %w", err)
	}

	// filter file
	if h.EntityType == heartbeat.FileType {
		err := filterFileEntity(h.Entity, config.IncludeOnlyWithProjectFile)
		if err != nil {
			return fmt.Errorf("filter file: %w", err)
		}
	}

	return nil
}

// filterByPattern determines if a heartbeat should be skipped by checking an
// entity against include and exclude patterns. Include will override exclude.
// Returns Err to signal to the caller to skip the heartbeat.
func filterByPattern(entity string, include, exclude []regex.Regex) error {
	if entity == "" {
		return nil
	}

	// filter by include pattern
	for _, pattern := range include {
		if pattern.MatchString(entity) {
			return nil
		}
	}

	// filter by  exclude pattern
	for _, pattern := range exclude {
		if pattern.MatchString(entity) {
			return Err(fmt.Sprintf("skipping because matches exclude pattern %q", pattern.String()))
		}
	}

	return nil
}

// filterFileEntity determines if a heartbeat should be skipped, by verifying
// the existence of the passed in filepath, and optionally by checking if a
// wakatime project file can be detected in the filepath directory tree.
// Returns Err to signal to the caller to skip the heartbeat.
func filterFileEntity(filepath string, includeOnlyWithProjectFile bool) error {
	// check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return Err(fmt.Sprintf("skipping because of non-existing file %q", filepath))
	}

	// check wakatime project file exists
	if includeOnlyWithProjectFile {
		_, ok := project.FindFileOrDirectory(filepath, "", ".wakatime-project")
		if !ok {
			return Err("skipping because of missing .wakatime-project file in parent path")
		}
	}

	return nil
}
