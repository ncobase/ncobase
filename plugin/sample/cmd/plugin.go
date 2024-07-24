//go:build plugin

package main

import "ncobase/plugin/sample"

// Instance is the exported symbol that will be looked up by the plugin loader
var Instance sample.Plugin

// This main function is required for the Go compiler,
// but it won't be executed when the plugin is loaded
func main() {}
