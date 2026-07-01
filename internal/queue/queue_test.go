package queue

import (
	"context"
	"testing"
)

func TestNewQueue(t *testing.T) {
	q := New()
	if q.Len() != 0 {
		t.Fatalf("expected empty queue, got len %d", q.Len())
	}
	if _, idx := q.Current(); idx != -1 {
		t.Fatalf("expected current -1, got %d", idx)
	}
}

func TestAddAndLen(t *testing.T) {
	q := New()
	q.Add("track one")
	q.Add("track two")
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
}

func TestAddTracks(t *testing.T) {
	q := New()
	q.AddTracks([]string{"a", "b", "c"})
	if q.Len() != 3 {
		t.Fatalf("expected len 3, got %d", q.Len())
	}
}

func TestCurrentOnEmpty(t *testing.T) {
	q := New()
	_, idx := q.Current()
	if idx != -1 {
		t.Fatalf("expected -1, got %d", idx)
	}
}

func TestSetCurrent(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(1)
	_, idx := q.Current()
	if idx != 1 {
		t.Fatalf("expected current 1, got %d", idx)
	}
}

func TestNext(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)

	track, idx, ok := q.Next()
	if !ok {
		t.Fatal("expected ok")
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
	if track.Query != "b" {
		t.Fatalf("expected query 'b', got %q", track.Query)
	}
}

func TestNextAtEnd(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	_, _, ok := q.Next()
	if ok {
		t.Fatal("expected not ok at end")
	}
}

func TestPrev(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(1)

	track, idx, ok := q.Prev()
	if !ok {
		t.Fatal("expected ok")
	}
	if idx != 0 {
		t.Fatalf("expected idx 0, got %d", idx)
	}
	if track.Query != "a" {
		t.Fatalf("expected query 'a', got %q", track.Query)
	}
}

func TestPrevAtStart(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	_, _, ok := q.Prev()
	if ok {
		t.Fatal("expected not ok at start")
	}
}

func TestRemove(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.SetCurrent(1)

	ok, removedCurrent := q.Remove(0)
	if !ok {
		t.Fatal("expected ok")
	}
	if removedCurrent {
		t.Fatal("expected removedCurrent false")
	}
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
	// Current should have shifted from 1 to 0
	_, idx := q.Current()
	if idx != 0 {
		t.Fatalf("expected current 0 after removal, got %d", idx)
	}
}

func TestRemoveCurrent(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.SetCurrent(1)

	ok, removedCurrent := q.Remove(1)
	if !ok {
		t.Fatal("expected ok")
	}
	if !removedCurrent {
		t.Fatal("expected removedCurrent true")
	}
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
}

func TestRemoveLastItem(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	ok, removedCurrent := q.Remove(0)
	if !ok {
		t.Fatal("expected ok")
	}
	if !removedCurrent {
		t.Fatal("expected removedCurrent true")
	}
	if q.Len() != 0 {
		t.Fatalf("expected len 0, got %d", q.Len())
	}
	_, idx := q.Current()
	if idx != -1 {
		t.Fatalf("expected current -1, got %d", idx)
	}
}

func TestRemoveInvalidIndex(t *testing.T) {
	q := New()
	q.Add("a")

	ok, _ := q.Remove(5)
	if ok {
		t.Fatal("expected not ok")
	}
	ok, _ = q.Remove(-1)
	if ok {
		t.Fatal("expected not ok for negative index")
	}
}

func TestRemaining(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.SetCurrent(0)

	if rem := q.Remaining(); rem != 2 {
		t.Fatalf("expected 2 remaining, got %d", rem)
	}
	q.SetCurrent(2)
	if rem := q.Remaining(); rem != 0 {
		t.Fatalf("expected 0 remaining, got %d", rem)
	}
}

func TestRemainingEmpty(t *testing.T) {
	q := New()
	if rem := q.Remaining(); rem != 0 {
		t.Fatalf("expected 0 remaining, got %d", rem)
	}
}

func TestPeekNext(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)

	track, idx, ok := q.PeekNext()
	if !ok {
		t.Fatal("expected ok")
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
	if track.Query != "b" {
		t.Fatalf("expected 'b', got %q", track.Query)
	}
	// PeekNext should not advance
	_, curIdx := q.Current()
	if curIdx != 0 {
		t.Fatalf("expected current still 0, got %d", curIdx)
	}
}

func TestPeekNextAtEnd(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	_, _, ok := q.PeekNext()
	if ok {
		t.Fatal("expected not ok at end")
	}
}

func TestTitles(t *testing.T) {
	q := New()
	q.Add("query one")
	q.Add("query two")

	titles := q.Titles()
	if len(titles) != 2 {
		t.Fatalf("expected 2 titles, got %d", len(titles))
	}
	if titles[0] != "query one" {
		t.Fatalf("expected 'query one', got %q", titles[0])
	}
}

func TestTracksCopy(t *testing.T) {
	q := New()
	q.Add("a")
	tracks := q.Tracks()
	if len(tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(tracks))
	}
	// Modify original queue
	q.Add("b")
	if len(tracks) != 1 {
		t.Fatal("returned slice should be a copy")
	}
}

func TestResolveTrackOutOfRange(t *testing.T) {
	q := New()
	_, err := q.ResolveTrack(context.Background(), 0)
	if err != nil {
		t.Fatalf("expected nil error for out of range, got %v", err)
	}
}

func TestConcurrentSafe(t *testing.T) {
	q := New()
	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			q.Add("track")
		}
		close(done)
	}()
	for i := 0; i < 100; i++ {
		q.Len()
	}
	<-done
	if q.Len() != 100 {
		t.Fatalf("expected 100, got %d", q.Len())
	}
}
