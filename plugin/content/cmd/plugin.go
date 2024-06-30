//go:build plugin

package main

import (
	"ncobase/plugin"
	"ncobase/plugin/content"
)

// PluginInstance is the exported symbol that will be looked up by the plugin loader
var PluginInstance = &plugin.Wrapper{
	Instance: &content.Plugin{},
	Metadata: plugin.Metadata{
		Name:         "content",
		Version:      "1.0.0",
		Dependencies: []string{},
		Description:  "Content management plugin",
	},
}

func init() {}

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
