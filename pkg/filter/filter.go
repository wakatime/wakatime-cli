package filter

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/project"

	jww "github.com/spf13/jwalterweatherman"
)

// Err is a heartbeat filtering error, signaling to skip the heartbeat.
type Err string

// Error implements error interface.
func (e Err) Error() string {
	return string(e)
}

// Config contains filtering configurations.
type Config struct {
	Exclude                    []*regexp.Regexp
	ExcludeUnknownProject      bool
	Include                    []*regexp.Regexp
	IncludeOnlyWithProjectFile bool
}

// WithFiltering initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline.
func WithFiltering(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			var filtered []heartbeat.Heartbeat

			for _, h := range hh {
				err := Filter(h, config)
				if err != nil {
					var errv Err
					if errors.As(err, &errv) {
						jww.DEBUG.Println(errv.Error())
						continue
					}

					return nil, fmt.Errorf("error filtergin heartbeat: %w", err)
				}

				filtered = append(filtered, h)
			}

			return next(filtered)
		}
	}
}

// Filter filters a heartbeat following the passed in configuration.
// Will return Err to signaling to the caller to skip the heartbeat.
func Filter(h heartbeat.Heartbeat, config Config) error {
	// unknown language
	if h.Language == nil || *h.Language == "" {
		return Err("skipping because of unknown language")
	}

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

// filterByPattern filters an entity by include and exclude pattern. Include
// will overwrite exclude.
// Will return Err to signaling to the caller to skip the heartbeat.
func filterByPattern(entity string, include, exclude []*regexp.Regexp) error {
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
			return Err(fmt.Sprintf("skipping because matching exclude pattern %q", pattern.String()))
		}
	}

	return nil
}

// filterFileEntity filters a filepath by verifying existence of file, and optionally
// checking if wakatime project file exists.
// Will return Err to signaling to the caller to skip the heartbeat.
func filterFileEntity(filepath string, includeOnlyWithProjectFile bool) error {
	// check if file exists
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return Err(fmt.Sprintf("skipping because of non-existing file %q", filepath))
	}

	// check wakatime project file exists
	if includeOnlyWithProjectFile {
		_, ok, err := project.FindFile(filepath)
		if err != nil {
			return fmt.Errorf("error detecting project file: %s", err)
		}

		if !ok {
			return Err("skipping because of non-existing project file in parent path")
		}
	}

	return nil
}
