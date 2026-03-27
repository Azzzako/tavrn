# tavrn.sh

A public SSH terminal tavern. Connect, chat, listen to music, and hang out — all from your terminal.

```
ssh tavrn.sh
```

No signup. No account. Your SSH key is your identity.

## What is this

tavrn.sh is a multi-user terminal application accessible over SSH. It runs a TUI built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) served through [Wish](https://github.com/charmbracelet/wish), so anyone with an SSH key can connect and interact in real time.

**Rooms** — Chat in the lounge, post sticky notes on the gallery board, or drop ideas in suggestions.

**Jukebox** — A shared music player powered by [Jamendo](https://www.jamendo.com/). Search for tracks, request songs, and vote on what plays next. Everyone in the tavern sees the same Now Playing bar.

**Audio** — Install the `tavrn` client binary to hear music through your speakers. Bare SSH gives you the full TUI without audio.

**Gallery** — A collaborative sticky note board. Post notes, drag them around, read what others left behind.

Everything resets weekly. Nothing is permanent.

## Connect

SSH in directly:

```
ssh tavrn.sh
```

Or install the client binary for audio:

```
go install tavrn.sh/cmd/tavrn-client@latest
tavrn
```

The client requires [mpv](https://mpv.io/) for audio playback.

## Keybinds

| Key | Action |
|-----|--------|
| F1 | Help |
| F2 | Change nickname |
| F3 | Switch rooms |
| F4 | Jukebox (search, request, vote) |
| F5 | Post a note (gallery) |
| ESC | Close modal |

## Stack

- Go
- [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Wish v2](https://github.com/charmbracelet/wish) — SSH server
- [Lipgloss v2](https://github.com/charmbracelet/lipgloss) — Terminal styling
- [Jamendo API](https://developer.jamendo.com/) — Music catalog
- SQLite — Data storage

## Self-hosting

```bash
git clone https://github.com/neur0map/tavrn.git
cd tavrn
JAMENDO_CLIENT_ID=your_key go run ./cmd/tavrn
```

Get a free Jamendo API key at [devportal.jamendo.com](https://devportal.jamendo.com/). The server runs on port 2222 by default.

## Development

### Local testing

```bash
# Terminal 1 — server
JAMENDO_CLIENT_ID=your_key go run ./cmd/tavrn

# Terminal 2 — client with audio
go run ./cmd/tavrn-client --dev

# Or bare SSH without audio
ssh localhost -p 2222
```

The `--dev` flag connects the client to `localhost:2222` instead of the production server.

### Branch workflow

```
feature/* ──PR──> dev ──merge──> main (deploy)
```

| Branch | Purpose |
|--------|---------|
| `main` | Production. Runs on the VPS. |
| `dev` | All development. PRs target here. |
| `feature/*` | Short-lived feature branches created from dev. |

1. Create a feature branch from `dev`
2. Open a PR targeting `dev`
3. Test locally with `--dev`
4. When dev is stable, merge to `main` during planned downtime

### Project structure

```
cmd/
  tavrn/           Server binary (SSH server, jukebox engine)
  tavrn-client/    Client binary (SSH client + mpv audio)
internal/
  chat/            Message parsing and storage types
  hub/             Connection management, broadcasting
  identity/        Nickname generation, flair, color assignment
  jukebox/         Music backends, engine, streamer, wire protocol
  ratelimit/       Chat rate limiting
  room/            Room definitions
  sanitize/        Input sanitization
  server/          Wish SSH server setup, channel handlers
  session/         Session state, message types
  store/           SQLite persistence
ui/
  app.go           Main Bubble Tea model
  modal.go         Modal system (help, nick, rooms, jukebox)
  jukebox_modal.go Three-tab jukebox UI (now playing, search, vote)
  topbar.go        Top bar with Now Playing wave animation
  sidebar.go       Rooms panel, online users, up next queue
  chatview.go      Chat message rendering
  gallery.go       Sticky note board
  overlay.go       Modal overlay compositor
  styles.go        Cantina color palette
  splash.go        Welcome screen
```

### Architecture

**Server** — Wish-based SSH server. Each connection gets a Bubble Tea TUI. A shared hub broadcasts messages between sessions. The jukebox engine manages track playback state (current track, requests, votes, phase transitions).

**Audio streaming** — The server registers a `tavrn-audio` SSH channel type. When the client binary connects, it opens two channels: `session` (TUI) and `tavrn-audio` (MP3 stream). The streamer downloads tracks from Jamendo, buffers them, and sends them to connected audio channels using a length-prefixed wire protocol.

**Client binary** — Custom SSH client that opens both channels. Pipes MP3 data to a headless mpv subprocess for playback. Late-joining clients receive audio from the current playback position.

**Wire protocol** — Per track: `[4-byte header len][JSON metadata][4-byte audio len][MP3 bytes]`.

### Admin commands

```bash
# Send banner to all connected users (server-side only)
./tavrn --message "Maintenance in 10 minutes"

# Purge all data
./tavrn purge
```

### Tests

```bash
go test ./...
```

## License

MIT
