package migrate

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "migrate",
		Args:    cobra.NoArgs,
		Aliases: []string{"m"},
		Short:   "Database migration commands",
		Long:    `Manage database migrations.`,
	}

	cmd.AddCommand(
		newUpCommand(),
		newDownCommand(),
		newCreateCommand(),
	)

	return cmd
}
