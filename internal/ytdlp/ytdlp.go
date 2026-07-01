package ytdlp

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Track represents a resolved or unresolved music track.
type Track struct {
	ID    string // YouTube video ID (empty if unresolved)
	Title string // video title (empty if unresolved)
	URL   string // direct audio stream URL (empty if unresolved)
	Query string // original search query
}

// Resolved returns true if the track has been resolved via yt-dlp.
func (t Track) Resolved() bool {
	return t.ID != ""
}

// StreamURL returns the direct URL for mpv to stream from.
func (t Track) StreamURL() string {
	if t.URL != "" {
		return t.URL
	}
	if t.ID == "" {
		return ""
	}
	return "https://www.youtube.com/watch?v=" + t.ID
}

// Resolve runs yt-dlp to resolve a search query into a Track.
// It uses ytsearch1 to find the top result and extracts the video ID, title, and direct URL.
func Resolve(ctx context.Context, query string) (Track, error) {
	args := []string{
		"ytsearch1:" + query,
		"--print", "id",
		"--print", "title",
		"--print", "urls",
		"--format", "251/140/bestaudio/best",
		"--extractor-args", "youtube:player_client=android;skip=webpage",
		"--force-ipv4",
		"--no-warnings",
		"--no-playlist",
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return Track{}, fmt.Errorf("yt-dlp resolve %q: %w", query, err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 3 {
		return Track{}, fmt.Errorf("yt-dlp resolve %q: unexpected output (got %d lines)", query, len(lines))
	}

	return Track{
		ID:    strings.TrimSpace(lines[0]),
		Title: strings.TrimSpace(lines[1]),
		URL:   strings.TrimSpace(lines[2]),
		Query: query,
	}, nil
}

// FetchMixTracks fetches track titles from the YouTube Mix (RD<id>) for the
// given video ID. Returns up to maxResults track titles as search queries.
func FetchMixTracks(ctx context.Context, videoID string, maxResults int) ([]string, error) {
	mixURL := "https://www.youtube.com/watch?v=" + videoID + "&list=RD" + videoID

	args := []string{
		mixURL,
		"--flat-playlist",
		"--print", "title",
		"--no-warnings",
		"--force-ipv4",
		"--playlist-end", fmt.Sprintf("%d", maxResults),
	}

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp fetch mix for %s: %w", videoID, err)
	}

	var titles []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			titles = append(titles, line)
		}
	}
	return titles, nil
}
