package jukebox

import (
	"context"
	"math/rand/v2"
	"sync"
	"time"
)

type EngineState struct {
	Current   *Track
	Position  time.Duration
	Listeners int
}

type Engine struct {
	mu            sync.RWMutex
	queue         []Track
	queuePos      int
	current       *Track
	playStart     time.Time
	onlineCount   func() int
	onStateChange func()
	onTrackChange func(Track)
}

func NewEngineWithCatalog(c *Catalog) *Engine {
	tracks := c.AllTracks()
	rand.Shuffle(len(tracks), func(i, j int) {
		tracks[i], tracks[j] = tracks[j], tracks[i]
	})
	return &Engine{queue: tracks}
}

func (e *Engine) SetOnStateChange(fn func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onStateChange = fn
}

func (e *Engine) SetOnTrackChange(fn func(Track)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onTrackChange = fn
}

func (e *Engine) SetOnlineCount(fn func() int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onlineCount = fn
}

func (e *Engine) State() EngineState {
	e.mu.RLock()
	defer e.mu.RUnlock()

	online := 0
	if e.onlineCount != nil {
		online = e.onlineCount()
	}

	state := EngineState{
		Current:   e.current,
		Listeners: online,
	}
	if e.current != nil {
		state.Position = time.Since(e.playStart)
	}
	return state
}

func (e *Engine) UpdateDuration(seconds int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.current != nil && seconds > 0 {
		e.current.Duration = seconds
		e.notifyChange()
	}
}

func (e *Engine) RetryTrack() {
	e.mu.Lock()
	e.playNext()
}

func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.tick()
		}
	}
}

func (e *Engine) tick() {
	e.mu.Lock()

	if e.current == nil {
		e.playNext()
		return
	}

	duration := time.Duration(e.current.Duration) * time.Second
	if duration == 0 {
		e.mu.Unlock()
		return
	}

	if time.Since(e.playStart) >= duration {
		e.playNext()
		return
	}

	e.mu.Unlock()
}

func (e *Engine) playNext() {
	if len(e.queue) == 0 {
		e.mu.Unlock()
		return
	}

	if e.queuePos >= len(e.queue) {
		rand.Shuffle(len(e.queue), func(i, j int) {
			e.queue[i], e.queue[j] = e.queue[j], e.queue[i]
		})
		e.queuePos = 0
	}

	pick := e.queue[e.queuePos]
	e.queuePos++
	e.current = &pick
	e.playStart = time.Now()
	e.notifyChange()

	fn := e.onTrackChange
	e.mu.Unlock()

	if fn != nil {
		fn(pick)
	}
}

func (e *Engine) notifyChange() {
	if e.onStateChange != nil {
		e.onStateChange()
	}
}
