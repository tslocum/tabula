package tabula

import (
	"fmt"
	"runtime"
	"sync"
)

const queueBufferSize = 4096000

var (
	WeightBlot     = 1.1
	WeightHit      = -0.9
	WeightOppScore = -1.5
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

var analysisQueue = make(chan *Analysis, queueBufferSize)

func init() {
	cpus := runtime.NumCPU()
	if cpus < 1 {
		cpus = 1
	}
	for i := 0; i < cpus; i++ {
		go analyzer()
	}
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

	result      *[]*Analysis
	resultMutex *sync.Mutex

	OppPips  float64
	OppBlots float64
	OppHits  float64
	OppScore float64

	player   int
	hitScore int
	chance   int
	wg       *sync.WaitGroup
}

func (a *Analysis) _analyze() {
	var hs int
	o := opponent(a.player)
	for i := 0; i < len(a.Moves); i++ {
		move := a.Moves[i]
		checkers := a.Board.Checkers(o, move[1])
		if checkers == 1 {
			hs += PseudoPips(o, move[1])
		}
		a.Board = a.Board.Move(move[0], move[1], a.player).UseRoll(move[0], move[1], a.player)
	}
	a.Board.evaluate(a.player, hs, a)

	if a.player == 1 && !a.Past {
		a.wg.Add(21)
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
					{
						a := &Analysis{
							Board:       bc,
							Past:        a.Past,
							player:      2,
							chance:      check[2],
							result:      a.result,
							resultMutex: a.resultMutex,
						}
						bc.evaluate(a.player, 0, a)
						a.resultMutex.Lock()
						for i := 0; i < a.chance; i++ {
							*a.result = append(*a.result, a)
						}
						a.resultMutex.Unlock()
					}
					a.wg.Done()
					return
				}
				a.wg.Add(len(available))
				for _, moves := range available {
					a := &Analysis{
						Board:       bc,
						Moves:       moves,
						Past:        a.Past,
						player:      2,
						chance:      check[2],
						result:      a.result,
						resultMutex: a.resultMutex,
						wg:          a.wg,
					}
					analysisQueue <- a
				}
				a.wg.Done()
			}()
		}
	} else if a.player == 2 {
		a.resultMutex.Lock()
		for i := 0; i < a.chance; i++ {
			*a.result = append(*a.result, a)
		}
		a.resultMutex.Unlock()
	}
	a.wg.Done()
}

func (a *Analysis) String() string {
	return fmt.Sprintf("Moves: %s Score: %.2f - Score: %.2f Pips: %d Blots: %d Hits: %d /  Score: %.2f Pips: %.2f Blots: %.2f Hits: %.2f Past: %v", fmt.Sprint(a.Moves), a.Score, a.PlayerScore, a.Pips, a.Blots, a.Hits, a.OppScore, a.OppPips, a.OppBlots, a.OppHits, a.Past)
}

func analyzer() {
	var a *Analysis
	for {
		a = <-analysisQueue
		a._analyze()
	}
}
