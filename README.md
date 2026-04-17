# GoTrack 🚀

A lightweight, local-first, TUI-based daily tracker and personal dashboard built in Go.

## Features

- **Local-First:** All data stays on your machine in a SQLite database.
- **Dynamic Configuration:** Build your own tracking system with custom categories and trackers (binary, duration, numeric, rating, text).
- **Rich TUI:** A full-screen interactive dashboard with:
  - **Heatmaps:** GitHub-style contribution grids for habit consistency.
  - **Trends:** Rolling line charts and vertical bar charts.
  - **Insights:** Correlation analysis (Scatter plots and A/B impact charts).
- **Portable:** Zero dependencies outside of the Go ecosystem (pure Go SQLite).

## Installation

```bash
go build -o gotrack main.go
./gotrack
```

## How it works

On first launch, GoTrack will walk you through a setup wizard where you can:
1. Choose a **Data Directory** to store your workspace (config and database).
2. Choose a **Setup Mode** (Default guided or Custom).
3. Configure your Categories (e.g., Languages, Health, Productivity).

Use the following keys to navigate the dashboard:
- `a`: Add a new daily entry.
- `h` / `l` or `←` / `→`: Switch between tabs.
- `j` / `k` or `↓` / `↑`: Scroll through your visualizations.
- `q`: Quit.

## Tech Stack

- **Go**
- **Bubble Tea / Lipgloss / Huh** (Terminal UI)
- **SQLite** (modernc.org/sqlite)
- **Asciigraph** (Line charts)
