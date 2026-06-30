package app

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// playlistDir returns the directory for storing playlists,
// creating it if necessary.
func playlistDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".config", "musicr", "playlists")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// savePlaylistFile writes track titles to a playlist file.
func savePlaylistFile(name string, titles []string) error {
	dir, err := playlistDir()
	if err != nil {
		return err
	}
	path := filepath.Join(dir, name+".txt")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, t := range titles {
		fmt.Fprintln(w, t)
	}
	return w.Flush()
}

// loadPlaylistFile reads track queries from a playlist file.
func loadPlaylistFile(name string) ([]string, error) {
	dir, err := playlistDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, name+".txt")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("playlist %q not found", name)
	}
	defer f.Close()

	var queries []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			queries = append(queries, line)
		}
	}
	return queries, scanner.Err()
}

// listPlaylists returns the names of all saved playlists.
func listPlaylists() ([]string, error) {
	dir, err := playlistDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			names = append(names, strings.TrimSuffix(e.Name(), ".txt"))
		}
	}
	return names, nil
}
