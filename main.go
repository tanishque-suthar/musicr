package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/tanishque-suthar/musicr/internal/app"
	"github.com/tanishque-suthar/musicr/internal/update"
)

var version = "dev"

func main() {
	cfg, action := parseArgs(os.Args[1:])

	switch action {
	case actionPlay:
		checkDeps()
		a := app.New(cfg)
		if err := a.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "musicr: %v\n", err)
			os.Exit(1)
		}

	case actionList:
		listPlaylists()

	case actionSave:
		fmt.Fprintf(os.Stderr, "musicr save: use 's' key in the player to save\n")
		os.Exit(1)

	case actionHelp:
		printUsage()

	case actionVersion:
		fmt.Println("musicr version", version)

	case actionUpdate:
		doUpdate()

	case actionUpdateCheck:
		doUpdateCheck()
	}
}

type actionType int

const (
	actionPlay actionType = iota
	actionList
	actionSave
	actionHelp
	actionVersion
	actionUpdate
	actionUpdateCheck
)

func parseArgs(args []string) (app.Config, actionType) {
	cfg := app.Config{}

	if len(args) == 0 {
		return cfg, actionHelp
	}

	i := 0
	for i < len(args) {
		switch args[i] {
		case "-h", "--help":
			return cfg, actionHelp

		case "-v", "--version":
			return cfg, actionVersion

		case "--update":
			return cfg, actionUpdate

		case "--update-check":
			return cfg, actionUpdateCheck

		case "-p", "--playlist":
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "musicr: -p requires a playlist name\n")
				os.Exit(1)
			}
			cfg.Playlist = args[i+1]
			i += 2

		case "--no-radio":
			cfg.NoRadio = true
			i++

		case "list":
			if i == 0 {
				return cfg, actionList
			}
			cfg.Queries = append(cfg.Queries, args[i])
			i++

		case "save":
			if i == 0 {
				return cfg, actionSave
			}
			cfg.Queries = append(cfg.Queries, args[i])
			i++

		default:
			raw := strings.Join(args[i:], " ")
			raw = strings.ReplaceAll(raw, "\"", "")
			for _, q := range strings.Split(raw, ",") {
				if trimmed := strings.TrimSpace(q); trimmed != "" {
					cfg.Queries = append(cfg.Queries, trimmed)
				}
			}
			i = len(args)
		}
	}

	return cfg, actionPlay
}

// checkDeps verifies that mpv and yt-dlp are installed.
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
		fmt.Fprintf(os.Stderr, "  Linux:  use your distro's package manager (e.g. apt install mpv yt-dlp,\n")
		fmt.Fprintf(os.Stderr, "          or pip install -U yt-dlp for the latest yt-dlp)\n")
		os.Exit(1)
	}
}

// listPlaylists prints all saved playlists to stdout.
func listPlaylists() {
	names, err := app.ListPlaylists()
	if err != nil {
		fmt.Fprintf(os.Stderr, "musicr: %v\n", err)
		os.Exit(1)
	}
	if len(names) == 0 {
		fmt.Println("No saved playlists.")
		return
	}
	for _, n := range names {
		fmt.Printf("  %s\n", n)
	}
}

func doUpdate() {
	latest, needsUpdate, err := update.Check(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "musicr: update check failed: %v\n", err)
		os.Exit(1)
	}
	if !needsUpdate {
		if version == "dev" {
			fmt.Println("musicr: unknown current version, performing update anyway...")
		} else {
			fmt.Printf("Already up to date (%s)\n", version)
			return
		}
	}
	fmt.Printf("Updating to %s...\n", latest)
	if err := update.Apply(latest); err != nil {
		fmt.Fprintf(os.Stderr, "musicr: update failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Updated to %s. Restart musicr to use the new version.\n", latest)
}

func doUpdateCheck() {
	latest, needsUpdate, err := update.Check(version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "musicr: update check failed: %v\n", err)
		os.Exit(1)
	}
	if needsUpdate {
		fmt.Printf("Update available: %s (current: %s)\n", latest, version)
	} else {
		fmt.Printf("Already up to date (v%s)\n", version)
	}
}

func printUsage() {
	fmt.Println(`musicr — stream music from YouTube via mpv + yt-dlp

Usage:
  musicr <query>            Play a track (comma-separated for multiple, enters interactive mode, radio on)
  musicr -p <playlist>     Load and play a saved playlist
  musicr --no-radio        Disable auto-extend for this session
  musicr list              List saved playlists
  musicr --version         Show version
  musicr --update          Update to the latest release
  musicr --update-check    Check if an update is available

Interactive keys:
  a       Add a track (prompts for search query)
  n       Next track
  p       Previous track
  z       Cycle repeat mode (off/one/all)
  x       Toggle shuffle
  space   Pause / resume
  s       Save queue as a playlist
  l       Load a playlist
  r       Toggle radio mode
  d       Delete a track from the queue
  +/=     Volume up
  -/_     Volume down
  ]/.     Seek forward 5s
  [/,     Seek backward 5s
  q       Quit`)
}
