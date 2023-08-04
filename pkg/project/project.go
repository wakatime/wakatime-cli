package project

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/wakatime/wakatime-cli/pkg/heartbeat"
	"github.com/wakatime/wakatime-cli/pkg/log"
	"github.com/wakatime/wakatime-cli/pkg/regex"
	"github.com/wakatime/wakatime-cli/pkg/windows"

	"github.com/gandarez/go-realpath"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	driveLetterRegex = regexp.MustCompile(`^[a-zA-Z]:\\$`)
)

const (
	// WakaTimeProjectFile is the special file which if present should contain the project name and optional branch name
	// that will be used instead of the auto-detected project and branch names.
	WakaTimeProjectFile = ".wakatime-project"
	// maxRecursiveIteration limits the number of a func will be called recursively.
	maxRecursiveIteration = 500
)

// DetectorID represents a detector ID.
type DetectorID int

const (
	// UnknownDetector is the detector ID used when not detected.
	UnknownDetector DetectorID = iota
	// FileDetector is the detector ID for file detector.
	FileDetector
	// MapDetector is the detector ID for map detector.
	MapDetector
	// GitDetector is the detector ID for git detector.
	GitDetector
	// MercurialDetector is the detector ID for mercurial detector.
	MercurialDetector
	// SubversionDetector is the detector ID for subversion detector.
	SubversionDetector
	// TfvcDetector is the detector ID for tfvc detector.
	TfvcDetector
)

const (
	fileDetectorString       = "project-file-detector"
	mapDetectorString        = "project-map-detector"
	gitDetectorString        = "git-detector"
	mercurialDetectorString  = "mercurial-detector"
	subversionDetectorString = "svn-detector"
	tfvcDetectorString       = "tfvc-detector"
)

// String implements fmt.Stringer interface.
func (d DetectorID) String() string {
	switch d {
	case FileDetector:
		return fileDetectorString
	case MapDetector:
		return mapDetectorString
	case GitDetector:
		return gitDetectorString
	case MercurialDetector:
		return mercurialDetectorString
	case SubversionDetector:
		return subversionDetectorString
	case TfvcDetector:
		return tfvcDetectorString
	default:
		return ""
	}
}

type (
	// Detecter is a common interface for project.
	Detecter interface {
		Detect() (Result, bool, error)
		ID() DetectorID
	}

	// DetecterArg determines for a given path if it needs to run.
	DetecterArg struct {
		Filepath  string
		ShouldRun bool
	}

	// Result contains the result of Detect().
	Result struct {
		Project string
		Branch  string
		Folder  string
	}

	// Config contains project detection configurations.
	Config struct {
		// HideProjectNames determines if the project name should be obfuscated by matching its path.
		HideProjectNames []regex.Regex
		// Patterns contains the overridden project name per path.
		MapPatterns []MapPattern
		// ProjectFromGitRemote when enabled uses the git remote as the project name instead of local git folder.
		ProjectFromGitRemote bool
		// Submodule contains the submodule configurations.
		Submodule Submodule
	}

	// MapPattern contains the project name and regular expression for a specific path.
	MapPattern struct {
		// Name is the project name.
		Name string
		// Regex is the regular expression for a specific path.
		Regex regex.Regex
	}

	// Submodule contains the submodule configurations.
	Submodule struct {
		// DisabledPatterns contains the paths to match against submodules
		// and if matched it will skip the project detection.
		DisabledPatterns []regex.Regex
		// MapPatterns contains the overridden project name per path for submodule.
		MapPatterns []MapPattern
	}
)

// WithDetection finds the current project and branch.
// First looks for a .wakatime-project file or project map. Second, uses the
// --project arg. Third, try to auto-detect using a revision control repository.
// Last, uses the --alternate-project arg.
func WithDetection(config Config) heartbeat.HandleOption {
	return func(next heartbeat.Handle) heartbeat.Handle {
		return func(hh []heartbeat.Heartbeat) ([]heartbeat.Result, error) {
			for n, h := range hh {
				log.Debugln("execute project detection for:", h.Entity)

				// first, use .wakatime-project or [projectmap] section with entity path.
				// Then, detect with project folder. This tries to use the same project name
				// across all IDEs instead of sometimes using alternate project when file is unsaved
				result, detector := Detect(config.MapPatterns,
					DetecterArg{Filepath: h.Entity, ShouldRun: h.EntityType == heartbeat.FileType},
					DetecterArg{Filepath: h.ProjectPathOverride, ShouldRun: true},
				)

				// second, use project override
				if result.Project == "" && h.ProjectOverride != "" {
					result.Project = h.ProjectOverride
					result.Folder = h.ProjectPathOverride
				}

				// third, autodetect with revision control with entity path.
				// Then, autodetect with project folder. This tries to use the same project name
				// across all IDEs instead of sometimes using alternate project when file is unsaved
				if result.Project == "" || result.Branch == "" || result.Folder == "" {
					revControlResult := DetectWithRevControl(
						config.Submodule.DisabledPatterns,
						config.Submodule.MapPatterns,
						config.ProjectFromGitRemote,
						DetecterArg{Filepath: h.Entity, ShouldRun: h.EntityType == heartbeat.FileType},
						DetecterArg{Filepath: h.ProjectPathOverride, ShouldRun: true},
					)

					result.Project = firstNonEmptyString(result.Project, revControlResult.Project)
					result.Branch = firstNonEmptyString(result.Branch, revControlResult.Branch)
					result.Folder = firstNonEmptyString(result.Folder, revControlResult.Folder)
				}

				// fourth, use alternate project
				if result.Project == "" && h.ProjectAlternate != "" {
					result.Project = h.ProjectAlternate
					result.Folder = firstNonEmptyString(h.ProjectPathOverride, result.Folder)
				}

				// fifth, use alternate branch
				if result.Branch == "" && h.BranchAlternate != "" {
					result.Branch = h.BranchAlternate
				}

				// sixth, use project folder found or entity's path
				result.Folder = firstNonEmptyString(result.Folder, h.ProjectPathOverride)

				// seventh, if no folder is found, use entity's directory
				if h.EntityType == heartbeat.FileType && result.Folder == "" {
					result.Folder = filepath.Dir(h.Entity)
				}

				if runtime.GOOS == "windows" && result.Folder != "" {
					result.Folder = windows.FormatFilePath(result.Folder)
				}

				// finally, obfuscate project name if necessary
				if heartbeat.ShouldSanitize(result.Folder, config.HideProjectNames) &&
					result.Project != "" && detector != FileDetector {
					result.Project = obfuscateProjectName(result.Folder)
				}

				result.Folder = FormatProjectFolder(result.Folder)

				// count total subfolders in project's path
				if result.Folder != "" && strings.HasPrefix(h.Entity, result.Folder) {
					subfolders := CountSlashesInProjectFolder(result.Folder)
					if subfolders > 0 {
						hh[n].ProjectRootCount = &subfolders
					}
				}

				hh[n].Project = &result.Project
				hh[n].Branch = &result.Branch
				hh[n].ProjectPath = result.Folder
			}

			return next(hh)
		}
	}
}

// Detect finds the current project and branch from config plugins.
func Detect(patterns []MapPattern, args ...DetecterArg) (Result, DetectorID) {
	for _, arg := range args {
		if !arg.ShouldRun || arg.Filepath == "" {
			continue
		}

		var configPlugins = []Detecter{
			File{
				Filepath: arg.Filepath,
			},
			Map{
				Filepath: arg.Filepath,
				Patterns: patterns,
			},
		}

		for _, p := range configPlugins {
			log.Debugln("execute", p.ID().String())

			result, detected, err := p.Detect()
			if err != nil {
				log.Errorf("unexpected error occurred at %q: %s", p.ID().String(), err)
				continue
			}

			if detected {
				return result, p.ID()
			}
		}
	}

	return Result{}, UnknownDetector
}

// DetectWithRevControl finds the current project and branch from rev control.
func DetectWithRevControl(
	submoduleDisabledPatterns []regex.Regex,
	submoduleProjectMapPatterns []MapPattern,
	projectFromGitRemote bool,
	args ...DetecterArg) Result {
	for _, arg := range args {
		if !arg.ShouldRun || arg.Filepath == "" {
			continue
		}

		var revControlPlugins = []Detecter{
			Git{
				Filepath:                    arg.Filepath,
				ProjectFromGitRemote:        projectFromGitRemote,
				SubmoduleDisabledPatterns:   submoduleDisabledPatterns,
				SubmoduleProjectMapPatterns: submoduleProjectMapPatterns,
			},
			Mercurial{
				Filepath: arg.Filepath,
			},
			Subversion{
				Filepath: arg.Filepath,
			},
			Tfvc{
				Filepath: arg.Filepath,
			},
		}

		for _, p := range revControlPlugins {
			log.Debugln("execute", p.ID().String())

			result, detected, err := p.Detect()
			if err != nil {
				log.Errorf("unexpected error occurred at %q: %s", p.ID().String(), err)
				continue
			}

			if detected {
				return Result{
					Project: result.Project,
					Branch:  result.Branch,
					Folder:  result.Folder,
				}
			}
		}
	}

	return Result{}
}

func obfuscateProjectName(folder string) string {
	project := generateProjectName()

	err := Write(folder, project)
	if err != nil {
		log.Warnf("failed to write: %s", err)
	}

	return project
}

// Write saves wakatime project file.
func Write(folder, project string) error {
	err := os.WriteFile(filepath.Join(folder, WakaTimeProjectFile), []byte(project+"\n"), 0600)
	if err != nil {
		return fmt.Errorf("failed to save wakatime project file: %s", err)
	}

	return nil
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

	c := cases.Title(language.AmericanEnglish)

	r := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec
	str = append(str, c.String(adjectives[r.Intn(len(adjectives))]))
	r = rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec
	str = append(str, c.String(nouns[r.Intn(len(nouns))]))
	r = rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec
	str = append(str, strconv.Itoa(r.Intn(100)))

	return strings.Join(str, " ")
}

// CountSlashesInProjectFolder counts the number of slashes in a folder path.
func CountSlashesInProjectFolder(directory string) int {
	if directory == "" {
		return 0
	}

	directory = windows.FormatFilePath(directory)

	// Add trailing slash if not present.
	if !strings.HasSuffix(directory, `/`) {
		directory += `/`
	}

	return strings.Count(directory, `/`)
}

// FindFileOrDirectory searches for a file or directory named `filename`.
// Search starts in `directory` and will traverse through all parent directories.
// `directory` may also be a file, and in that case will start from the file's directory.
func FindFileOrDirectory(directory, filename string) (string, bool) {
	i := 0
	for i < maxRecursiveIteration {
		if isRootPath(directory) {
			return "", false
		}

		if fileExists(filepath.Join(directory, filename)) {
			return filepath.Join(directory, filename), true
		}

		directory = filepath.Clean(filepath.Join(directory, ".."))

		i++
	}

	log.Warnf("max %d iterations reached without finding %s", maxRecursiveIteration, filename)

	return "", false
}

func isRootPath(directory string) bool {
	return (directory == "" ||
		directory == "." ||
		directory == string(filepath.Separator) ||
		directory == "\\\\wsl$" ||
		driveLetterRegex.MatchString(directory) ||
		directory == filepath.VolumeName(directory) ||
		directory == filepath.Dir(directory))
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

// FormatProjectFolder returns the abs and real path for the given directory path.
func FormatProjectFolder(fp string) string {
	if fp == "" {
		return ""
	}

	if runtime.GOOS == "windows" {
		return windows.FormatFilePath(fp)
	}

	formatted, err := filepath.Abs(fp)
	if err != nil {
		log.Debugf("failed to resolve absolute path for %q: %s", fp, err)
		return formatted
	}

	// evaluate any symlinks
	formatted, err = realpath.Realpath(formatted)
	if err != nil {
		log.Debugf("failed to resolve real path for %q: %s", formatted, err)
	}

	return formatted
}
