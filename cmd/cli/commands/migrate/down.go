package migrate

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDownCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Rollback the last migration",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement migration rollback
			fmt.Println("Rolling back last migration...")
			return nil
		},
	}
}
