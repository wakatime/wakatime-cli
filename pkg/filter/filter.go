package filter

import (
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
					log.Debugf(err.Error())

					continue
				}

				filtered = append(filtered, h)
			}

			return next(filtered)
		}
	}
}

// WithLengthValidator initializes and returns a heartbeat handle option, which
// can be used to abort execution if all heartbeats were filtered and the list is empty.
func WithLengthValidator() heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			if len(hh) == 0 {
				log.Debugln("no heartbeats left after filtering. abort heartbeat handling.")
				return []heartbeat.Result{}, nil
			}

			return next(hh)
		}
	}
}

// Filter determines, following the passed in configurations, if a heartbeat
// should be skipped.
func Filter(h heartbeat.Heartbeat, config Config) error {
	// filter by pattern
	if err := filterByPattern(h.Entity, config.Include, config.Exclude); err != nil {
		return fmt.Errorf(fmt.Sprintf("filter by pattern: %s", err))
	}

	err := filterFileEntity(h, config)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("filter file: %s", err))
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
			return fmt.Errorf(fmt.Sprintf("skipping because matches exclude pattern %q", pattern.String()))
		}
	}

	return nil
}

// filterFileEntity determines if a heartbeat of type file should be skipped, by verifying
// the existence of the passed in filepath, and optionally by checking if a
// wakatime project file can be detected in the filepath directory tree.
// Returns an error to signal to the caller to skip the heartbeat.
func filterFileEntity(h heartbeat.Heartbeat, config Config) error {
	if h.EntityType != heartbeat.FileType {
		return nil
	}

	if h.IsUnsavedEntity {
		return nil
	}

	if h.IsRemote() {
		return nil
	}

	entity := h.Entity
	if h.LocalFile != "" {
		entity = h.LocalFile
	}

	// skip files that don't exist on disk
	if _, err := os.Stat(entity); os.IsNotExist(err) {
		return fmt.Errorf(fmt.Sprintf("skipping because of non-existing file %q", entity))
	}

	// when including only with project file, skip files when the project doesn't have a .wakatime-project file
	if config.IncludeOnlyWithProjectFile {
		_, ok := project.FindFileOrDirectory(entity, project.WakaTimeProjectFile)
		if !ok {
			return fmt.Errorf("skipping because missing .wakatime-project file in parent path")
		}
	}

	return nil
}
