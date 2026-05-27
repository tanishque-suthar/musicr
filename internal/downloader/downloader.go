package downloader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// EnsureDependency checks if a binary exists in PATH or cache, and downloads it if necessary
func EnsureDependency(binaryName, cachePath string) (string, error) {
	// First, check if binary is already in PATH
	pathExe, err := exec.LookPath(binaryName)
	if err == nil {
		return pathExe, nil
	}

	// Not in PATH, check cache
	cachedExe := getCachedBinaryPath(binaryName, cachePath)
	if _, err := os.Stat(cachedExe); err == nil {
		// Make it executable (Unix-like)
		if runtime.GOOS != "windows" {
			os.Chmod(cachedExe, 0755)
		}
		return cachedExe, nil
	}

	// Not in PATH or cache, download it
	fmt.Printf("Downloading %s (this may take a moment)...\n", binaryName)
	if err := downloadBinary(binaryName, cachePath); err != nil {
		return "", fmt.Errorf("failed to download %s: %w", binaryName, err)
	}

	cachedExe = getCachedBinaryPath(binaryName, cachePath)
	if runtime.GOOS != "windows" {
		os.Chmod(cachedExe, 0755)
	}

	return cachedExe, nil
}

// getCachedBinaryPath returns the path where a binary should be cached
func getCachedBinaryPath(binaryName, cachePath string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(cachePath, binaryName+".exe")
	}
	return filepath.Join(cachePath, binaryName)
}

// downloadBinary downloads the latest stable version of a binary
func downloadBinary(binaryName, cachePath string) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(cachePath, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	downloadURL := getDownloadURL(binaryName)
	if downloadURL == "" {
		return fmt.Errorf("unsupported binary: %s", binaryName)
	}

	cachedPath := getCachedBinaryPath(binaryName, cachePath)

	switch binaryName {
	case "yt-dlp":
		return downloadYTDLP(downloadURL, cachedPath)
	case "mpv":
		return downloadMPV(downloadURL, cachedPath)
	default:
		return fmt.Errorf("unsupported binary: %s", binaryName)
	}
}

// getDownloadURL returns the appropriate download URL for the current platform
func getDownloadURL(binaryName string) string {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch binaryName {
	case "yt-dlp":
		return getYTDLPURL(os, arch)
	case "mpv":
		return getMPVURL(os, arch)
	default:
		return ""
	}
}

// getYTDLPURL returns the latest yt-dlp download URL for the platform
func getYTDLPURL(os, arch string) string {
	// yt-dlp releases: https://github.com/yt-dlp/yt-dlp/releases
	baseURL := "https://github.com/yt-dlp/yt-dlp/releases/latest/download"

	switch os {
	case "darwin":
		if arch == "arm64" {
			return baseURL + "/yt-dlp_macos_legacy"
		}
		return baseURL + "/yt-dlp_macos"
	case "windows":
		return baseURL + "/yt-dlp.exe"
	case "linux":
		return baseURL + "/yt-dlp_linux"
	default:
		return ""
	}
}

// getMPVURL returns the latest mpv download URL for the platform
func getMPVURL(os, arch string) string {
	// mpv releases: https://github.com/mpv-player/mpv/releases
	baseURL := "https://github.com/mpv-player/mpv/releases/download"

	// Note: This is a simplified version. For production, we'd use the latest release API
	// For now, we'll direct to a known stable build source or use GitHub API

	switch os {
	case "darwin":
		// macOS: use homebrew or pre-built binaries (simplified)
		// In practice, recommend: brew install mpv
		return "" // Will be handled by system package manager or bundled
	case "windows":
		// Windows: https://sourceforge.net/projects/mpv-player-windows/files/
		// or use GitHub releases for mpv-build
		return "https://sourceforge.net/projects/mpv-player-windows/files/64bit-latest/mpv-x86_64-latest.7z/download"
	case "linux":
		// Linux: recommend system package manager
		// For now, return empty (will use apt/dnf/pacman)
		return ""
	default:
		return ""
	}
}

// IsInPath checks if a command is available in the system PATH
func IsInPath(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// GetExecutableName returns the platform-specific name of an executable
func GetExecutableName(baseName string) string {
	if runtime.GOOS == "windows" {
		return baseName + ".exe"
	}
	return baseName
}
