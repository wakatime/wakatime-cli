package project

import (
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	// Alternate sets an alternate project name. Auto-discovered project takes priority.
	Alternate string
	// Patterns contains the overridden project name per path.
	MapPatterns []MapPattern
	// SubmodulePatterns contains the paths to validate for submodules.
	SubmodulePatterns []*regexp.Regexp
	// ShouldObfuscateProject determines if project should be obfuscated according some rules.
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
					project := firstNonEmptyString(c.Override, c.Alternate)
					hh[n].Project = &project

					continue
				}

				result := Detect(h.Entity, c.MapPatterns)

				if result.Project == "" {
					result.Project = c.Override
				}

				if result.Project == "" || result.Branch == "" {
					result = DetectWithRevControl(h.Entity, c.SubmodulePatterns,
						c.ShouldObfuscateProject, result.Project, result.Branch)
					if result.Project == "" {
						result.Project = setProjectName(c.Alternate, c.ShouldObfuscateProject, result.Folder)
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
	shouldObfuscateProject bool, project string, branch string) Result {
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
			result := Result{
				Project: firstNonEmptyString(project, result.Project),
				Branch:  firstNonEmptyString(branch, result.Branch),
				Folder:  result.Folder,
			}

			if shouldObfuscateProject {
				result.Project = ""
			}

			return result
		}
	}

	return Result{}
}

func setProjectName(alternate string, shouldObfuscateProject bool, folder string) string {
	if !shouldObfuscateProject {
		return alternate
	}

	project := generateProjectName()

	fc := FileControl{
		Path:    folder,
		Project: project,
	}

	err := fc.Write()
	if err != nil {
		jww.WARN.Printf("failed to write: %s", err)
	}

	return project
}

func generateProjectName() string {
	adjectives := []string{
		"aged", "ancient", "autumn", "billowing", "bitter", "black", "blue", "bold",
		"broad", "broken", "calm", "cold", "cool", "crimson", "curly", "damp",
		"dark", "dawn", "delicate", "divine", "dry", "empty", "falling", "fancy",
		"flat", "floral", "fragrant", "frosty", "gentle", "green", "hidden", "holy",
		"icy", "jolly", "late", "lingering", "little", "lively", "long", "lucky",
		"misty", "morning", "muddy", "mute", "nameless", "noisy", "odd", "old",
		"orange", "patient", "plain", "polished", "proud", "purple", "quiet", "rapid",
		"raspy", "red", "restless", "rough", "round", "royal", "shiny", "shrill",
		"shy", "silent", "small", "snowy", "soft", "solitary", "sparkling", "spring",
		"square", "steep", "still", "summer", "super", "sweet", "throbbing", "tight",
		"tiny", "twilight", "wandering", "weathered", "white", "wild", "winter", "wispy",
		"withered", "yellow", "young"}

	nouns := []string{
		"art", "band", "bar", "base", "bird", "block", "boat", "bonus",
		"bread", "breeze", "brook", "bush", "butterfly", "cake", "cell", "cherry",
		"cloud", "credit", "darkness", "dawn", "dew", "disk", "dream", "dust",
		"feather", "field", "fire", "firefly", "flower", "fog", "forest", "frog",
		"frost", "glade", "glitter", "grass", "hall", "hat", "haze", "heart",
		"hill", "king", "lab", "lake", "leaf", "limit", "math", "meadow",
		"mode", "moon", "morning", "mountain", "mouse", "mud", "night", "paper",
		"pine", "poetry", "pond", "queen", "rain", "recipe", "resonance", "rice",
		"river", "salad", "scene", "sea", "shadow", "shape", "silence", "sky",
		"smoke", "snow", "snowflake", "sound", "star", "sun", "sun", "sunset",
		"surf", "term", "thunder", "tooth", "tree", "truth", "union", "unit",
		"violet", "voice", "water", "waterfall", "wave", "wildflower", "wind", "wood"}

	str := []string{}

	rand.Seed(time.Now().UnixNano())
	str = append(str, strings.Title(adjectives[rand.Intn(len(adjectives))]))
	rand.Seed(time.Now().UnixNano())
	str = append(str, strings.Title(nouns[rand.Intn(len(nouns))]))
	rand.Seed(time.Now().UnixNano())
	str = append(str, strconv.Itoa(rand.Intn(100)))

	return strings.Join(str, " ")
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
