package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

// downloadYTDLP downloads yt-dlp binary from GitHub releases
func downloadYTDLP(url, targetPath string) error {
	if url == "" {
		return fmt.Errorf("yt-dlp download URL not available for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading from: %s\n", url)
	return downloadFile(url, targetPath)
}

// downloadMPV downloads mpv binary from release sources
func downloadMPV(url, targetPath string) error {
	if url == "" {
		return fmt.Errorf("mpv download not supported for %s/%s - please install mpv manually:\n"+
			"  macOS: brew install mpv\n"+
			"  Windows: choco install mpv or download from https://sourceforge.net/projects/mpv-player-windows/\n"+
			"  Linux: apt install mpv (Debian/Ubuntu) or pacman -S mpv (Arch)", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Downloading from: %s\n", url)

	// Note: Windows mpv is usually a .7z archive, needs extraction
	if runtime.GOOS == "windows" {
		tempPath := targetPath + ".7z"
		if err := downloadFile(url, tempPath); err != nil {
			return err
		}
		// TODO: Extract .7z file
		// For now, this is a placeholder - user should install via chocolatey or manually
		return fmt.Errorf("mpv for Windows requires manual installation or extraction from .7z archive")
	}

	return downloadFile(url, targetPath)
}

// downloadFile downloads a file from a URL to a local path
func downloadFile(url, targetPath string) error {
	// Create the file
	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Get the file from the URL
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Make executable on Unix-like systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to make file executable: %w", err)
		}
	}

	return nil
}
