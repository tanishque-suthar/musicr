package ytdlp

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// VideoInfo represents information about a video
type VideoInfo struct {
	ID  string
	URL string
}

// Client represents a yt-dlp client
type Client struct {
	YTDLPPath string
	Config    YTDLPConfig
}

// YTDLPConfig holds yt-dlp configuration
type YTDLPConfig struct {
	VideoFormat  string
	PlayerClient string
}

// NewClient creates a new yt-dlp client
func NewClient(ytdlpPath string, cfg YTDLPConfig) *Client {
	return &Client{
		YTDLPPath: ytdlpPath,
		Config:    cfg,
	}
}

// SearchVideo searches for a video and returns ID and title
// Returns (videoID, title, streamURL, error)
func (c *Client) SearchVideo(query string) (string, string, string, error) {
	args := []string{
		"--print", "id",
		"--print", "title",
		"--print", "urls",
		"--format", c.Config.VideoFormat,
		"--extractor-args", fmt.Sprintf("youtube:player_client=%s;skip=webpage", c.Config.PlayerClient),
		"--force-ipv4",
		"--no-warnings",
		fmt.Sprintf("ytsearch1:%s", query),
	}

	cmd := exec.Command(c.YTDLPPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", "", fmt.Errorf("yt-dlp search failed: %w\nstderr: %s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 3 {
		return "", "", "", fmt.Errorf("could not find video for query: %s", query)
	}

	return lines[0], lines[1], lines[2], nil
}

// FetchMixIDs fetches video IDs from a YouTube Mix/Recommendation list
// Returns a slice of video IDs
func (c *Client) FetchMixIDs(radioURL string, startIdx, endIdx int) ([]string, error) {
	args := []string{
		"--flat-playlist",
		"--playlist-items", fmt.Sprintf("%d-%d", startIdx, endIdx),
		"--print", "id",
		"--no-warnings",
		radioURL,
	}

	cmd := exec.Command(c.YTDLPPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("yt-dlp fetch mix failed: %w\nstderr: %s", err, stderr.String())
	}

	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}

	return lines, nil
}

// FetchRawURL fetches the raw streaming URL and title for a video ID
// Returns (title, streamURL, error)
func (c *Client) FetchRawURL(videoID string) (string, string, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)

	args := []string{
		"--print", "%(title)s<SEP>%(url)s",
		"--format", c.Config.VideoFormat,
		"--extractor-args", fmt.Sprintf("youtube:player_client=%s;skip=webpage", c.Config.PlayerClient),
		"--force-ipv4",
		"--no-warnings",
		videoURL,
	}

	cmd := exec.Command(c.YTDLPPath, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", "", fmt.Errorf("yt-dlp fetch raw URL failed: %w\nstderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())
	parts := strings.Split(output, "<SEP>")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("unexpected yt-dlp output format")
	}

	title := strings.TrimSpace(parts[0])
	streamURL := strings.TrimSpace(parts[1])

	return title, streamURL, nil
}

// CleanTitle removes problematic characters from a title for display
func CleanTitle(title string) string {
	// Remove commas and equals signs which can cause issues with mpv
	title = strings.ReplaceAll(title, ",", "")
	title = strings.ReplaceAll(title, "=", "")
	return title
}

// GetVersion returns the version of yt-dlp
func (c *Client) GetVersion() (string, error) {
	cmd := exec.Command(c.YTDLPPath, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get yt-dlp version: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// CheckHealth verifies yt-dlp is working
func (c *Client) CheckHealth() error {
	_, err := c.GetVersion()
	return err
}
