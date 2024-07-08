//go:build plugin

package main

import "ncobase/feature/content"

// Instance is the exported symbol that will be looked up by the plugin loader
var Instance content.Plugin

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
