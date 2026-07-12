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
	if err == nil {
		t.Fatal("expected error for out of range, got nil")
	}
}

func TestInsertAt(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("c")
	q.SetCurrent(0)

	q.InsertAt(1, "b")
	if q.Len() != 3 {
		t.Fatalf("expected len 3, got %d", q.Len())
	}
	titles := q.Titles()
	if titles[1] != "b" {
		t.Fatalf("expected 'b' at index 1, got %q", titles[1])
	}
	if titles[2] != "c" {
		t.Fatalf("expected 'c' at index 2, got %q", titles[2])
	}
}

func TestInsertAtBeforeCurrent(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(1)

	q.InsertAt(0, "x")
	_, idx := q.Current()
	if idx != 2 {
		t.Fatalf("expected current 2 after insert before, got %d", idx)
	}
}

func TestInsertAtAfterCurrent(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("c")
	q.SetCurrent(0)

	q.InsertAt(1, "b")
	_, idx := q.Current()
	if idx != 0 {
		t.Fatalf("expected current unchanged at 0, got %d", idx)
	}
}

func TestInsertAtClampNegative(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	q.InsertAt(-5, "first")
	_, idx := q.Current()
	if idx != 1 {
		t.Fatalf("expected current 1 after insert clamped to 0, got %d", idx)
	}
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
}

func TestInsertAtClampBeyond(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	q.InsertAt(10, "last")
	if q.Len() != 2 {
		t.Fatalf("expected len 2, got %d", q.Len())
	}
}

func TestRepeatModeString(t *testing.T) {
	if RepeatOff.String() != "off" {
		t.Fatalf("expected 'off', got %q", RepeatOff.String())
	}
	if RepeatOne.String() != "one" {
		t.Fatalf("expected 'one', got %q", RepeatOne.String())
	}
	if RepeatAll.String() != "all" {
		t.Fatalf("expected 'all', got %q", RepeatAll.String())
	}
}

func TestNextRepeatAllWraps(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(1)
	q.SetRepeat(RepeatAll)

	track, idx, ok := q.Next()
	if !ok {
		t.Fatal("expected ok for RepeatAll wrap")
	}
	if idx != 0 {
		t.Fatalf("expected idx 0, got %d", idx)
	}
	if track.Query != "a" {
		t.Fatalf("expected query 'a', got %q", track.Query)
	}
}

func TestNextRepeatAllSingleTrack(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)
	q.SetRepeat(RepeatAll)

	track, idx, ok := q.Next()
	if !ok {
		t.Fatal("expected ok with single track RepeatAll")
	}
	if idx != 0 {
		t.Fatalf("expected idx 0, got %d", idx)
	}
	if track.Query != "a" {
		t.Fatalf("expected 'a', got %q", track.Query)
	}
}

func TestPrevRepeatAllWraps(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)
	q.SetRepeat(RepeatAll)

	track, idx, ok := q.Prev()
	if !ok {
		t.Fatal("expected ok for RepeatAll wrap")
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
	if track.Query != "b" {
		t.Fatalf("expected query 'b', got %q", track.Query)
	}
}

func TestNextRepeatOneAdvances(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)
	q.SetRepeat(RepeatOne)

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

func TestPeekNextRepeatAllWraps(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(1)
	q.SetRepeat(RepeatAll)

	track, idx, ok := q.PeekNext()
	if !ok {
		t.Fatal("expected ok for RepeatAll peek")
	}
	if idx != 0 {
		t.Fatalf("expected idx 0, got %d", idx)
	}
	if track.Query != "a" {
		t.Fatalf("expected query 'a', got %q", track.Query)
	}
	_, cur := q.Current()
	if cur != 1 {
		t.Fatalf("expected current unchanged at 1, got %d", cur)
	}
}

func TestRemainingRepeatAll(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)
	q.SetRepeat(RepeatAll)

	if rem := q.Remaining(); rem != 2 {
		t.Fatalf("expected 2 remaining with RepeatAll, got %d", rem)
	}
}

func TestShuffle(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.Add("d")
	q.SetCurrent(0)

	original := q.Titles()
	q.Shuffle()

	if !q.IsShuffled() {
		t.Fatal("expected shuffled")
	}
	if q.Len() != 4 {
		t.Fatalf("expected len 4, got %d", q.Len())
	}
	_, cur := q.Current()
	if cur != 0 {
		t.Fatalf("expected current 0 after shuffle, got %d", cur)
	}
	shuffled := q.Titles()
	sameOrder := true
	for i := range original {
		if original[i] != shuffled[i] {
			sameOrder = false
			break
		}
	}
	if sameOrder {
		t.Fatal("shuffle should have changed the order")
	}
}

func TestUnshuffle(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.Add("c")
	q.SetCurrent(0)

	original := q.Titles()
	q.Shuffle()
	q.Unshuffle()

	if q.IsShuffled() {
		t.Fatal("expected not shuffled after unshuffle")
	}
	restored := q.Titles()
	for i := range original {
		if original[i] != restored[i] {
			t.Fatalf("expected %q at %d, got %q", original[i], i, restored[i])
		}
	}
}

func TestUnshuffleWithoutShuffle(t *testing.T) {
	q := New()
	q.Add("a")
	q.Unshuffle() // should be a no-op
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestShuffleSingleTrack(t *testing.T) {
	q := New()
	q.Add("a")
	q.SetCurrent(0)

	q.Shuffle()
	// Shuffle with 1 track is a no-op
	if q.IsShuffled() {
		t.Fatal("expected not shuffled for single track")
	}
	if q.Len() != 1 {
		t.Fatalf("expected len 1, got %d", q.Len())
	}
}

func TestAddDuringShuffle(t *testing.T) {
	q := New()
	q.Add("a")
	q.Add("b")
	q.SetCurrent(0)
	q.Shuffle()
	before := q.Len()

	q.Add("c")
	if q.Len() != before+1 {
		t.Fatalf("expected len %d, got %d", before+1, q.Len())
	}
	q.Unshuffle()
	titles := q.Titles()
	if titles[2] != "c" {
		t.Fatalf("expected 'c' at index 2 after unshuffle, got %q", titles[2])
	}
}

func TestSetRepeat(t *testing.T) {
	q := New()
	if q.Repeat() != RepeatOff {
		t.Fatal("expected RepeatOff initially")
	}
	q.SetRepeat(RepeatOne)
	if q.Repeat() != RepeatOne {
		t.Fatal("expected RepeatOne")
	}
	q.SetRepeat(RepeatAll)
	if q.Repeat() != RepeatAll {
		t.Fatal("expected RepeatAll")
	}
	q.SetRepeat(RepeatOff)
	if q.Repeat() != RepeatOff {
		t.Fatal("expected RepeatOff")
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
