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

func TestParseArgsMultipleQueries(t *testing.T) {
	cfg, action := parseArgs([]string{"song a", "song b"})
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
	if len(cfg.Queries) != 2 {
		t.Fatalf("expected 2 queries, got %d", len(cfg.Queries))
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
}
