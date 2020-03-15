package configs

// GetConfigForPlugin GetConfigForPlugin
func (cf ConfigFile) GetConfigForPlugin(pluginName string) map[string]string {
	values, _ := cf.GetSectionMap(pluginName)
	return values
}
