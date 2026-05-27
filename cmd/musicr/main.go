package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"musicr/internal/config"
	"musicr/internal/downloader"
	"musicr/internal/player"
	"musicr/internal/queue"
	"musicr/internal/ui"
	"musicr/internal/ytdlp"
)

const Version = "1.0.0"

func main() {
	// Define command-line flags
	fs := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	interactive := fs.Bool("i", false, "Start interactive search mode")
	interactiveLong := fs.Bool("interactive", false, "Start interactive search mode")
	searchQuery := fs.String("s", "", "Search for a song and play it")
	searchLong := fs.String("search", "", "Search for a song and play it")
	showConfig := fs.Bool("c", false, "Show config directory location")
	configLong := fs.Bool("config", false, "Show config directory location")
	setupConfig := fs.Bool("setup", false, "Initialize/reset configuration")
	showVersion := fs.Bool("v", false, "Show version")
	versionLong := fs.Bool("version", false, "Show version")
	showHelp := fs.Bool("h", false, "Show help message")
	helpLong := fs.Bool("help", false, "Show help message")

	fs.Parse(os.Args[1:])
	
	// Merge short and long flags
	*interactive = *interactive || *interactiveLong
	*searchQuery = orString(*searchQuery, *searchLong)
	*showConfig = *showConfig || *configLong
	*showVersion = *showVersion || *versionLong
	*showHelp = *showHelp || *helpLong

	userUI := ui.NewUI()

	// Handle flags
	if *showHelp {
		userUI.RenderHelp()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("musicr v%s\n", Version)
		os.Exit(0)
	}

	// Load configuration
	configPath, err := config.ConfigPath()
	if err != nil {
		userUI.RenderError(fmt.Sprintf("Failed to determine config path: %v", err))
		os.Exit(1)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		userUI.RenderError(fmt.Sprintf("Failed to load config: %v", err))
		os.Exit(1)
	}

	if *showConfig {
		fmt.Println(configPath)
		os.Exit(0)
	}

	if *setupConfig {
		if err := cfg.Save(configPath); err != nil {
			userUI.RenderError(fmt.Sprintf("Failed to save config: %v", err))
			os.Exit(1)
		}
		fmt.Printf("✓ Config saved to: %s\n", configPath)
		os.Exit(0)
	}

	// Get search query
	var query string
	if *searchQuery != "" {
		query = *searchQuery
	} else if *interactive {
		query = userUI.RenderInteractivePrompt()
	} else if fs.NArg() > 0 {
		query = strings.Join(fs.Args(), " ")
	} else {
		userUI.RenderHelp()
		os.Exit(0)
	}

	// Check for and download dependencies
	ytdlpPath, err := downloader.EnsureDependency("yt-dlp", cfg.BinCachePath)
	if err != nil {
		userUI.RenderError(fmt.Sprintf("Failed to ensure yt-dlp: %v", err))
		os.Exit(1)
	}

	mpvPath, err := downloader.EnsureDependency("mpv", cfg.BinCachePath)
	if err != nil {
		userUI.RenderDependencyMissing("mpv")
		fmt.Println()
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Create yt-dlp client
	ytdlpConfig := ytdlp.YTDLPConfig{
		VideoFormat:  cfg.VideoFormat,
		PlayerClient: cfg.PlayerClient,
	}
	ytdlpClient := ytdlp.NewClient(ytdlpPath, ytdlpConfig)

	// Create mpv player
	playerConfig := player.PlayerConfig{
		VideoFormat:     cfg.VideoFormat,
		PlayerClient:    cfg.PlayerClient,
		DemuxerMaxBytes: cfg.DemuxerMaxBytes,
		TermStatusMsg:   cfg.TermStatusMsg,
		CachePath:       cfg.CachePath,
	}
	mpvPlayer := player.NewPlayer(mpvPath, playerConfig)

	// Show searching UI
	userUI.RenderSearching(query)

	// Search for the video
	videoID, title, streamURL, err := ytdlpClient.SearchVideo(query)
	if err != nil {
		userUI.RenderError(fmt.Sprintf("Could not find: %s\nError: %v", query, err))
		os.Exit(1)
	}

	// Clean title for display
	cleanedTitle := ytdlp.CleanTitle(title)

	// Create queue with radio mix
	radioURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=RD%s", videoID, videoID)
	q := queue.NewQueue(ytdlpClient, radioURL)

	// Add first track to queue
	q.AddTrack(cleanedTitle, streamURL)

	// Start background fetching of playlist
	q.StartBackgroundFetch()

	// Play the first track
	if err := mpvPlayer.Play(streamURL, radioURL, cleanedTitle); err != nil {
		userUI.RenderError(fmt.Sprintf("Failed to start player: %v", err))
		os.Exit(1)
	}

	// Update UI with tracks from queue as they're added
	go func() {
		for update := range q.Updates() {
			if update.Type == queue.UpdateTypeAdded {
				userUI.SetTracks(q.GetTracks())
				userUI.Render()
			}
		}
	}()

	// Render initial UI
	userUI.SetTracks(q.GetTracks())
	userUI.Render()

	// Wait for playback to finish
	if err := mpvPlayer.Wait(); err != nil {
		// Player was closed by user, exit gracefully
	}

	q.Close()
}

// Helper functions for flag merging
func orBool(a, b bool) bool {
	return a || b
}

func orString(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
