# AI assistants and GoTrack

GoTrack ships a **local MCP server** (`gotrack mcp`) so tools like Cursor, Claude Desktop, or any MCP-capable client can read your tracker schema, pull bounded entry history, compute per-tracker insights, and upsert daily fields — without a hosted API.

## Requirements

- GoTrack must be initialized once (`gotrack` TUI setup wizard completed).
- The `gotrack` binary must be on `PATH` (or use an absolute path in client config).
- **SQLite**: avoid heavy writes while the full-screen TUI is open on the same machine; the CLI and MCP tools are short transactions, but concurrent writers can still contend.

## Tools exposed

| Tool | Purpose |
|------|---------|
| `gotrack_schema` | Full `config.json` envelope (categories, tracker ids/names/types/units/targets). Call this first in a session. |
| `gotrack_entries` | Entries between `from` and `to` (ISO dates), newest first. Defaults to roughly the last 60 days through today; caps row count (default 500, max 2000). |
| `gotrack_insights` | One tracker: binary streak/consistency; numeric/rating/duration/count series tail, momentum windows, optional target hit rate; text “latest” from newest row. Optional date range; omit both dates for all-time. |
| `gotrack_log` | Upsert `values` for one `date` (same coercion rules as `gotrack log`). Keys are tracker **names** or **UUIDs**; values are JSON scalars or strings. |

## Cursor (example)

Add to your MCP configuration (e.g. Cursor **Settings → MCP** or project `.cursor/mcp.json`), adjusting the binary path if needed:

```json
{
  "mcpServers": {
    "gotrack": {
      "command": "gotrack",
      "args": ["mcp"]
    }
  }
}
```

The server uses **stdio only** (no HTTP). Do not wrap stdout.

## Shell-only agents

You can still drive GoTrack without MCP:

- `gotrack log sleep=7.5 code=true --date yesterday`
- `gotrack export --format json` for full backups (large)

For large histories, prefer MCP `gotrack_entries` with a date range instead of exporting everything.

## Workspace on another path

GoTrack resolves the database via `~/.gotrack_workspace` (see [`db/path.go`](../db/path.go)). If you sync the workspace directory across machines, use the same pointer file or the same absolute workspace path so `gotrack mcp` sees the intended data.
