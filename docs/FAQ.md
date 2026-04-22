# GoTrack FAQ

### Configuration

**How do I add custom trackers?**  
Press `s` to open settings, then navigate to the tracker configuration section. You can add new trackers with names and select their type.

**What tracker types are available?**

- **Binary**: yes/no, true/false (checkbox)
- **Duration**: time span (hours:minutes)
- **Numeric**: any number (integer or decimal)
- **Rating**: 1–5 star rating
- **Text**: free-form text notes

**How do I enable Obsidian export?**  
In settings (`s`), find the Obsidian section. Enter your vault path and optionally a template. Each day's entry will be mirrored as a markdown file.

### Controls and navigation

**What are all the keyboard shortcuts?**

- `a`: add/edit entry
- `d`: delete entry
- `s`: settings
- `esc`: cancel/back
- `h`/`l` or `←`/`→`: switch tabs
- `j`/`k` or `↓`/`↑`: scroll
- `q`: quit

**How do I navigate between views?**  
Use `h`/`l` or the arrow keys to switch between dashboard tabs (overview, trends, correlations).

### Visual customization

**How do I change themes?**  
Open settings (`s`) and navigate to the theme section. Choose between GoTrack (default), Catppuccin, Nord, or Accessible.

**What is the starfield mode?**  
The starfield is an ambient animated background for the dashboard. Toggle it on/off in settings (`s` → appearance).

### Reading the Review tab

**Why does the Review card show huge negative deltas mid-week?**  
The Review tab compares the *current calendar week* against the *previous calendar week*. On Tuesday the current week only has 2 days of data while the previous week has 7, so any cumulative tracker (minutes, counts) will show a large ▼ arrow — you're comparing 2 days of sums to 7 days of sums. This is expected. The delta becomes meaningful only on Sunday when both weeks are full, or on the Monthly toggle (`w`) once a month completes.

**Why does Weight show "178.4 lb" instead of my latest reading?**  
Numeric (measurement-type) trackers show the week's **average**, not the sum or the latest value. This makes weight, resting heart rate, etc. comparable across weeks even when you skip days. Duration and Count trackers still show the total — those *are* cumulative by nature.

**Why do I see "Chinese · CI / Input" and "Japanese · CI / Input" as separate rows?**  
When two trackers share a name across different categories, GoTrack prefixes the category so you can tell them apart. Categories roll up independently, so the Power pack's per-language CI / Input minutes are tracked separately.

**Why do some trackers not appear in Review?**  
Text trackers (notes, wins, blockers) aren't aggregated — they're only visible on the Overview / entry form. Rating trackers also skip the numeric card. A numeric tracker also won't appear if it has zero data in both the current and previous period.

**Why does the Biggest Mover card sometimes name something surprising?**  
"Biggest mover" is sorted by absolute delta across *all* numeric trackers, so a tracker you just started logging this week will always look like a huge mover vs the prior zero. The effect disappears once you have 2+ weeks of data for that tracker.

### Data and storage

**Where is my data stored?**  
Your data lives in the workspace directory you chose during first launch (default under your home directory). It contains:

- `data.db` — SQLite database with all entries
- `config.json` — your configuration

**How do I backup or restore?**  
Copy the workspace directory, or use `gotrack export --format json` and `gotrack import`. To restore from a folder, point GoTrack at that directory on next launch (or replace the workspace path in `~/.gotrack_workspace`).

**Does reinstalling wipe my data?**  
No. The binary and data are separate. Replacing `gotrack` won't affect your workspace database.

### AI assistants and MCP

See [AI_AGENTS.md](AI_AGENTS.md) for connecting Cursor and other MCP clients to `gotrack mcp`, and for SQLite concurrency notes.
