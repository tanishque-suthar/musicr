package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// InputMode represents the current input mode of the UI.
type InputMode int

const (
	ModeKeystroke InputMode = iota // single-keystroke mode (default)
	ModeLineInput                  // line editing mode (for search/save/load prompts)
)

// State holds the data the UI needs to render a frame.
type State struct {
	TrackTitle  string
	TimePos     float64
	Duration    float64
	Paused      bool
	RadioOn     bool
	Queue       []string // display names for all tracks
	QueueIdx    int      // index of currently playing track (-1 if none)
	InputMode   InputMode
	InputPrompt string // e.g. "Search: ", "Save as: ", "Load: "
	InputBuf    string // current line input buffer
	StatusMsg   string // transient status message (e.g. "Saved!", "Loading...")
	Playlists   []string // available playlist names (shown during load)
}

// KeyEvent represents a key pressed by the user.
type KeyEvent struct {
	Key  byte   // single key in keystroke mode
	Line string // completed line in line-input mode
	Mode InputMode
}

// UI manages the terminal: raw mode, rendering, and key reading.
type UI struct {
	oldState *term.State
	fd       int
	width    int
	height   int
}

// New creates a new UI and enters raw terminal mode.
func New() (*UI, error) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, fmt.Errorf("failed to enter raw mode: %w", err)
	}

	w, h, _ := term.GetSize(fd)
	if w == 0 {
		w = 80
	}
	if h == 0 {
		h = 24
	}

	u := &UI{
		oldState: oldState,
		fd:       fd,
		width:    w,
		height:   h,
	}

	// Hide cursor
	fmt.Fprint(os.Stdout, "\033[?25l")

	return u, nil
}

// Restore restores the terminal to its original state.
// Safe to call multiple times.
func (u *UI) Restore() {
	if u.oldState != nil {
		// Show cursor, clear screen
		fmt.Fprint(os.Stdout, "\033[?25h\033[2J\033[H")
		term.Restore(u.fd, u.oldState)
		u.oldState = nil
	}
}

// ReadKey reads a single byte from stdin (blocking).
func (u *UI) ReadKey() (byte, error) {
	buf := make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	return buf[0], err
}

// Render draws the entire UI based on the given state.
func (u *UI) Render(s State) {
	// Refresh terminal size
	if w, h, err := term.GetSize(u.fd); err == nil && w > 0 && h > 0 {
		u.width = w
		u.height = h
	}

	var b strings.Builder

	// Move cursor to top-left, clear screen
	b.WriteString("\033[H\033[2J")

	// Header
	radioStatus := "off"
	radioColor := "\033[90m" // dim gray
	if s.RadioOn {
		radioStatus = "on"
		radioColor = "\033[92m" // bright green
	}
	b.WriteString(fmt.Sprintf("  \033[1;95mmusicr\033[0m  —  radio mode: %s%s\033[0m\n\n", radioColor, radioStatus))

	// Now playing
	if s.TrackTitle != "" {
		pauseIndicator := "♪"
		titleColor := "\033[96m" // cyan
		if s.Paused {
			pauseIndicator = "⏸"
			titleColor = "\033[93m" // yellow when paused
		}
		b.WriteString(fmt.Sprintf("  %s \033[1m%s%s\033[0m\n", pauseIndicator, titleColor, s.TrackTitle))
	} else {
		b.WriteString("  \033[90m♪ No track playing\033[0m\n")
	}

	// Progress bar
	b.WriteString("  ")
	b.WriteString(renderProgressBar(s.TimePos, s.Duration, u.width-4))
	b.WriteString("\n\n")

	// Queue
	b.WriteString("  \033[1;37mQueue:\033[0m\n")
	if len(s.Queue) == 0 {
		b.WriteString("    \033[90m(empty)\033[0m\n")
	} else {
		maxVisible := u.height - 12 // reserve lines for header, progress, help bar
		if maxVisible < 3 {
			maxVisible = 3
		}

		start := 0
		if s.QueueIdx > maxVisible-2 {
			start = s.QueueIdx - maxVisible + 2
		}
		end := start + maxVisible
		if end > len(s.Queue) {
			end = len(s.Queue)
		}

		if start > 0 {
			b.WriteString(fmt.Sprintf("    \033[90m  ↑ %d more\033[0m\n", start))
		}

		for i := start; i < end; i++ {
			num := fmt.Sprintf("%2d", i+1)
			title := s.Queue[i]
			if len(title) > u.width-16 {
				title = title[:u.width-19] + "..."
			}
			if i == s.QueueIdx {
				b.WriteString(fmt.Sprintf("    \033[1;92m%s  %s  (playing)\033[0m\n", num, title))
			} else {
				b.WriteString(fmt.Sprintf("    \033[37m%s\033[0m  \033[90m%s\033[0m\n", num, title))
			}
		}

		if end < len(s.Queue) {
			b.WriteString(fmt.Sprintf("    \033[90m  ↓ %d more\033[0m\n", len(s.Queue)-end))
		}
	}

	// Status message
	if s.StatusMsg != "" {
		b.WriteString(fmt.Sprintf("\n  \033[93m%s\033[0m\n", s.StatusMsg))
	}

	// Input area (line-input mode)
	if s.InputMode == ModeLineInput {
		b.WriteString(fmt.Sprintf("\n  \033[1;97m%s\033[0m%s\033[5m▎\033[0m\n", s.InputPrompt, s.InputBuf))
	}

	// Help bar at bottom - move cursor to the bottom
	b.WriteString(fmt.Sprintf("\033[%d;1H", u.height))
	if s.InputMode == ModeLineInput {
		b.WriteString("  \033[90m[enter] confirm  [esc] cancel\033[0m")
	} else {
		b.WriteString("  \033[90m[a]dd  [n]ext  [p]rev  [space]pause  [s]ave  [l]oad  [r]adio  [d]elete  [q]uit\033[0m")
	}

	// In raw terminal mode, \n only moves the cursor down (Line Feed), but does
	// not return to the beginning of the line (Carriage Return).
	// Translate all \n to \r\n to ensure correct layout rendering.
	output := b.String()
	output = strings.ReplaceAll(output, "\n", "\r\n")
	fmt.Fprint(os.Stdout, output)
}

// renderProgressBar creates a text progress bar.
func renderProgressBar(pos, dur float64, width int) string {
	timeStr := fmt.Sprintf(" %s / %s", formatTime(pos), formatTime(dur))
	
	// If the terminal is too narrow to even show a tiny bar, just return the time
	if width < len(timeStr)+5 {
		return timeStr
	}

	barWidth := width - len(timeStr) - 2 // 2 for brackets
	if barWidth < 1 {
		barWidth = 1
	}

	var pct float64
	if dur > 0 {
		pct = pos / dur
		if pct > 1 {
			pct = 1
		}
		if pct < 0 {
			pct = 0
		}
	}

	filled := int(float64(barWidth) * pct)
	if filled > barWidth {
		filled = barWidth
	}

	bar := "\033[95m" + strings.Repeat("█", filled) +
		"\033[90m" + strings.Repeat("░", barWidth-filled) + "\033[0m"

	return fmt.Sprintf("[%s]%s", bar, timeStr)
}

// formatTime formats seconds as mm:ss.
func formatTime(secs float64) string {
	if secs < 0 {
		secs = 0
	}
	m := int(secs) / 60
	s := int(secs) % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}
