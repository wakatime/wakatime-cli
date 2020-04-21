package configs

// ConfigFileMock ConfigFileMock
type ConfigFileMock struct{}

// Get Get an option value for a given section
func (cf ConfigFileMock) Get(section string, key string) (string, error) {
	switch key {
	case "api_key", "apikey":
		return "167fc48a-fe6f-4893-8621-90dc1489fbe4", nil
	case "ignore":
		return "file1.txt\n*.log", nil
	case "exclude":
		return ".data\n*.inf\nlog*.log", nil
	case "include":
		return ".waka\n*.abc", nil
	case "include_only_with_project_file":
		return "true", nil
	case "exclude_unknown_project":
		return "true", nil
	case "offline":
		return "true", nil
	case "proxy":
		return "https://waka:time@domain.be:8080", nil
	case "no_ssl_verify":
		return "true", nil
	case "verbose":
		return "true", nil
	case "log_file":
		return "~/f/w/.wakatime.log", nil
	case "api_url":
		return "https://proxy.wakatime.com/api/v1", nil
	case "timeout":
		return "30", nil
	default:
		return key, nil
	}
}

// GetSectionMap Get a map for a given section
func (cf ConfigFileMock) GetSectionMap(section string) (map[string]string, error) {
	return map[string]string{}, nil
}

// Set Set an option
func (cf ConfigFileMock) Set(section string, keyValue map[string]string) []string {
	return []string{}
}

// GetConfigForPlugin Get config for specific plugin
func (cf ConfigFileMock) GetConfigForPlugin(pluginName string) map[string]string {
	return map[string]string{}
}
