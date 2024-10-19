package main

import (
	"fmt"
	"os"

	"ncobase/cmd/cli/feature"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nco",
	Short: "Ncobase CLI tool",
	Long:  `A CLI tool for managing Ncobase.`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(feature.Cmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
