package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/tanishque-suthar/musicr/internal/mpv"
	"github.com/tanishque-suthar/musicr/internal/queue"
	"github.com/tanishque-suthar/musicr/internal/ui"
	"github.com/tanishque-suthar/musicr/internal/ytdlp"
)

// Config holds the app configuration from CLI args.
type Config struct {
	Queries  []string // initial search queries
	Playlist string   // playlist name to load (-p flag)
	NoRadio  bool     // disable radio mode
}

// App is the central state owner. It coordinates all goroutines and owns
// all mutable state. All state mutations happen in the Run loop.
type App struct {
	config Config
	player *mpv.Player
	queue  *queue.Queue
	ui     *ui.UI

	// State
	radioOn        bool
	radioFetching  bool
	paused         bool
	volume     float64
	timePos    float64
	duration   float64
	inputMode  ui.InputMode
	inputBuf   string
	inputType  inputAction
	statusMsg  string
	playlists  []string

	// Channels
	keyCh      chan byte
	prefetchCh chan prefetchResult
	radioCh    chan []string
	ctx        context.Context
	cancel     context.CancelFunc
}

// inputAction represents what line-input mode is being used for.
type inputAction int

const (
	inputNone inputAction = iota
	inputAdd              // typing a search query
	inputSave             // typing a playlist name to save
	inputLoad             // typing a playlist name to load
	inputDelete           // typing an index to delete
)

// prefetchResult is the result of a background track resolution.
type prefetchResult struct {
	index int
	track ytdlp.Track
	err   error
}

// New creates a new App with the given config. Does not start anything yet.
func New(cfg Config) *App {
	return &App{
		config:     cfg,
		queue:      queue.New(),
		radioOn:    !cfg.NoRadio,
		volume:     100,
		keyCh:      make(chan byte, 16),
		prefetchCh: make(chan prefetchResult, 4),
		radioCh:    make(chan []string, 1),
	}
}

// Run starts the app: spawns mpv, enters raw terminal mode, and runs the
// central event loop. Blocks until quit.
func (a *App) Run() error {
	a.ctx, a.cancel = context.WithCancel(context.Background())
	defer a.cancel()

	// Start mpv
	var err error
	a.player, err = mpv.Start()
	if err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	// Start UI (raw terminal mode)
	a.ui, err = ui.New()
	if err != nil {
		a.player.Quit()
		return fmt.Errorf("failed to start UI: %w", err)
	}

	// Guarantee terminal restoration on any exit path
	defer a.cleanup()

	// Handle signals for clean exit
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		a.cancel()
	}()

	// Start key reader goroutine
	go a.readKeys()

	// Add initial queries to queue
	if a.config.Playlist != "" {
		a.loadPlaylist(a.config.Playlist)
	}
	for _, q := range a.config.Queries {
		a.queue.Add(q)
	}

	// Start playing the first track if we have any
	if a.queue.Len() > 0 {
		a.playTrack(0)
	}

	// Initial render
	a.render()

	// Central event loop — single goroutine owns all state
	return a.eventLoop()
}

// eventLoop is the central select loop that processes all events.
func (a *App) eventLoop() error {
	for {
		select {
		case <-a.ctx.Done():
			return nil

		case key, ok := <-a.keyCh:
			if !ok {
				return nil
			}
			if a.handleKey(key) {
				return nil // quit requested
			}

		case evt, ok := <-a.player.Events:
			if !ok {
				return nil // mpv exited
			}
			a.handleMpvEvent(evt)

		case result := <-a.prefetchCh:
			if result.err != nil && strings.HasPrefix(result.err.Error(), "radio:") {
				a.radioFetching = false
			}
			a.handlePrefetchResult(result)

		case titles := <-a.radioCh:
			a.radioFetching = false
			a.queue.AddTracks(titles)
			a.statusMsg = fmt.Sprintf("Radio: added %d tracks", len(titles))
			a.render()

		case <-a.player.Done():
			return nil // mpv process exited
		}
	}
}

// handleKey processes a keystroke and returns true if quit was requested.
func (a *App) handleKey(key byte) bool {
	if a.inputMode == ui.ModeLineInput {
		return a.handleLineInput(key)
	}

	switch key {
	case 'q':
		return true

	case 'a':
		a.enterLineInput(inputAdd)

	case 'n':
		a.nextTrack()

	case 'p':
		a.prevTrack()

	case ' ':
		a.togglePause()

	case 's':
		a.enterLineInput(inputSave)

	case 'l':
		a.listAndLoadPlaylist()

	case 'r':
		a.radioOn = !a.radioOn
		if a.radioOn {
			a.statusMsg = "Radio mode: on"
			a.checkRadio()
		} else {
			a.statusMsg = "Radio mode: off"
		}

	case 'd':
		a.enterLineInput(inputDelete)

	case '+', '=':
		a.volume = min(a.volume+1, 100)
		a.player.SetVolume(a.volume)
		a.statusMsg = fmt.Sprintf("Volume: %.0f%%", a.volume)

	case '-', '_':
		a.volume = max(a.volume-1, 0)
		a.player.SetVolume(a.volume)
		a.statusMsg = fmt.Sprintf("Volume: %.0f%%", a.volume)

	case ']', '.':
		a.player.Seek(5)
		a.statusMsg = "Seek +5s"

	case '[', ',':
		a.player.Seek(-5)
		a.statusMsg = "Seek -5s"
	}

	a.render()
	return false
}

// handleLineInput processes keys in line-input mode.
func (a *App) handleLineInput(key byte) bool {
	switch key {
	case 13: // Enter
		line := a.inputBuf
		action := a.inputType
		a.exitLineInput()
		a.processLineInput(action, line)

	case 27: // Escape
		a.exitLineInput()

	case 127, 8: // Backspace
		if len(a.inputBuf) > 0 {
			a.inputBuf = a.inputBuf[:len(a.inputBuf)-1]
		}

	default:
		if key >= 32 && key < 127 {
			a.inputBuf += string(key)
		}
	}

	a.render()
	return false
}

// processLineInput handles the completed line input for the given action.
func (a *App) processLineInput(action inputAction, line string) {
	switch action {
	case inputAdd:
		if line != "" {
			a.queue.Add(line)
			a.statusMsg = fmt.Sprintf("Added: %s", line)
			// If nothing was playing, start playing
			if _, idx := a.queue.Current(); idx == -1 {
				a.playTrack(0)
			}
		}

	case inputSave:
		if line != "" {
			a.savePlaylist(line)
		}

	case inputLoad:
		if line != "" {
			a.loadPlaylist(line)
		}

	case inputDelete:
		if idx, err := strconv.Atoi(line); err == nil {
			idx-- // convert from 1-based display to 0-based
			if ok, wasCurrent := a.queue.Remove(idx); ok {
				a.statusMsg = fmt.Sprintf("Removed track #%d", idx+1)
				if wasCurrent {
					a.player.Stop()
					if _, currentIdx := a.queue.Current(); currentIdx != -1 {
						a.playTrack(currentIdx)
					} else {
						a.statusMsg = "Removed playing track (queue empty)"
					}
				}
			} else {
				a.statusMsg = "Invalid track number"
			}
		}
	}
}

// handleMpvEvent processes an event from the mpv IPC connection.
func (a *App) handleMpvEvent(evt mpv.Event) {
	switch evt.Type {
	case mpv.EventTimePos:
		a.timePos = evt.FloatVal
	case mpv.EventDuration:
		a.duration = evt.FloatVal
	case mpv.EventPause:
		a.paused = evt.BoolVal
	case mpv.EventVolume:
		a.volume = evt.FloatVal
	case mpv.EventEndFile:
		// Only auto-advance if the file ended naturally (eof) or due to an error.
		// Ignore "stop" (manual stop/skip) or "quit".
		if evt.TextVal == "eof" || evt.TextVal == "error" {
			a.onTrackEnd()
		}
	case mpv.EventStartFile:
		a.timePos = 0
		a.duration = 0
	}
	a.render()
}

// handlePrefetchResult processes a completed track resolution.
func (a *App) handlePrefetchResult(result prefetchResult) {
	if result.err != nil {
		a.statusMsg = fmt.Sprintf("Resolve error: %v", result.err)
		a.render()
		return
	}
	// If the resolved track is the current playing track, trigger radio check.
	_, currentIdx := a.queue.Current()
	if result.index == currentIdx {
		a.checkRadio()
	}
}

// playTrack resolves and starts playing the track at the given index.
func (a *App) playTrack(index int) {
	a.queue.SetCurrent(index)

	go func(idx int) {
		track, err := a.queue.ResolveTrack(a.ctx, idx)
		if err != nil {
			select {
			case a.prefetchCh <- prefetchResult{index: idx, err: err}:
			case <-a.ctx.Done():
			}
			return
		}

		url := track.StreamURL()
		if url == "" {
			select {
			case a.prefetchCh <- prefetchResult{index: idx, err: fmt.Errorf("track has no stream URL")}:
			case <-a.ctx.Done():
			}
			return
		}
		if err := a.player.LoadFile(url, "replace"); err != nil {
			select {
			case a.prefetchCh <- prefetchResult{index: idx, err: fmt.Errorf("loadfile: %w", err)}:
			case <-a.ctx.Done():
			}
			return
		}
		select {
		case a.prefetchCh <- prefetchResult{index: idx, track: track}:
		case <-a.ctx.Done():
		}

		if nextTrack, nextIdx, ok := a.queue.PeekNext(); ok && !nextTrack.Resolved() {
			_, err := a.queue.ResolveTrack(a.ctx, nextIdx)
			if err != nil && err != context.Canceled {
				select {
				case a.prefetchCh <- prefetchResult{index: nextIdx, err: fmt.Errorf("prefetch: %w", err)}:
				case <-a.ctx.Done():
				}
			}
		}
	}(index)

	a.render()
}

// onTrackEnd handles the end of the current track.
func (a *App) onTrackEnd() {
	// Check if we need more radio tracks
	a.checkRadio()

	// Advance to next track
	if _, nextIdx, ok := a.queue.Next(); ok {
		a.playTrack(nextIdx)
	} else {
		a.statusMsg = "Queue finished"
		a.timePos = 0
		a.duration = 0
	}
}

// nextTrack skips to the next track.
func (a *App) nextTrack() {
	if _, nextIdx, ok := a.queue.Next(); ok {
		a.player.Stop()
		a.playTrack(nextIdx)
	} else {
		a.statusMsg = "No next track"
	}
}

// prevTrack goes to the previous track.
func (a *App) prevTrack() {
	if _, prevIdx, ok := a.queue.Prev(); ok {
		a.player.Stop()
		a.playTrack(prevIdx)
	} else {
		a.statusMsg = "No previous track"
	}
}

// togglePause toggles the pause state.
func (a *App) togglePause() {
	a.player.TogglePause()
}

// enterLineInput switches to line-input mode.
// The prompt is derived from inputType by render().
func (a *App) enterLineInput(action inputAction) {
	a.inputMode = ui.ModeLineInput
	a.inputType = action
	a.inputBuf = ""
	a.statusMsg = ""
	a.render()
}

// exitLineInput returns to single-keystroke mode.
func (a *App) exitLineInput() {
	a.inputMode = ui.ModeKeystroke
	a.inputType = inputNone
	a.inputBuf = ""
}

// checkRadio triggers radio fetch if needed.
func (a *App) checkRadio() {
	if !a.radioOn || a.radioFetching {
		return
	}
	if a.queue.Remaining() > 5 {
		return
	}
	current, _ := a.queue.Current()
	if current.ID == "" {
		return
	}

	a.radioFetching = true
	currentTitle := current.Title
	go func(videoID string) {
		titles, err := ytdlp.FetchMixTracks(a.ctx, videoID, 20)
		if err != nil {
			select {
			case a.prefetchCh <- prefetchResult{err: fmt.Errorf("radio: %w", err)}:
			case <-a.ctx.Done():
			}
			return
		}
		filtered := make([]string, 0, len(titles))
		for _, t := range titles {
			if t != currentTitle {
				filtered = append(filtered, t)
			}
		}
		if len(filtered) > 0 {
			select {
			case a.radioCh <- filtered:
			case <-a.ctx.Done():
			}
		}
	}(current.ID)
}

// listAndLoadPlaylist shows available playlists and enters load mode.
func (a *App) listAndLoadPlaylist() {
	a.loadPlaylistList()
	a.enterLineInput(inputLoad)
}

// loadPlaylistList loads the list of available playlists.
func (a *App) loadPlaylistList() {
	entries, err := ListPlaylists()
	if err != nil {
		a.playlists = nil
		a.statusMsg = fmt.Sprintf("Playlist error: %v", err)
		return
	}
	a.playlists = entries
}

// savePlaylist saves the current queue to a playlist file.
func (a *App) savePlaylist(name string) {
	titles := a.queue.Titles()
	if len(titles) == 0 {
		a.statusMsg = "Queue is empty, nothing to save"
		return
	}

	err := savePlaylistFile(name, titles)
	if err != nil {
		a.statusMsg = fmt.Sprintf("Save error: %v", err)
		return
	}
	a.statusMsg = fmt.Sprintf("Saved playlist: %s (%d tracks)", name, len(titles))
}

// loadPlaylist loads a playlist from file and replaces the queue.
func (a *App) loadPlaylist(name string) {
	queries, err := loadPlaylistFile(name)
	if err != nil {
		a.statusMsg = fmt.Sprintf("Load error: %v", err)
		return
	}
	if len(queries) == 0 {
		a.statusMsg = "Playlist is empty"
		return
	}

	// Clear current queue and add loaded tracks
	a.player.Stop()
	newQueue := queue.New()
	for _, q := range queries {
		newQueue.Add(q)
	}
	a.queue = newQueue
	a.statusMsg = fmt.Sprintf("Loaded: %s (%d tracks)", name, len(queries))
	a.playTrack(0)
}

// render updates the UI with the current state.
func (a *App) render() {
	if a.ui == nil {
		return
	}

	current, idx := a.queue.Current()
	title := current.Title
	if title == "" {
		title = current.Query
	}

	prompt := ""
	switch a.inputType {
	case inputAdd:
		prompt = "Search: "
	case inputSave:
		prompt = "Save as: "
	case inputLoad:
		prompt = "Load: "
	case inputDelete:
		prompt = "Delete track #: "
	}

	a.ui.Render(ui.State{
		TrackTitle:  title,
		TimePos:     a.timePos,
		Duration:    a.duration,
		Paused:      a.paused,
		Volume:      a.volume,
		RadioOn:     a.radioOn,
		Queue:       a.queue.Titles(),
		QueueIdx:    idx,
		InputMode:   a.inputMode,
		InputPrompt: prompt,
		InputBuf:    a.inputBuf,
		StatusMsg:   a.statusMsg,
		Playlists:   a.playlists,
	})
}

// cleanup restores the terminal and quits mpv.
func (a *App) cleanup() {
	if a.ui != nil {
		a.ui.Restore()
	}
	if a.player != nil {
		a.player.Quit()
	}
}

// readKeys reads single bytes from stdin and sends them on keyCh.
// Uses a helper goroutine so ctx cancellation unblocks immediately.
func (a *App) readKeys() {
	defer close(a.keyCh)
	readCh := make(chan byte, 1)
	go func() {
		for {
			key, err := a.ui.ReadKey()
			if err != nil {
				close(readCh)
				return
			}
			select {
			case readCh <- key:
			case <-a.ctx.Done():
				return
			}
		}
	}()
	for {
		select {
		case key, ok := <-readCh:
			if !ok {
				return
			}
			select {
			case a.keyCh <- key:
			case <-a.ctx.Done():
				return
			}
		case <-a.ctx.Done():
			return
		}
	}
}
