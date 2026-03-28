package jukebox

import (
	"context"
	"testing"
)

func TestLofiCatalogParsed(t *testing.T) {
	tracks := parseLofiCatalog()
	if len(tracks) < 100 {
		t.Errorf("expected at least 100 tracks, got %d", len(tracks))
	}
	// Spot check first track
	if tracks[0].id == "" || tracks[0].title == "" {
		t.Errorf("first track has empty fields: %+v", tracks[0])
	}
}

func TestLofiEnabled(t *testing.T) {
	l := NewLofi()
	if !l.Enabled() {
		t.Error("lofi should be enabled with embedded catalog")
	}
}

func TestLofiName(t *testing.T) {
	l := NewLofi()
	if l.Name() != "lofi" {
		t.Errorf("name = %q, want lofi", l.Name())
	}
}

func TestLofiSearchPopular(t *testing.T) {
	l := NewLofi()
	tracks, err := l.Search(context.Background(), "popular", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(tracks) != 5 {
		t.Errorf("expected 5 tracks, got %d", len(tracks))
	}
	for _, tr := range tracks {
		if tr.Source != "lofi" {
			t.Errorf("source = %q, want lofi", tr.Source)
		}
		if tr.URL == "" {
			t.Error("track has empty URL")
		}
	}
}

func TestLofiSearchByTitle(t *testing.T) {
	l := NewLofi()
	// Search for a common word that should match something
	tracks, err := l.Search(context.Background(), "sun", 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	// Should find at least one match in 1349 tracks
	if len(tracks) == 0 {
		t.Error("expected at least one match for 'sun'")
	}
	for _, tr := range tracks {
		if tr.ID == "" || tr.Title == "" {
			t.Errorf("track has empty fields: %+v", tr)
		}
	}
}

func TestLofiStreamURL(t *testing.T) {
	l := NewLofi()
	// Use the first track's ID
	tracks := l.Tracks()
	if len(tracks) == 0 {
		t.Fatal("no tracks")
	}
	url, err := l.StreamURL(context.Background(), tracks[0].id)
	if err != nil {
		t.Fatalf("StreamURL: %v", err)
	}
	if url == "" {
		t.Error("empty URL")
	}
}

func TestLofiStreamURLNotFound(t *testing.T) {
	l := NewLofi()
	_, err := l.StreamURL(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent track")
	}
}

func TestLofiRandomTracksNoDuplicates(t *testing.T) {
	l := NewLofi()
	tracks := l.randomTracks(20)
	seen := make(map[string]bool)
	for _, tr := range tracks {
		if seen[tr.ID] {
			t.Errorf("duplicate track ID: %s", tr.ID)
		}
		seen[tr.ID] = true
	}
}
