package player

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Player controls mpv playback
type Player struct {
	MPVPath string
	Config  PlayerConfig
	cmd     *exec.Cmd
	done    chan error
}

// PlayerConfig holds mpv configuration
type PlayerConfig struct {
	VideoFormat     string
	PlayerClient    string
	DemuxerMaxBytes string
	TermStatusMsg   string
	CachePath       string
}

// NewPlayer creates a new mpv player
func NewPlayer(mpvPath string, cfg PlayerConfig) *Player {
	return &Player{
		MPVPath: mpvPath,
		Config:  cfg,
		done:    make(chan error, 1),
	}
}

// Play starts playing a stream URL
func (p *Player) Play(streamURL, radioURL, firstTitle string) error {
	args := []string{
		"--no-video",
	}

	// Add status message if configured
	if p.Config.TermStatusMsg != "" {
		args = append(args, fmt.Sprintf("--term-status-msg=%s", p.Config.TermStatusMsg))
	}

	// Add script options for the radio/queue script if radioURL is provided
	if radioURL != "" {
		scriptOpts := fmt.Sprintf("radio_url=%s,first_title=%s", radioURL, firstTitle)
		// Note: We would use a script here, but for now we'll just queue tracks via playlist
		// args = append(args, "--script=<path-to-lazy-radio-equivalent>")
		// args = append(args, fmt.Sprintf("--script-opts=%s", scriptOpts))
	}

	// Add yt-dlp integration options
	ytdlpOpts := fmt.Sprintf(
		"ignore-config=,no-warnings=,no-check-certificates=,format=%s,force-ipv4=,extractor-args=youtube:player_client=%s",
		p.Config.VideoFormat, p.Config.PlayerClient,
	)
	args = append(args, fmt.Sprintf("--ytdl-raw-options=%s", ytdlpOpts))

	// Add cache settings
	args = append(args,
		"--cache=yes",
		fmt.Sprintf("--demuxer-max-bytes=%s", p.Config.DemuxerMaxBytes),
		"--prefetch-playlist=yes",
	)

	// Add the stream URL to play
	args = append(args, streamURL)

	p.cmd = exec.Command(p.MPVPath, args...)
	p.cmd.Stdin = os.Stdin
	p.cmd.Stdout = os.Stdout
	p.cmd.Stderr = os.Stderr

	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	return nil
}

// Wait blocks until playback finishes
func (p *Player) Wait() error {
	if p.cmd == nil {
		return fmt.Errorf("player not started")
	}

	err := p.cmd.Wait()
	return err
}

// IsRunning checks if the player is currently running
func (p *Player) IsRunning() bool {
	if p.cmd == nil || p.cmd.ProcessState == nil {
		return false
	}
	return p.cmd.ProcessState.Exited() == false
}

// Stop terminates playback
func (p *Player) Stop() error {
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Kill()
	}
	return nil
}

// GetVersion returns the version of mpv
func (p *Player) GetVersion() (string, error) {
	cmd := exec.Command(p.MPVPath, "--version")
	var output strings.Builder
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get mpv version: %w", err)
	}

	// First line is usually "mpv <version>"
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) > 0 {
		return lines[0], nil
	}

	return "", nil
}

// CheckHealth verifies mpv is working
func (p *Player) CheckHealth() error {
	cmd := exec.Command(p.MPVPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mpv health check failed: %w", err)
	}
	return nil
}

// BuildPlaylistFile creates an m3u playlist file for mpv to use
// This can be used to queue multiple tracks at once
func BuildPlaylistFile(tracks []string, outputPath string) error {
	content := "#EXTM3U\n"
	for _, track := range tracks {
		content += track + "\n"
	}

	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write playlist: %w", err)
	}

	return nil
}

// IsSupported checks if the current OS is supported for mpv auto-download
func IsSupported() bool {
	switch runtime.GOOS {
	case "darwin", "linux":
		return true // macOS and Linux can download yt-dlp, and mpv can be installed via package managers
	case "windows":
		return true // Windows can download yt-dlp
	default:
		return false
	}
}
