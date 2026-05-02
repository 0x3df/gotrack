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

## Quick Start

1.  **Install:** `go install github.com/0x3df/gotrack@latest`
2.  **Initialize:** Run `gotrack` and follow the guided setup.
3.  **Track:** Press `a` to log your first entry or `x` for a quick entry.
4.  **Visualize:** Use `h`/`l` to browse your dashboard and see trends emerge.

## Controls

GoTrack is designed for speed. Most actions are a single keypress away:

### Navigation
- `h` / `l` or `←` / `→`: Switch between dashboard tabs (Overview, Categories, Insights, Review)
- `j` / `k` or `↓` / `↑`: Scroll through visualizations
- `[` / `]`: Cycle between Overview hero visuals (Tracker Wall, Yearly Pulse, etc.)
- `?`: Open the help popup with all keybinds

### Entry & Editing
- `a`: **Add or edit** an entry for any date (supports shortcuts like `t`, `y`, `-3`)
- `x`: **Quick entry** for a single tracker today
- `p`: Start a **Pomodoro** timer to log focused work automatically
- `e`: **Edit** one of your 20 most recent entries
- `u`: **Undo** the last save (if you made a mistake)
- `d`: **Delete** an entry for a specific date

### Settings & Data
- `s`: Open **Settings** to manage trackers, themes, and integrations
- `E`: **Export** all data to a JSON file in your workspace
- `q`: **Quit** safely (triggers backup if configured)

## Power User Features

### CLI Logging
Log data without even opening the TUI. Perfect for scripts, shortcuts, or terminal-heavy workflows:
```bash
gotrack log code=true sleep=7.5
gotrack log rating=4 --date yesterday
gotrack log "Deep Work=90" "Main Win=shipped feature"
```

### Obsidian Integration
Mirror your daily entries into your Obsidian vault as Markdown notes. Each entry gets its own file, and you can even generate a **weekly summary report**:
```bash
gotrack report
```

### AI Context (MCP)
Expose your tracking data to AI assistants like Cursor or Claude Desktop using the Model Context Protocol:
```bash
gotrack mcp
```
See [docs/AI_AGENTS.md](docs/AI_AGENTS.md) for setup.

## Documentation

- **[Detailed Usage Guide](docs/USAGE.md)**: Deep dive into Pomodoro, Quick Entry, and more.
- **[Customization Guide](docs/CUSTOMIZATION.md)**: How to structure your trackers and categories.
- **[FAQ](docs/FAQ.md)**: Troubleshooting and common questions.
- **[Remote Logging](docs/REMOTE_LOGGING.md)**: Logging from your phone or other devices.

## Stack

- **Go**
- **Bubble Tea / Lipgloss / Huh**
- **SQLite** via `modernc.org/sqlite`
- **Asciigraph**
- **Cobra**
- **MCP** ([modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk)) for `gotrack mcp`
