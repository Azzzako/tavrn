package jukebox

import (
	"strings"
	"testing"
)

func TestCatalogTrackCount(t *testing.T) {
	c := NewCatalog()
	if c.TrackCount() < 3500 {
		t.Errorf("total tracks = %d, want >= 3500", c.TrackCount())
	}
}

func TestCatalogAllTracks(t *testing.T) {
	c := NewCatalog()
	tracks := c.AllTracks()
	if len(tracks) != c.TrackCount() {
		t.Errorf("AllTracks len = %d, TrackCount = %d", len(tracks), c.TrackCount())
	}
	for i, tr := range tracks[:10] {
		if tr.URL == "" {
			t.Errorf("track %d has empty URL", i)
		}
		if tr.Title == "" {
			t.Errorf("track %d has empty title", i)
		}
	}
}

func TestCatalogAllTracksIsCopy(t *testing.T) {
	c := NewCatalog()
	a := c.AllTracks()
	b := c.AllTracks()
	if &a[0] == &b[0] {
		t.Error("AllTracks should return a copy, not a reference")
	}
}

func TestCatalogHasMultipleArtists(t *testing.T) {
	c := NewCatalog()
	artists := make(map[string]bool)
	for _, tr := range c.AllTracks() {
		artists[tr.Artist] = true
	}
	if len(artists) < 3 {
		t.Errorf("expected at least 3 distinct artists, got %d", len(artists))
	}
}

func TestCatalogURLFormats(t *testing.T) {
	c := NewCatalog()
	var hasChillhop, hasArchive bool
	for _, tr := range c.AllTracks() {
		if strings.HasPrefix(tr.URL, "https://stream.chillhop.com/") {
			hasChillhop = true
		}
		if strings.HasPrefix(tr.URL, "https://archive.org/download/") || strings.HasPrefix(tr.URL, "https://ia601004") {
			hasArchive = true
		}
		if hasChillhop && hasArchive {
			break
		}
	}
	if !hasChillhop {
		t.Error("expected some chillhop URLs")
	}
	if !hasArchive {
		t.Error("expected some archive.org URLs")
	}
}
