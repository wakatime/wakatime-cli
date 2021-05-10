package project

import (
	"math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
)

// nolint: gochecknoglobals
var driveLetterRegex = regexp.MustCompile(`^[a-zA-Z]:\\$`)

// maxRecursiveIteration limits the number of a func will be called recursively.
const maxRecursiveIteration = 500

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
	// Patterns contains the overridden project name per path.
	MapPatterns []MapPattern
	// SubmodulePatterns contains the paths to validate for submodules.
	SubmodulePatterns []regex.Regex
	// ShouldObfuscateProject determines if the project name should be obfuscated according some rules.
	ShouldObfuscateProject bool
}

// MapPattern contains [projectmap] data.
type MapPattern struct {
	// Name is the project name.
	Name string
	// Regex is the regular expression for a specific path.
	Regex regex.Regex
}

// WithDetection finds the current project and branch.
// First looks for a .wakatime-project file. Second, uses the --project arg.
// Third, uses the folder name from a revision control repository. Last, uses
// the --alternate-project arg.
func WithDetection(c Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			log.Debugln("execute project detection")

			for n, h := range hh {
				if h.EntityType != heartbeat.FileType {
					project := firstNonEmptyString(h.ProjectOverride, h.ProjectAlternate)
					hh[n].Project = &project

					continue
				}

				var result Result

				result.Project, result.Branch = Detect(h.Entity, c.MapPatterns)

				if result.Project == "" {
					result.Project = h.ProjectOverride
				}

				if result.Project == "" || result.Branch == "" {
					revControlResult := DetectWithRevControl(h.Entity, c.SubmodulePatterns, c.ShouldObfuscateProject)

					result.Project = firstNonEmptyString(result.Project, revControlResult.Project)
					result.Branch = firstNonEmptyString(result.Branch, revControlResult.Branch)

					if result.Project == "" {
						result.Project = setProjectName(h.ProjectAlternate, c.ShouldObfuscateProject, revControlResult.Folder)
					}
				}

				hh[n].Project = &result.Project
				hh[n].Branch = &result.Branch
			}

			return next(hh)
		}
	}
}

// Detect finds the current project and branch from config plugins.
func Detect(entity string, patterns []MapPattern) (project, branch string) {
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
			log.Errorf("unexpected error occurred at %q: %s", p.String(), err)
			continue
		} else if detected {
			return result.Project, result.Branch
		}
	}

	return "", ""
}

// DetectWithRevControl finds the current project and branch from rev control.
func DetectWithRevControl(entity string, submodulePatterns []regex.Regex, shouldObfuscate bool) Result {
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
		Tfvc{
			Filepath: entity,
		},
	}

	for _, p := range revControlPlugins {
		result, detected, err := p.Detect()
		if err != nil {
			log.Errorf("unexpected error occurred at %q: %s", p.String(), err)
			continue
		}

		if detected {
			result := Result{
				Project: result.Project,
				Branch:  result.Branch,
				Folder:  result.Folder,
			}

			if shouldObfuscate {
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

	err := Write(folder, project)
	if err != nil {
		log.Warnf("failed to write: %s", err)
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
	str = append(str, strings.Title(adjectives[rand.Intn(len(adjectives))])) //nolint:gosec
	rand.Seed(time.Now().UnixNano())
	str = append(str, strings.Title(nouns[rand.Intn(len(nouns))])) //nolint:gosec
	rand.Seed(time.Now().UnixNano())
	str = append(str, strconv.Itoa(rand.Intn(100))) //nolint:gosec

	return strings.Join(str, " ")
}

// findFileOrDirectory searches for a file or directory with name `filename`.
// Search starts in `startDir` and will traverse through all parent directories until the file is found,
// root directory is reached or `maxRecursiveIteration` is exceeded.
func findFileOrDirectory(startDir, fileDir, filename string) (string, bool) {
	i := 0
	for i < maxRecursiveIteration {
		if fileExists(filepath.Join(startDir, fileDir, filename)) {
			return filepath.Join(startDir, fileDir, filename), true
		}

		startDir = filepath.Clean(filepath.Join(startDir, ".."))
		if startDir == "." || startDir == "/" || driveLetterRegex.MatchString(startDir) {
			return "", false
		}

		i++
	}

	log.Warnf("didn't find %s after %d iterations", filename, maxRecursiveIteration)

	return "", false
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
