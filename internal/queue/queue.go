package queue

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"github.com/tanishque-suthar/musicr/internal/ytdlp"
)

type RepeatMode int

const (
	RepeatOff RepeatMode = iota
	RepeatOne
	RepeatAll
)

func (m RepeatMode) String() string {
	switch m {
	case RepeatOne:
		return "one"
	case RepeatAll:
		return "all"
	default:
		return "off"
	}
}

type Queue struct {
	mu            sync.RWMutex
	tracks        []ytdlp.Track
	originalOrder []ytdlp.Track
	current       int
	repeat        RepeatMode
}

func New() *Queue {
	return &Queue{
		current: -1,
	}
}

func (q *Queue) Add(query string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, ytdlp.Track{Query: query})
	if q.originalOrder != nil {
		q.originalOrder = append(q.originalOrder, ytdlp.Track{Query: query})
	}
}

func (q *Queue) AddTrack(t ytdlp.Track) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.tracks = append(q.tracks, t)
	if q.originalOrder != nil {
		q.originalOrder = append(q.originalOrder, t)
	}
}

func (q *Queue) AddTracks(queries []string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for _, query := range queries {
		q.tracks = append(q.tracks, ytdlp.Track{Query: query})
		if q.originalOrder != nil {
			q.originalOrder = append(q.originalOrder, ytdlp.Track{Query: query})
		}
	}
}

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
	if q.originalOrder != nil {
		q.originalOrder = append(q.originalOrder, ytdlp.Track{})
		copy(q.originalOrder[index+1:], q.originalOrder[index:])
		q.originalOrder[index] = ytdlp.Track{Query: query}
	}
}

func (q *Queue) Remove(index int) (ok bool, removedCurrent bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if index < 0 || index >= len(q.tracks) {
		return false, false
	}
	wasCurrent := index == q.current
	q.tracks = append(q.tracks[:index], q.tracks[index+1:]...)
	if q.originalOrder != nil {
		q.originalOrder = append(q.originalOrder[:index], q.originalOrder[index+1:]...)
	}
	if index < q.current {
		q.current--
	} else if index == q.current {
		if q.current >= len(q.tracks) {
			q.current = len(q.tracks) - 1
		}
	}
	return true, wasCurrent
}

func (q *Queue) Current() (ytdlp.Track, int) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.current < 0 || q.current >= len(q.tracks) {
		return ytdlp.Track{}, -1
	}
	return q.tracks[q.current], q.current
}

func (q *Queue) SetCurrent(index int) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.current = index
}

func (q *Queue) Next() (ytdlp.Track, int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	next := q.current + 1
	if next >= len(q.tracks) {
		if q.repeat == RepeatAll && len(q.tracks) > 0 {
			next = 0
		} else {
			return ytdlp.Track{}, -1, false
		}
	}
	q.current = next
	return q.tracks[next], next, true
}

func (q *Queue) Prev() (ytdlp.Track, int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	prev := q.current - 1
	if prev < 0 {
		if q.repeat == RepeatAll && len(q.tracks) > 0 {
			prev = len(q.tracks) - 1
		} else {
			return ytdlp.Track{}, -1, false
		}
	}
	q.current = prev
	return q.tracks[prev], prev, true
}

func (q *Queue) Tracks() []ytdlp.Track {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]ytdlp.Track, len(q.tracks))
	copy(out, q.tracks)
	return out
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.tracks)
}

func (q *Queue) Remaining() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.current < 0 {
		return len(q.tracks)
	}
	if q.repeat == RepeatAll && len(q.tracks) > 0 {
		return len(q.tracks)
	}
	rem := len(q.tracks) - q.current - 1
	if rem < 0 {
		return 0
	}
	return rem
}

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
	if index < len(q.tracks) {
		q.tracks[index] = resolved
	}
	return resolved, nil
}

func (q *Queue) PeekNext() (ytdlp.Track, int, bool) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	next := q.current + 1
	if next >= len(q.tracks) {
		if q.repeat == RepeatAll && len(q.tracks) > 0 {
			next = 0
		} else {
			return ytdlp.Track{}, -1, false
		}
	}
	return q.tracks[next], next, true
}

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

func (q *Queue) SetRepeat(m RepeatMode) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.repeat = m
}

func (q *Queue) Repeat() RepeatMode {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.repeat
}

func (q *Queue) Shuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.tracks) <= 1 {
		return
	}
	q.originalOrder = make([]ytdlp.Track, len(q.tracks))
	copy(q.originalOrder, q.tracks)
	rand.Shuffle(len(q.tracks), func(i, j int) {
		q.tracks[i], q.tracks[j] = q.tracks[j], q.tracks[i]
	})
	q.current = 0
}

func (q *Queue) Unshuffle() {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.originalOrder == nil {
		return
	}
	currentID := ""
	if q.current >= 0 && q.current < len(q.tracks) {
		currentID = q.tracks[q.current].ID
	}
	q.tracks = q.originalOrder
	q.originalOrder = nil
	q.current = 0
	if currentID != "" {
		for i, t := range q.tracks {
			if t.ID == currentID {
				q.current = i
				break
			}
		}
	}
}

func (q *Queue) IsShuffled() bool {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.originalOrder != nil
}
