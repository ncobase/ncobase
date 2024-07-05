//go:build plugin

package main

import (
	"ncobase/plugin/asset"
)

// PluginInstance is the exported symbol that will be looked up by the plugin loader
var PluginInstance asset.Plugin

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
