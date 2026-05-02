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

**What are the key keyboard shortcuts?**
- `a`: Add/edit full entry
- `x`: Quick entry (one tracker only)
- `p`: Pomodoro timer (logs to a duration tracker)
- `e`: Edit recent entries
- `u`: Undo last save
- `d`: Delete entry
- `s`: Settings & Setup
- `?`: Toggle help overlay
- `h`/`l`: Switch tabs
- `j`/`k`: Scroll view
- `q`: Quit

**How do I use the date shortcuts?**
When prompted for a date, you can use:
- `t` or `today`
- `y` or `yesterday`
- `-N` (e.g., `-2` for two days ago, `-7` for a week ago)
- `MM-DD` (e.g., `05-01` for May 1st)
- `YYYY-MM-DD` (e.g., `2026-05-01`)

**What is the difference between "Add Entry" (`a`) and "Quick Entry" (`x`)?**
- **Add Entry (`a`)**: Opens a full form for the chosen date, showing all trackers in their respective categories. Best for your end-of-day reflection.
- **Quick Entry (`x`)**: A streamlined flow to log a single value for *today* only. Use this for "fire and forget" tracking throughout the day (e.g., logging a cup of coffee or a quick workout).

**How does the Pomodoro timer work?**
Press `p` to start a session. You'll pick a duration tracker (e.g., "Deep Work" or "Coding") and a time (default 25m). While active, GoTrack shows a countdown. When finished (or if you stop early with `enter`), the elapsed time is automatically added to that tracker for today.

### Visuals and Insights

**What do the different symbols in heatmaps mean?**
- `■` (or `#`): Activity recorded/Goal met
- `·` (or `.`): No activity recorded
- `▪` (or `o`): Partial progress (for numeric trackers)

**How are correlations calculated in the Insights tab?**
GoTrack calculates the **Pearson Correlation Coefficient (r)** between pairs of numeric trackers. 
- `r = +1.0`: Strong positive correlation (as X goes up, Y goes up)
- `r = -1.0`: Strong negative correlation (as X goes up, Y goes down)
- `r = 0.0`: No linear correlation
It also shows a scatter plot to help you visualize the relationship.

**What is the "Biggest Mover" in the Review tab?**
This identifies the tracker with the largest absolute percentage change compared to the previous period (week or month). It's a great way to spot sudden shifts in your habits.

### Settings and Data

**Where is my config file?**
GoTrack stores everything in your workspace directory (chosen during setup). The default is usually `~/.gotrack`. Look for `config.json` and `data.db`.

**Can I change the colors?**
Yes! Go to Settings (`s`) -> Appearance to choose a theme (GoTrack, Catppuccin, Nord, etc.). You can also enable/disable the "Starfield" background animation.

**How do I back up my data?**
GoTrack supports a "Backup Command" that runs every time you save. See the **Backup Recipes** section below for examples using Git, rclone, or Syncthing.

**Does GoTrack support multiple workspaces?**
Currently, GoTrack uses a single workspace defined in `~/.gotrack_workspace`. You can manually edit this file to point to a different directory if needed.

---

### Backup recipes

#### Git (local version history)

Turn your workspace into a git repo and auto-commit every save:

```bash
# One-time setup — run in your terminal
cd ~/.gotrack          # or wherever your workspace is
git init
git add -A
git commit -m "initial"
```

Backup command to set in GoTrack:

```
git -C ~/.gotrack add -A && git -C ~/.gotrack commit -m "backup $(date +%Y-%m-%d)"
```

Every entry save creates a new commit. Run `git log` in `~/.gotrack` to see the history.

#### Git + remote (GitHub/Gitea for off-site backup)

After the git setup above, add a remote:

```bash
cd ~/.gotrack
git remote add origin git@github.com:yourname/gotrack-data.git
```

Backup command:

```
git -C ~/.gotrack add -A && git -C ~/.gotrack commit -m "backup $(date +%Y-%m-%d)" ; git -C ~/.gotrack push
```

> **Privacy note:** this will push your raw tracking data to the remote. Use a private repo.

#### rclone (cloud: S3, Dropbox, Google Drive, etc.)

```bash
# One-time: configure a remote named "mycloud"
rclone config
```

Backup command:

```
rclone sync ~/.gotrack mycloud:gotrack-backup
```

#### Syncthing (peer-to-peer sync, no cloud required)

Add `~/.gotrack` as a shared folder in the [Syncthing GUI](https://syncthing.net/). No backup command needed — Syncthing watches the directory and syncs on change automatically.

#### Restic (encrypted snapshots)

```bash
# One-time: initialise a repo (local or remote)
restic -r /path/to/restic-repo init
```

Backup command:

```
restic -r /path/to/restic-repo backup ~/.gotrack
```

---

**Does reinstalling wipe my data?**  
No. The binary and data are separate. Replacing `gotrack` won't affect your workspace database.

### AI assistants and MCP

See [AI_AGENTS.md](AI_AGENTS.md) for connecting Cursor and other MCP clients to `gotrack mcp`, and for SQLite concurrency notes.
