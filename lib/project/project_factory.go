package project

import (
	"log"
)

// GetProjectPlugin GetProjectPlugin
func GetProjectPlugin(projectType string, entity string, configItems map[string]string) Project {
	switch projectType {
	case "file":
		return &ProjectFile{
			Entity: entity,
		}
	case "map":
		return &ProjectMap{
			Entity:      entity,
			ConfigItems: configItems,
		}
	case "git":
		return &Git{
			Entity:      entity,
			ConfigItems: configItems,
		}
	default:
		log.Printf("Project plugin of type '%s' is not implemented.", projectType)
		return nil
	}
}
