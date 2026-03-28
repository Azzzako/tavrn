package jukebox

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

const lofiBaseURL = "https://stream.chillhop.com/mp3/"

// lofiTrack is a track from the embedded chillhop catalog.
type lofiTrack struct {
	id    string
	title string
}

// Lofi implements MusicBackend using chillhop.com's public MP3 stream.
// No API key required. Tracks are from the embedded catalog.
type Lofi struct {
	tracks []lofiTrack
	client *http.Client
}

// NewLofi creates a Lofi backend with the embedded track catalog.
func NewLofi() *Lofi {
	return &Lofi{
		tracks: parseLofiCatalog(),
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (l *Lofi) Name() string        { return "lofi" }
func (l *Lofi) Enabled() bool       { return len(l.tracks) > 0 }
func (l *Lofi) Tracks() []lofiTrack { return l.tracks }

func (l *Lofi) Search(_ context.Context, query string, limit int) ([]Track, error) {
	query = strings.ToLower(query)

	// For autoplay ("popular"), return random tracks
	if query == "popular" {
		return l.randomTracks(limit), nil
	}

	// Search by title
	var matches []Track
	for _, t := range l.tracks {
		if strings.Contains(strings.ToLower(t.title), query) {
			matches = append(matches, Track{
				ID:       t.id,
				Title:    t.title,
				Artist:   "Chillhop",
				Duration: 180, // estimate — actual length determined by MP3
				URL:      lofiBaseURL + t.id,
				Source:   "lofi",
			})
			if len(matches) >= limit {
				break
			}
		}
	}
	return matches, nil
}

func (l *Lofi) StreamURL(_ context.Context, trackID string) (string, error) {
	for _, t := range l.tracks {
		if t.id == trackID {
			return lofiBaseURL + t.id, nil
		}
	}
	return "", fmt.Errorf("lofi: track %s not found", trackID)
}

func (l *Lofi) randomTracks(n int) []Track {
	if n > len(l.tracks) {
		n = len(l.tracks)
	}
	// Fisher-Yates partial shuffle
	indices := make([]int, len(l.tracks))
	for i := range indices {
		indices[i] = i
	}
	for i := 0; i < n; i++ {
		j := i + rand.IntN(len(indices)-i)
		indices[i], indices[j] = indices[j], indices[i]
	}

	tracks := make([]Track, n)
	for i := 0; i < n; i++ {
		t := l.tracks[indices[i]]
		tracks[i] = Track{
			ID:       t.id,
			Title:    t.title,
			Artist:   "Chillhop",
			Duration: 180,
			URL:      lofiBaseURL + t.id,
			Source:   "lofi",
		}
	}
	return tracks
}
