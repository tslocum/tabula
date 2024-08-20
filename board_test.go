package tabula

import (
	"fmt"
	"testing"
)

func TestAvailableHighRoll(t *testing.T) {
	b := Board{0, 0, 0, 0, -2, -2, -2, -2, -2, -1, -2, 0, -2, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2, 4, 0, 0, 1, 1, 0}
	available, _ := b.Available(1)

	type testCase struct {
		moves [4][2]int8
		found int
	}
	var testCases = []*testCase{
		{
			moves: [4][2]int8{{13, 9}},
		},
	}
	for i, testCase := range testCases {
		for _, moves := range available {
			if movesEqual(moves, testCase.moves) {
				testCase.found++
				continue
			}
			t.Errorf("unexpected available moves for test case %d: %+v", i, moves)
		}
		if testCase.found != 1 {
			t.Errorf("unexpected number of found moves for test case %d: expected 1 move, got %d", i, testCase.found)
		}
	}
}

func TestMoveBackgammon(t *testing.T) {
	b := NewBoard(VariantBackgammon)
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

func TestBearOffBackgammon(t *testing.T) {
	b := Board{0, 0, 2, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -1, 0, 0, -3, 0, 0, -3, -6, -2, 0, 0, 0, 3, 3, 0, 0, 1, 1, 0}
	available, _ := b.Available(1)
	got, expected := len(available), 1
	if got != expected {
		t.Errorf("unexpected number of move combinations: expected %d: got %d", expected, got)
	}
	var foundMoves [][2]int8
	for i := 0; i < 4; i++ {
		move := available[0][i]
		if move[0] == 0 && move[1] == 0 {
			break
		}
		foundMoves = append(foundMoves, move)
	}
	got, expected = len(foundMoves), 2
	if got != expected {
		t.Errorf("unexpected number of legal moves: expected %d: got %d", expected, got)
	}
	var found30 bool
	var found41 bool
	for _, move := range foundMoves {
		if move[0] == 3 && move[1] == 0 {
			found30 = true
		} else if move[0] == 4 && move[1] == 1 {
			found41 = true
		}
	}
	if !found30 {
		t.Errorf("expected move 3/0 was not found")
	}
	if !found41 {
		t.Errorf("expected move 4/1 was not found")
	}
}

func TestMoveTabula(t *testing.T) {
	{
		b := NewBoard(VariantTabula)
		b[SpaceRoll1] = 1
		b[SpaceRoll2] = 2

		bc := b.Move(SpaceHomePlayer, 23, 1)
		got, expected := bc[SpaceHomePlayer], int8(14)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", SpaceHomePlayer, expected, got)
		}
		got, expected = bc[23], int8(1)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 23, expected, got)
		}
		bc = bc.Move(SpaceHomePlayer, 22, 1)
		got, expected = bc[SpaceHomePlayer], int8(13)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", SpaceHomePlayer, expected, got)
		}
		got, expected = bc[22], int8(1)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 22, expected, got)
		}
	}

	{
		b := Board{0, 1, 2, 3, 4, 4, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -15, 0, 0, 0, 0, 1, 0, 2}
		b[SpaceRoll1] = 1
		b[SpaceRoll2] = 2

		bc := b.UseRoll(1, 2, 1).Move(1, 2, 1)
		got, expected := bc[1], int8(0)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 1, expected, got)
		}
		got, expected = bc[2], int8(3)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 2, expected, got)
		}

		if !bc.HaveRoll(5, 7, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, true)
		}

		if bc.HaveRoll(12, 13, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, true)
		}

		if !bc.HaveRoll(12, 14, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, true)
		}
	}

	{
		b := Board{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -15, 0, 0, 0, 0, 0, 0, 2}
		b[SpaceRoll1] = 1
		b[SpaceRoll2] = 2

		if b.HaveRoll(12, 13, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, true)
		}

		bc := b.UseRoll(0, 1, 1).Move(0, 1, 1)
		got, expected := bc[0], int8(0)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 0, expected, got)
		}
		got, expected = bc[1], int8(1)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 1, expected, got)
		}

		if bc.HaveRoll(12, 13, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, false)
		}

		if !bc.HaveRoll(12, 14, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, false)
		}
	}

	{
		b := Board{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -15, 0, 0, 0, 0, 0, 0, 1, 0, 2}
		b[SpaceRoll1] = 1
		b[SpaceRoll2] = 2

		if !b.HaveRoll(12, 14, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, false)
		}

		if b.MayBearOff(1) {
			t.Errorf("unexpected bear off value: expected %v: got %v", false, true)
		}

		b = b.Move(12, 14, 1).UseRoll(12, 14, 1)
		got, expected := b[12], int8(0)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 12, expected, got)
		}
		got, expected = b[14], int8(1)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 14, expected, got)
		}

		if !b.MayBearOff(1) {
			t.Errorf("unexpected bear off value: expected %v: got %v", true, false)
		}

		if !b.HaveRoll(14, 15, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, false)
		}

		if b.HaveRoll(14, 16, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, true)
		}
	}

	{
		b := Board{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, -15, 0, 0, 0, 0, 0, 0, 1, 0, 2}
		b[SpaceRoll1] = 1
		b[SpaceRoll2] = 2

		if !b.HaveRoll(12, 14, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, false)
		}

		if b.MayBearOff(1) {
			t.Errorf("unexpected bear off value: expected %v: got %v", false, true)
		}

		if b.HaveRoll(24, 0, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", false, true)
		}

		b = b.UseRoll(12, 14, 1).Move(12, 14, 1)
		got, expected := b[12], int8(0)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 12, expected, got)
		}
		got, expected = b[14], int8(1)
		if got != expected {
			t.Errorf("unexpected space %d value: expected %d: got %d", 14, expected, got)
		}

		if !b.MayBearOff(1) {
			t.Errorf("unexpected bear off value: expected %v: got %v", true, false)
		}

		if !b.HaveRoll(24, 0, 1) {
			t.Errorf("unexpected have roll value: expected %v: got %v", true, false)
		}
	}
}

func TestPastBackgammon(t *testing.T) {
	b := NewBoard(VariantBackgammon)
	got, expected := b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, 1, 3, -1, 0, -1, 0, -2, 0, 0, -1, 0, -3, 3, 0, 0, 0, -2, 0, -5, 4, 0, 2, 2, 0, 0, 0, 0, 5, 5, 5, 5, 1, 1, 0}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, -1, 1, -2, -1, 2, 4, 0, 0, 0, 0, 0, -1, 2, -1, 0, 0, -1, 3, -3, 0, 3, -1, -3, -1, 0, 0, 0, 4, 3, 0, 0, 1, 1, 0}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{7, 2, 2, 4, 0, -2, 0, 0, -1, 0, -1, 0, 0, 0, 0, 0, -1, -1, 0, -4, 0, -2, -1, -1, -1, 0, 0, 0, 6, 2, 0, 0, 1, 1, 0}
	got, expected = b.Past(), true
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}
}

func TestPastAcey(t *testing.T) {
	b := NewBoard(VariantAceyDeucey)
	got, expected := b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, 1, 3, -1, 0, -1, 0, -2, 0, 0, -1, 0, -3, 3, 0, 0, 0, -2, 0, -5, 4, 0, 2, 2, 0, 0, 0, 0, 5, 5, 5, 5, 1, 1, 1}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, -1, 1, -2, -1, 2, 4, 0, 0, 0, 0, 0, -1, 2, -1, 0, 0, -1, 3, -3, 0, 3, -1, -3, -1, 0, 0, 0, 4, 3, 0, 0, 1, 1, 1}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{7, 2, 2, 4, 0, -2, 0, 0, -1, 0, -1, 0, 0, 0, 0, 0, -1, -1, 0, -4, 0, -2, -1, -1, -1, 0, 0, 0, 6, 2, 0, 0, 0, 1, 1}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{7, 2, 2, 4, 0, -2, 0, 0, -1, 0, -1, 0, 0, 0, 0, 0, -1, -1, 0, -4, 0, -2, -1, -1, -1, 0, 0, 0, 6, 2, 0, 0, 1, 1, 1}
	got, expected = b.Past(), true
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}
}

func TestPastTabula(t *testing.T) {
	b := NewBoard(VariantTabula)
	got, expected := b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, 1, 3, -1, 0, -1, 0, -2, 0, 0, -1, 0, -3, 3, 0, 0, 0, -2, 0, -5, 4, 0, 2, 2, 0, 0, 0, 0, 5, 5, 5, 5, 1, 1, 2}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{0, -1, 1, -2, -1, 2, 4, 0, 0, 0, 0, 0, -1, 2, -1, 0, 0, -1, 3, -3, 0, 3, -1, -3, -1, 0, 0, 0, 4, 3, 0, 0, 1, 1, 2}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}

	b = Board{7, 2, 2, 4, 0, -2, 0, 0, -1, 0, -1, 0, 0, 0, 0, 0, -1, -1, 0, -4, 0, -2, -1, -1, -1, 0, 0, 0, 6, 2, 0, 0, 1, 1, 2}
	got, expected = b.Past(), false
	if got != expected {
		t.Errorf("unexpected past value: expected %v: got %v", expected, got)
	}
}

func TestBlots(t *testing.T) {
	b := NewBoard(VariantBackgammon)
	got, expected := b.Blots(1), 0
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}
	got, expected = b.Blots(2), 0
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}

	b = b.Move(24, 23, 1)

	got, expected = b.Blots(1), 19
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}
	got, expected = b.Blots(2), 0
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}

	b = b.Move(1, 2, 2)

	got, expected = b.Blots(1), 19
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}
	got, expected = b.Blots(2), 19
	if got != expected {
		t.Errorf("unexpected blots value: expected %v: got %v", expected, got)
	}
}

func TestHitScore(t *testing.T) {
	b := Board{0, 0, -2, -2, -2, 4, 0, -1, 0, 0, -2, 4, 0, -2, -1, 0, -2, 4, 0, 2, 0, 0, 0, 0, -1, 0, 1, 0, 4, 1, 0, 0, 1, 1, 1}
	available, _ := b.Available(1)
	analysis := make([]*Analysis, 0, AnalysisBufferSize)
	b.Analyze(available, &analysis, false)

	var reached bool
	minHitScore := 200
	for _, a := range analysis {
		if a.hitScore >= minHitScore {
			reached = true
			break
		}
	}
	if !reached {
		t.Errorf("unexpected hit score for analysis: expected hit score of at least %d", minHitScore)
	}
}

func TestAnalyze(t *testing.T) {
	b := NewBoard(VariantBackgammon)
	b = b.Move(24, 23, 1)
	b = b.Move(1, 2, 2)
	b[SpaceRoll1], b[SpaceRoll2] = 1, 2

	available, _ := b.Available(1)
	analysis := make([]*Analysis, 0, AnalysisBufferSize)
	b.Analyze(available, &analysis, false)
	var blots int
	for _, r := range analysis {
		blots += r.Blots
	}
	if blots <= 0 {
		t.Errorf("expected >0 blots in results, got %d", blots)
	}

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
		t.Run(fmt.Sprintf("%d-%d", c.roll1, c.roll2), func(t *testing.T) {
			board := NewBoard(VariantBackgammon)
			board[SpaceRoll1] = c.roll1
			board[SpaceRoll2] = c.roll2
			board[SpaceRoll3] = c.roll3
			board[SpaceRoll4] = c.roll4
			available, _ := board.Available(1)

			board.Analyze(available, &analysis, false)
		})
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
			board := NewBoard(VariantBackgammon)
			board[SpaceRoll1] = c.roll1
			board[SpaceRoll2] = c.roll2
			board[SpaceRoll3] = c.roll3
			board[SpaceRoll4] = c.roll4

			var available [][4][2]int8
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				available, _ = board.Available(1)
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
			board := NewBoard(VariantBackgammon)
			board[SpaceRoll1] = c.roll1
			board[SpaceRoll2] = c.roll2
			board[SpaceRoll3] = c.roll3
			board[SpaceRoll4] = c.roll4
			available, _ := board.Available(1)
			analysis := make([]*Analysis, 0, AnalysisBufferSize)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				board.Analyze(available, &analysis, false)
			}

			_ = analysis
		})
	}
}

func BenchmarkChooseDoubles(b *testing.B) {
	board := Board{0, -2, -2, -3, -1, -2, -3, 0, -2, 0, 0, 0, 0, 0, 0, 0, 2, 4, 0, 2, 2, 5, 0, 0, 0, 0, 0, 0, 1, 2, 0, 0, 1, 1, 1}
	analysis := make([]*Analysis, 0, AnalysisBufferSize)

	var doubles int
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doubles = board.ChooseDoubles(&analysis)
	}
	_ = doubles
}

func BenchmarkMayBearOff(b *testing.B) {
	board := NewBoard(VariantBackgammon)

	var allowed bool
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		allowed = board.MayBearOff(1)
	}

	_ = allowed
}
