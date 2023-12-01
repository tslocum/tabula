package tabula

import (
	"fmt"
	"log"
	"sort"
	"sync"
)

var (
	WeightBlot     = 1.0
	WeightHit      = -1.0
	WeightOppScore = -0.5
)

const (
	SpaceHomePlayer   = 0
	SpaceHomeOpponent = 25
	SpaceBarPlayer    = 26
	SpaceBarOpponent  = 27
	SpaceRoll1        = 28
	SpaceRoll2        = 29
	SpaceRoll3        = 30
	SpaceRoll4        = 31
)

const (
	boardSpaces = 32
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
}

func (a *Analysis) String() string {
	return fmt.Sprintf("Moves: %s Score: %.2f - Score: %.2f Pips: %d Blots: %d Hits: %d /  Score: %.2f Pips: %.2f Blots: %.2f Hits: %.2f", fmt.Sprint(a.Moves), a.Score, a.PlayerScore, a.Pips, a.Blots, a.Hits, a.OppScore, a.OppPips, a.OppBlots, a.OppHits)
}

// Board represents the state of a game. It contains spaces for the checkers,
// as well as four "spaces" which contain the available die rolls.
type Board [boardSpaces]int8

// NewBoard returns a new board with checkers placed in their starting positions.
func NewBoard() Board {
	return Board{0, -2, 0, 0, 0, 0, 5, 0, 3, 0, 0, 0, -5, 5, 0, 0, 0, -3, 0, -5, 0, 0, 0, 0, 2, 0, 0, 0}
}

func (b Board) SetValue(space int, value int8) Board {
	b[space] = value
	return b
}

// Move moves a checker on the board.
func (b Board) Move(from int, to int, player int) Board {
	if b[from] == 0 || (player == 1 && b[from] < 0) || (player == 2 && b[from] > 0) {
		log.Panic("illegal move: no from checker", from, to, player)
	} else if b[to] != 0 {
		if (player == 1 && b[to] == -1) || (player == 2 && b[to] == 1) {
			b[to] = 0
			if player == 1 {
				b[SpaceBarOpponent] -= 1
			} else {
				b[SpaceBarPlayer] += 1
			}
		} else if (player == 1 && b[to] < 0) || (player == 2 && b[to] > 0) {
			b.Print()
			log.Panic("illegal move: existing checkers at to space", from, to, player, b[to])
		}
	}
	delta := int8(1)
	if player == 2 {
		delta = int8(-1)
	}
	b[from], b[to] = b[from]-delta, b[to]+delta
	return b
}

// Checkers returns the number of checkers at a space. It always returns a positive number.
func (b Board) Checkers(space int, player int) int8 {
	v := b[space]
	if player == 1 && v > 0 {
		return v
	} else if player == 2 && v < 0 {
		return v * -1
	}
	return 0
}

func (b Board) MayBearOff(player int) bool {
	homeStart := 1
	homeEnd := 6
	barSpace := SpaceBarPlayer
	if player == 2 {
		homeStart = 19
		homeEnd = 24
		barSpace = SpaceBarOpponent
	}
	if b.Checkers(barSpace, player) != 0 {
		return false
	}
	for space := 1; space < 25; space++ {
		if space >= homeStart && space <= homeEnd {
			continue
		} else if b.Checkers(space, player) != 0 {
			return false
		}
	}
	return true
}

// HaveRoll returns whether the player has a sufficient die roll for the specified move.
func (b Board) HaveRoll(from int, to int, player int) bool {
	delta := int8(spaceDiff(from, to))
	if delta == 0 {
		return false
	}
	playerDelta := -1
	playerHomeEnd := 6
	if player == 2 {
		playerDelta = 1
		playerHomeEnd = 19
	}
	if b.MayBearOff(player) {
		allowGreater := true
		for checkSpace := int8(0); checkSpace < 6-delta; checkSpace++ {
			if b.Checkers(playerHomeEnd+int(checkSpace)*playerDelta, player) != 0 {
				allowGreater = false
				break
			}
		}
		if allowGreater {
			return (b[SpaceRoll1] >= delta || b[SpaceRoll2] >= delta || b[SpaceRoll3] >= delta || b[SpaceRoll4] >= delta)
		}
	}
	return (b[SpaceRoll1] == delta || b[SpaceRoll2] == delta || b[SpaceRoll3] == delta || b[SpaceRoll4] == delta)
}

// UseRoll uses a die roll.
func (b Board) UseRoll(from int, to int, player int) Board {
	delta := int8(spaceDiff(from, to))
	if delta == 0 {
		b.Print()
		log.Panic("unknown space diff", from, to, player)
	}
	playerDelta := -1
	playerHomeEnd := 6
	if player == 2 {
		playerDelta = 1
		playerHomeEnd = 19
	}
	var allowGreater bool
	if b.MayBearOff(player) {
		allowGreater = true
		for checkSpace := int8(0); checkSpace < 6-delta; checkSpace++ {
			if b.Checkers(playerHomeEnd+int(checkSpace)*playerDelta, player) != 0 {
				allowGreater = false
				break
			}
		}
	}
	if allowGreater {
		switch {
		case b[SpaceRoll1] >= delta:
			b[SpaceRoll1] = 0
		case b[SpaceRoll2] >= delta:
			b[SpaceRoll2] = 0
		case b[SpaceRoll3] >= delta:
			b[SpaceRoll3] = 0
		case b[SpaceRoll4] >= delta:
			b[SpaceRoll4] = 0
		default:
			b.Print()
			log.Panic("no available roll for move", from, to, player)
		}
	} else {
		switch {
		case b[SpaceRoll1] == delta:
			b[SpaceRoll1] = 0
		case b[SpaceRoll2] == delta:
			b[SpaceRoll2] = 0
		case b[SpaceRoll3] == delta:
			b[SpaceRoll3] = 0
		case b[SpaceRoll4] == delta:
			b[SpaceRoll4] = 0
		default:
			b.Print()
			log.Panic("no available roll for move", from, to, player)
		}
	}
	return b
}

// Available returns legal moves available.
func (b Board) Available(player int) [][]int {
	barSpace := SpaceBarPlayer
	opponentBarSpace := SpaceBarOpponent
	if player == 2 {
		barSpace = SpaceBarOpponent
		opponentBarSpace = SpaceBarPlayer
	}
	mayBearOff := b.MayBearOff(player)
	onBar := b[barSpace] != 0
	var moves [][]int
	for from := 0; from < 28; from++ {
		if from == SpaceHomePlayer || from == SpaceHomeOpponent || from == opponentBarSpace || b.Checkers(from, player) == 0 || (onBar && from != barSpace) {
			continue
		}
		if player == 1 {
			for to := 0; to < from; to++ {
				if to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomeOpponent || (to == SpaceHomePlayer && !mayBearOff) {
					continue
				}
				v := b[to]
				if (player == 1 && v < -1) || (player == 2 && v > 1) || !b.HaveRoll(from, to, player) {
					continue
				}
				moves = append(moves, []int{from, to})
			}
		} else { // TODO clean up
			start := from + 1
			if from == SpaceBarOpponent {
				start = 0
			}
			for to := start; to <= 25; to++ {
				if to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomeOpponent || (to == SpaceHomeOpponent && !mayBearOff) {
					continue
				}
				v := b[to]
				if (player == 1 && v < -1) || (player == 2 && v > 1) || !b.HaveRoll(from, to, player) {
					continue
				}
				moves = append(moves, []int{from, to})
			}
		}
	}
	return moves
}

func (b Board) Pips(player int) int {
	var pips float64
	var spaceValue float64
	if player == 1 {
		pips += float64(b.Checkers(SpaceBarPlayer, player)) * 25
	} else {
		pips += float64(b.Checkers(SpaceBarOpponent, player)) * 25
	}
	for space := 1; space < 25; space++ {
		if player == 1 {
			spaceValue = float64(space)
			if space <= 6 {
				spaceValue /= 4
			} else {
				spaceValue += 6
			}
		} else {
			spaceValue = float64(25 - space)
			if space >= 19 {
				spaceValue /= 4
			} else {
				spaceValue += 6
			}
		}
		pips += float64(b.Checkers(space, player)) * spaceValue
	}
	return int(pips)
}

func (b Board) Blots(player int) int {
	var pips int
	var spaceValue int
	for space := 1; space < 25; space++ {
		checkers := b.Checkers(space, player)
		if checkers != 1 {
			continue
		}
		if player == 1 {
			spaceValue = 25 - space
		} else {
			spaceValue = space
		}
		pips += int(checkers) * spaceValue
	}
	return pips
}

func (b Board) Score(player int, hitScore int) float64 {
	pips := b.Pips(player)
	blots := b.Blots(player)
	return float64(pips) + float64(blots)*WeightBlot + float64(hitScore)*WeightHit
}

func (b Board) Evaluation(player int, hitScore int, moves [][]int) *Analysis {
	pips := b.Pips(player)
	blots := b.Blots(player)
	score := float64(pips) + float64(blots)*WeightBlot + float64(hitScore)*WeightHit
	return &Analysis{
		Board:       b,
		Moves:       moves,
		Pips:        pips,
		Blots:       blots,
		Hits:        hitScore,
		PlayerScore: score,
	}
}

func (b Board) _analyze(player int, hitScore int, available [][]int, moves [][]int) []*Analysis {
	m := &sync.Mutex{}
	w := &sync.WaitGroup{}
	var result []*Analysis
	for _, move := range available {
		if !b.HaveRoll(move[0], move[1], player) {
			log.Panic("NO ROLL", move[0], move[1], player, b)
		}
		move := move
		w.Add(1)
		go func() {
			var hs = hitScore
			var bc Board
			bc = b
			checkers := bc.Checkers(move[1], opponent(player))
			if checkers == 1 {
				if player == 1 {
					hs += move[1]
				} else {
					hs += 25 - move[1]
				}
			}
			bc = bc.Move(move[0], move[1], player)
			bc = bc.UseRoll(move[0], move[1], player)

			evaluation := bc.Evaluation(player, hs, append(append([][]int{}, moves...), move))
			subEvaluation := bc._analyze(player, hs, bc.Available(player), append(append([][]int{}, moves...), move))

			m.Lock()
			result = append(result, evaluation)
			result = append(result, subEvaluation...)
			m.Unlock()
			w.Done()
		}()
	}
	w.Wait()
	return result
}

func (b Board) Analyze(player int, available [][]int) []*Analysis {
	if len(available) == 0 {
		return nil
	}
	result := b._analyze(player, 0, available, nil)
	var maxMoves int
	for i := range result {
		l := len(result[i].Moves)
		if l > maxMoves {
			maxMoves = l
		}
	}
	var newResult []*Analysis
	for i := 0; i < len(result); i++ {
		if len(result[i].Moves) == maxMoves {
			newResult = append(newResult, result[i])
		}
	}
	result = newResult
	if player == 1 {
		m := &sync.Mutex{}
		for i := range result {
			var oppPips float64
			var oppBlots float64
			var oppHits float64
			var oppScore float64
			w := &sync.WaitGroup{}
			w.Add(6)
			for roll1 := int8(1); roll1 <= 6; roll1++ {
				roll1 := roll1
				go func() {
					for roll2 := int8(1); roll2 <= 6; roll2++ {
						bc := Board{}
						bc = result[i].Board
						bc[SpaceRoll1], bc[SpaceRoll2] = roll1, roll2
						if roll1 == roll2 {
							bc[SpaceRoll3], bc[SpaceRoll4] = roll1, roll2
						} else {
							bc[SpaceRoll3], bc[SpaceRoll4] = 0, 0
						}
						opponentAvailable := bc.Available(2)
						if len(opponentAvailable) == 0 {
							continue
						}
						result2 := bc._analyze(2, 0, opponentAvailable, nil)
						var averagePips float64
						var averageBlots float64
						var averageHits float64
						var averageScore float64
						for _, r := range result2 {
							averagePips += float64(r.Pips)
							averageBlots += float64(r.Blots)
							averageHits += float64(r.Hits)
							averageScore += r.PlayerScore
						}
						averagePips /= float64(len(result2))
						averageBlots /= float64(len(result2))
						averageHits /= float64(len(result2))
						averageScore /= float64(len(result2))
						m.Lock()
						oppPips += averagePips
						oppBlots += averageBlots
						oppHits += averageHits
						oppScore += averageScore
						m.Unlock()
					}
					w.Done()
				}()
			}
			w.Wait()
			result[i].OppPips = (oppPips / 36)
			result[i].OppBlots = (oppBlots / 36)
			result[i].OppHits = (oppHits / 36)
			result[i].OppScore = (oppScore / 36)
			result[i].Score = result[i].PlayerScore + result[i].OppScore*WeightOppScore
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Score < result[j].Score
	})
	return result
}

func (b Board) Print() {
	log.Printf("%+v", b)
}

func opponent(player int) int {
	if player == 1 {
		return 2
	} else {
		return 1
	}
}
func spaceDiff(from int, to int) int {
	if from < 0 || from > 27 || to < 0 || to > 27 {
		return 0
	} else if to == SpaceBarPlayer || to == SpaceBarOpponent {
		return 0
	} else if from == SpaceHomePlayer || from == SpaceHomeOpponent {
		return 0
	}

	if (from == SpaceBarPlayer || from == SpaceBarOpponent) && (to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomePlayer || to == SpaceHomeOpponent) {
		return 0
	}

	if from == SpaceBarPlayer {
		return 25 - to
	} else if from == SpaceBarOpponent {
		return to
	}

	if to == SpaceHomePlayer {
		return from
	} else if to == SpaceHomeOpponent {
		return 25 - from
	}

	diff := to - from
	if diff < 0 {
		return diff * -1
	}
	return diff
}
