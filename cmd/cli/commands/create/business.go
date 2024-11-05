package create

import (
	"ncobase/cmd/cli/generator"

	"github.com/spf13/cobra"
)

func newBusinessCommand() *cobra.Command {
	opts := &generator.Options{}

	cmd := &cobra.Command{
		Use:     "business [name]",
		Aliases: []string{"b"},
		Short:   "Create a new business module",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Name = args[0]
			opts.Type = "module"
			opts.Group = cmd.Flag("group").Value.String()
			return generator.Generate(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.UseEnt, "use-ent", false, "use Ent as ORM")
	cmd.Flags().BoolVar(&opts.WithTest, "with-test", false, "generate test files")
	cmd.Flags().StringVar(&opts.Group, "group", "", "module belongs domain group (optional)")

	return cmd
}
