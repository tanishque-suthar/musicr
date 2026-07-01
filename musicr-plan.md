# musicr — Project Plan

A lightweight, cross-platform (Linux/macOS) CLI tool for streaming music via
YouTube, built on top of `mpv` + `yt-dlp`. Single static Go binary, no
language runtime to install, minimal dependencies.

## Goals

- Replace the current bash + mpv-Lua prototype with a proper, distributable
  Go binary.
- Keep runtime dependencies to exactly two external tools: `mpv` and
  `yt-dlp`.
- Interactive in-player controls (add to queue, save/load playlists,
  pause/skip) without needing a second terminal.
- Custom-rendered status bar + live queue list (replaces mpv's own terminal
  output, since we own the terminal).
- Auto-extending "radio" mode on by default (mirrors current `RD<id>` mix
  behavior).
- Simple, human-readable playlist files (search queries, not URLs/IDs).

## Non-goals (for v1)

- Windows support.
- A GUI.
- Multi-user/server mode (single local process, single user).
- Authenticated/private YouTube content.

---

## Architecture overview

```
                 ┌──────────────────────┐
                 │        main           │
                 │  parse args, bootstrap │
                 └──────────┬────────────┘
                            │
                 ┌──────────▼────────────┐
                 │     app state owner     │   single goroutine,
                 │ (queue, now playing,    │   owns all mutable state,
                 │  radio mode, ui state)  │   receives events via channel
                 └───┬───────┬────────┬───┘
                     │       │        │
        ┌────────────▼─┐ ┌──▼─────┐ ┌▼─────────────┐
        │ key listener  │ │ mpv IPC │ │ prefetch /    │
        │ (raw mode)    │ │ listener│ │ radio worker  │
        └───────────────┘ └────┬───┘ └───────┬───────┘
                                │             │
                          ┌─────▼─────┐ ┌─────▼─────┐
                          │   mpv     │ │  yt-dlp    │
                          │ (subprocess,│ │ (resolve   │
                          │  IPC socket)│ │  queries)  │
                          └───────────┘ └────────────┘
```

A single goroutine owns all mutable state (queue, now-playing, radio
mode, UI). The other three goroutines (key listener, mpv event listener,
prefetch/radio worker) only communicate with it via channels — avoids lock
juggling, keeps a single source of truth for what gets rendered.

---

## Components

### `internal/mpv` — IPC client
- Spawns `mpv` headless: `--no-video --idle=yes
  --input-ipc-server=<tmp socket> --cache=yes --demuxer-max-bytes=67MiB`.
- No terminal control handed to mpv — we render 100% of output ourselves.
- Thin wrapper for JSON IPC commands: `loadfile`, `playlist-next`,
  `playlist-prev`, `set_property pause`, `get_property playlist`, etc.
- Subscribes to property changes (`time-pos`, `duration`, `pause`) and
  events (`start-file`, `end-file`) and forwards them as typed Go events.

### `internal/ytdlp` — resolver
- `Resolve(query string) (Track, error)`:
  ```
  yt-dlp ytsearch1:<query> \
    --print id --print title \
    --format "251/140/bestaudio/best" \
    --extractor-args "youtube:player_client=android;skip=webpage" \
    --force-ipv4 --no-warnings
  ```
- Returns `{ID, Title, StreamURL}`.
- Only place resolution happens. Called just-in-time, not in bulk, since
  resolved stream URLs expire.

### `internal/queue` — playback queue state
- Ordered list of `Track`, each starting as "unresolved" (query string
  only).
- Prefetch worker resolves the *next* track ~5–10s before the current one
  ends, so playback never stalls on a `yt-dlp` call — but doesn't resolve
  further ahead than that, since URLs expire quickly.

### `internal/radio` — auto-extend (on by default)
- When the queue runs low (within 5 tracks of the end) and radio mode is
  on, fetch the next batch (20 tracks) from the `RD<id>` YouTube mix tied
  to the currently playing video — same mechanism as the current Lua
  script's `fetch_ids`/`fetch_raw_url`.
- Auto-fetched tracks are stored as titles, same as manually queued
  tracks, so they save/load identically.
- Toggleable at runtime (`r` key) and via `--no-radio` flag at launch.

### `internal/playlist` — save/load
- Plain text, one search query (track title) per line.
- Location: `~/.config/musicr/playlists/<name>.txt`.
- `s` key (or `musicr save <name>`) dumps current queue titles to file.
- `l` key (or `musicr -p <name>`) reads file, re-resolves every line via
  `yt-dlp` fresh on load.

### `internal/ui` — terminal interface
- Raw terminal mode via `golang.org/x/term`, single redraw function driven
  by app-state changes (mpv property events, track changes, key actions).
- Layout:
  ```
  musicr — radio mode: on

    ♪ Tame Impala - The Less I Know The Better
    [████████████████░░░░░░░░░░] 02:14 / 03:36 (61%)

  Queue:
    1  Tame Impala - The Less I Know The Better   (playing)
    2  Mac DeMarco - Chamber of Reflection
    3  Khruangbin - Time (You and I)
    4  ...

  [a]dd  [n]ext  [p]rev  [space]pause  [s]ave  [l]oad  [r]adio  [d]elete  [q]uit
  ```
- Two input sub-modes:
  - **Single-keystroke mode** (default): one key = one action.
  - **Line-input mode** (entered by `a`, `s`, `l`): full line editing with
    backspace, for typing a query/playlist name, then returns to
    single-keystroke mode.
- Guaranteed terminal restoration (deferred + signal-handled for Ctrl+C,
  SIGTERM, and panics) — raw mode must never strand the user's shell.

---

## Keybindings (v1)

| Key     | Action                                          |
|---------|--------------------------------------------------|
| `a`     | Prompt for a query, append to queue              |
| `n`     | Next track                                       |
| `p`     | Previous track                                   |
| `space` | Pause / resume                                   |
| `s`     | Save current queue as a playlist (prompts name)  |
| `l`     | Load a playlist (prompts name, lists existing)   |
| `r`     | Toggle radio/auto-extend mode                    |
| `d`     | Remove an item from the queue (prompts index)    |
| `q`     | Quit (cleanly kills mpv, restores terminal)      |

---

## CLI surface

```
musicr <query...>        # resolve + play, enters interactive mode, radio on by default
musicr -p <playlist>     # load and play a saved playlist
musicr save <name>       # save current state (also available as 's' in-player)
musicr list              # list saved playlists
musicr --no-radio        # disable auto-extend for this session
```

---

## Startup checks

On launch, verify `mpv` and `yt-dlp` are present on `$PATH`. If missing,
fail fast with a clear install hint:

```
musicr requires mpv and yt-dlp.
  macOS:  brew install mpv yt-dlp
  Linux:  use your distro's package manager (e.g. apt install mpv yt-dlp,
          or pip install -U yt-dlp for the latest yt-dlp)
```

---

## Distribution

- Single static Go binary per OS/arch (`linux/amd64`, `linux/arm64`,
  `darwin/amd64`, `darwin/arm64`), built via GoReleaser, published to
  GitHub Releases.
- Install script: `curl -fsSL https://.../install.sh | sh` → drops binary
  into `~/.local/bin` (or detects a sensible PATH dir).
- Homebrew tap once stable (post-v1).
- MIT or Apache-2.0 license (OSS).

---

## Build order / milestones

1. **Scaffolding** — `go.mod`, directory layout, dependency-check on
   startup.
2. **`internal/mpv`** — spawn mpv headless, IPC client (commands +
   event subscription). Validate with a hardcoded single track.
3. **`internal/ytdlp`** — resolver, validate against the mpv client from
   step 2 (manual single-track play end-to-end).
4. **`internal/ui`** — raw-mode key loop + line-input sub-mode + status
   bar/queue rendering, wired to a static fake queue first (no mpv/yt-dlp
   yet) to get rendering and terminal-restoration right in isolation.
5. **`internal/queue`** — real queue + prefetch worker, wire into UI and
   mpv/yt-dlp from steps 2–3.
6. **`internal/radio`** — auto-extend, on by default, toggle key.
7. **`internal/playlist`** — save/load, `s`/`l` keys, `-p`/`save`/`list`
   CLI subcommands.
8. **Polish** — signal handling, error states (network failure, video
   unavailable, empty search), README, install script, CI build matrix.
9. **Release** — tag v0.1.0, GoReleaser, GitHub Release with prebuilt
   binaries.

---

## Open items to revisit later (not blocking v1)

- Volume control keybinding.
- Search result disambiguation (currently always takes the top
  `ytsearch1` result, same as the original script).
- Possible `--format` override flag for users who want a different
  audio quality tier than `251/140/bestaudio/best`.
- Windows support (would need a different IPC transport — named pipes
  instead of Unix sockets — and a different raw-terminal-mode library).
