# musicr — YouTube Music Radio CLI

A lightweight, **cross-platform** Go rewrite of musicr: an infinite YouTube Music radio experience with **zero setup**.

**Features:**
- 🎵 Stream YouTube Music with a single command
- 🎼 Infinite auto-loading playlist (YouTube Mix/Recommendations)
- 💻 Works on **macOS**, **Windows**, and **Linux**
- ⚡ Fast, lightweight binary (~9MB)
- 🔧 **Minimal setup**: auto-downloads dependencies on first run
- 🎨 Beautiful TUI with ANSI colors (graceful fallback to plain text)
- 🔌 No setup required—just download and run

---

## 🚀 Quick Start

### Installation (Pick One)

**macOS & Linux (Bash)**
```bash
curl -fL https://raw.githubusercontent.com/yourusername/musicr/main/install.sh | bash
```

**Windows (PowerShell)**
```powershell
iex (iwr https://raw.githubusercontent.com/yourusername/musicr/main/install.ps1).Content
```

**Or manually:** Download from [GitHub Releases](https://github.com/yourusername/musicr/releases)

### Setup (One-Time)

Install mpv (the audio player):

```bash
# macOS
brew install mpv

# Ubuntu/Debian
sudo apt install mpv

# Arch Linux
sudo pacman -S mpv

# Windows
choco install mpv
# OR download: https://sourceforge.net/projects/mpv-player-windows/
```

### Usage

```bash
musicr "taylor swift"
musicr -i              # Interactive search
musicr --help          # Show all options
```

**That's it!** On first run, musicr automatically downloads `yt-dlp` and sets up configuration.

---

## 📖 Full Documentation

See [README_GO.md](README_GO.md) for:
- Detailed installation instructions
- Configuration options
- Playback controls
- Architecture overview
- Troubleshooting
- Building from source

---

## 🏗️ Requirements

- **mpv** — audio player (install via package manager)
- **yt-dlp** — auto-downloaded on first run
- **Internet connection**

| OS | Status | Supported |
|----|--------|-----------|
| macOS (Intel & Apple Silicon) | ✅ | Yes |
| Linux (x86_64) | ✅ | Yes |
| Windows (x86_64) | ✅ | Yes |

---

## 📋 Usage

```bash
# Basic search and play
musicr "taylor swift"

# Interactive mode (prompts for search)
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

Once playing, use these mpv keybindings:

- `SPACE` — Play/Pause
- `←/→` — Seek backward/forward
- `</>`  — Previous/Next track
- `0`/`9` — Volume up/down
- `m` — Mute
- `q` — Quit

Full mpv controls: https://mpv.io/manual/stable/#interactive-control

---

## ⚙️ Configuration

Config is stored in `config.json` alongside the binary (portable). Edit it to customize:

```json
{
  "video_format": "251/140/bestaudio/best",
  "player_client": "android",
  "demuxer_max_bytes": "67MiB",
  "term_status_msg": "Status: ${time-pos} / ${duration} (${percent-pos}%)"
}
```

See [README_GO.md](README_GO.md#configuration) for full configuration details.

---

## 🛠️ Building from Source

```bash
git clone https://github.com/yourusername/musicr.git
cd musicr

# Build locally
go build -o musicr ./cmd/musicr

# Run
./musicr "search query"
```

### Cross-Platform Release Build

Requires [GoReleaser](https://goreleaser.com/):

```bash
goreleaser build --clean
```

---

## 🐛 Troubleshooting

### "Command not found: musicr"
- Ensure the binary is in your `$PATH`
- On Windows, restart PowerShell after install

### "Could not find video"
- Check internet connection
- Try a different search query
- YouTube may be rate-limiting your IP (wait or use VPN)

### "Error: mpv not found"
- Install mpv using your package manager (see Requirements)
- Restart musicr

### Poor audio quality
- Edit `config.json` and change `video_format` to:
  - `251` — Best (Opus format)
  - `140` — Good (m4a)
  - `bestaudio` — Fallback

See [README_GO.md](README_GO.md#troubleshooting) for more help.

---

## 🔄 Migration from Bash Version

If you were using the old Bash version:

1. **Install Go version** — See Quick Start above
2. **Install mpv** — See Requirements
3. **First run** — musicr will auto-download yt-dlp and set up config
4. **No other changes needed** — All commands work the same!

Old Bash files are archived in `original script/` folder.

---

## 🤝 Contributing

Contributions welcome! Areas for enhancement:
- [ ] Better error messages
- [ ] macOS code signing
- [ ] Package manager integration
- [ ] Lyrics display
- [ ] Streaming to other devices

---

## 📄 License

Dual-licensed:
- **GPL-3.0** (see [LICENSE.GPL](LICENSE.GPL))
- **LGPL-3.0** (see [LICENSE.LGPL](LICENSE.LGPL))

---

## 🙏 Credits

- **yt-dlp** — YouTube downloader: https://github.com/yt-dlp/yt-dlp
- **mpv** — Media player: https://mpv.io/
- **Original Bash version** — See `original script/` folder

---

## 📞 Support

- 🐛 **Report Issues**: [GitHub Issues](https://github.com/yourusername/musicr/issues)
- 💬 **Discussions**: [GitHub Discussions](https://github.com/yourusername/musicr/discussions)
- ⭐ **Liked it?** Please star the repo!

Enjoy your infinite music radio! 🎵