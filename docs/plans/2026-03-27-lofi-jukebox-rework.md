# Jukebox Rework (completed)

## Summary

Stripped the jukebox down from Jamendo/YouTube multi-backend to a genre radio.
Removed search, voting, multi-backend machinery. Replaced with Tavern Radio —
4 genre channels (Lofi, Jazz, Electronic, Cantina) with a tab bar UI.

## Current state

See [2026-03-28-genre-radio.md](2026-03-28-genre-radio.md) for the genre radio design.

## What was removed

- `jamendo.go`, `jamendo_test.go`
- `youtube.go`, `youtube_test.go`
- `MusicBackend` interface
- Search tab, Vote tab, all search/vote messages
- `JAMENDO_CLIENT_ID` env var handling
- Vote-to-skip system (never implemented)

## Engine flow

1. Startup → random track from active genre → play
2. Streamer downloads, ffprobe gets duration, broadcasts
3. Duration expires → next random track (from pending genre if changed)
4. Download fails → try another random track immediately
5. No idle phase — always playing
