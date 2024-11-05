package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement migration execution
			fmt.Println("Running migrations...")
			return nil
		},
	}
}
