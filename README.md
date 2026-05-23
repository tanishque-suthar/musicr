# musicr

A lightweight YouTube Music Radio CLI tool that creates a Spotify-like infinite radio experience from YouTube content.

## Features

- **Search & Play**: Instantly search YouTube and play any song
- **Radio Mode**: Automatically generates an infinite radio mix (YouTube Mix style)
- **Smart Buffering**: Pre-loads upcoming tracks seamlessly
- **TUI Display**: Beautiful terminal UI showing current playlist
- **Fast**: Minimal dependencies, zero bloat
- **Customizable**: Easy config for audio format, buffering, and more

## Requirements

- **yt-dlp** - YouTube content downloader
- **mpv** - Audio player
- **bash** 4.0+
- Linux (primary support)

## Installation

### Quick Install

```bash
git clone https://github.com/yourusername/musicr.git
cd musicr
bash install.sh
# OR install to a custom location
bash install.sh ~/.local/opt/musicr
```

### Add to PATH

After installation, make musicr accessible globally:

```bash
# Option 1: Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:$HOME/.local/opt/musicr/bin"

# Option 2: Create a symlink (system-wide)
sudo ln -s /path/to/musicr/bin/musicr /usr/local/bin/musicr
```

### Install Dependencies

#### Ubuntu/Debian
```bash
sudo apt install yt-dlp mpv
```

#### Fedora
```bash
sudo dnf install yt-dlp mpv
```

#### Arch Linux
```bash
sudo pacman -S yt-dlp mpv
```

#### macOS (Homebrew)
```bash
brew install yt-dlp mpv
```

## Usage

### Basic Search and Play

```bash
# Search for a song
musicr "taylor swift lover"

# Or use the -s flag
musicr -s "the weeknd blinding lights"
```

### Interactive Mode

```bash
# Start interactive search
musicr -i
# Then type your search query
```

### Show Config Location

```bash
# Print the config directory
musicr -c
# Example output: ~/.config/musicr
```

### Initialize Configuration

```bash
# Set up default config files
musicr --setup
```

### Help and Version

```bash
musicr --help
musicr --version
```

## Configuration

musicr stores configuration in `~/.config/musicr/config`. You can customize:

- **VIDEO_FORMAT**: Audio format quality (default: `251/140/bestaudio/best`)
- **PLAYER_CLIENT**: YouTube client to use (default: `android`)
- **TERM_STATUS_MSG**: Status message format for mpv
- **DEMUXER_MAX_BYTES**: Buffer size (default: `67MiB`) 67 :P

Example config:
```bash
# ~/.config/musicr/config
VIDEO_FORMAT="251/140/bestaudio/best"
PLAYER_CLIENT="android"
TERM_STATUS_MSG="Status: ${time-pos} / ${duration} (${percent-pos}%)"
DEMUXER_MAX_BYTES="67MiB"
PREFETCH_PLAYLIST="yes"
CACHE="yes"
```

## How It Works

1. **Search**: You provide a search query
2. **Fetch**: musicr finds the video on YouTube using yt-dlp
3. **Play**: The first track starts playing with mpv
4. **Mix**: The lazy-radio.lua script automatically:
   - Fetches the YouTube Mix/Recommendation list
   - Pre-loads upcoming tracks in the background
   - Updates the TUI with track titles
   - Automatically loads more when you're near the end (infinite loop)

## Keybindings (mpv)

Since musicr uses mpv under the hood, you have access to all mpv keybindings:

- `SPACE` - Play/Pause
- `←/→` - Seek backward/forward
- `</>`- Previous/Next track
- `9` - Decrease Volume
- `0` - Increase Volume
- `m` - Mute
- `q` - Quit
- `h` - Show help

Full mpv keybindings: https://mpv.io/manual/stable/#interactive-control

## Troubleshooting

### "Missing dependencies"
Make sure yt-dlp and mpv are installed and in your PATH:
```bash
which yt-dlp mpv
```

### "Could not find video"
- Check your internet connection
- Try a different search query
- YouTube may be blocking your IP (try a VPN or wait)

### Poor audio quality
Edit `~/.config/musicr/config` and change `VIDEO_FORMAT`:
- `251` = Best quality (Opus, requires recent mpv)
- `140` = Good quality (m4a)
- `bestaudio` = Fallback, varies by video

### Playback issues
Try resetting configuration:
```bash
musicr --setup
```

## Development

### Project Structure

```
musicr/
├── bin/
│   └── musicr              # Main entry point
├── musicr/
│   └── core.sh            # Core functionality
├── config/
│   ├── config.default     # Default config
│   └── lazy-radio.lua     # Radio script
├── install.sh             # Installation script
├── README.md
└── LICENSE
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT License - see LICENSE file for details

## Disclaimer

musicr is for personal use only. Respect YouTube's Terms of Service and copyright laws.
this tool is just to avoid bloat of web browsers to play songs. i by no means encourage unfair use of this tool.

## Inspiration

Chrome hogging my ram just to play songs was a put off. I made a script for myself, later thought to make it available for everyone.

The script is pretty simple. If you wish you can skip installing musicr and use my original script, it is available in original script directory.

## Support

Found a bug? Have a feature request? [Open an issue!](https://github.com/yourusername/musicr/issues)

---