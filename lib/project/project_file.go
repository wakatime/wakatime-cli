package project

import (
	"log"
	"strings"

	"github.com/wakatime/wakatime-cli/lib/utils"
)

// ProjectFile Information from a .wakatime-project file about the project for
// a given file. First line of .wakatime-project sets the project
// name. Second line sets the current branch name.
type ProjectFile struct {
	Entity string
}

var (
	projectFileProjectName *string
	projectFileBranchName  *string
)

// Process Process
func (p ProjectFile) Process() bool {
	projectFile := utils.FindProjectFile(p.Entity)
	if projectFile != nil {
		lines, err := utils.ReadFile(*projectFile)
		if err != nil {
			log.Printf("Error while opening file '%s' (%s)", *projectFile, err)
			return false
		}

		var name *string
		var branch *string

		if utils.Isset(lines, 0) {
			*name = strings.TrimSpace(lines[0])
		}
		if utils.Isset(lines, 1) {
			*branch = strings.TrimSpace(lines[1])
		}

		return true
	}

	return false
}

// ProjectName ProjectName
func (p ProjectFile) ProjectName() *string {
	return projectFileProjectName
}

// BranchName BranchName
func (p ProjectFile) BranchName() *string {
	return projectFileBranchName
}
