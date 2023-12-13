package tabula

import (
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
)

var (
	AnalysisBufferSize    = 128
	SubAnalysisBufferSize = 3072
)

const (
	SpaceHomePlayer      = 0
	SpaceHomeOpponent    = 25
	SpaceBarPlayer       = 26
	SpaceBarOpponent     = 27
	SpaceRoll1           = 28
	SpaceRoll2           = 29
	SpaceRoll3           = 30
	SpaceRoll4           = 31
	SpaceEnteredPlayer   = 32 // Whether the player has fully entered the board. Only used in acey-deucey games.
	SpaceEnteredOpponent = 33 // Whether the opponent has fully entered the board. Only used in acey-deucey games.
	SpaceAcey            = 34 // 0 - Backgammon, 1 - Acey-deucey.
)

const (
	boardSpaces = 35
)

// Board represents the state of a game. It contains spaces for the checkers,
// as well as four "spaces" which contain the available die rolls.
type Board [boardSpaces]int8

// NewBoard returns a new board with checkers placed in their starting positions.
func NewBoard(acey bool) Board {
	if acey {
		return Board{15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -15, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	}
	return Board{0, -2, 0, 0, 0, 0, 5, 0, 3, 0, 0, 0, -5, 5, 0, 0, 0, -3, 0, -5, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0}
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
				b[SpaceBarOpponent]--
			} else {
				b[SpaceBarPlayer]++
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

// checkers returns the number of checkers that belong to the spcified player at the provided space.
func checkers(player int, v int8) int8 {
	if player == 1 && v > 0 {
		return v
	} else if player == 2 && v < 0 {
		return v * -1
	}
	return 0
}

func (b Board) MayBearOff(player int) bool {
	if b[SpaceAcey] == 1 && ((player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0)) {
		return false
	}
	barSpace := SpaceBarPlayer
	if player == 2 {
		barSpace = SpaceBarOpponent
	}
	if checkers(player, b[barSpace]) != 0 {
		return false
	}
	if player == 1 {
		for space := 24; space > 6; space-- {
			if checkers(player, b[space]) != 0 {
				return false
			}
		}
	} else {
		for space := 1; space < 19; space++ {
			if checkers(player, b[space]) != 0 {
				return false
			}
		}
	}
	return true
}

func (b Board) spaceDiff(player int, from int, to int) int {
	switch {
	case from < 0 || from > 27 || to < 0 || to > 27:
		return 0
	case to == SpaceBarPlayer || to == SpaceBarOpponent:
		return 0
	case (from == SpaceBarPlayer || from == SpaceBarOpponent) && (to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomePlayer || to == SpaceHomeOpponent):
		return 0
	case to == SpaceHomePlayer:
		return from
	case to == SpaceHomeOpponent:
		return 25 - from
	case from == SpaceHomePlayer || from == SpaceHomeOpponent:
		if b[SpaceAcey] == 1 {
			if player == 1 && from == SpaceHomePlayer && b[SpaceEnteredPlayer] == 0 {
				return 25 - to
			} else if player == 2 && from == SpaceHomeOpponent && b[SpaceEnteredOpponent] == 0 {
				return to
			}
		}
		return 0
	case from == SpaceBarPlayer:
		return 25 - to
	case from == SpaceBarOpponent:
		return to
	default:
		diff := to - from
		if diff < 0 {
			return diff * -1
		}
		return diff
	}
}

// HaveRoll returns whether the player has a sufficient die roll for the specified move.
func (b Board) HaveRoll(from int, to int, player int) bool {
	barSpace := SpaceBarPlayer
	if player == 2 {
		barSpace = SpaceBarOpponent
	}
	if b[barSpace] != 0 && from != barSpace {
		return false
	}

	delta := int8(b.spaceDiff(player, from, to))
	if delta == 0 {
		return false
	}

	if b[SpaceRoll1] == delta || b[SpaceRoll2] == delta || b[SpaceRoll3] == delta || b[SpaceRoll4] == delta {
		return true
	}

	playerDelta := -1
	playerHomeEnd := 6
	if player == 2 {
		playerDelta = 1
		playerHomeEnd = 19
	}
	if b.MayBearOff(player) && b[SpaceAcey] == 0 {
		allowGreater := true
		for checkSpace := 0; checkSpace < 6-int(delta); checkSpace++ {
			if checkers(player, b[playerHomeEnd+checkSpace*playerDelta]) != 0 {
				allowGreater = false
				break
			}
		}
		if allowGreater {
			return (b[SpaceRoll1] >= delta || b[SpaceRoll2] >= delta || b[SpaceRoll3] >= delta || b[SpaceRoll4] >= delta)
		}
	}
	return false
}

// UseRoll uses a die roll.
func (b Board) UseRoll(from int, to int, player int) Board {
	delta := int8(b.spaceDiff(player, from, to))
	if delta == 0 {
		b.Print()
		log.Panic("unknown space diff", from, to, player)
	}

	switch {
	case b[SpaceRoll1] == delta:
		b[SpaceRoll1] = 0
		return b
	case b[SpaceRoll2] == delta:
		b[SpaceRoll2] = 0
		return b
	case b[SpaceRoll3] == delta:
		b[SpaceRoll3] = 0
		return b
	case b[SpaceRoll4] == delta:
		b[SpaceRoll4] = 0
		return b
	}

	playerDelta := -1
	playerHomeEnd := 6
	if player == 2 {
		playerDelta = 1
		playerHomeEnd = 19
	}
	var allowGreater bool
	if b.MayBearOff(player) && b[SpaceAcey] == 0 {
		allowGreater = true
		for checkSpace := int8(0); checkSpace < 6-delta; checkSpace++ {
			if checkers(player, b[playerHomeEnd+int(checkSpace)*playerDelta]) != 0 {
				allowGreater = false
				break
			}
		}
	}
	if !allowGreater {
		b.Print()
		log.Panic(fmt.Sprint(b), "no available roll for move", from, to, player, delta)
	}

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
		log.Panic(fmt.Sprint(b), "no available roll for move", from, to, player, delta)
	}
	return b
}

func (b Board) _available(player int) [][2]int {
	homeSpace := SpaceHomePlayer
	barSpace := SpaceBarPlayer
	opponentBarSpace := SpaceBarOpponent
	if player == 2 {
		homeSpace = SpaceHomeOpponent
		barSpace = SpaceBarOpponent
		opponentBarSpace = SpaceBarPlayer
	}
	mayBearOff := b.MayBearOff(player)
	onBar := b[barSpace] != 0

	var moves [][2]int

	if b[SpaceAcey] == 1 && ((player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0)) && b[homeSpace] != 0 {
		for space := 1; space < 25; space++ {
			v := b[space]
			if ((player == 1 && v >= -1) || (player == 2 && v <= 1)) && b.HaveRoll(homeSpace, space, player) {
				moves = append(moves, [2]int{homeSpace, space})
			}
		}
	}

	for from := 0; from < 28; from++ {
		if from == SpaceHomePlayer || from == SpaceHomeOpponent || from == opponentBarSpace || checkers(player, b[from]) == 0 || (onBar && from != barSpace) {
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
				moves = append(moves, [2]int{from, to})
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
				moves = append(moves, [2]int{from, to})
			}
		}
	}

	return moves
}

// Available returns legal moves available.
func (b Board) Available(player int) ([][4][2]int, []Board) {
	var allMoves [][4][2]int

	resultMutex := &sync.Mutex{}
	movesFound := func(moves [4][2]int) bool {
		resultMutex.Lock()
		for i := range allMoves {
			if movesEqual(allMoves[i], moves) {
				resultMutex.Unlock()
				return true
			}
		}
		resultMutex.Unlock()
		return false
	}

	var boards []Board
	a := b._available(player)
	maxLen := 1
	for _, move := range a {
		newBoard := b.Move(move[0], move[1], player).UseRoll(move[0], move[1], player)
		newAvailable := newBoard._available(player)
		if len(newAvailable) == 0 {
			moves := [4][2]int{move}
			if !movesFound(moves) {
				allMoves = append(allMoves, moves)
				boards = append(boards, newBoard)
			}
			continue
		}
		for _, move2 := range newAvailable {
			newBoard2 := newBoard.Move(move2[0], move2[1], player).UseRoll(move2[0], move2[1], player)
			newAvailable2 := newBoard2._available(player)
			if len(newAvailable2) == 0 {
				moves := [4][2]int{move, move2}
				if !movesFound(moves) {
					allMoves = append(allMoves, moves)
					boards = append(boards, newBoard2)
					maxLen = 2
				}
				continue
			}
			for _, move3 := range newAvailable2 {
				newBoard3 := newBoard2.Move(move3[0], move3[1], player).UseRoll(move3[0], move3[1], player)
				newAvailable3 := newBoard3._available(player)
				if len(newAvailable3) == 0 {
					moves := [4][2]int{move, move2, move3}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard3)
						maxLen = 3
					}
					continue
				}
				for _, move4 := range newAvailable3 {
					newBoard4 := newBoard3.Move(move4[0], move4[1], player).UseRoll(move4[0], move4[1], player)
					moves := [4][2]int{move, move2, move3, move4}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard4)
						maxLen = 4
					}
				}
			}
		}
	}
	var newMoves [][4][2]int
	for i := 0; i < len(allMoves); i++ {
		l := 0
		for j := 0; j < 4; j++ {
			if allMoves[i][j][0] == 0 && allMoves[i][j][1] == 0 {
				break
			}
			l = j + 1
		}
		if l >= maxLen {
			newMoves = append(newMoves, allMoves[i])
		}
	}
	return newMoves, boards
}

func (b Board) Past() bool {
	if b[SpaceBarPlayer] != 0 || b[SpaceBarOpponent] != 0 {
		return false
	}
	var playerFirst, opponentLast int
	for space := 1; space < 25; space++ {
		v := b[space]
		if v == 0 {
			continue
		} else if v > 0 {
			if space > playerFirst {
				playerFirst = space
			}
		} else {
			if opponentLast == 0 {
				opponentLast = space
			}
		}
	}
	return playerFirst < opponentLast
}

func (b Board) Pips(player int) int {
	var pips int
	if b[SpaceAcey] == 1 {
		if player == 1 && b[SpaceEnteredPlayer] == 0 {
			pips += int(checkers(player, b[SpaceHomePlayer])) * PseudoPips(player, SpaceHomePlayer)
		} else if player == 2 && b[SpaceEnteredOpponent] == 0 {
			pips += int(checkers(player, b[SpaceHomeOpponent])) * PseudoPips(player, SpaceHomeOpponent)
		}
	}
	if player == 1 {
		pips += int(checkers(player, b[SpaceBarPlayer])) * PseudoPips(player, SpaceBarPlayer)
	} else {
		pips += int(checkers(player, b[SpaceBarOpponent])) * PseudoPips(player, SpaceBarOpponent)
	}
	for space := 1; space < 25; space++ {
		pips += int(checkers(player, b[space])) * PseudoPips(player, space)
	}
	return pips
}

func (b Board) Blots(player int) int {
	o := opponent(player)
	var pips int
	for space := 1; space < 25; space++ {
		if checkers(player, b[space]) == 1 {
			pips += PseudoPips(o, space)
		}
	}
	return pips
}

func (b Board) evaluate(player int, hitScore int, a *Analysis) {
	pips := b.Pips(player)
	score := float64(pips)
	var blots int
	if !a.Past {
		blots = b.Blots(player)
		score += float64(blots)*WeightBlot + float64(hitScore)*WeightHit
	}
	a.Pips = pips
	a.Blots = blots
	a.Hits = hitScore
	a.PlayerScore = score
	a.hitScore = hitScore
}

func (b Board) Evaluation(player int, hitScore int, moves [4][2]int) *Analysis {
	a := &Analysis{
		Board:  b,
		Moves:  moves,
		Past:   b.Past(),
		player: player,
		chance: 1,
	}
	b.evaluate(player, hitScore, a)
	return a
}

func (b Board) Analyze(available [][4][2]int, result *[]*Analysis) {
	if len(available) == 0 {
		*result = (*result)[:0]
		return
	}

	var reuse []*[]*Analysis
	for _, r := range *result {
		if r.result != nil {
			reuse = append(reuse, r.result)
		}
	}
	*result = (*result)[:0]
	reuseLen := len(reuse)
	var reuseIndex int

	w := &sync.WaitGroup{}

	past := b.Past()
	w.Add(len(available))
	for _, moves := range available {
		var r *[]*Analysis
		if reuseIndex < reuseLen {
			r = reuse[reuseIndex]
			*r = (*r)[:0]
			reuseIndex++
		} else {
			v := make([]*Analysis, 0, SubAnalysisBufferSize)
			r = &v
		}
		a := &Analysis{
			Board:       b,
			Moves:       moves,
			Past:        past,
			player:      1,
			chance:      1,
			result:      r,
			resultMutex: &sync.Mutex{},
			wg:          w,
		}
		*result = append(*result, a)
		analysisQueue <- a
	}
	w.Wait()

	for _, a := range *result {
		if a.player == 1 && !a.Past {
			var oppPips float64
			var oppBlots float64
			var oppHits float64
			var oppScore float64
			var count float64
			for _, r := range *a.result {
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
	}

	sort.Slice(*result, func(i, j int) bool {
		return (*result)[i].Score < (*result)[j].Score
	})
}

func (b Board) ChooseDoubles(result *[]*Analysis) int {
	if b[SpaceAcey] == 0 {
		return 0
	}

	bestDoubles := 6
	bestScore := math.MaxFloat64

	var available [][4][2]int
	for i := 0; i < 6; i++ {
		doubles := int8(i + 1)
		bc := b
		bc[SpaceRoll1], bc[SpaceRoll2], bc[SpaceRoll3], bc[SpaceRoll4] = doubles, doubles, doubles, doubles

		available, _ = bc.Available(1)
		bc.Analyze(available, result)
		if len(*result) > 0 && (*result)[0].Score < bestScore {
			bestDoubles = i + 1
			bestScore = (*result)[0].Score
		}
	}

	return bestDoubles
}

func (b Board) Print() {
	log.Printf("%+v", b)
}

func opponent(player int) int {
	if player == 1 {
		return 2
	}
	return 1
}

func spaceValue(player int, space int) int {
	if space == SpaceHomePlayer || space == SpaceHomeOpponent || space == SpaceBarPlayer || space == SpaceBarOpponent {
		return 25
	} else if player == 1 {
		return space
	} else {
		return 25 - space
	}
}

func PseudoPips(player int, space int) int {
	v := 6 + spaceValue(player, space) + int(math.Exp(float64(spaceValue(player, space))*0.2))*2
	if space == SpaceHomePlayer || space == SpaceHomeOpponent || (player == 1 && (space > 6 || space == SpaceBarPlayer)) || (player == 2 && (space < 19 || space == SpaceBarOpponent)) {
		v += 24
	}
	return v
}

func movesEqual(a [4][2]int, b [4][2]int) bool {
	if a[0][0] == b[0][0] && a[0][1] == b[0][1] { // 1
		if a[1][0] == b[1][0] && a[1][1] == b[1][1] { // 2
			if (a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 3,4
				(a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) { // 4,3
				return true
			}
		}
		if a[1][0] == b[2][0] && a[1][1] == b[2][1] { // 3
			if (a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 2,4
				(a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) { // 4,2
				return true
			}
		}
		if a[1][0] == b[3][0] && a[1][1] == b[3][1] { // 4
			if (a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 3,2
				(a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) { // 2,3
				return true
			}
		}
	}
	if a[0][0] == b[1][0] && a[0][1] == b[1][1] { // 2
		if a[1][0] == b[0][0] && a[1][1] == b[0][1] { // 1
			if (a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 3,4
				(a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) { // 4,3
				return true
			}
		}
		if a[1][0] == b[2][0] && a[1][1] == b[2][1] { // 3
			if (a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 4,1
				(a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) { // 1,4
				return true
			}
		}
		if a[1][0] == b[3][0] && a[1][1] == b[3][1] { // 4
			if (a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 3,1
				(a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) { // 1,3
				return true
			}
		}
	}
	if a[0][0] == b[2][0] && a[0][1] == b[2][1] { // 3
		if a[1][0] == b[0][0] && a[1][1] == b[0][1] { // 1
			if (a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 2,4
				(a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) { // 4,2
				return true
			}
		}
		if a[1][0] == b[1][0] && a[1][1] == b[1][1] { // 2
			if (a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 1,4
				(a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) { // 4,1
				return true
			}
		}
		if a[1][0] == b[3][0] && a[1][1] == b[3][1] { // 4
			if (a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 2,1
				(a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) { // 1,2
				return true
			}
		}
	}
	if a[0][0] == b[3][0] && a[0][1] == b[3][1] { // 4
		if a[1][0] == b[0][0] && a[1][1] == b[0][1] { // 1
			if (a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 3,2
				(a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) { // 2,3
				return true
			}
		}
		if a[1][0] == b[1][0] && a[1][1] == b[1][1] { // 2
			if (a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 1,3
				(a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) { // 3,1
				return true
			}
		}
		if a[1][0] == b[2][0] && a[1][1] == b[2][1] { // 3
			if (a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 1,2
				(a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) { // 2,1
				return true
			}
		}
	}
	return false
}
