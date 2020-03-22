package scm

import (
	"log"
)

// GetScmPlugin GetProjectPlugin
func GetScmPlugin(scmType string, entity string, configItems map[string]string) Scm {
	switch scmType {
	case "git":
		return &Git{
			Entity:      entity,
			ConfigItems: configItems,
		}
	default:
		log.Printf("Scm plugin of type '%s' is not implemented.", scmType)
		return nil
	}
}
