package main

import (
	"fmt"
	"ncobase/cmd/cli/extension"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nco",
	Short: "Ncobase CLI tool",
	Long:  `A CLI tool for managing Ncobase.`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(extension.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
