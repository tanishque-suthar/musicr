# musicr — YouTube Music Radio CLI (Go Edition)

A lightweight, cross-platform Bash-to-Go rewrite of musicr: an infinite YouTube Music radio experience with zero setup.

**Features:**
- 🎵 Stream YouTube Music with a single command
- 🎼 Infinite auto-loading playlist (YouTube Mix/Recommendations)
- 💻 Works on **macOS**, **Windows**, and **Linux**
- ⚡ Fast, lightweight binary (~9MB)
- 🔧 Minimal setup: auto-downloads dependencies on first run
- 🎨 Beautiful TUI with ANSI colors (graceful fallback to plain text)

---

## Quick Start

### Installation

#### macOS & Linux (Bash)
```bash
curl -fL https://raw.githubusercontent.com/yourusername/musicr/main/install.sh | bash
```

#### Windows (PowerShell)
```powershell
iex (iwr https://raw.githubusercontent.com/yourusername/musicr/main/install.ps1).Content
```

Or download pre-built binaries from [GitHub Releases](https://github.com/yourusername/musicr/releases).

### First Run

```bash
# Install mpv (one-time setup)
# macOS:
brew install mpv

# Linux (Debian/Ubuntu):
sudo apt install mpv

# Windows:
choco install mpv
# OR download from https://sourceforge.net/projects/mpv-player-windows/
```

Then:
```bash
musicr "taylor swift"
```

On first run, musicr will automatically download `yt-dlp` to cache and set up configuration.

---

## Usage

### Basic Commands

```bash
# Search and play
musicr "the weeknd blinding lights"

# Interactive search
musicr -i

# Show help
musicr --help

# Show version
musicr --version

# Show config location
musicr --config

# Reset configuration
musicr --setup
```

### Playback Controls (via mpv)

- `SPACE` — Play/Pause
- `←/→` — Seek backward/forward
- `</>`  — Previous/Next track
- `0`/`9` — Volume up/down
- `m` — Mute
- `q` — Quit

---

## Configuration

Configuration is stored in `config.json` located alongside the binary (portable):

```json
{
  "video_format": "251/140/bestaudio/best",
  "player_client": "android",
  "demuxer_max_bytes": "67MiB",
  "term_status_msg": "Status: ${time-pos} / ${duration} (${percent-pos}%)",
  "cache_path": "/tmp/musicr-cache",
  "bin_cache_path": "/tmp/musicr-bins"
}
```

Edit `config.json` directly to customize:
- **video_format**: Audio quality (251=best Opus, 140=m4a, bestaudio=fallback)
- **player_client**: YouTube client (android recommended)
- **demuxer_max_bytes**: mpv buffer size for smooth playback
- **term_status_msg**: Status bar format (leave empty to disable)

---

## How It Works

1. **Search**: You provide a search query
2. **Fetch**: musicr finds the video on YouTube using yt-dlp
3. **Play**: The first track starts playing with mpv
4. **Mix**: musicr automatically:
   - Fetches the YouTube Mix/Recommendation list
   - Pre-loads upcoming tracks in the background
   - Updates the TUI with track titles
   - Loads more when you're near the end (infinite loop)

---

## Architecture

| Component | Purpose |
|-----------|---------|
| **cmd/musicr/** | Main entry point, CLI orchestration |
| **internal/config/** | Configuration loading/saving |
| **internal/downloader/** | Auto-download yt-dlp & mpv per platform |
| **internal/ytdlp/** | yt-dlp subprocess wrapper (search, fetch URLs) |
| **internal/player/** | mpv subprocess wrapper (playback control) |
| **internal/queue/** | Playlist management, background fetching |
| **internal/ui/** | TUI renderer (ANSI codes + plain text fallback) |

---

## Requirements

- **mpv** (audio player) — install via package manager
  - macOS: `brew install mpv`
  - Linux: `apt install mpv` (Debian/Ubuntu) | `pacman -S mpv` (Arch)
  - Windows: `choco install mpv` or download from [sourceforge](https://sourceforge.net/projects/mpv-player-windows/)

- **yt-dlp** — auto-downloaded on first run

---

## Building from Source

```bash
# Clone the repo
git clone https://github.com/yourusername/musicr.git
cd musicr

# Build binary
go build -o musicr ./cmd/musicr

# Run
./musicr "search query"
```

### Cross-Platform Builds (requires GoReleaser)

```bash
# Install GoReleaser: https://goreleaser.com/
goreleaser build --clean

# Create release (requires git tag)
git tag v1.1.0
goreleaser release --clean
```

---

## Troubleshooting

### "Command not found: musicr"
- Ensure the binary is in your PATH
- On macOS/Linux, add to `.bashrc` or `.zshrc`:
  ```bash
  export PATH="/usr/local/bin:$PATH"
  ```

### "Could not find video"
- Check internet connection
- Try a different search query
- YouTube may be blocking your IP (try VPN or wait)

### "Error: mpv not found"
- Install mpv using the package manager for your OS (see Requirements)
- Restart musicr after installation

### Poor audio quality
- Edit `config.json` and try different `video_format`:
  - `251` — Best quality (Opus, requires recent mpv)
  - `140` — Good quality (m4a)
  - `bestaudio` — Fallback, varies by video

### Player won't start on Windows
- Ensure mpv is installed and in PATH
- Try running `mpv --version` in PowerShell to verify

---

## Version History

### v1.0.0 (2026-05-27)
- 🎉 Initial Go release
- ✅ Cross-platform support (macOS, Windows, Linux)
- ✅ Auto-download dependencies on first run
- ✅ Pure ANSI TUI with plain text fallback
- ✅ Portable configuration (config.json in binary dir)
- ✅ Feature parity with Bash version

---

## Contributing

Contributions welcome! Areas for enhancement:
- [ ] Better error messages & troubleshooting
- [ ] macOS code signing & Notarization
- [ ] Package manager integration (Homebrew, apt, choco)
- [ ] Lyrics display integration
- [ ] Streaming to other devices (Chromecast, AirPlay)

---

## License

dual-licensed:
- GPL-3.0 (see [LICENSE.GPL](LICENSE.GPL))
- LGPL-3.0 (see [LICENSE.LGPL](LICENSE.LGPL))

---

## Credits

- **musicr** — Original Bash version
- **yt-dlp** — YouTube downloader
- **mpv** — Audio/video player

## Support

- 🐛 **Issues**: [GitHub Issues](https://github.com/yourusername/musicr/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/yourusername/musicr/discussions)
- 📖 **Docs**: [Full Documentation](https://github.com/yourusername/musicr/wiki)
