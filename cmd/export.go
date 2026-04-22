package cmd

import (
	"fmt"
	"os"

	"dailytrack/db"

	"github.com/spf13/cobra"
)

var exportFormat string

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export config and entries to stdout",
	Long: `Dump all entries and (for JSON) the current config to stdout.
Redirect to a file to save a backup:

  gotrack export --format json > backup.json
  gotrack export --format csv  > entries.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch exportFormat {
		case "json":
			return db.ExportJSON(os.Stdout)
		case "csv":
			return db.ExportCSV(os.Stdout)
		}
		return fmt.Errorf("unknown --format %q (want json|csv)", exportFormat)
	},
}

func init() {
	exportCmd.Flags().StringVar(&exportFormat, "format", "json", "export format: json or csv")
	rootCmd.AddCommand(exportCmd)
}
