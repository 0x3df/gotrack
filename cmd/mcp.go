package cmd

import (
	"context"
	"os"

	"dailytrack/mcpconn"

	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run the Model Context Protocol server (stdio) for AI assistants",
	Long: `Runs an MCP server on stdin/stdout so Cursor, Claude Desktop, and other
clients can read your schema, query bounded entry ranges, fetch per-tracker
insights, and upsert daily values (same semantics as gotrack log).

Do not wrap stdout — the MCP JSON-RPC stream must be pristine. Log diagnostics
to stderr only.

See docs/AI_AGENTS.md for client configuration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := mcpconn.Run(context.Background()); err != nil {
			return err
		}
		return nil
	},
	SilenceUsage: true,
}

func init() {
	// MCP owns stdout; route cobra errors to stderr so clients are not polluted.
	mcpCmd.SetOut(os.Stderr)
	mcpCmd.SetErr(os.Stderr)
	rootCmd.AddCommand(mcpCmd)
}
