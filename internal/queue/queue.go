package queue

import (
	"fmt"
	"sync"
	"time"

	"musicr/internal/ytdlp"
)

// Track represents a track in the queue
type Track struct {
	Title   string
	URL     string
	Index   int
	Fetched bool
}

// QueueUpdate represents an update to the queue
type QueueUpdate struct {
	Type   UpdateType
	Tracks []Track
	Error  error
}

// UpdateType indicates the type of queue update
type UpdateType int

const (
	UpdateTypeAdded UpdateType = iota
	UpdateTypeError
)

// Queue manages the playlist and background fetching
type Queue struct {
	mu              sync.RWMutex
	tracks          []Track
	currentIndex    int
	ytdlpClient     *ytdlp.Client
	radioURL        string
	batchSize       int
	updates         chan QueueUpdate
	done            chan struct{}
	fetching        bool
	lastFetchedIdx  int
}

// NewQueue creates a new queue manager
func NewQueue(client *ytdlp.Client, radioURL string) *Queue {
	return &Queue{
		tracks:       []Track{},
		currentIndex: 0,
		ytdlpClient:  client,
		radioURL:     radioURL,
		batchSize:    20,
		updates:      make(chan QueueUpdate, 10),
		done:         make(chan struct{}),
		fetching:     false,
		lastFetchedIdx: 1,
	}
}

// AddTrack adds a track to the queue
func (q *Queue) AddTrack(title, url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	track := Track{
		Title:   title,
		URL:     url,
		Index:   len(q.tracks),
		Fetched: true,
	}
	q.tracks = append(q.tracks, track)

	q.updates <- QueueUpdate{
		Type:   UpdateTypeAdded,
		Tracks: []Track{track},
	}
}

// GetTracks returns all tracks currently in the queue
func (q *Queue) GetTracks() []Track {
	q.mu.RLock()
	defer q.mu.RUnlock()

	tracksCopy := make([]Track, len(q.tracks))
	copy(tracksCopy, q.tracks)
	return tracksCopy
}

// SetCurrentIndex updates the currently playing track index
func (q *Queue) SetCurrentIndex(index int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if index >= 0 && index < len(q.tracks) {
		q.currentIndex = index
	}

	// Check if we need to fetch more tracks
	if index > len(q.tracks)-5 && !q.fetching {
		go q.fetchNextBatch()
	}
}

// GetCurrentIndex returns the currently playing track index
func (q *Queue) GetCurrentIndex() int {
	q.mu.RLock()
	defer q.mu.RUnlock()

	return q.currentIndex
}

// Updates returns the channel for queue updates
func (q *Queue) Updates() <-chan QueueUpdate {
	return q.updates
}

// StartBackgroundFetch starts background fetching of tracks
func (q *Queue) StartBackgroundFetch() {
	go q.fetchNextBatch()
}

// fetchNextBatch fetches the next batch of tracks from the YouTube Mix
func (q *Queue) fetchNextBatch() {
	q.mu.Lock()
	if q.fetching {
		q.mu.Unlock()
		return
	}
	q.fetching = true
	startIdx := q.lastFetchedIdx
	q.mu.Unlock()

	// Fetch video IDs from YouTube Mix
	endIdx := startIdx + q.batchSize - 1
	videoIDs, err := q.ytdlpClient.FetchMixIDs(q.radioURL, startIdx, endIdx)
	if err != nil {
		q.updates <- QueueUpdate{
			Type:  UpdateTypeError,
			Error: fmt.Errorf("failed to fetch mix IDs: %w", err),
		}

		q.mu.Lock()
		q.fetching = false
		q.mu.Unlock()
		return
	}

	// Fetch raw URLs for each video ID
	var newTracks []Track
	for i, videoID := range videoIDs {
		if videoID == "" {
			continue
		}

		title, streamURL, err := q.ytdlpClient.FetchRawURL(videoID)
		if err != nil {
			// Log error but continue with next track
			fmt.Printf("Warning: failed to fetch URL for %s: %v\n", videoID, err)
			continue
		}

		q.mu.Lock()
		idx := len(q.tracks)
		q.mu.Unlock()

		track := Track{
			Title:   title,
			URL:     streamURL,
			Index:   idx,
			Fetched: true,
		}
		newTracks = append(newTracks, track)

		q.mu.Lock()
		q.tracks = append(q.tracks, track)
		q.mu.Unlock()

		// Small delay to avoid hammering yt-dlp
		if i < len(videoIDs)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if len(newTracks) > 0 {
		q.updates <- QueueUpdate{
			Type:   UpdateTypeAdded,
			Tracks: newTracks,
		}
	}

	q.mu.Lock()
	q.lastFetchedIdx = endIdx + 1
	q.fetching = false
	q.mu.Unlock()
}

// Close stops the queue manager and closes the updates channel
func (q *Queue) Close() {
	close(q.done)
	close(q.updates)
}
