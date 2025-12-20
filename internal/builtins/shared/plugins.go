/*
 *
 * RR2 - internal/builtins/shared/plugins.go
 *
 */

package shared

type LoadedPlugin struct {
	Name      string
	Version   string
	Path      string
	Functions []string
	Constants []string
}

var loadedPlugins = make(map[string]*LoadedPlugin)

func RegisterPlugin(plugin *LoadedPlugin) {
	loadedPlugins[plugin.Name] = plugin
}

func GetPlugin(name string) (*LoadedPlugin, bool) {
	plugin, exists := loadedPlugins[name]

	return plugin, exists
}

func GetAllPlugins() []*LoadedPlugin {
	plugins := make([]*LoadedPlugin, 0, len(loadedPlugins))

	for _, plugin := range loadedPlugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}
