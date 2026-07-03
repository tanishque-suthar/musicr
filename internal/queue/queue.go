package queue

import (
	"context"
	"fmt"
	"sync"

	"github.com/tanishque-suthar/musicr/internal/ytdlp"
)

// Queue manages an ordered list of tracks for playback.
// Tracks start as unresolved (query string only) and are resolved
// just-in-time via yt-dlp before playback.
type Queue struct {
	mu      sync.RWMutex
	tracks  []ytdlp.Track
	current int // index of the currently playing track, -1 if none
}

// New creates an empty queue.
func New() *Queue {
	return &Queue{
		current: -1,
	}
}

// Add appends an unresolved track (search query) to the queue.
func (q *Queue) Add(query string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, ytdlp.Track{Query: query})
}

// AddTrack appends a pre-populated track to the queue.
func (q *Queue) AddTrack(t ytdlp.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, t)
}

// AddTracks appends multiple queries as unresolved tracks.
func (q *Queue) AddTracks(queries []string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, query := range queries {
		q.tracks = append(q.tracks, ytdlp.Track{Query: query})
	}
}

// InsertAt inserts an unresolved track at the given index, shifting
// existing tracks right. Adjusts current index if needed.
func (q *Queue) InsertAt(index int, query string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 {
		index = 0
	}
	if index > len(q.tracks) {
		index = len(q.tracks)
	}
	q.tracks = append(q.tracks, ytdlp.Track{})
	copy(q.tracks[index+1:], q.tracks[index:])
	q.tracks[index] = ytdlp.Track{Query: query}
	if index <= q.current {
		q.current++
	}
}

// Remove removes the track at the given index (0-based).
// Returns false if the index is out of range, and whether the removed
// track was the currently playing one.
func (q *Queue) Remove(index int) (ok bool, removedCurrent bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 || index >= len(q.tracks) {
		return false, false
	}
	wasCurrent := index == q.current
	q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)
	if index < q.current {
		q.current--
	} else if index == q.current {
		if q.current >= len(q.tracks) {
			q.current = len(q.tracks) - 1
		}
	}
	return true, wasCurrent
}

// Current returns the currently playing track and its index.
// Returns an empty track and -1 if nothing is playing.
func (q *Queue) Current() (ytdlp.Track, int) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.current < 0 || q.current >= len(q.tracks) {
		return ytdlp.Track{}, -1
	}
	return q.tracks[q.current], q.current
}

// SetCurrent sets the current track index.
func (q *Queue) SetCurrent(index int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = index
}

// Next advances to the next track and returns it.
// Returns false if already at the end.
func (q *Queue) Next() (ytdlp.Track, int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	next := q.current + 1
	if next >= len(q.tracks) {
		return ytdlp.Track{}, -1, false
	}
	q.current = next
	return q.tracks[next], next, true
}

// Prev goes to the previous track and returns it.
// Returns false if already at the beginning.
func (q *Queue) Prev() (ytdlp.Track, int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	prev := q.current - 1
	if prev < 0 {
		return ytdlp.Track{}, -1, false
	}
	q.current = prev
	return q.tracks[prev], prev, true
}

// Tracks returns a copy of all tracks in the queue.
func (q *Queue) Tracks() []ytdlp.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]ytdlp.Track, len(q.tracks))
	copy(out, q.tracks)
	return out
}

// Len returns the number of tracks in the queue.
func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks)
}

// Remaining returns how many tracks are left after the current one.
func (q *Queue) Remaining() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.current < 0 {
		return len(q.tracks)
	}
	rem := len(q.tracks) - q.current - 1
	if rem < 0 {
		return 0
	}
	return rem
}

// ResolveTrack resolves the track at the given index using yt-dlp.
// Updates the track in-place with the resolved ID and Title.
func (q *Queue) ResolveTrack(ctx context.Context, index int) (ytdlp.Track, error) {
	q.mu.RLock()
	if index < 0 || index >= len(q.tracks) {
		q.mu.RUnlock()
		return ytdlp.Track{}, fmt.Errorf("track index %d out of range (len=%d)", index, len(q.tracks))
	}
	track := q.tracks[index]
	q.mu.RUnlock()

	if track.Resolved() {
		return track, nil
	}

	resolved, err := ytdlp.Resolve(ctx, track.Query)
	if err != nil {
		return ytdlp.Track{}, err
	}

	q.mu.Lock()
	defer q.mu.Unlock()
	// Double-check index is still valid (queue might have changed)
	if index < len(q.tracks) {
		q.tracks[index] = resolved
	}
	return resolved, nil
}

// PeekNext returns the next track without advancing.
func (q *Queue) PeekNext() (ytdlp.Track, int, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	next := q.current + 1
	if next >= len(q.tracks) {
		return ytdlp.Track{}, -1, false
	}
	return q.tracks[next], next, true
}

// Titles returns all track display names (Title if resolved, Query otherwise).
func (q *Queue) Titles() []string {
	q.mu.RLock()
	defer q.mu.RUnlock()
	titles := make([]string, len(q.tracks))
	for i, t := range q.tracks {
		if t.Title != "" {
			titles[i] = t.Title
		} else {
			titles[i] = t.Query
		}
	}
	return titles
}
