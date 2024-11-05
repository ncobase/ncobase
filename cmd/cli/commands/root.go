package commands

import (
	"ncobase/cmd/cli/commands/create"
	"ncobase/cmd/cli/commands/migrate"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nco",
	Short: "Ncobase CLI tool",
	Long:  `A CLI tool for managing Ncobase.`,
}

func init() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.AddCommand(create.NewCommand(), migrate.NewCommand())
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}
