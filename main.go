package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tsvd/musicr/internal/ytdlp"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "-h", "--help":
		printUsage()
		return
	}

	checkDeps()

	// Quick smoke test: resolve the query
	query := strings.Join(os.Args[1:], " ")
	fmt.Printf("Resolving: %s\n", query)
	track, err := ytdlp.Resolve(context.Background(), query)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found: %s (%s)\n", track.Title, track.StreamURL())
}

func checkDeps() {
	missing := []string{}
	if _, err := exec.LookPath("mpv"); err != nil {
		missing = append(missing, "mpv")
	}
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		missing = append(missing, "yt-dlp")
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "musicr requires %s.\n", strings.Join(missing, " and "))
		fmt.Fprintf(os.Stderr, "  macOS:  brew install mpv yt-dlp\n")
		fmt.Fprintf(os.Stderr, "  Linux:  use your distro's package manager\n")
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`musicr — stream music from YouTube via mpv + yt-dlp

Usage:
  musicr <query...>        Play a track
  musicr -p <playlist>     Load and play a saved playlist
  musicr --no-radio        Disable auto-extend
  musicr list              List saved playlists`)
}
