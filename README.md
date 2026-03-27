<div align="center">

# tavrn.sh

A terminal tavern over SSH.

Chat, vote on music, and hang out with strangers — right from your terminal.

No signup. No account. Your SSH key is your identity.

Everything resets weekly. Nothing is permanent.

</div>

---

<div align="center">

### Quick connect

Chat, gallery, jukebox controls, voting — no install needed.

```
ssh tavrn.sh
```

### Full experience

Same as above, plus music through your speakers.

```
go install tavrn.sh/cmd/tavrn@latest
tavrn
```

Requires [mpv](https://mpv.io/) for audio playback.
The binary checks on launch and tells you how to install it.

</div>

---

### What's inside

**Rooms** — Hang out in the lounge, post notes on the gallery board, or leave ideas in suggestions.

**Jukebox** — Search for music, add songs to the queue, vote on what plays next. Powered by [Jamendo](https://www.jamendo.com/)'s CC-licensed catalog.

**Gallery** — A shared sticky note board. Post, drag, read what others left behind.

**Now Playing** — A live animated bar shows the current track for everyone.

### Keybinds

```
F1  help        F2  nickname     F3  rooms
F4  jukebox     F5  post note    ESC close
```

### Contributing

See [docs/CONTRIBUTING.md](docs/CONTRIBUTING.md) for development setup, branch workflow, and architecture.

### Built with

[Bubble Tea](https://github.com/charmbracelet/bubbletea) · [Wish](https://github.com/charmbracelet/wish) · [Lipgloss](https://github.com/charmbracelet/lipgloss) · [Jamendo](https://www.jamendo.com/) · Go · SQLite

### License

MIT
