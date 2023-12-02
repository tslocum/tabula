package tabula

import (
	"fmt"
)

type Analysis struct {
	Board Board
	Moves [][]int
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
	past     bool
}

func (a *Analysis) String() string {
	return fmt.Sprintf("Moves: %s Score: %.2f - Score: %.2f Pips: %d Blots: %d Hits: %d /  Score: %.2f Pips: %.2f Blots: %.2f Hits: %.2f", fmt.Sprint(a.Moves), a.Score, a.PlayerScore, a.Pips, a.Blots, a.Hits, a.OppScore, a.OppPips, a.OppBlots, a.OppHits)
}
