# Logging from your phone

GoTrack is local-first — there's no server. But with the `gotrack log` CLI
(see `gotrack log --help`) and a sync tool, you can log from anywhere.

## Option 1 — Syncthing / iCloud Drive / Dropbox

Put your GoTrack workspace inside a synced folder and point the pointer file
at it.

```sh
# One-time: move the workspace into a synced folder.
mv ~/.gotrack ~/Sync/gotrack
echo ~/Sync/gotrack > ~/.gotrack-location
```

Any machine that has `gotrack` installed and the same synced folder will now
read and write the same database. Launch the TUI as usual, or use the CLI:

```sh
gotrack log sleep=7.5 code=true note="shipped feature"
gotrack log --date y mood=3
```

**Caveats:**
- Don't run the TUI on two machines at the same time — SQLite will not love
  it. The CLI is fast enough that collisions are unlikely.
- iCloud Drive may delay uploads; Syncthing is faster for frequent logging.

## Option 2 — iOS Shortcut → SSH → `gotrack log`

If you keep a home machine always-on (desktop, mini, Raspberry Pi), an iOS
Shortcut can SSH in and run `gotrack log`.

1. Install the [Shortcuts](https://apps.apple.com/app/shortcuts/id915249334)
   app.
2. New Shortcut → add **Run Script Over SSH**.
   - Host / User / Password (or key) pointing at your home machine.
   - Script: `gotrack log rating="$1" note="$2"` (or similar — customize to
     whichever trackers you want to log from phone).
3. Add **Ask for Input** actions above the SSH step to prompt for the values.
4. Add the Shortcut to your Home Screen or Lock Screen.

Tap, type, done — entry is written to the database on your home machine.

## Backup & restore

Regardless of sync method, `gotrack export` and `gotrack import` give you a
portable JSON or CSV snapshot:

```sh
gotrack export --format json > backup.json
gotrack export --format csv  > entries.csv
gotrack import backup.json --dry-run
gotrack import backup.json
```
