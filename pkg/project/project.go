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
	Folder  string
}

// Config contains project detection configurations.
type Config struct {
	// Override sets an optional project name.
	Override string
	// Alternative sets an alternate project name. Auto-discovered project takes priority.
	Alternative string
	// Patterns contains the overridden project name per path.
	MapPatterns []MapPattern
	// SubmodulePatterns contains the paths to validate for submodules.
	SubmodulePatterns []*regexp.Regexp
	// ShouldObfuscateProject if true will take Alternative string, otherwise will be ignored.
	ShouldObfuscateProject bool
}

// MapPattern contains [projectmap] data.
type MapPattern struct {
	// Name is the project name.
	Name string
	// Regex is the regular expression for a specific path.
	Regex *regexp.Regexp
}

// WithDetection finds the current project and branch.
// First looks for a .wakatime-project file. Second, uses the --project arg.
// Third, uses the folder name from a revision control repository. Last, uses
// the --alternate-project arg.
func WithDetection(c Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				if h.EntityType != heartbeat.FileType {
					project := firstNonEmptyString(c.Override, c.Alternative)
					hh[n].Project = &project

					continue
				}

				result := Detect(h.Entity, c.MapPatterns)

				if result.Project == "" {
					result.Project = c.Override
				}

				if result.Project == "" || result.Branch == "" {
					result = DetectWithRevControl(h.Entity, c.SubmodulePatterns, result.Project, result.Branch)
					if c.ShouldObfuscateProject {
						result.Project = ""
					}
				}

				hh[n].Branch = &result.Branch
				hh[n].Project = &result.Project
			}

			return next(hh)
		}
	}
}

// Detect finds the current project and branch from config plugins.
func Detect(entity string, patterns []MapPattern) Result {
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
			return result
		}
	}

	return Result{}
}

// DetectWithRevControl finds the current project and branch from rev control.
func DetectWithRevControl(entity string, submodulePatterns []*regexp.Regexp,
	project string, branch string) Result {
	var revControlPlugins []Detecter = []Detecter{
		Git{
			Filepath:          entity,
			SubmodulePatterns: submodulePatterns,
		},
		Mercurial{
			Filepath: entity,
		},
		Subversion{
			Filepath: entity,
		},
	}

	for _, p := range revControlPlugins {
		result, detected, err := p.Detect()
		if err != nil {
			jww.ERROR.Printf("unexpected error occurred at %q: %s", p.String(), err)
			continue
		} else if detected {
			return Result{
				Project: firstNonEmptyString(project, result.Project),
				Branch:  firstNonEmptyString(branch, result.Branch),
				Folder:  result.Folder,
			}
		}
	}

	return Result{}
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
