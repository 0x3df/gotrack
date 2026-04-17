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

## Quick Start

```bash
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
- `s`: open settings and tracking setup
- `s` settings also manages theme, Obsidian export, and starfield background
- `h` / `l` or `←` / `→`: switch tabs
- `j` / `k` or `↓` / `↑`: scroll visualizations
- `q`: quit

## Stack

- **Go**
- **Bubble Tea / Lipgloss / Huh**
- **SQLite** via `modernc.org/sqlite`
- **Asciigraph**
