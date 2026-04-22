package cmd

import (
	"fmt"
	"strings"

	"dailytrack/db"
	"dailytrack/models"
	"dailytrack/tui"

	"github.com/spf13/cobra"
)

var (
	initPackName string
	initForce    bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Bootstrap config.json from a preset pack (non-interactive)",
	Long: `Write a fresh config.json populated from a named preset pack,
skipping the TUI setup wizard. Use --force to overwrite an existing config.

Examples:
  gotrack init --pack Power
  gotrack init --pack Focus --force
  gotrack init --list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		listOnly, _ := cmd.Flags().GetBool("list")
		if listOnly {
			fmt.Println("Available packs:")
			for _, p := range tui.Packs {
				fmt.Printf("  %-10s  %s\n", p.Name, p.Description)
			}
			return nil
		}

		if strings.TrimSpace(initPackName) == "" {
			return fmt.Errorf("--pack is required (use --list to see options)")
		}
		pack := tui.PackByName(initPackName)
		if pack == nil {
			// case-insensitive fallback
			for i := range tui.Packs {
				if strings.EqualFold(tui.Packs[i].Name, initPackName) {
					pack = &tui.Packs[i]
					break
				}
			}
		}
		if pack == nil {
			names := make([]string, 0, len(tui.Packs))
			for _, p := range tui.Packs {
				names = append(names, p.Name)
			}
			return fmt.Errorf("unknown pack %q; try one of: %s", initPackName, strings.Join(names, ", "))
		}

		existing, _ := db.LoadConfig()
		if existing != nil && existing.SetupComplete && !initForce {
			return fmt.Errorf("config.json already exists; pass --force to overwrite")
		}

		cfg := &models.Config{
			SetupComplete: true,
			App:           models.DefaultAppSettings(),
			Categories:    pack.Build(),
		}
		if strings.EqualFold(pack.Name, "Power") {
			cfg.NonNegotiables = tui.PowerNonNegotiables()
		}
		for i := range cfg.Categories {
			cfg.Categories[i].Order = i
		}
		models.NormalizeConfig(cfg)

		if err := db.SaveConfig(cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}
		if err := db.InitDB(); err != nil {
			return fmt.Errorf("init database: %w", err)
		}
		fmt.Printf("Wrote config with %d categories from pack %q.\n", len(cfg.Categories), pack.Name)
		return nil
	},
}

func init() {
	initCmd.Flags().StringVar(&initPackName, "pack", "", "preset pack name (see --list)")
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing config")
	initCmd.Flags().Bool("list", false, "list available packs and exit")
	rootCmd.AddCommand(initCmd)
}
