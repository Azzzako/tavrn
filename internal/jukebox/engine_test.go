package jukebox

import (
	"testing"
	"time"
)

func newTestEngine() *Engine {
	return NewEngineWithCatalog(NewCatalog())
}

func TestEngineInitialState(t *testing.T) {
	e := newTestEngine()
	state := e.State()
	if state.Current != nil {
		t.Error("expected no current track before first tick")
	}
}

func TestEngineAutoPicksOnFirstTick(t *testing.T) {
	e := newTestEngine()
	e.tick()
	state := e.State()
	if state.Current == nil {
		t.Error("expected a track after first tick")
	}
}

func TestEngineAutoNextOnTrackEnd(t *testing.T) {
	e := newTestEngine()
	e.tick()

	e.mu.Lock()
	e.current.Duration = 1
	e.playStart = time.Now().Add(-2 * time.Second)
	e.mu.Unlock()

	e.tick()

	state := e.State()
	if state.Current == nil {
		t.Fatal("expected a track after auto-next")
	}
}

func TestEngineWaitsForDuration(t *testing.T) {
	e := newTestEngine()
	e.tick()

	state := e.State()
	if state.Current == nil {
		t.Fatal("expected a track")
	}
	if state.Current.Duration != 0 {
		t.Errorf("expected duration 0, got %d", state.Current.Duration)
	}

	firstID := state.Current.ID
	e.tick()
	if e.State().Current.ID != firstID {
		t.Error("should not change track while duration is unknown")
	}
}

func TestEngineUpdateDuration(t *testing.T) {
	e := newTestEngine()
	e.tick()

	e.UpdateDuration(200)
	if e.State().Current.Duration != 200 {
		t.Errorf("duration = %d, want 200", e.State().Current.Duration)
	}
}

func TestEngineListeners(t *testing.T) {
	e := newTestEngine()
	e.SetOnlineCount(func() int { return 7 })
	state := e.State()
	if state.Listeners != 7 {
		t.Errorf("listeners = %d, want 7", state.Listeners)
	}
}

func TestEngineTrackChangeCallback(t *testing.T) {
	e := newTestEngine()
	called := false
	e.SetOnTrackChange(func(track Track) {
		called = true
	})
	e.tick()
	if !called {
		t.Error("expected onTrackChange to be called on first tick")
	}
}

func TestEngineAdvancesThroughQueue(t *testing.T) {
	e := newTestEngine()
	seen := make(map[string]bool)
	for i := 0; i < 10; i++ {
		e.mu.Lock()
		e.playNext() // releases lock
		e.mu.RLock()
		seen[e.current.ID] = true
		e.mu.RUnlock()
	}
	if len(seen) != 10 {
		t.Errorf("expected 10 distinct tracks in first 10 plays, got %d", len(seen))
	}
}

func TestEngineReshufflesOnExhaustion(t *testing.T) {
	c := &Catalog{tracks: []Track{
		{ID: "a", Title: "A", Artist: "X", URL: "http://a"},
		{ID: "b", Title: "B", Artist: "X", URL: "http://b"},
		{ID: "c", Title: "C", Artist: "X", URL: "http://c"},
	}}
	e := NewEngineWithCatalog(c)

	for i := 0; i < 3; i++ {
		e.mu.Lock()
		e.playNext() // releases lock
	}
	e.mu.RLock()
	pos := e.queuePos
	e.mu.RUnlock()
	if pos != 3 {
		t.Fatalf("queuePos = %d, want 3", pos)
	}

	e.mu.Lock()
	e.playNext() // releases lock
	e.mu.RLock()
	pos = e.queuePos
	e.mu.RUnlock()
	if pos != 1 {
		t.Errorf("after reshuffle queuePos = %d, want 1", pos)
	}
}

func TestEngineEmptyQueue(t *testing.T) {
	c := &Catalog{tracks: []Track{}}
	e := NewEngineWithCatalog(c)
	e.tick() // should not panic
	if e.State().Current != nil {
		t.Error("expected no track with empty catalog")
	}
}
