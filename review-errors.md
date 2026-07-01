# musicr — Code Review: Bugs, Duplications & Memory Leaks

## Critical Issues

### 1. Goroutine leak — `playTrack` sends to `prefetchCh` without `ctx.Done()` guard

**File:** `internal/app/app.go:324-341`

```go
go func(idx int) {
    track, err := a.queue.ResolveTrack(a.ctx, idx)
    if err != nil {
        a.prefetchCh <- prefetchResult{index: idx, err: err}  // blocks forever if event loop exited
        return
    }
    url := track.StreamURL()
    if url != "" {
        a.player.LoadFile(url, "replace")
    }
    a.prefetchCh <- prefetchResult{index: idx, track: track}  // blocks forever if event loop exited
}(index)
```

If the event loop exits (user presses `q`, SIGINT, mpv crash) **after** `ResolveTrack` returns but **before** the send to `prefetchCh` completes, the goroutine blocks forever. The event loop selects on `ctx.Done()` first (line 141) and returns, so nobody ever drains `prefetchCh` again.

**Fix:** Wrap each send in a `select` with `ctx.Done()`:

```go
select {
case a.prefetchCh <- result:
case <-a.ctx.Done():
}
```

---

### 2. Goroutine leak — `checkRadio` sends to `prefetchCh`/`radioCh` without `ctx.Done()` guard

**File:** `internal/app/app.go:418-429`

```go
go func(videoID string) {
    titles, err := ytdlp.FetchMixTracks(a.ctx, videoID, 20)
    if err != nil {
        a.prefetchCh <- prefetchResult{err: fmt.Errorf("radio: %w", err)}  // can block forever
        return
    }
    if len(titles) > 0 {
        a.radioCh <- titles  // can block forever
    }
}(current.ID)
```

Same pattern as #1. After `FetchMixTracks` completes but before the channel send, the context may be cancelled. The goroutine leaks.

**Fix:** Same `select { ... case <-ctx.Done() }` guard around all sends.

---

### 3. Data race — `Quit()` closes `conn` without mutex

**File:** `internal/mpv/mpv.go:166` vs `:196`

```go
func (p *Player) Quit() {
    p.sendCommand("quit")     // line 165: acquires mu, writes to conn
    p.conn.Close()            // line 166: closes conn WITHOUT mu
    // ...
}
```

```go
func (p *Player) sendCommand(args ...interface{}) error {
    p.mu.Lock()
    defer p.mu.Unlock()
    _, err = p.conn.Write(data)  // line 196: writes to conn under mu
    return err
}
```

`Quit()` calls `p.conn.Close()` **without** holding `p.mu`, racing with the `conn.Write` inside `sendCommand`. This is undefined behavior on `net.Conn`.

**Fix:** Either hold `p.mu` during close, or restructure `Quit()` to just close the connection (unblocking `readEvents`) without calling `sendCommand`:

```go
func (p *Player) Quit() {
    p.conn.Close()
    os.Remove(p.socketPath)
    select {
    case <-p.done:
    case <-time.After(2 * time.Second):
        p.cmd.Process.Kill()
    }
}
```

---

### 4. Goroutine leak — `readKeys` blocks on `stdin.Read()` unresponsive to context cancellation

**File:** `internal/app/app.go:538-551`

```go
func (a *App) readKeys() {
    defer close(a.keyCh)
    for {
        key, err := a.ui.ReadKey()   // blocking read on stdin
        if err != nil {
            return
        }
        select {
        case a.keyCh <- key:
        case <-a.ctx.Done():
            return
        }
    }
}
```

`a.ui.ReadKey()` calls `os.Stdin.Read()`, a blocking system call. If SIGINT fires while waiting for input, `cancel()` is called but `ReadKey` stays blocked in `read(0)` until the user presses a key. The `defer close(a.keyCh)` never runs.

**Fix:** Run the read in a separate goroutine and select between its result and `ctx.Done()`:

```go
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
            readCh <- key
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
```

---

## Logic Bugs

### 5. Queue removal desyncs mpv from app state

**File:** `internal/queue/queue.go:51-68`, `internal/app/app.go:290-299`

When the user deletes the currently-playing track (via `d` key):

```go
case inputDelete:
    if idx, err := strconv.Atoi(line); err == nil {
        idx-- // 1-based to 0-based
        if a.queue.Remove(idx) {
            a.statusMsg = fmt.Sprintf("Removed track #%d", idx+1)
        }
    }
```

`queue.Remove` adjusts the internal `current` index but **never tells mpv to stop**. The old audio continues playing while the UI shows the next queued track as "playing". App state and mpv are out of sync.

**Fix:** After removing the current track, call `a.player.Stop()` and optionally `a.playTrack(q.current)` if tracks remain.

---

### 6. Empty URL in `playTrack` — silent no-op

**File:** `internal/app/app.go:331-334`

```go
url := track.StreamURL()
if url != "" {
    a.player.LoadFile(url, "replace")
}
```

If `track.StreamURL()` returns `""` (unresolved or malformed track), `LoadFile` is never called. The queue advances, the UI shows the track as "playing", but no audio plays. No error is reported to the user.

**Fix:** Add an else branch that logs an error and advances past the unplayable track:

```go
url := track.StreamURL()
if url != "" {
    if err := a.player.LoadFile(url, "replace"); err != nil {
        a.statusMsg = fmt.Sprintf("Playback error: %v", err)
    }
} else {
    a.statusMsg = "Track has no stream URL, skipping"
    a.nextTrack()
}
```

---

### 7. Prefetch goroutine may resolve wrong track after queue mutation

**File:** `internal/app/app.go:338-340`

```go
if nextTrack, nextIdx, ok := a.queue.PeekNext(); ok && !nextTrack.Resolved() {
    a.queue.ResolveTrack(a.ctx, nextIdx)
}
```

This runs in a goroutine spawned by `playTrack`. Between `PeekNext()` and `ResolveTrack()`, the event loop may have:
- Advanced `q.current` via `nextTrack()`/`prevTrack()`/`onTrackEnd()`
- Replaced the entire queue via `loadPlaylist()`

The prefetch resolves a track at an index that may no longer be the "next" track or that belongs to a completely different playlist.

**Fix:** Re-check the queue state after resolution completes instead of assuming the index is still valid.

---

### 8. Dead parameter in `enterLineInput`

**File:** `internal/app/app.go:387-395`

```go
func (a *App) enterLineInput(action inputAction, prompt string) {
    a.inputMode = ui.ModeLineInput
    a.inputType = action
    a.inputBuf = ""
    a.statusMsg = ""
    _ = prompt    // <-- discarded! The comment "render will pick it up" is wrong
    a.render()
}
```

The `prompt` parameter is stored to the blank identifier. The actual prompt is re-derived from `a.inputType` inside `render()` (lines 500-509). This is dead code. If a new input type is added, both the caller and the switch in `render()` must be updated.

**Fix:** Remove the `prompt` parameter entirely.

---

## Missing Error Handling

### 9. `LoadFile` error silently ignored

**File:** `internal/app/app.go:333`

```go
a.player.LoadFile(url, "replace")  // return value discarded
```

If mpv IPC fails or the URL is invalid, the user gets no feedback.

---

### 10. Prefetch `ResolveTrack` error silently ignored

**File:** `internal/app/app.go:339`

```go
a.queue.ResolveTrack(a.ctx, nextIdx)  // return value (both Track and error) discarded
```

If yt-dlp fails (network down, unavailable video), the error is invisible to the user. The next track will be resolved again when the user attempts to play it, so it isn't fatal, but the user sees no diagnostic.

---

### 11. `cmd.Wait()` error discarded

**File:** `internal/mpv/mpv.go:124`

```go
go func() {
    cmd.Wait()      // return value discarded
    close(p.done)
}()
```

If mpv crashes with a non-zero exit code, the error is lost.

---

### 12. `os.Remove(socketPath)` error discarded on startup

**File:** `internal/mpv/mpv.go:72`

```go
os.Remove(socketPath)  // error discarded
```

If a stale Unix socket cannot be removed, mpv's subsequent `net.Dial` will fail. The user gets a "timed out waiting for mpv IPC socket" error, but the real cause (permission denied removing old socket) is hidden.

---

### 13. `loadPlaylistList` silently drops errors

**File:** `internal/app/app.go:439-446`

```go
func (a *App) loadPlaylistList() {
    entries, err := listPlaylists()
    if err != nil {
        a.playlists = nil
        return   // error silently discarded
    }
    a.playlists = entries
}
```

Permission errors or filesystem issues silently become an empty playlist list with no user feedback.

---

### 14. `sendCommand("quit")` error ignored

**File:** `internal/mpv/mpv.go:165`

```go
p.sendCommand("quit")  // error discarded
```

If the IPC socket is already dead, this fails silently. Not a functional bug (we're quitting anyway), but the error should at least be logged.

---

## Code Duplication

### 15. `listPlaylists` in two places

**File:** `main.go:115-144` and `internal/app/playlist.go:70-86`

Both functions enumerate `.txt` files from `~/.config/musicr/playlists/`:

- Same directory path construction (though `main.go` uses string concatenation while `playlist.go` uses `filepath.Join`)
- Same `.txt` suffix filter
- Same `.txt` suffix strip

The `main.go` version prints results to stdout; the `playlist.go` version returns a `[]string`. If the playlist directory or file naming convention changes, both must be updated.

**Fix:** Have `main.go`'s `listPlaylists` call the `app` package's `listPlaylists()` and iterate over the returned slice to print results.

---

### 16. Overlapping `yt-dlp` flags in `Resolve` and `FetchMixTracks`

**File:** `internal/ytdlp/ytdlp.go:36-66` and `:70-96`

Both functions build yt-dlp command lines with duplicated flags:
- `--no-warnings`
- `--force-ipv4`

Minor duplication — could be extracted into shared constants or a helper function.

---

## Resource Leaks

### 17. `signal.Notify` never stopped

**File:** `internal/app/app.go:107-112`

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
```

`signal.Notify` is never followed by `defer signal.Stop(sigCh)`. After `Run()` returns, the signal notification remains registered. If a signal arrives before `main()` exits, the orphan goroutine calls `a.cancel()` on an already-finished `App`. Harmless but technically a resource leak.

**Fix:** Add `defer signal.Stop(sigCh)` after `signal.Notify`.

---

## Rendering Issues

### 18. Progress bar can overflow narrow terminals

**File:** `internal/ui/ui.go:220-236`

```go
barWidth := width - len(timeStr) - 2
```

`timeStr` is at least `" 00:00 / 00:00"` (15 characters). On terminals narrower than ~20 columns, `barWidth` becomes negative. The `if barWidth < 10 { barWidth = 10 }` clamp prevents the crash but the rendered output exceeds the terminal width.

---

## Priority Summary

| Priority | Issues |
|----------|--------|
| **Fix immediately** | 1, 2, 3, 4 (goroutine leaks + data race) |
| **Fix before release** | 5, 6, 7, 9, 10 (logic bugs + missing errors) |
| **Clean up** | 8, 15, 16, 17, 18 (dead code, duplication, leak) |
| **Nice to have** | 11, 12, 13, 14 (discarded errors, low impact) |
