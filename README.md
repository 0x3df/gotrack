<div align="center">
  <pre>
  ______     ______     ______   ______     ______     ______     __  __
  /\  ___\   /\  __ \   /\__  _\ /\  == \   /\  __ \   /\  ___\   /\ \/ /
  \ \ \__ \  \ \ \/\ \  \/_/\ \/ \ \  __<   \ \  __ \  \ \ \____  \ \  _"-.
   \ \_____\  \ \_____\    \ \_\  \ \_\ \_\  \ \_\ \_\  \ \_____\  \ \_\ \_\
    \/_____/   \/_____/     \/_/   \/_/ /_/   \/_/\/_/   \/_____/   \/_/\/_/
  </pre>
  <h1>GoTrack</h1>
  <p><strong>Local-first terminal dashboard for daily tracking, trends, and personal insights.</strong></p>
</div>

<div align="center">
  <a href="https://github.com/0x3df/gotrack"><img src="https://img.shields.io/badge/status-active-22c55e?style=for-the-badge" alt="Status" /></a>
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/language-go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/tui-bubble%20tea-ff75b5?style=for-the-badge" alt="Bubble Tea" /></a>
  <a href="https://modernc.org/sqlite"><img src="https://img.shields.io/badge/storage-sqlite-0f766e?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite" /></a>
  <a href="https://github.com/guptarohit/asciigraph"><img src="https://img.shields.io/badge/charts-asciigraph-111827?style=for-the-badge" alt="Asciigraph" /></a>
  <a href="https://github.com/spf13/cobra"><img src="https://img.shields.io/badge/cli-cobra-2563eb?style=for-the-badge" alt="Cobra" /></a>
</div>

## Highlights

- **Local-first:** all data stays on your machine in a SQLite database.
- **Dynamic configuration:** build your own tracking system with custom categories and trackers for binary, duration, numeric, rating, and text inputs.
- **Obsidian export:** optionally mirror dated entries into Markdown notes inside your vault.
- **Theme support:** switch between GoTrack, Catppuccin, and Nord from settings.
- **Rich TUI:** full-screen dashboard with heatmaps, rolling trend lines, vertical bar charts, and simple correlation views.
- **Ambient mode:** optional omnidirectional starfield background for the dashboard.
- **Portable:** zero external runtime dependencies outside the Go ecosystem.
- **MCP for AI tools:** run `gotrack mcp` to attach Cursor, Claude Desktop, or other MCP clients to your local workspace ([docs/AI_AGENTS.md](docs/AI_AGENTS.md)).

## Installation

### Option 1: `go install` (Recommended)

```bash
go install github.com/0x3df/gotrack@latest
```

This installs the `gotrack` binary to your Go bin directory (typically `$HOME/go/bin`).

### Option 2: Build from source

```bash
git clone https://github.com/0x3df/gotrack.git
cd gotrack
go build -o gotrack .
./gotrack
```

## First Launch

On first launch, GoTrack walks you through a setup wizard so you can:

1. Choose a data directory for your workspace, config, and database.
2. Pick a setup mode: guided defaults or custom from scratch.
3. Define the categories and trackers you want on the dashboard.

## Controls

- `a`: add or edit an entry for any date
- `x`: quick entry for one tracker value today
- `p`: start a pomodoro and allocate elapsed minutes to a duration tracker
- `d`: delete the entry for a given date
- `esc`: cancel date/entry form and return to dashboard
- `s`: open settings and tracking setup
- `s` settings also manages theme, Obsidian export, and starfield background
- `h` / `l` or `←` / `→`: switch tabs
- `j` / `k` or `↓` / `↑`: scroll visualizations
- `[` / `]`: cycle between overview hero visuals (Yearly Pulse, Tracker Wall, Weekday Rhythm, Momentum Podium)
- `?`: open the keybinds popup window
- `q`: quit

Date prompts accept shortcuts: `t` / `today`, `y` / `yesterday`, or `-N` (e.g. `-3` for three days ago).

### Logging from your phone

You can log from iOS or another machine using the `gotrack log` CLI together with a sync tool (Syncthing / iCloud) or an SSH-based iOS Shortcut. See [docs/REMOTE_LOGGING.md](docs/REMOTE_LOGGING.md) for setup recipes.

## Data Persistence

- GoTrack stores your database and config in the workspace you chose on first launch.
- Rebuilding or replacing the `gotrack` binary does not delete this workspace or wipe your data.

## FAQ

Common questions, troubleshooting, and Review-tab behavior: **[docs/FAQ.md](docs/FAQ.md)**.  
AI assistants (MCP): **[docs/AI_AGENTS.md](docs/AI_AGENTS.md)**.

## Stack

- **Go**
- **Bubble Tea / Lipgloss / Huh**
- **SQLite** via `modernc.org/sqlite`
- **Asciigraph**
- **Cobra**
- **MCP** ([modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)) for `gotrack mcp`
