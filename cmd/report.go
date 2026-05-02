package cmd

import (
	"fmt"
	"time"

	"dailytrack/db"
	"dailytrack/integrations"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Write a weekly Markdown report to your Obsidian vault",
	Long: `Generates a weekly summary Markdown file (week_YYYY-WW.md) and writes it
to the configured Obsidian vault folder.

Requires obsidian.enabled=true and obsidian.vault_path set in config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := db.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg == nil {
			return fmt.Errorf("no config found — run gotrack to complete setup first")
		}
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("init db: %w", err)
		}
		entries, err := db.GetAllEntries()
		if err != nil {
			return fmt.Errorf("load entries: %w", err)
		}
		path, err := integrations.WriteWeeklyReport(cfg, entries, time.Now())
		if err != nil {
			return err
		}
		fmt.Printf("Weekly report written: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
