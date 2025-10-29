package cmd

import (
	"fmt"
	"os"

	"github.com/Rana718/Graft/internal/config"
	"github.com/Rana718/Graft/internal/raft"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate SQL from .raft schema file",
	Long:  `Converts your schema.raft file into SQL schema and creates a migration`,
	Run:   runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		color.Red("❌ Failed to load config: %v", err)
		os.Exit(1)
	}

	raftFile := "schema.raft"
	if _, err := os.Stat(raftFile); os.IsNotExist(err) {
		color.Red("❌ schema.raft file not found")
		color.Yellow("💡 Create a schema.raft file with your models")
		os.Exit(1)
	}

	color.Cyan("📖 Parsing schema.raft...")
	schema, err := raft.ParseRaftFile(raftFile)
	if err != nil {
		color.Red("❌ Failed to parse schema: %v", err)
		os.Exit(1)
	}

	if len(schema.Models) == 0 {
		color.Yellow("⚠️  No models found in schema.raft")
		os.Exit(0)
	}

	color.Cyan("🔨 Generating SQL for %d model(s)...", len(schema.Models))
	sql, err := raft.GenerateSQL(schema, cfg.Database.Provider)
	if err != nil {
		color.Red("❌ Failed to generate SQL: %v", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(cfg.GetSchemaDir(), 0755); err != nil {
		color.Red("❌ Failed to create schema directory: %v", err)
		os.Exit(1)
	}

	if err := os.WriteFile(cfg.SchemaPath, []byte(sql), 0644); err != nil {
		color.Red("❌ Failed to write schema file: %v", err)
		os.Exit(1)
	}

	color.Green("✅ Generated SQL schema at %s", cfg.SchemaPath)
	color.Cyan("\n📝 Next steps:")
	color.White("  1. Review the generated SQL in %s", cfg.SchemaPath)
	color.White("  2. Run 'graft migrate \"initial schema\"' to create migration")
	color.White("  3. Run 'graft apply' to apply the migration")
}
