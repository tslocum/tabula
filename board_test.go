package tabula

import (
	"fmt"
	"testing"
)

func TestBoard(t *testing.T) {
	b := NewBoard()
	b[SpaceRoll1] = 1
	b[SpaceRoll2] = 2
	got, expected := b[24], int8(2)
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 24, expected, got)
	}
	got, expected = b[22], 0
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 22, expected, got)
	}
	bc := b.Move(24, 23, 1)
	got, expected = b[24], int8(2)
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 24, expected, got)
	}
	got, expected = bc[23], int8(1)
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 23, expected, got)
	}
	got, expected = bc[24], 1
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 24, expected, got)
	}
	got, expected = bc[22], 0
	if got != expected {
		t.Errorf("unexpected space %d value: expected %d: got %d", 22, expected, got)
	}
}

func BenchmarkAvailable(b *testing.B) {
	type testCase struct {
		roll1, roll2, roll3, roll4 int8
	}
	cases := []*testCase{
		{1, 1, 1, 1},
		{2, 2, 2, 2},
		{3, 3, 3, 3},
		{4, 4, 4, 4},
		{5, 5, 5, 5},
		{6, 6, 6, 6},
		{1, 2, 0, 0},
		{2, 3, 0, 0},
		{3, 4, 0, 0},
		{4, 5, 0, 0},
		{5, 6, 0, 0},
	}
	for _, c := range cases {
		b.Run(fmt.Sprintf("%d-%d", c.roll1, c.roll2), func(b *testing.B) {
			board := NewBoard()
			board[SpaceRoll1] = c.roll1
			board[SpaceRoll2] = c.roll2
			board[SpaceRoll3] = c.roll3
			board[SpaceRoll4] = c.roll4

			var available [][]int
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				available = board.Available(1)
			}

			_ = available
		})
	}
}

func BenchmarkAnalyze(b *testing.B) {
	type testCase struct {
		roll1, roll2, roll3, roll4 int8
	}
	cases := []*testCase{
		{1, 1, 1, 1},
		{2, 2, 2, 2},
		{3, 3, 3, 3},
		{4, 4, 4, 4},
		{5, 5, 5, 5},
		{6, 6, 6, 6},
		{1, 2, 0, 0},
		{2, 3, 0, 0},
		{3, 4, 0, 0},
		{4, 5, 0, 0},
		{5, 6, 0, 0},
	}
	for _, c := range cases {
		b.Run(fmt.Sprintf("%d-%d", c.roll1, c.roll2), func(b *testing.B) {
			board := NewBoard()
			board[SpaceRoll1] = c.roll1
			board[SpaceRoll2] = c.roll2
			board[SpaceRoll3] = c.roll3
			board[SpaceRoll4] = c.roll4
			available := board.Available(1)

			var analysis []*Analysis
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				analysis = board.Analyze(1, available)
			}

			_ = analysis
		})
	}
}
