package app

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestDir(t *testing.T) func() {
	orig := playlistDirFn
	dir := t.TempDir()
	playlistDirFn = func() (string, error) {
		return dir, nil
	}
	return func() {
		playlistDirFn = orig
	}
}

func TestSaveAndLoadPlaylist(t *testing.T) {
	defer setupTestDir(t)()

	err := savePlaylistFile("test_playlist", []string{"track one", "track two"})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	tracks, err := loadPlaylistFile("test_playlist")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks, got %d", len(tracks))
	}
	if tracks[0] != "track one" {
		t.Fatalf("expected 'track one', got %q", tracks[0])
	}
	if tracks[1] != "track two" {
		t.Fatalf("expected 'track two', got %q", tracks[1])
	}
}

func TestSaveAndLoadSkipsEmptyLines(t *testing.T) {
	defer setupTestDir(t)()

	err := savePlaylistFile("test", []string{"a", "", "b"})
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}
	tracks, err := loadPlaylistFile("test")
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(tracks) != 2 {
		t.Fatalf("expected 2 tracks (empty lines skipped), got %d", len(tracks))
	}
}

func TestListPlaylists(t *testing.T) {
	defer setupTestDir(t)()

	savePlaylistFile("alpha", []string{"a"})
	savePlaylistFile("beta", []string{"b"})

	names, err := ListPlaylists()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestListPlaylistsEmpty(t *testing.T) {
	defer setupTestDir(t)()

	names, err := ListPlaylists()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(names) != 0 {
		t.Fatalf("expected empty list, got %d", len(names))
	}
}

func TestListPlaylistsSkipsNonTxt(t *testing.T) {
	defer setupTestDir(t)()

	os.WriteFile(filepath.Join(t.TempDir(), "notes.md"), []byte("hello"), 0644)
	// Re-point to this dir with the extra file
	playlistDirFn = func() (string, error) { return t.TempDir(), nil }
	savePlaylistFile("mymusic", []string{"track"})
	playlistDirFn = func() (string, error) { return t.TempDir(), nil }

	names, err := ListPlaylists()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	for _, n := range names {
		if n == "notes" {
			t.Fatal("should not list non-.txt files")
		}
	}
}

func TestSaveEmptyPlaylist(t *testing.T) {
	defer setupTestDir(t)()

	err := savePlaylistFile("empty", nil)
	if err != nil {
		t.Fatalf("save empty failed: %v", err)
	}
	tracks, err := loadPlaylistFile("empty")
	if err != nil {
		t.Fatalf("load empty failed: %v", err)
	}
	if len(tracks) != 0 {
		t.Fatalf("expected 0 tracks, got %d", len(tracks))
	}
}

func TestLoadNonexistentPlaylist(t *testing.T) {
	defer setupTestDir(t)()

	_, err := loadPlaylistFile("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent playlist")
	}
}
