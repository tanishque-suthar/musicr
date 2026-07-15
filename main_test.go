package main

import (
	"reflect"
	"testing"
)

func TestParseArgsEmpty(t *testing.T) {
	_, action := parseArgs([]string{})
	if action != actionHelp {
		t.Fatalf("expected help, got %d", action)
	}
}

func TestParseArgsHelp(t *testing.T) {
	for _, flag := range []string{"-h", "--help"} {
		_, action := parseArgs([]string{flag})
		if action != actionHelp {
			t.Fatalf("expected help for %s", flag)
		}
	}
}

func TestParseArgsQuery(t *testing.T) {
	cfg, action := parseArgs([]string{"some song"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if len(cfg.Queries) != 1 || cfg.Queries[0] != "some song" {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsMultipleWords(t *testing.T) {
	cfg, action := parseArgs([]string{"song a", "song b"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if !reflect.DeepEqual(cfg.Queries, []string{"song a song b"}) {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsCommaSeparated(t *testing.T) {
	cfg, action := parseArgs([]string{"song a, song b"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if !reflect.DeepEqual(cfg.Queries, []string{"song a", "song b"}) {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsCommaSeparatedMultipleArgs(t *testing.T) {
	cfg, action := parseArgs([]string{"song a,", "song", "b"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if !reflect.DeepEqual(cfg.Queries, []string{"song a", "song b"}) {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsPlaylist(t *testing.T) {
	cfg, action := parseArgs([]string{"-p", "mylist"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if cfg.Playlist != "mylist" {
		t.Fatalf("expected playlist 'mylist', got %q", cfg.Playlist)
	}
}

func TestParseArgsNoRadio(t *testing.T) {
	cfg, action := parseArgs([]string{"--no-radio", "song"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if !cfg.NoRadio {
		t.Fatal("expected NoRadio true")
	}
	if len(cfg.Queries) != 1 {
		t.Fatalf("expected 1 query, got %d", len(cfg.Queries))
	}
}

func TestParseArgsList(t *testing.T) {
	_, action := parseArgs([]string{"list"})
	if action != actionList {
		t.Fatalf("expected list, got %d", action)
	}
}

func TestParseArgsSave(t *testing.T) {
	_, action := parseArgs([]string{"save"})
	if action != actionSave {
		t.Fatalf("expected save, got %d", action)
	}
}

func TestParseArgsListNotFirst(t *testing.T) {
	cfg, action := parseArgs([]string{"song", "list"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if len(cfg.Queries) != 1 || cfg.Queries[0] != "song list" {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsMixed(t *testing.T) {
	cfg, action := parseArgs([]string{"--no-radio", "-p", "jazz", "song"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if !cfg.NoRadio {
		t.Fatal("expected NoRadio true")
	}
	if cfg.Playlist != "jazz" {
		t.Fatalf("expected playlist 'jazz', got %q", cfg.Playlist)
	}
	if !reflect.DeepEqual(cfg.Queries, []string{"song"}) {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsCommaAndFlags(t *testing.T) {
	cfg, action := parseArgs([]string{"-p", "jazz", "song a, song b"})
	if action != actionPlay {
		t.Fatalf("expected play, got %d", action)
	}
	if cfg.Playlist != "jazz" {
		t.Fatalf("expected playlist 'jazz', got %q", cfg.Playlist)
	}
	if !reflect.DeepEqual(cfg.Queries, []string{"song a", "song b"}) {
		t.Fatalf("unexpected queries: %v", cfg.Queries)
	}
}

func TestParseArgsVersion(t *testing.T) {
	for _, flag := range []string{"-v", "--version"} {
		_, action := parseArgs([]string{flag})
		if action != actionVersion {
			t.Fatalf("expected version for %s, got %d", flag, action)
		}
	}
}

func TestParseArgsUpdate(t *testing.T) {
	_, action := parseArgs([]string{"--update"})
	if action != actionUpdate {
		t.Fatalf("expected update, got %d", action)
	}
}

func TestParseArgsUpdateCheck(t *testing.T) {
	_, action := parseArgs([]string{"--update-check"})
	if action != actionUpdateCheck {
		t.Fatalf("expected update-check, got %d", action)
	}
}

func TestParseArgsUpdateWithQuery(t *testing.T) {
	_, action := parseArgs([]string{"--update", "song"})
	if action != actionUpdate {
		t.Fatalf("expected update, got %d", action)
	}
}
