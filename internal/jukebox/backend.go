package jukebox

import (
	"context"
	"time"
)

// Track represents a playable music track from any backend.
type Track struct {
	ID       string // backend-specific ID
	Title    string
	Artist   string
	Duration int    // seconds
	URL      string // direct stream/download URL
	Source   string // "jamendo", "radio", "youtube"
}

// DurationTime returns the track duration as time.Duration.
func (t Track) DurationTime() time.Duration {
	return time.Duration(t.Duration) * time.Second
}

// MusicBackend is the interface all music sources implement.
type MusicBackend interface {
	Name() string
	Search(ctx context.Context, query string, limit int) ([]Track, error)
	StreamURL(ctx context.Context, trackID string) (string, error)
	Enabled() bool
}
