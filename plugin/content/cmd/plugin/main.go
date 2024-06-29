//go:build plugin

package main

import (
	"context"
	"ncobase/common/log"
	"ncobase/plugin/content/cmd"
)

// Plugin is the exported symbol that will be looked up by the plugin loader
var Plugin cmd.Plugin

func init() {
	log.Infof(context.Background(), "%s plugin initialized", Plugin.Name())
}

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
