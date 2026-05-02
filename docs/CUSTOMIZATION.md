# GoTrack Customization Guide

GoTrack is highly flexible. You can tailor it to track exactly what matters to you.

## Categories

Categories are the top-level containers for your trackers. They appear as separate tabs in the dashboard.

- **Name**: Short and descriptive (e.g., "Fitness", "Learning", "Deep Work").
- **Reflection**: A special category type. Trackers in a Reflection category are always shown in the entry form (not gated). Perfect for "Daily Win" or "Mood".

## Trackers

Each category can contain multiple trackers of different types:

### Tracker Types
- **Binary**: A simple Yes/No checkbox. Great for habits like "Read 20 mins" or "No Sugar".
- **Duration**: Tracks time in minutes. Used by the Pomodoro timer.
- **Numeric**: Any decimal number. Use this for weight, temperature, or calories.
- **Count**: Whole numbers for things like "Cups of Coffee" or "Pushups".
- **Rating**: A 1-5 star scale. Ideal for energy levels or sleep quality.
- **Text**: Free-form notes. Use this for journaling or logging specific events.

### Targets
You can set **Daily**, **Weekly**, and **Monthly** targets for Numeric, Duration, and Count trackers.
- **Daily Target**: Shown on the entry form and category charts.
- **Weekly/Monthly Targets**: Shown as progress bars in the dashboard.

## Themes

Change the look and feel of GoTrack in **Settings (`s`) -> Appearance**.
Available themes:
- **GoTrack (Default)**: Clean and high-contrast.
- **Catppuccin**: Soft, pastel colors (Latte, Frappe, Macchiato, Mocha).
- **Nord**: Cool, arctic blue tones.
- **Accessible**: High-contrast black and white.

## Starfield

The **Starfield** is an ambient background animation that adds a touch of "zen" to your dashboard. You can toggle it on/off and adjust the density in the Appearance settings.

## Integration: Obsidian

GoTrack can automatically mirror your entries to Markdown files.

1. **Vault Path**: Set the absolute path to your Obsidian vault.
2. **Folder**: (Optional) Specify a subfolder inside your vault.
3. **Template**: (Optional) Customize how the data is formatted in Markdown.

### Weekly Reports
Run `gotrack report` to generate a beautiful weekly summary in your Obsidian vault, perfect for your weekly review.
