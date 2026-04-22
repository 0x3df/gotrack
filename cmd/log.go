package cmd

import (
	"fmt"
	"os"
	"strings"

	"dailytrack/db"

	"github.com/spf13/cobra"
)

var logDate string

var logCmd = &cobra.Command{
	Use:   "log key=value [key=value ...]",
	Short: "Quick-log an entry without opening the TUI",
	Long: `Upsert an entry for today (or --date) by passing tracker values as
key=value pairs. Tracker names match case-insensitively against your config.

Duration trackers are stored in their configured unit (minutes for the
Power pack). Pass raw numbers.

Examples:
  gotrack log code=true sleep=7.5
  gotrack log rating=4 --date yesterday
  gotrack log "C++ Study=45" "DS&A / LeetCode=30" "Diet On Track=true"
  gotrack log "Deep Work=90" "Main Win=shipped feature"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := db.LoadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if cfg == nil || !cfg.SetupComplete {
			return fmt.Errorf("gotrack is not set up yet — run `gotrack` first")
		}

		date, err := db.NormalizeDate(logDate)
		if err != nil {
			return fmt.Errorf("invalid --date: %w", err)
		}

		values := map[string]interface{}{}
		for _, a := range args {
			eq := strings.Index(a, "=")
			if eq <= 0 {
				return fmt.Errorf("expected key=value, got %q", a)
			}
			key := strings.TrimSpace(a[:eq])
			raw := strings.TrimSpace(a[eq+1:])
			values[key] = raw
		}

		if err := db.UpsertEntryLog(cfg, date, values); err != nil {
			return fmt.Errorf("save: %w", err)
		}
		after, err := db.GetEntryForDate(date)
		if err != nil {
			return fmt.Errorf("read back entry: %w", err)
		}
		n := 0
		if after != nil && after.Data != nil {
			n = len(after.Data)
		}
		fmt.Fprintf(os.Stdout, "logged %s (%d fields)\n", date, n)
		return nil
	},
}

func init() {
	logCmd.Flags().StringVar(&logDate, "date", "today", "date to log (YYYY-MM-DD, t, y, or -N)")
	rootCmd.AddCommand(logCmd)
}
