package project

import (
	"log"
	"path/filepath"
	"regexp"

	"github.com/slongfield/pyfmt"
)

// ProjectMap Use the ~/.wakatime.cfg file to set custom project names by matching files
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
type ProjectMap struct {
	Entity      string
	ConfigItems map[string]string
	Name        *string
}

// Process Process
func (p ProjectMap) Process() bool {
	name := p.findProject()

	if name != nil {
		p.Name = name
		return true
	}
	return false
}

func (p ProjectMap) findProject() *string {
	if absPath, _ := filepath.Abs(p.Entity); len(absPath) > 0 {
		for pattern, newProjName := range p.ConfigItems {
			re, err := regexp.Compile("(?i)" + pattern)
			if err != nil {
				log.Printf("Regex error (%s) for projectmap pattern: %s", err, pattern)
				continue
			}

			if re.MatchString(p.Entity) {
				groups := re.SubexpNames()
				str, err := pyfmt.Fmt(newProjName, groups)
				if err != nil {
					log.Printf("Regex error (%s) for projectmap pattern: %s", err, pattern)
					continue
				}
				return &str
			}
		}
	}

	return nil
}

// ProjectName ProjectName
func (p ProjectMap) ProjectName() *string {
	return p.Name
}

// BranchName BranchName
func (p ProjectMap) BranchName() *string {
	return nil
}
