package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Rana718/Graft/internal/config"
	"github.com/Rana718/Graft/internal/migrator"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [name]",
	Short: "Create a new migration",
	Long: `Create a new migration file with the specified name.
	If no name is provided, you will be prompted to enter one.

	The migration file will include:
	- Timestamp and migration name header
	- Up migration section (forward changes)
	- Down migration section (rollback changes)
	- Auto-generated SQL based on schema differences (if --auto flag is used)

	Examples:
	  graft migrate "create users table"
	  graft migrate "add email index" --auto
	  graft migrate --empty "custom migration"
	  graft migrate  # Interactive mode`,

	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid config: %w", err)
		}

		if err := cfg.EnsureDirectories(); err != nil {
			return fmt.Errorf("failed to create directories: %w", err)
		}

		var migrationName string
		if len(args) > 0 {
			migrationName = strings.Join(args, " ")
		} else {
			fmt.Print("Enter migration name: ")
			reader := bufio.NewReader(os.Stdin)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			migrationName = strings.TrimSpace(input)
		}

		if migrationName == "" {
			return fmt.Errorf("migration name cannot be empty")
		}

		ctx := context.Background()

		m, err := migrator.NewMigrator(cfg)
		if err != nil {
			return fmt.Errorf("failed to create migrator: %w", err)
		}
		defer m.Close()

		empty, _ := cmd.Flags().GetBool("empty")

		if empty {
			// Force empty migration template
			if err := m.GenerateEmptyMigration(ctx, migrationName); err != nil {
				return err
			}
		} else {
			// Default behavior - try schema diff, fallback to empty
			if err := m.GenerateMigration(ctx, migrationName, cfg.SchemaPath); err != nil {
				return err
			}
		}

		fmt.Println("✅ Migration generated successfully")
		fmt.Println("📝 Edit the migration file to add your SQL statements")
		fmt.Println("💡 Run 'graft apply' to apply the migration")
		fmt.Println("🔧 Run 'graft gen' to generate SQLC types after applying migrations")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	
	migrateCmd.Flags().BoolP("empty", "e", false, "Create an empty migration template without schema diff")
}
