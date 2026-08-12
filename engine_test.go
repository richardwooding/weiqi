package weiqi

import (
	"errors"
	"testing"
)

func idx(row, col int) int { return row*N + col }

// build makes a board from (row, col, colour) triples.
func build(stones ...[3]int) Board {
	var b Board
	for _, s := range stones {
		b[idx(s[0], s[1])] = int8(s[2])
	}
	return b
}

// --- engine: captures --------------------------------------------------------

func TestCaptureSingleStone(t *testing.T) {
	// White at (0,0) with black at (1,0); black plays (0,1) → white captured.
	b := build([3]int{0, 0, 2}, [3]int{1, 0, 1})
	nb, captured, err := ApplyMove(b, 1, idx(0, 1), Board{})
	if err != nil {
		t.Fatalf("capture move: %v", err)
	}
	if len(captured) != 1 || captured[0] != idx(0, 0) {
		t.Fatalf("captured = %v, want [%d]", captured, idx(0, 0))
	}
	if nb[idx(0, 0)] != 0 {
		t.Fatalf("captured stone still on board")
	}
	if nb[idx(0, 1)] != 1 {
		t.Fatalf("placed stone missing")
	}
}

func TestCaptureGroup(t *testing.T) {
	// White group (0,0)-(0,1) with black on all liberties but (0,2); black
	// plays (0,2) → the two-stone group is captured.
	b := build([3]int{0, 0, 2}, [3]int{0, 1, 2}, [3]int{1, 0, 1}, [3]int{1, 1, 1})
	nb, captured, err := ApplyMove(b, 1, idx(0, 2), Board{})
	if err != nil {
		t.Fatalf("capture group: %v", err)
	}
	if len(captured) != 2 {
		t.Fatalf("captured %d stones, want 2 (%v)", len(captured), captured)
	}
	if nb[idx(0, 0)] != 0 || nb[idx(0, 1)] != 0 {
		t.Fatalf("group not fully removed: %v", nb[:3])
	}
}

func TestCapturePrecedesSuicide(t *testing.T) {
	// The placed stone would have no liberties, but it captures first and so
	// gains one — legal, not suicide.
	b := build([3]int{0, 0, 2}, [3]int{1, 0, 1})
	if _, captured, err := ApplyMove(b, 1, idx(0, 1), Board{}); err != nil || len(captured) != 1 {
		t.Fatalf("capture-not-suicide: err=%v captured=%v", err, captured)
	}
}

// --- engine: suicide & occupancy --------------------------------------------

func TestSuicideRejected(t *testing.T) {
	// Corner (0,0) surrounded by white with other liberties → black self-fill
	// captures nothing and dies → suicide.
	b := build([3]int{0, 1, 2}, [3]int{1, 0, 2})
	if _, _, err := ApplyMove(b, 1, idx(0, 0), Board{}); !errors.Is(err, ErrSuicide) {
		t.Fatalf("suicide: got %v", err)
	}
}

func TestOccupiedAndOffBoard(t *testing.T) {
	b := build([3]int{4, 4, 1})
	if _, _, err := ApplyMove(b, 2, idx(4, 4), Board{}); !errors.Is(err, ErrOccupied) {
		t.Fatalf("occupied: %v", err)
	}
	if _, _, err := ApplyMove(b, 1, -1, Board{}); !errors.Is(err, ErrOffBoard) {
		t.Fatalf("off board: %v", err)
	}
	if _, _, err := ApplyMove(b, 1, N*N, Board{}); !errors.Is(err, ErrOffBoard) {
		t.Fatalf("off board high: %v", err)
	}
}

// --- engine: ko --------------------------------------------------------------

// koSetup returns the board just before black captures into a ko, plus the
// index black plays and the index white would recapture at.
func koSetup() (b0 Board, blackPlay, whiteRecapture int) {
	b0 = build(
		[3]int{3, 4, 1}, [3]int{4, 3, 1}, [3]int{5, 4, 1}, // black diamond
		[3]int{3, 5, 2}, [3]int{4, 4, 2}, [3]int{4, 6, 2}, [3]int{5, 5, 2}, // white
	)
	return b0, idx(4, 5), idx(4, 4)
}

func TestKoRecaptureRejected(t *testing.T) {
	b0, blackPlay, whiteRecapture := koSetup()

	// Black captures the single white stone (no ko constraint yet).
	nb1, captured, err := ApplyMove(b0, 1, blackPlay, Board{})
	if err != nil || len(captured) != 1 {
		t.Fatalf("black capture: err=%v captured=%v", err, captured)
	}

	// White immediately recapturing would recreate b0 → ko, illegal.
	if _, _, err := ApplyMove(nb1, 2, whiteRecapture, b0); !errors.Is(err, ErrKo) {
		t.Fatalf("ko recapture should be rejected, got %v", err)
	}

	// The identical recapture is legal once the ko no longer forbids it.
	if _, captured, err := ApplyMove(nb1, 2, whiteRecapture, Board{}); err != nil || len(captured) != 1 {
		t.Fatalf("recapture without ko: err=%v captured=%v", err, captured)
	}
}

func TestLegalMovesExcludeSuicideAndKo(t *testing.T) {
	b0, blackPlay, whiteRecapture := koSetup()
	nb1, _, err := ApplyMove(b0, 1, blackPlay, Board{})
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	legal := LegalMoves(nb1, 2, b0)
	for _, m := range legal {
		if int(m) == whiteRecapture {
			t.Fatalf("LegalMoves included the ko recapture %d", whiteRecapture)
		}
	}
}

// --- engine: scoring ---------------------------------------------------------

func TestScoreFilledPartition(t *testing.T) {
	// Columns 0-4 black, 5-8 white; every point a stone, no territory.
	var b Board
	for r := 0; r < N; r++ {
		for c := 0; c < N; c++ {
			if c <= 4 {
				b[idx(r, c)] = 1
			} else {
				b[idx(r, c)] = 2
			}
		}
	}
	bs, ws, winner := FinalScore(b)
	if bs != 45 || ws != 36+Komi || winner != 1 {
		t.Fatalf("score = black %.1f white %.1f winner %d; want 45 / 42.5 / 1", bs, ws, winner)
	}
}

func TestScoreTerritory(t *testing.T) {
	// Black seals the (0,0) corner, white seals (8,8); the rest is one big
	// region touching both → neutral. Each side: 2 stones + 1 point.
	b := build(
		[3]int{0, 1, 1}, [3]int{1, 0, 1},
		[3]int{8, 7, 2}, [3]int{7, 8, 2},
	)
	black, white := score(b)
	if black != 3 || white != 3 {
		t.Fatalf("area = black %d white %d, want 3 / 3", black, white)
	}
	bs, ws, winner := FinalScore(b)
	if bs != 3 || ws != 3+Komi || winner != 2 {
		t.Fatalf("final = %.1f / %.1f winner %d; want 3 / 9.5 / 2 (komi)", bs, ws, winner)
	}
}
