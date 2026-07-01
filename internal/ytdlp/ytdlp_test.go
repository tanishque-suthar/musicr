package ytdlp

import "testing"

func TestTrackResolved(t *testing.T) {
	track := Track{ID: "abc123", Title: "Test", Query: "test song"}
	if !track.Resolved() {
		t.Fatal("expected resolved")
	}
}

func TestTrackUnresolved(t *testing.T) {
	track := Track{Query: "test song"}
	if track.Resolved() {
		t.Fatal("expected unresolved")
	}
}

func TestStreamURL(t *testing.T) {
	track := Track{ID: "abc123"}
	expected := "https://www.youtube.com/watch?v=abc123"
	if got := track.StreamURL(); got != expected {
		t.Fatalf("StreamURL() = %q, want %q", got, expected)
	}
}

func TestStreamURLDirectURL(t *testing.T) {
	track := Track{ID: "abc123", URL: "https://example.com/stream"}
	if got := track.StreamURL(); got != "https://example.com/stream" {
		t.Fatalf("StreamURL() should prefer URL field, got %q", got)
	}
}

func TestStreamURLEmpty(t *testing.T) {
	track := Track{}
	if got := track.StreamURL(); got != "" {
		t.Fatalf("StreamURL() = %q, want empty", got)
	}
}
