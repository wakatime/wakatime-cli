package project

import (
	"strings"

	"github.com/wakatime/wakatime-cli/lib/configs"
	"github.com/wakatime/wakatime-cli/lib/utils"
)

// Project Project interface
type Project interface {
	Process() bool
	ProjectName() *string
	BranchName() *string
	SetEntity(string)
	SetConfigItems(map[string]string)
}

// ProjectInfo ProjectInfo
type ProjectInfo struct {
	Entity           string
	EntityType       string
	Project          string
	AlternateProject string
	HideProjectNames []string
	Branch           string
}

var (
	configPlugins map[string]Project = map[string]Project{
		"file": ProjectFile{},
		"map":  ProjectMap{},
	}
	revControlPlugins map[string]Project = map[string]Project{
		"git":         Git{},
		"mercurial":   Mercurial{},
		"subversions": Subversion{},
	}
)

// GetProjectInfo Find the current project and branch.
// First looks for a .wakatime-project file. Second, uses the --project arg.
// Third, uses the folder name from a revision control repository. Last, uses
// the --alternate-project arg.
func GetProjectInfo(pi ProjectInfo, cfg configs.WakaTimeConfig) (string, string) {
	projectName, branchName := strings.TrimSpace(pi.Project), strings.TrimSpace(pi.Branch)
	if pi.EntityType != "file" {
		if len(projectName) == 0 {
			projectName = pi.AlternateProject
		}
		return projectName, branchName
	}

	if len(projectName) == 0 || len(branchName) == 0 {
		for key, p := range configPlugins {
			pluginConfigs := cfg.GetConfigForPlugin(key)
			p.SetEntity(pi.Entity)
			p.SetConfigItems(pluginConfigs)

			if p.Process() {
				projectName = *p.ProjectName()
				branchName = *p.BranchName()
				break
			}
		}
	}

	hideProject := utils.ShouldObfuscateProject(pi.Entity, pi.HideProjectNames)

	if len(projectName) == 0 || len(branchName) == 0 {
		for key, p := range revControlPlugins {
			pluginConfigs := cfg.GetConfigForPlugin(key)
			p.SetEntity(pi.Entity)
			p.SetConfigItems(pluginConfigs)

			if p.Process() {
				if !hideProject {
					projectName = *p.ProjectName()
				}
				branchName = *p.BranchName()
				break
			}
		}
	}

	if len(projectName) == 0 {
		if !hideProject {
			projectName = pi.AlternateProject
		} else {
			projectName = utils.GenerateProjectName()
			//todo: https://github.com/wakatime/wakatime/blob/master/wakatime/project.py#L92
		}
	}

	return projectName, branchName
}
