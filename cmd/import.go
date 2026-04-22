package cmd

import (
	"fmt"
	"os"

	"dailytrack/db"

	"github.com/spf13/cobra"
)

var (
	importDryRun          bool
	importOverwriteConfig bool
)

var importCmd = &cobra.Command{
	Use:   "import <file.json>",
	Short: "Import entries (and optionally config) from a JSON export",
	Long: `Read a JSON bundle produced by "gotrack export --format json" and
upsert its entries into the database. By default the on-disk config is
left alone — pass --overwrite-config to replace it with the imported one.

  gotrack import backup.json --dry-run
  gotrack import backup.json --overwrite-config`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		n, err := db.ImportJSON(f, importOverwriteConfig, importDryRun)
		if err != nil {
			return err
		}
		if importDryRun {
			fmt.Printf("dry-run: would import %d entries\n", n)
			return nil
		}
		fmt.Printf("imported %d entries\n", n)
		return nil
	},
}

func init() {
	importCmd.Flags().BoolVar(&importDryRun, "dry-run", false, "report what would be imported without writing")
	importCmd.Flags().BoolVar(&importOverwriteConfig, "overwrite-config", false, "also replace config.json with the imported config")
	rootCmd.AddCommand(importCmd)
}
