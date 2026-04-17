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
- `esc`: cancel date/entry form and return to dashboard
- `s`: open settings and tracking setup
- `s` settings also manages theme, Obsidian export, and starfield background
- `h` / `l` or `←` / `→`: switch tabs
- `j` / `k` or `↓` / `↑`: scroll visualizations
- `q`: quit

## Data Persistence

- GoTrack stores your database and config in the workspace you chose on first launch.
- Rebuilding or replacing the `gotrack` binary does not delete this workspace or wipe your data.

## FAQ

### Configuration

**How do I add custom trackers?**
Press `s` to open settings, then navigate to the tracker configuration section. You can add new trackers with names and select their type.

**What tracker types are available?**
- **Binary**: yes/no, true/false (checkbox)
- **Duration**: time span (hours:minutes)
- **Numeric**: any number (integer or decimal)
- **Rating**: 1-5 star rating
- **Text**: free-form text notes

**How do I enable Obsidian export?**
In settings (`s`), find the Obsidian section. Enter your vault path and optionally a template. Each day's entry will be mirrored as a markdown file.

### Controls & Navigation

**What are all the keyboard shortcuts?**
- `a`: add/edit entry
- `s`: settings
- `esc`: cancel/back
- `h`/`l` or `←`/`→`: switch tabs
- `j`/`k` or `↓`/`↑`: scroll
- `q`: quit

**How do I navigate between views?**
Use `h`/`l` or the arrow keys to switch between dashboard tabs (overview, trends, correlations).

### Visual Customization

**How do I change themes?**
Open settings (`s`) and navigate to the theme section. Choose between GoTrack (default), Catppuccin, or Nord.

**What is the starfield mode?**
The starfield is an ambient animated background for the dashboard. Toggle it on/off in settings (`s` → appearance).

### Data & Storage

**Where is my data stored?**
Your data lives in the workspace directory you chose during first launch (default: `~/.gotrack`). It contains:
- `gotrack.db` - SQLite database with all entries
- `config.json` - your configuration

**How do I backup or restore?**
Simply copy the workspace directory. To restore, point GoTrack to your backup directory on next launch.

**Does reinstalling wipe my data?**
No. The binary and data are separate. Replacing `gotrack` won't affect your database in `~/.gotrack`.

## Stack

- **Go**
- **Bubble Tea / Lipgloss / Huh**
- **SQLite** via `modernc.org/sqlite`
- **Asciigraph**
