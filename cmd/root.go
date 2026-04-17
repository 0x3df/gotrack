package cmd

import (
	"fmt"
	"os"

	"dailytrack/db"
	"dailytrack/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gotrack",
	Short: "A local-first daily tracking dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := db.LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
		// If cfg is nil or SetupComplete is false, InitialModel launches setup wizard
		p := tea.NewProgram(tui.InitialModel(cfg), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
