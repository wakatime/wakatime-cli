package project

import (
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"

	jww "github.com/spf13/jwalterweatherman"
)

// Detecter is a common interface for project.
type Detecter interface {
	Detect() (Result, bool, error)
	String() string
}

// Result contains the result of Detect().
type Result struct {
	Project string
	Branch  string
}

// Config contains project data.
type Config struct {
	// Override sets an optional project name.
	Override string
	// Alternative sets an alternate project name. Auto-discovered project takes priority.
	Alternative string
	// Patterns contains the overridden project name per path.
	Patterns []Pattern
}

// Pattern contains [projectmap] data.
type Pattern struct {
	// Name is the project name.
	Name string
	// Regex is the regular expression for a specific path.
	Regex *regexp.Regexp
}

// WithDetection find the current project and branch.
// First looks for a .wakatime-project file. Second, uses the --project arg.
// Third, uses the folder name from a revision control repository. Last, uses
// the --alternate-project arg.
func WithDetection(c Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				if h.EntityType == heartbeat.FileType {
					project, branch := detect(h.Entity, c.Patterns)

					hh[n].Branch = &branch
					hh[n].Project = &project
				} else {
					project := firstNonEmptyString(c.Override, c.Alternative)
					hh[n].Project = &project
				}
			}

			return next(hh)
		}
	}
}

func detect(entity string, patterns []Pattern) (project, branch string) {
	var configPlugins []Detecter = []Detecter{
		File{
			Filepath: entity,
		},
		Map{
			Filepath: entity,
			Patterns: patterns,
		},
	}

	for _, p := range configPlugins {
		result, detected, err := p.Detect()
		if err != nil {
			jww.ERROR.Printf("unexpected error occurred at %q: %s", p.String(), err)
			continue
		} else if detected {
			return result.Project, result.Branch
		}
	}

	return "", ""
}

// firstNonEmptyString accepts multiple values and return the first non empty string value.
func firstNonEmptyString(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
