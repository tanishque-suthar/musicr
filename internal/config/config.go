package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	VideoFormat      string `json:"video_format"`      // e.g., "251/140/bestaudio/best"
	PlayerClient     string `json:"player_client"`     // e.g., "android"
	DemuxerMaxBytes  string `json:"demuxer_max_bytes"` // e.g., "67MiB"
	TermStatusMsg    string `json:"term_status_msg"`   // optional status message
	CachePath        string `json:"cache_path"`        // directory for caching streams
	BinCachePath     string `json:"bin_cache_path"`    // directory for cached binaries (yt-dlp, mpv)
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() Config {
	return Config{
		VideoFormat:     "251/140/bestaudio/best",
		PlayerClient:    "android",
		DemuxerMaxBytes: "67MiB",
		TermStatusMsg:   "Status: ${time-pos} / ${duration} (${percent-pos}%)",
		CachePath:       getTempDir(),
		BinCachePath:    filepath.Join(getTempDir(), "musicr-bins"),
	}
}

// Load loads configuration from a JSON file at the given path
// If the file doesn't exist, it creates it with default config
func Load(configPath string) (Config, error) {
	cfg := DefaultConfig()

	// If config file exists, load from it
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return cfg, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := json.Unmarshal(data, &cfg); err != nil {
			return cfg, fmt.Errorf("failed to parse config JSON: %w", err)
		}

		return cfg, nil
	}

	// File doesn't exist, create it with defaults
	if err := cfg.Save(configPath); err != nil {
		return cfg, fmt.Errorf("failed to create default config: %w", err)
	}

	return cfg, nil
}

// Save writes the configuration to a JSON file
func (c Config) Save(configPath string) error {
	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ConfigPath returns the path where config.json should be stored
// (in the same directory as the executable)
func ConfigPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to determine executable path: %w", err)
	}

	executableDir := filepath.Dir(exe)
	return filepath.Join(executableDir, "config.json"), nil
}

// getTempDir returns a platform-appropriate temporary directory for caching
func getTempDir() string {
	tmpDir := os.Getenv("TEMP")
	if tmpDir == "" {
		tmpDir = os.Getenv("TMP")
	}
	if tmpDir == "" {
		tmpDir = "/tmp"
	}
	return filepath.Join(tmpDir, "musicr-cache")
}
