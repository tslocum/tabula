package tabula

import (
	"fmt"
	"math"
	"sync"
)

var (
	WeightBlot     = 1.03
	WeightHit      = -0.5
	WeightOppScore = -3.0
)

// rollProbabilities is a table of the probability of each roll combination.
var rollProbabilities = [21][3]int{
	{1, 1, 1},
	{1, 2, 2},
	{1, 3, 2},
	{1, 4, 2},
	{1, 5, 2},
	{1, 6, 2},
	{2, 2, 1},
	{2, 3, 2},
	{2, 4, 2},
	{2, 5, 2},
	{2, 6, 2},
	{3, 3, 1},
	{3, 4, 2},
	{3, 5, 2},
	{3, 6, 2},
	{4, 4, 1},
	{4, 5, 2},
	{4, 6, 2},
	{5, 5, 1},
	{5, 6, 2},
	{6, 6, 1},
}

type Analysis struct {
	Board Board
	Moves [][]int
	Past  bool
	Score float64

	Pips        int
	Blots       int
	Hits        int
	PlayerScore float64

	OppPips  float64
	OppBlots float64
	OppHits  float64
	OppScore float64

	player   int
	hitScore int
	chance   int
}

func (a *Analysis) _analyze(result *[]*Analysis, resultMutex *sync.Mutex, w *sync.WaitGroup) {
	var hs int
	o := opponent(a.player)
	for i := 0; i < len(a.Moves); i++ {
		move := a.Moves[i]
		checkers := a.Board.Checkers(o, move[1])
		if checkers == 1 {
			hs += pseudoPips(o, move[1])
		}
		a.Board = a.Board.Move(move[0], move[1], a.player).UseRoll(move[0], move[1], a.player)
	}
	a.Board.evaluate(a.player, hs, a)

	if a.player == 1 && !a.Past {
		const bufferSize = 1024
		oppResults := make([]*Analysis, 0, bufferSize)
		oppResultMutex := &sync.Mutex{}
		wg := &sync.WaitGroup{}
		wg.Add(21)
		for j := 0; j < 21; j++ {
			j := j
			go func() {
				check := rollProbabilities[j]
				bc := a.Board
				bc[SpaceRoll1], bc[SpaceRoll2] = int8(check[0]), int8(check[1])
				if check[0] == check[1] {
					bc[SpaceRoll3], bc[SpaceRoll4] = int8(check[0]), int8(check[1])
				} else {
					bc[SpaceRoll3], bc[SpaceRoll4] = 0, 0
				}
				available, _ := bc.Available(2)
				if len(available) == 0 {
					a := &Analysis{
						Board:  bc,
						Past:   a.Past,
						player: 2,
						chance: check[2],
					}
					bc.evaluate(a.player, 0, a)
					oppResultMutex.Lock()
					for i := 0; i < check[2]; i++ {
						oppResults = append(oppResults, a)
					}
					oppResultMutex.Unlock()
					wg.Done()
					return
				}
				wg.Add(len(available) - 1)
				for _, moves := range available {
					a := &Analysis{
						Board:  bc,
						Moves:  moves,
						Past:   a.Past,
						player: 2,
						chance: check[2],
					}
					go a._analyze(&oppResults, oppResultMutex, wg)
				}
			}()
		}
		wg.Wait()

		var oppPips float64
		var oppBlots float64
		var oppHits float64
		var oppScore float64
		var count float64
		for _, r := range oppResults {
			oppPips += float64(r.Pips)
			oppBlots += float64(r.Blots)
			oppHits += float64(r.Hits)
			oppScore += r.PlayerScore
			count++
		}
		if count == 0 {
			a.Score = a.PlayerScore
		} else {
			a.OppPips = (oppPips / count)
			a.OppBlots = (oppBlots / count)
			a.OppHits = (oppHits / count)
			a.OppScore = (oppScore / count)
			score := a.PlayerScore
			if !math.IsNaN(oppScore) {
				score += a.OppScore * WeightOppScore
			}
			a.Score = score
		}
	} else {
		a.Score = a.PlayerScore
	}

	resultMutex.Lock()
	for i := 0; i < a.chance; i++ {
		*result = append(*result, a)
	}
	resultMutex.Unlock()
	w.Done()
}

func (a *Analysis) String() string {
	return fmt.Sprintf("Moves: %s Score: %.2f - Score: %.2f Pips: %d Blots: %d Hits: %d /  Score: %.2f Pips: %.2f Blots: %.2f Hits: %.2f Past: %v", fmt.Sprint(a.Moves), a.Score, a.PlayerScore, a.Pips, a.Blots, a.Hits, a.OppScore, a.OppPips, a.OppBlots, a.OppHits, a.Past)
}
