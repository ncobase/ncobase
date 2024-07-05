//go:build plugin

package main

import "ncobase/plugin/content"

// PluginInstance is the exported symbol that will be looked up by the plugin loader
var PluginInstance content.Plugin

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
