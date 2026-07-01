# musicr

A lightweight CLI tool for streaming music via YouTube, built on `mpv` and `yt-dlp`.
Single Go binary. No GUI. No bloat.

## Features

- **Interactive TUI** — progress bar, live queue, single-keystroke controls
- **Radio mode** — auto-extends your queue from YouTube Mix (on by default)
- **Volume and seek controls** — adjust volume and skip forward/backward without leaving the terminal
- **Playlists** — save/load queues as plain text files (one search query per line)
- **Prefetch** — resolves the next track before the current one ends, no stalls
- **Minimal runtime deps** — just `mpv` and `yt-dlp`, both widely packaged

## Install

### Prerequisites

```bash
# Ubuntu / Debian (only tested on Ubuntu)
sudo apt install mpv
sudo add-apt-repository ppa:tomtomtom/yt-dlp    # Add ppa repo to apt
sudo apt update                                 # Update package list
sudo apt install yt-dlp                         # Install yt-dlp

# macOS
brew install mpv yt-dlp

# Arch
sudo pacman -S mpv yt-dlp

# Fedora
sudo dnf install mpv
pip install -U yt-dlp
```

### From source

```bash
go install github.com/tanishque-suthar/musicr@latest
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/tanishque-suthar/musicr/releases) or:

```bash
curl -fsSL https://raw.githubusercontent.com/tanishque-suthar/musicr/main/install.sh | sh
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

| Key       | Action                                         |
|-----------|-------------------------------------------------|
| `a`       | Add a track (prompts for search query)          |
| `n`       | Next track                                      |
| `p`       | Previous track                                  |
| `space`   | Pause / resume                                  |
| `s`       | Save queue as a playlist (prompts for name)     |
| `l`       | Load a playlist (prompts for name)              |
| `r`       | Toggle radio mode (auto-extend)                 |
| `d`       | Delete a track from the queue                   |
| `+` / `=` | Volume up (1 step)                              |
| `-` / `_` | Volume down (1 step)                            |
| `]` / `.` | Seek forward 5 seconds                          |
| `[` / `,` | Seek backward 5 seconds                         |
| `q`       | Quit                                            |

## How It Works

```
musicr tame impala
  -> yt-dlp resolves the search query to a YouTube video ID
  -> mpv streams the audio (no video) via JSON IPC
  -> The TUI renders a progress bar, queue, and keybinding help
  -> Radio mode fetches more tracks from YouTube Mix when the queue runs low
```

All mutable state lives in a single goroutine. Background goroutines (key listener,
mpv event listener, prefetch/radio worker) communicate via channels — no locks, no
races.

Playlists are stored as plain text at `~/.config/musicr/playlists/<name>.txt`,
one search query per line. They are re-resolved on load, so they never go stale.

## Credits

musicr is a thin wrapper around two incredible open-source projects:

- **[mpv](https://mpv.io/)** — a free, open-source, and cross-platform media player.
  musicr spawns mpv headless and controls it via JSON IPC over a Unix socket.
  mpv handles all audio decoding and playback.

- **[yt-dlp](https://github.com/yt-dlp/yt-dlp)** — a feature-rich command-line
  audio/video downloader with extensive site support. musicr uses yt-dlp to
  resolve search queries into video IDs and to fetch YouTube Mix playlists for
  radio mode.

Both tools are used as external subprocesses; musicr would not exist without them.

i was using mpv and yt-dlp script locally. seeing how well the script worked, i decided to vibe-code this :P 

## License

MIT
