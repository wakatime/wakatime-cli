package configs

import (
	"fmt"
)

// ConfigFileMockFail ConfigFileMock
type ConfigFileMockFail struct{}

// Get Get an option value for a given section
func (cf ConfigFileMockFail) Get(section string, key string) (string, error) {
	switch key {
	case "api_key", "apikey":
		return "167fc48a-fe6f-4893-8621-90dc1489fbe4", nil
	case "hostname":
		return "", fmt.Errorf("failed")
	case "ignore":
		return "", fmt.Errorf("failed")
	case "exclude":
		return "", fmt.Errorf("failed")
	case "include_only_with_project_file":
		return "fail", nil
	case "exclude_unknown_project":
		return "fail", nil
	case "offline":
		return "fail", nil
	case "proxy":
		return "", fmt.Errorf("failed")
	case "no_ssl_verify":
		return "fail", nil
	case "ssl_certs_file":
		return "", fmt.Errorf("failed")
	case "api_url":
		return "", fmt.Errorf("failed")
	default:
		return key, nil
	}
}

// GetSectionMap Get a map for a given section
func (cf ConfigFileMockFail) GetSectionMap(section string) (map[string]string, error) {
	return map[string]string{}, nil
}

// Set Set an option
func (cf ConfigFileMockFail) Set(section string, keyValue map[string]string) []string {
	return []string{}
}

// GetConfigForPlugin Get config for specific plugin
func (cf ConfigFileMockFail) GetConfigForPlugin(pluginName string) map[string]string {
	return map[string]string{}
}
