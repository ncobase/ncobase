package create

import (
	"ncobase/cmd/cli/generator"

	"github.com/spf13/cobra"
)

func newPluginCommand() *cobra.Command {
	opts := &generator.Options{}

	cmd := &cobra.Command{
		Use:     "plugin [name]",
		Aliases: []string{"p"},
		Short:   "Create a new plugin",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Type = "plugin"
			return generator.Generate(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.UseEnt, "use-ent", false, "use Ent as ORM")
	cmd.Flags().BoolVar(&opts.UseGorm, "use-gorm", false, "use Gorm as ORM")
	cmd.Flags().BoolVar(&opts.WithTest, "with-test", false, "generate test files")
	cmd.Flags().StringVar(&opts.Group, "group", "", "plugin belongs domain group (optional)")

	return cmd
}
