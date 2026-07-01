# musicr

A lightweight CLI tool for streaming music via YouTube, built on `mpv` + `yt-dlp`.  
Single Go binary. No GUI. No bloat.

## Features

- **Interactive TUI** — progress bar, live queue, single-keystroke controls
- **Radio mode** — auto-extends your queue from YouTube Mix (on by default)
- **Playlists** — save/load queues as plain text files (one search query per line)
- **Prefetch** — resolves the next track before the current one ends, no stalls
- **Minimal deps** — just `mpv` and `yt-dlp`, both widely packaged

## Install

### Prerequisites

```bash
# macOS
brew install mpv yt-dlp

# Ubuntu / Debian
sudo apt install mpv
pip install -U yt-dlp

# Arch
sudo pacman -S mpv yt-dlp
```

### From source

```bash
go install github.com/tsvd/musicr@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/tsvd/musicr/releases) or:

```bash
curl -fsSL https://raw.githubusercontent.com/tsvd/musicr/main/install.sh | sh
```

## Usage

```bash
# Play a track
musicr tame impala the less i know the better

# Play multiple tracks
musicr "tame impala" "mac demarco chamber of reflection" "khruangbin time"

# Load a saved playlist
musicr -p chill

# Disable radio mode
musicr --no-radio "boards of canada"

# List saved playlists
musicr list
```

## Interactive Keys

| Key     | Action                                         |
|---------|-------------------------------------------------|
| `a`     | Add a track (prompts for search query)          |
| `n`     | Next track                                      |
| `p`     | Previous track                                  |
| `space` | Pause / resume                                  |
| `s`     | Save queue as a playlist (prompts for name)     |
| `l`     | Load a playlist (prompts for name)              |
| `r`     | Toggle radio mode (auto-extend)                 |
| `d`     | Delete a track from the queue                   |
| `q`     | Quit                                            |

## How It Works

```
musicr "tame impala"
  → yt-dlp resolves the search query to a YouTube video ID
  → mpv streams the audio (no video) via IPC
  → The TUI renders a progress bar, queue, and keybinding help
  → Radio mode fetches more tracks from YouTube Mix when the queue runs low
```

**Playlists** are stored as plain text at `~/.config/musicr/playlists/<name>.txt`, one search query per line.  
They're re-resolved on load, so they never go stale.

## License

MIT
