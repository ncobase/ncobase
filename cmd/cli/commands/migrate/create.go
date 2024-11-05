package migrate

import (
	"fmt"
	"ncobase/cmd/cli/utils"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			timestamp := time.Now().Format("20060102150405")
			filename := fmt.Sprintf("%s_%s.sql", timestamp, name)

			content := fmt.Sprintf(`-- migrate:up

-- TODO: Add your UP migration SQL here

-- migrate:down

-- TODO: Add your DOWN migration SQL here
`)

			migrationDir := "migrations"
			if err := utils.EnsureDir(migrationDir); err != nil {
				return fmt.Errorf("failed to create migrations directory: %v", err)
			}

			path := filepath.Join(migrationDir, filename)
			if err := utils.WriteTemplateFile(path, content, nil); err != nil {
				return fmt.Errorf("failed to create migration file: %v", err)
			}

			fmt.Printf("Created migration file: %s\n", path)
			return nil
		},
	}
}
