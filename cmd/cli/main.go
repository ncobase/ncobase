package main

import (
	"log"
	"ncobase/cmd/cli/commands"
	"os"
)

func main() {
	if err := commands.Execute(); err != nil {
		log.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
