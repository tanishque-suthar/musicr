package ui

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"musicr/internal/queue"
)

// SupportsANSI checks if the terminal supports ANSI escape codes
func SupportsANSI() bool {
	// On Windows, check if we're in Windows Terminal or similar
	if runtime.GOOS == "windows" {
		// Windows 10+ with Windows Terminal or ConPTY supports ANSI
		term := os.Getenv("TERM")
		if term == "xterm" || term == "xterm-256color" {
			return true
		}

		// Check for Windows Terminal
		wt := os.Getenv("WT_SESSION")
		if wt != "" {
			return true
		}

		// Default to false for legacy CMD
		return false
	}

	// Unix-like systems generally support ANSI
	return true
}

// ANSI color codes
const (
	colorReset = "\033[0m"
	colorBold  = "\033[1m"
	colorMagenta = "\033[35m"
	colorCyan   = "\033[36m"
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
)

// UI manages the terminal user interface
type UI struct {
	supportsANSI bool
	currentIndex int
	tracks       []queue.Track
}

// NewUI creates a new UI instance
func NewUI() *UI {
	return &UI{
		supportsANSI: SupportsANSI(),
		currentIndex: 0,
		tracks:       []queue.Track{},
	}
}

// Clear clears the terminal screen
func (u *UI) Clear() {
	if u.supportsANSI {
		// ANSI escape code to clear screen
		fmt.Print("\033[H\033[J")
	} else {
		// Fallback: print newlines
		for i := 0; i < 50; i++ {
			fmt.Println()
		}
	}
}

// SetCurrentTrack updates the currently playing track index
func (u *UI) SetCurrentTrack(index int) {
	u.currentIndex = index
}

// SetTracks updates the track list
func (u *UI) SetTracks(tracks []queue.Track) {
	u.tracks = tracks
}

// Render draws the UI
func (u *UI) Render() {
	u.Clear()

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Radio      ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	if len(u.tracks) == 0 {
		fmt.Println("Loading playlist...")
		return
	}

	fmt.Println("Now Playing Queue:")
	fmt.Println("─────────────────────────────────────")

	// Show current and next 10 tracks
	startIdx := u.currentIndex
	if startIdx > 0 && u.currentIndex > 0 {
		startIdx = u.currentIndex - 1
	}

	endIdx := startIdx + 11
	if endIdx > len(u.tracks) {
		endIdx = len(u.tracks)
	}

	for i := startIdx; i < endIdx; i++ {
		track := u.tracks[i]
		isCurrentTrack := i == u.currentIndex

		// Truncate long titles to fit terminal
		title := track.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		if isCurrentTrack {
			u.renderTrackLine(fmt.Sprintf("%d", i+1), title, true)
		} else {
			u.renderTrackLine(fmt.Sprintf("%d", i+1), title, false)
		}
	}

	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Loaded: %d / ~%d tracks\n", len(u.tracks), len(u.tracks)+100)
	fmt.Println()
	fmt.Println("Controls:")
	fmt.Println("  SPACE - Play/Pause | ←/→ - Seek | </> - Previous/Next")
	fmt.Println("  0 - Volume Up | 9 - Volume Down | m - Mute | q - Quit")
}

// renderTrackLine renders a single track line
func (u *UI) renderTrackLine(index, title string, isCurrent bool) {
	if u.supportsANSI {
		if isCurrent {
			fmt.Printf("%s●%s %s%s%s (playing)\n",
				colorMagenta, colorReset,
				colorBold, title, colorReset)
		} else {
			fmt.Printf("  ○ %s\n", title)
		}
	} else {
		// Fallback: plain text
		if isCurrent {
			fmt.Printf("●> %s (playing)\n", title)
		} else {
			fmt.Printf("   %s\n", title)
		}
	}
}

// RenderSearching shows a loading message
func (u *UI) RenderSearching(query string) {
	u.Clear()

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Radio      ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	if u.supportsANSI {
		fmt.Printf("%sSearching for:%s %s\n", colorCyan, colorReset, query)
		fmt.Println("Please wait...")
		fmt.Printf("%s(Loading first track and fetching playlist...)%s\n", colorYellow, colorReset)
	} else {
		fmt.Printf("Searching for: %s\n", query)
		fmt.Println("Please wait...")
		fmt.Println("(Loading first track and fetching playlist...)")
	}
}

// RenderError displays an error message
func (u *UI) RenderError(errMsg string) {
	u.Clear()

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Radio      ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	if u.supportsANSI {
		fmt.Printf("%s✗ Error:%s %s\n", colorRed, colorReset, errMsg)
	} else {
		fmt.Printf("Error: %s\n", errMsg)
	}

	fmt.Println()
	fmt.Println("Press any key to exit...")
}

// RenderInteractivePrompt shows the interactive search prompt
func (u *UI) RenderInteractivePrompt() string {
	u.Clear()

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Radio      ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	if u.supportsANSI {
		fmt.Printf("%sEnter song or artist name:%s ", colorCyan, colorReset)
	} else {
		fmt.Print("Enter song or artist name: ")
	}

	var query string
	fmt.Scanln(&query)
	return strings.TrimSpace(query)
}

// RenderDependencyMissing shows a message about missing dependencies
func (u *UI) RenderDependencyMissing(dep string) {
	u.Clear()

	fmt.Println("╔════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Radio      ║")
	fmt.Println("╚════════════════════════════════════╝")
	fmt.Println()

	if u.supportsANSI {
		fmt.Printf("%s✗ Missing dependency: %s%s\n", colorRed, dep, colorReset)
	} else {
		fmt.Printf("Missing dependency: %s\n", dep)
	}

	fmt.Println()
	fmt.Println("musicr requires:")
	fmt.Println("  - yt-dlp (downloading...)")
	fmt.Println("  - mpv (audio player)")
	fmt.Println()
	fmt.Println("Installation:")

	switch runtime.GOOS {
	case "windows":
		fmt.Println("  Windows:")
		fmt.Println("    choco install mpv")
		fmt.Println("    OR download from https://sourceforge.net/projects/mpv-player-windows/")
	case "darwin":
		fmt.Println("  macOS:")
		fmt.Println("    brew install mpv")
	case "linux":
		fmt.Println("  Linux:")
		fmt.Println("    apt install mpv          (Debian/Ubuntu)")
		fmt.Println("    pacman -S mpv            (Arch)")
		fmt.Println("    dnf install mpv          (Fedora)")
	}

	fmt.Println()
	fmt.Println("After installing mpv, run musicr again.")
}

// RenderHelp displays the help message
func (u *UI) RenderHelp() {
	u.Clear()

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║        musicr - YouTube Music Radio CLI                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("USAGE:")
	fmt.Println("  musicr [OPTIONS] [SEARCH_QUERY]")
	fmt.Println()

	fmt.Println("OPTIONS:")
	fmt.Println("  -s, --search <query>    Search for a song and play it")
	fmt.Println("  -i, --interactive       Start interactive search mode")
	fmt.Println("  -c, --config            Show config directory location")
	fmt.Println("  --setup                 Initialize/reset configuration")
	fmt.Println("  --version               Show version")
	fmt.Println("  -h, --help              Show this help message")
	fmt.Println()

	fmt.Println("EXAMPLES:")
	fmt.Println("  musicr taylor swift")
	fmt.Println("  musicr -s \"the weeknd blinding lights\"")
	fmt.Println("  musicr --interactive")
	fmt.Println()

	fmt.Println("PLAYBACK CONTROLS (via mpv):")
	fmt.Println("  SPACE      - Play/Pause")
	fmt.Println("  ←/→        - Seek backward/forward")
	fmt.Println("  </> or ,/. - Previous/Next track")
	fmt.Println("  0          - Increase volume")
	fmt.Println("  9          - Decrease volume")
	fmt.Println("  m          - Mute")
	fmt.Println("  q          - Quit")
	fmt.Println()

	fmt.Println("For more info, visit: https://github.com/yourusername/musicr")
}
