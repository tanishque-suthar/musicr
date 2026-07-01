# Implementation Plan: Volume and Seek Controls

This plan adds volume control and 5-second relative seeking (forward/backward) to `musicr`.

## User Review Required

> [!IMPORTANT]
> **Key Bindings**: Since `musicr` runs in a raw terminal, handling multi-byte Arrow keys (Up/Down/Left/Right) requires modifying our keystroke parser. 
> To keep the CLI robust and snappy, I propose using the following single-character keys (similar to standard mpv/YouTube shortcuts):
> - **Volume Up**: `+` (or `=`)
> - **Volume Down**: `-`
> - **Seek Forward (5s)**: `]` (or `.`)
> - **Seek Backward (5s)**: `[` (or `,`)
> 
> *Would you prefer these single-character keys, or should I implement ANSI escape sequence parsing so we can use the Arrow keys?*

## Proposed Changes

### 1. `internal/mpv/mpv.go`
**New methods for audio control:**
- `VolumeUp()`: Sends IPC command `add volume 5`
- `VolumeDown()`: Sends IPC command `add volume -5`
- `Seek(seconds float64)`: Sends IPC command `seek <seconds> relative`

**Volume state tracking:**
- Add `p.observeProperty(4, "volume")` in `Start()`.
- Add `EventVolume` to `EventType`.
- Parse `volume` property changes in `readEvents` and send `EventVolume` through the channel.

### 2. `internal/app/app.go`
**State and Events:**
- Add `volume float64` to the `App` struct.
- In `handleMpvEvent`, add a case for `mpv.EventVolume` to update `a.volume` and call `a.render()`.

**Key Handling (`handleKey`):**
- Add cases for the chosen volume keys (e.g. `+` / `-`) to call `a.player.VolumeUp()` and `a.player.VolumeDown()`.
- Add cases for the chosen seek keys (e.g. `]` / `[`) to call `a.player.Seek(5)` and `a.player.Seek(-5)`.

### 3. `internal/ui/ui.go`
**State:**
- Add `Volume float64` to `ui.State`.

**Render:**
- Modify the header to display the current volume.
  Example: `musicr  —  radio mode: on  |  vol: 85%`

### 4. `main.go`
**Help Text:**
- Update `printUsage()` to document the new volume and seek keybindings.

## Verification Plan
1. `go build ./...` and `go vet ./...`
2. Run `./musicr <query>`, press the volume up/down keys and verify the UI updates the volume percentage and the audio level changes.
3. Press the seek keys and verify the time position in the UI progress bar jumps forward/backward by 5 seconds and the audio skips accordingly.
