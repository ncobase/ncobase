package create

import (
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"gen", "generate"},
		Short:   "Generate new extension components",
		Long:    `Generate new extensions(core, business or plugin).`,
	}

	cmd.AddCommand(
		newCoreCommand(),
		newBusinessCommand(),
		newPluginCommand(),
	)

	return cmd
}
