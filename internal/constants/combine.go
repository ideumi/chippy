/*
 *
 * RR2 - internal/constants/combine.go
 *
 */

package constants

const (
	// Combine tool defaults

	COMBINE_DEFAULT_FILENAME = "combine.chp"
	COMBINE_SHEBANG          = "#!/usr/bin/chippy"

	// File permissions

	FILE_PERM_READABLE   = 0644
	FILE_PERM_EXECUTABLE = 0755

	// Config variable names

	CONFIG_PROJECT          = "Project"
	CONFIG_VERSION          = "Version"
	CONFIG_LICENCE          = "Licence"
	CONFIG_OUTPUT           = "Output"
	CONFIG_SOURCE           = "Source"
	CONFIG_PATHS            = "Paths"
	CONFIG_PLUGIN_PATHS     = "PluginPaths"
	CONFIG_ASSETS           = "Assets"
	CONFIG_STRIP_COMMENTS   = "StripComments"
	CONFIG_STRIP_WHITESPACE = "StripWhitespace"
	CONFIG_EXTERNAL         = "External"
	CONFIG_EXTERNAL_PLUGINS = "ExternalPlugins"
	CONFIG_PRODUCE_BUNDLE   = "ProduceBundle"

	COMBINE_TEMPLATE = `
#
# Combine configuration file
#

# Metadata

var Project = "myprogram";
var Version = "1.0.0";
var Licence = "licence.txt";
var Output = "myprogram";

# Entry point

var Source = "main.chp";

# Create the program as a CHIPBIN bundle
# This must be true if you are using plugins

var ProduceBundle = true;

# if ProduceBundle == true
# {
#	Output = Output + "-" + CHIPOS + "-" + CHIPAR;
# }

# Search paths for dependencies

var Paths = [];
Paths = Paths + ["src/lib"];

var PluginPaths = [];
PluginPaths = PluginPaths + ["src/lib/plugins"];

# Add user-local library path if HOME is available

var homePath = getenv("HOME");

if homePath != err
{
    Paths = Paths + [homePath + "/.local/share/chiplang/lib"];
    PluginPaths = PluginPaths + [homePath + "/.local/share/chiplang/lib/plugins"];
}

Paths = Paths + ["/usr/lib/chiplang"];
PluginPaths = PluginPaths + ["/usr/lib/chiplang/plugins"];

# Assets to bundle, requires ProduceBundle be true

var Assets = [];

# Dependencies combine shouldn't bundle and just keep as loads

var External = [];

# Plugins combine shouldn't bundle and just keep as ploads
# If you wish to produce a text bundle that retains pload calls, use this

var ExternalPlugins = [];

# Stylistic Options

var StripComments = true;
var StripWhitespace = true;
`
)
