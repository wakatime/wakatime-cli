package project

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/slongfield/pyfmt"
	"github.com/yookoala/realpath"
)

// Map contains map data.
type Map struct {
	Filepath string
	Patterns []MapPattern
}

// Detect use the ~/.wakatime.cfg file to set custom project names by matching files
// with regex patterns. Project maps go under the [projectmap] config section.
//
// For example:
// 	[projectmap]
//	/home/user/projects/foo = new project name
//	/home/user/projects/bar(\d+)/ = project{0}
//
// Will result in file '/home/user/projects/foo/src/main.c' to have
// project name 'new project name' and file '/home/user/projects/bar42/main.c'
// to have project name 'project42'.
func (m Map) Detect() (Result, bool, error) {
	log.Debugln("execute map project detection")

	if len(m.Patterns) == 0 {
		return Result{}, false, nil
	}

	result, ok, err := matchPattern(m.Filepath, m.Patterns)
	if err != nil {
		return Result{}, false,
			Err(fmt.Sprintf("error matching pattern: %s", err))
	} else if !ok {
		return Result{}, false, nil
	}

	return Result{
		Project: result,
	}, true, nil
}

// matchPattern matches regex against entity's path to find project name.
func matchPattern(fp string, patterns []MapPattern) (string, bool, error) {
	fp, err := realpath.Realpath(fp)
	if err != nil {
		return "", false,
			Err(fmt.Errorf("failed to get the real path: %w", err).Error())
	}

	for _, pattern := range patterns {
		if pattern.Regex.MatchString(fp) {
			matches := pattern.Regex.FindStringSubmatch(fp)
			if len(matches) > 0 {
				params := make([]interface{}, len(matches[1:]))
				for i, v := range matches[1:] {
					params[i] = v
				}

				result, err := pyfmt.Fmt(pattern.Name, params...)

				if err != nil {
					log.Errorf("error formatting %q: %s", pattern.Name, err)
					continue
				}

				return result, true, nil
			}
		}
	}

	return "", false, nil
}

// String returns its name.
func (m Map) String() string {
	return "project-map-detector"
}
