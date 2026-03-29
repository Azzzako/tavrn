package sudoku

import "testing"

// helper: create a game and find one clue cell and one empty cell.
func clueAndEmpty(g *Game) (clueR, clueC, emptyR, emptyC int) {
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].IsClue {
				clueR, clueC = r, c
			}
			if g.board[r][c].Value == 0 {
				emptyR, emptyC = r, c
			}
		}
	}
	return
}

func TestNewGame(t *testing.T) {
	g := NewGame("easy")

	// Board must have clues.
	clues := 0
	board := g.Board()
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if board[r][c].IsClue {
				clues++
			}
		}
	}
	if clues == 0 {
		t.Fatal("expected clues on the board")
	}

	// Scores should be empty.
	if len(g.Scores()) != 0 {
		t.Fatal("expected empty scores")
	}

	// Difficulty stored.
	if g.Difficulty() != "easy" {
		t.Fatalf("expected easy, got %s", g.Difficulty())
	}
}

func TestPlaceCorrect(t *testing.T) {
	g := NewGame("easy")

	// Find an empty cell and place the correct value.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].Value == 0 {
				correct := g.solution[r][c]
				pts := g.Place("player1", r, c, correct)
				if pts != 1 {
					t.Fatalf("expected +1, got %d", pts)
				}
				if g.Score("player1") != 1 {
					t.Fatalf("expected score 1, got %d", g.Score("player1"))
				}
				return
			}
		}
	}
	t.Fatal("no empty cell found")
}

func TestPlaceWrong(t *testing.T) {
	g := NewGame("easy")

	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].Value == 0 {
				correct := g.solution[r][c]
				wrong := correct%9 + 1 // guaranteed != correct
				pts := g.Place("player1", r, c, wrong)
				if pts != -1 {
					t.Fatalf("expected -1, got %d", pts)
				}
				if g.Score("player1") != -1 {
					t.Fatalf("expected score -1, got %d", g.Score("player1"))
				}
				return
			}
		}
	}
	t.Fatal("no empty cell found")
}

func TestPlaceOnClue(t *testing.T) {
	g := NewGame("easy")

	clueR, clueC, _, _ := clueAndEmpty(g)
	original := g.board[clueR][clueC]
	pts := g.Place("player1", clueR, clueC, 5)
	if pts != 0 {
		t.Fatalf("expected 0 on clue, got %d", pts)
	}
	// Cell must be unchanged.
	if g.board[clueR][clueC] != original {
		t.Fatal("clue cell was modified")
	}
}

func TestClearOwnPlacement(t *testing.T) {
	g := NewGame("easy")

	_, _, emptyR, emptyC := clueAndEmpty(g)
	// Place a wrong number (correct placements are locked and can't be cleared)
	wrongVal := g.solution[emptyR][emptyC]%9 + 1
	if wrongVal == g.solution[emptyR][emptyC] {
		wrongVal = wrongVal%9 + 1
	}
	g.Place("player1", emptyR, emptyC, wrongVal)

	ok := g.Clear("player1", emptyR, emptyC)
	if !ok {
		t.Fatal("expected clear to succeed")
	}
	if g.board[emptyR][emptyC].Value != 0 {
		t.Fatal("cell not cleared")
	}
}

func TestCorrectPlacementIsLocked(t *testing.T) {
	g := NewGame("easy")

	_, _, emptyR, emptyC := clueAndEmpty(g)
	g.Place("player1", emptyR, emptyC, g.solution[emptyR][emptyC])

	// Can't clear a locked cell
	if g.Clear("player1", emptyR, emptyC) {
		t.Fatal("should not be able to clear locked cell")
	}
	// Can't overwrite a locked cell
	pts := g.Place("player1", emptyR, emptyC, g.solution[emptyR][emptyC]%9+1)
	if pts != 0 {
		t.Fatal("should not score on locked cell")
	}
}

func TestClearOtherPlacement(t *testing.T) {
	g := NewGame("easy")

	_, _, emptyR, emptyC := clueAndEmpty(g)
	g.Place("player1", emptyR, emptyC, g.solution[emptyR][emptyC])

	ok := g.Clear("player2", emptyR, emptyC)
	if ok {
		t.Fatal("should not clear another player's placement")
	}
	if g.board[emptyR][emptyC].Value == 0 {
		t.Fatal("cell was wrongly cleared")
	}
}

func TestTopScores(t *testing.T) {
	g := NewGame("easy")
	g.RegisterNickname("p1", "alice")
	g.RegisterNickname("p2", "bob")

	// Place some correct numbers for p1
	placed := 0
	for r := 0; r < 9 && placed < 3; r++ {
		for c := 0; c < 9 && placed < 3; c++ {
			if g.board[r][c].Value == 0 {
				g.Place("p1", r, c, g.solution[r][c])
				placed++
			}
		}
	}

	top := g.TopScores(5)
	if len(top) == 0 {
		t.Fatal("expected at least one score entry")
	}
	if top[0].Nickname != "alice" {
		t.Errorf("top scorer = %q, want alice", top[0].Nickname)
	}
}

func TestReset(t *testing.T) {
	g := NewGame("easy")
	_, _, er, ec := clueAndEmpty(g)
	g.Place("p1", er, ec, g.solution[er][ec])

	g.Reset()

	if g.Score("p1") != 0 {
		t.Error("scores should reset")
	}
	if g.Filled() >= 81 {
		t.Error("board should have empty cells after reset")
	}
}

func TestIsSolved(t *testing.T) {
	g := NewGame("easy")

	if g.IsSolved() {
		t.Fatal("should not be solved at start")
	}

	// Fill every empty cell with the correct value.
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			if g.board[r][c].Value == 0 {
				g.Place("solver", r, c, g.solution[r][c])
			}
		}
	}

	if !g.IsSolved() {
		t.Fatal("should be solved after filling all cells correctly")
	}
}
