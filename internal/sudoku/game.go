package sudoku

import (
	"sync"
	"time"
)

// Cell represents a single cell on the shared game board.
type Cell struct {
	Value    int
	PlacedBy string // fingerprint, empty for clues
	IsClue   bool
	Locked   bool // correct placement — immutable
}

// Position identifies a row/col on the board.
type Position struct {
	Row, Col int
}

// ScoreEntry is a player's score for the scoreboard.
type ScoreEntry struct {
	Nickname string
	Score    int
}

// Game holds the shared multiplayer state for a single Sudoku puzzle.
type Game struct {
	mu         sync.RWMutex
	puzzle     Board      // starting state (clues only)
	solution   Board      // answer key
	board      [9][9]Cell // current state
	scores     map[string]int
	nicknames  map[string]string // fingerprint → nickname
	cursors    map[string]Position
	started    time.Time
	difficulty string
	onUpdate   func() // called after state changes
}

// NewGame creates a game with a freshly generated puzzle.
func NewGame(difficulty string) *Game {
	puzzle, solution := Generate(difficulty)
	g := &Game{
		puzzle:     puzzle,
		solution:   solution,
		scores:     make(map[string]int),
		nicknames:  make(map[string]string),
		cursors:    make(map[string]Position),
		started:    time.Now(),
		difficulty: difficulty,
	}
	// Initialize board from puzzle
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if puzzle[r][c] != 0 {
				g.board[r][c] = Cell{Value: puzzle[r][c], IsClue: true}
			}
		}
	}
	return g
}

// SetOnUpdate registers a callback invoked after every state change.
func (g *Game) SetOnUpdate(fn func()) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.onUpdate = fn
}

// RegisterNickname associates a fingerprint with a display name.
func (g *Game) RegisterNickname(fingerprint, nickname string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nicknames[fingerprint] = nickname
}

// Place puts a number on the board. Returns points earned (+1 or -1).
func (g *Game) Place(fingerprint string, row, col, value int) int {
	g.mu.Lock()
	defer g.mu.Unlock()
	if row < 0 || row > 8 || col < 0 || col > 8 || value < 1 || value > 9 {
		return 0
	}
	cell := g.board[row][col]
	if cell.IsClue || cell.Locked {
		return 0
	}
	// Same value already there — no-op
	if cell.Value == value {
		return 0
	}
	correct := g.solution[row][col] == value
	g.board[row][col] = Cell{Value: value, PlacedBy: fingerprint, Locked: correct}
	points := -1
	if correct {
		points = 1
	}
	g.scores[fingerprint] += points
	g.notify()
	return points
}

// Clear removes a player's own placement.
func (g *Game) Clear(fingerprint string, row, col int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if row < 0 || row > 8 || col < 0 || col > 8 {
		return false
	}
	cell := g.board[row][col]
	if cell.IsClue || cell.Locked || cell.PlacedBy != fingerprint || cell.Value == 0 {
		return false
	}
	g.board[row][col] = Cell{}
	g.notify()
	return true
}

// Reset generates a new puzzle, keeping nicknames and cursors but resetting scores and board.
func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	puzzle, solution := Generate(g.difficulty)
	g.puzzle = puzzle
	g.solution = solution
	g.board = [9][9]Cell{}
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if puzzle[r][c] != 0 {
				g.board[r][c] = Cell{Value: puzzle[r][c], IsClue: true}
			}
		}
	}
	g.scores = make(map[string]int)
	g.started = time.Now()
	g.notify()
}

// TopScores returns up to n top scorers sorted by score descending.
func (g *Game) TopScores(n int) []ScoreEntry {
	g.mu.RLock()
	defer g.mu.RUnlock()
	var entries []ScoreEntry
	for fp, sc := range g.scores {
		nick := fp
		if name, ok := g.nicknames[fp]; ok {
			nick = name
		}
		entries = append(entries, ScoreEntry{Nickname: nick, Score: sc})
	}
	// Sort descending by score
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].Score > entries[i].Score {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}

// SetCursor updates a player's cursor position.
func (g *Game) SetCursor(fingerprint string, row, col int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cursors[fingerprint] = Position{Row: row, Col: col}
}

// RemovePlayer removes a player's cursor (on disconnect).
func (g *Game) RemovePlayer(fingerprint string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.cursors, fingerprint)
}

// IsSolved returns true if all cells are correctly filled.
func (g *Game) IsSolved() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].Value != g.solution[r][c] {
				return false
			}
		}
	}
	return true
}

// Score returns a player's current score.
func (g *Game) Score(fingerprint string) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.scores[fingerprint]
}

// Scores returns a copy of all scores.
func (g *Game) Scores() map[string]int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cp := make(map[string]int, len(g.scores))
	for k, v := range g.scores {
		cp[k] = v
	}
	return cp
}

// Filled returns how many cells have values.
func (g *Game) Filled() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	n := 0
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].Value != 0 {
				n++
			}
		}
	}
	return n
}

// Difficulty returns the game difficulty.
func (g *Game) Difficulty() string {
	return g.difficulty
}

// Board returns a copy of the current board state.
func (g *Game) Board() [9][9]Cell {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.board
}

// Cursors returns a copy of cursor positions.
func (g *Game) Cursors() map[string]Position {
	g.mu.RLock()
	defer g.mu.RUnlock()
	cp := make(map[string]Position, len(g.cursors))
	for k, v := range g.cursors {
		cp[k] = v
	}
	return cp
}

// Started returns when the game started.
func (g *Game) Started() time.Time {
	return g.started
}

func (g *Game) notify() {
	if g.onUpdate != nil {
		g.onUpdate()
	}
}
