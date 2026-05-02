# GoTrack Usage Guide

This guide covers the core features of GoTrack to help you get the most out of your tracking.

## Daily Entry (`a`)

The primary way to log your day. Press `a` and choose a date (defaults to today). 

- **Categories**: Trackers are grouped into categories like "Work", "Health", or "Personal".
- **Gating**: Categories (except "Reflection") have a "Did you do this today?" toggle. If set to No, all trackers in that category are skipped for that day, keeping your database clean.
- **Auto-loading**: If you've already logged data for a date, GoTrack pre-fills the form so you can make quick updates.

## Quick Entry (`x`)

For tracking specific events as they happen without opening the full daily form.

1. Press `x`.
2. Select the tracker you want to update.
3. Enter the value.
4. GoTrack merges this value into your existing entry for today.

## Pomodoro Timer (`p`)

Boost your productivity with the built-in Pomodoro timer.

1. Press `p`.
2. Select a **Duration** tracker (e.g., "Deep Work").
3. Set the duration in minutes (default is 25).
4. A countdown timer will appear. You can minimize GoTrack; the timer keeps running.
5. When finished, press `enter` to log the time. If you stop early, you can still log the elapsed minutes.

## Visualizations

GoTrack provides several ways to see your data:

### Overview Tab
- **Hero Visuals**: Large charts at the top. Cycle through them with `[` and `]`.
- **Yearly Pulse**: A heatmap of your overall consistency over the year.
- **Tracker Wall**: A grid of all your binary trackers for the last 30 days.
- **Cards**: Summaries of each category, showing streaks and averages.

### Category Tabs
Each category gets its own tab with detailed charts:
- **Binary Trackers**: GitHub-style heatmaps.
- **Numeric/Duration**: Line charts showing trends and rolling averages.
- **Text**: A list of your most recent notes.

### Insights Tab
Find connections in your data:
- **Correlations**: See if your "Sleep" affects your "Mood" or "Productivity".
- **Scatter Plots**: Visualize the relationship between any two numeric trackers.

### Review Tab
Compare your current week or month against the previous one:
- **Deltas**: See if you're trending up or down.
- **Biggest Mover**: Spot which habit changed the most.
- **Averages vs. Totals**: Measurements (like Weight) show averages, while activities (like Exercise) show totals.

## Data Management

### Undo (`u`)
Made a mistake? Press `u` on the dashboard to immediately revert your last save or deletion.

### Export (`E`)
Press `E` to export your entire database to a JSON file. This is useful for custom analysis or migrating your data.

### Search & Edit (`e`)
Press `e` to see a list of your 20 most recent entries. Select one to open it in the edit form.
