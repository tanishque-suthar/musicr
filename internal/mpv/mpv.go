package mpv

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// EventType represents the type of mpv event.
type EventType int

const (
	EventStartFile      EventType = iota // a new file started playing
	EventEndFile                         // current file finished
	EventPause                           // pause state changed
	EventTimePos                         // playback position updated
	EventDuration                        // duration became known
	EventVolume                          // volume changed
	EventIdle                            // mpv entered idle state
	EventPlaybackRestart                 // playback restarted (seek finished, etc.)
)

// Event is a typed mpv event sent to the consumer.
type Event struct {
	Type     EventType
	FloatVal float64 // for TimePos, Duration
	BoolVal  bool    // for Pause
	TextVal  string  // for Reason
}

// Player wraps an mpv subprocess and communicates via JSON IPC.
type Player struct {
	cmd        *exec.Cmd
	socketPath string
	conn       net.Conn
	Events     chan Event

	reqID atomic.Int64
	mu    sync.Mutex // protects writes to conn

	done    chan struct{} // closed when mpv process exits
	exitErr error
}

// ipcRequest is a JSON IPC command sent to mpv.
type ipcRequest struct {
	Command   []interface{} `json:"command"`
	RequestID int64         `json:"request_id,omitempty"`
}

// ipcResponse is a JSON IPC response or event from mpv.
type ipcResponse struct {
	Error     string      `json:"error,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID int64       `json:"request_id,omitempty"`
	Event     string      `json:"event,omitempty"`
	Name      string      `json:"name,omitempty"` // for property-change events
	ID        int64       `json:"id,omitempty"`   // for property-change events
	Reason    string      `json:"reason,omitempty"`
}

// Start spawns a new mpv process in headless/idle mode and connects via IPC.
func Start() (*Player, error) {
	socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("musicr-mpv-%d.sock", os.Getpid()))

	// Remove stale socket if it exists
	if err := os.Remove(socketPath); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to remove stale socket: %w", err)
	}

	cmd := exec.Command("mpv",
		"--no-video",
		"--idle=yes",
		"--input-ipc-server="+socketPath,
		"--cache=yes",
		"--demuxer-max-bytes=67MiB",
		"--no-terminal",
		"--really-quiet",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start mpv: %w", err)
	}

	// Wait for the socket to appear
	var conn net.Conn
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var err error
		conn, err = net.Dial("unix", socketPath)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if conn == nil {
		cmd.Process.Kill()
		return nil, fmt.Errorf("timed out waiting for mpv IPC socket at %s", socketPath)
	}

	p := &Player{
		cmd:        cmd,
		socketPath: socketPath,
		conn:       conn,
		Events:     make(chan Event, 64),
		done:       make(chan struct{}),
	}

	// Observe properties we care about
	p.observeProperty(1, "pause")
	p.observeProperty(2, "time-pos")
	p.observeProperty(3, "duration")
	p.observeProperty(4, "volume")

	// Start event reader goroutine
	go p.readEvents()

	// Monitor mpv process exit
	go func() {
		p.exitErr = cmd.Wait()
		close(p.done)
	}()

	return p, nil
}

// LoadFile tells mpv to play the given URL/path.
// mode: "replace" (play now), "append" (add to mpv internal playlist),
// "append-play" (append and start if idle).
func (p *Player) LoadFile(url string, mode string) error {
	return p.sendCommand("loadfile", url, mode)
}

// Next skips to the next track in mpv's internal playlist.
func (p *Player) Next() error {
	return p.sendCommand("playlist-next", "weak")
}

// Prev goes to the previous track in mpv's internal playlist.
func (p *Player) Prev() error {
	return p.sendCommand("playlist-prev", "weak")
}

// SetPause sets the pause state.
func (p *Player) SetPause(paused bool) error {
	return p.setProperty("pause", paused)
}

// TogglePause toggles the pause state.
func (p *Player) TogglePause() error {
	return p.sendCommand("cycle", "pause")
}

// Stop stops playback and clears mpv's playlist.
func (p *Player) Stop() error {
	return p.sendCommand("stop")
}

// SetVolume sets the volume to an absolute value (0-100).
func (p *Player) SetVolume(vol float64) error {
	return p.setProperty("volume", vol)
}

// VolumeUp increases volume by the given amount (0-100).
func (p *Player) VolumeUp(delta float64) error {
	return p.sendCommand("add", "volume", delta)
}

// VolumeDown decreases volume by the given amount (0-100).
func (p *Player) VolumeDown(delta float64) error {
	return p.sendCommand("add", "volume", -delta)
}

// Seek seeks relative to current position by the given seconds.
// Positive = forward, negative = backward.
func (p *Player) Seek(seconds float64) error {
	return p.sendCommand("seek", seconds, "relative")
}

// Quit closes the connection and waits for mpv to exit.
func (p *Player) Quit() {
	p.sendCommand("quit")
	p.mu.Lock()
	if p.conn != nil {
		p.conn.Close()
	}
	p.mu.Unlock()
	os.Remove(p.socketPath)
	select {
	case <-p.done:
	case <-time.After(2 * time.Second):
		if p.cmd != nil && p.cmd.Process != nil {
			p.cmd.Process.Kill()
		}
	}
}

// Done returns a channel that is closed when the mpv process exits.
func (p *Player) Done() <-chan struct{} {
	return p.done
}

// ExitError returns the exit error of the mpv process, if any.
func (p *Player) ExitError() error {
	return p.exitErr
}

// sendCommand sends a JSON IPC command to mpv.
func (p *Player) sendCommand(args ...interface{}) error {
	id := p.reqID.Add(1)
	req := ipcRequest{
		Command:   args,
		RequestID: id,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}
	data = append(data, '\n')

	p.mu.Lock()
	defer p.mu.Unlock()
	_, err = p.conn.Write(data)
	return err
}

// setProperty sets an mpv property via IPC.
func (p *Player) setProperty(name string, value interface{}) error {
	return p.sendCommand("set_property", name, value)
}

// observeProperty tells mpv to send property-change events for the given property.
func (p *Player) observeProperty(id int64, name string) error {
	return p.sendCommand("observe_property", id, name)
}

// readEvents reads JSON IPC messages from mpv and converts them to typed events.
func (p *Player) readEvents() {
	defer close(p.Events)
	scanner := bufio.NewScanner(p.conn)
	// Increase buffer for large responses
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		var resp ipcResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			continue
		}

		switch resp.Event {
		case "start-file":
			p.Events <- Event{Type: EventStartFile}
		case "end-file":
			p.Events <- Event{Type: EventEndFile, TextVal: resp.Reason}
		case "idle":
			p.Events <- Event{Type: EventIdle}
		case "playback-restart":
			p.Events <- Event{Type: EventPlaybackRestart}
		case "property-change":
			switch resp.Name {
			case "pause":
				if v, ok := resp.Data.(bool); ok {
					p.Events <- Event{Type: EventPause, BoolVal: v}
				}
			case "time-pos":
				if v, ok := resp.Data.(float64); ok {
					p.Events <- Event{Type: EventTimePos, FloatVal: v}
				}
			case "duration":
				if v, ok := resp.Data.(float64); ok {
					p.Events <- Event{Type: EventDuration, FloatVal: v}
				}
			case "volume":
				if v, ok := resp.Data.(float64); ok {
					p.Events <- Event{Type: EventVolume, FloatVal: v}
				}
			}
		}
	}
}
