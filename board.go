package tabula

import (
	"log"
	"math"
	"sort"
	"sync"
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

// Board represents the state of a game. It contains spaces for the checkers,
// as well as four "spaces" which contain the available die rolls.
type Board [boardSpaces]int8

// NewBoard returns a new board with checkers placed in their starting positions.
func NewBoard() Board {
	return Board{0, -2, 0, 0, 0, 0, 5, 0, 3, 0, 0, 0, -5, 5, 0, 0, 0, -3, 0, -5, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0}
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

// Checkers returns the number of checkers at a space. It always returns a positive number.
func (b Board) Checkers(player int, space int) int {
	v := b[space]
	if player == 1 && v > 0 {
		return int(v)
	} else if player == 2 && v < 0 {
		return int(v * -1)
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
	if b.Checkers(player, barSpace) != 0 {
		return false
	}
	for space := 1; space < 25; space++ {
		if space >= homeStart && space <= homeEnd {
			continue
		} else if b.Checkers(player, space) != 0 {
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
			if b.Checkers(player, playerHomeEnd+int(checkSpace)*playerDelta) != 0 {
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
			if b.Checkers(player, playerHomeEnd+int(checkSpace)*playerDelta) != 0 {
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

func (b Board) _available(player int) [][]int {
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
		if from == SpaceHomePlayer || from == SpaceHomeOpponent || from == opponentBarSpace || b.Checkers(player, from) == 0 || (onBar && from != barSpace) {
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

// Available returns legal moves available.
func (b Board) Available(player int) ([][][]int, []Board) {
	var allMoves [][][]int

	resultMutex := &sync.Mutex{}
	movesFound := func(moves [][]int) bool {
		resultMutex.Lock()
		for _, f := range allMoves {
			if movesEqual(f, moves) {
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
			moves := [][]int{move}
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
				moves := [][]int{move, move2}
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
					moves := [][]int{move, move2, move3}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard3)
						maxLen = 3
					}
					continue
				}
				for _, move4 := range newAvailable3 {
					newBoard4 := newBoard3.Move(move4[0], move4[1], player).UseRoll(move4[0], move4[1], player)
					moves := [][]int{move, move2, move3, move4}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard4)
						maxLen = 4
					}
				}
			}
		}
	}
	var newMoves [][][]int
	for i := 0; i < len(allMoves); i++ {
		if len(allMoves[i]) >= maxLen {
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
	if player == 1 {
		pips += int(b.Checkers(player, SpaceBarPlayer)) * pseudoPips(player, SpaceBarPlayer)
	} else {
		pips += int(b.Checkers(player, SpaceBarOpponent)) * pseudoPips(player, SpaceBarOpponent)
	}
	for space := 1; space < 25; space++ {
		pips += int(b.Checkers(player, space)) * pseudoPips(player, space)
	}
	return pips
}

func (b Board) Blots(player int) int {
	o := opponent(player)
	var pips int
	for space := 1; space < 25; space++ {
		checkers := b.Checkers(player, space)
		if checkers != 1 {
			continue
		}
		pips += int(checkers) * pseudoPips(o, space)
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

func (b Board) Evaluation(player int, hitScore int, moves [][]int) *Analysis {
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

func (b Board) Analyze(available [][][]int) []*Analysis {
	if len(available) == 0 {
		return nil
	}

	const bufferSize = 128
	result := make([]*Analysis, 0, bufferSize)
	resultMutex := &sync.Mutex{}
	w := &sync.WaitGroup{}

	past := b.Past()
	w.Add(len(available))
	for _, moves := range available {
		a := &Analysis{
			Board:  b,
			Moves:  moves,
			Past:   past,
			player: 1,
			chance: 1,
		}
		go a._analyze(&result, resultMutex, w)
	}
	w.Wait()

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
	}
	return 1
}

func spaceValue(player int, space int) int {
	if space == SpaceBarPlayer || space == SpaceBarOpponent {
		return 25
	} else if player == 1 {
		return space
	} else {
		return 25 - space
	}
}

func pseudoPips(player int, space int) int {
	v := 6 + spaceValue(player, space) + int(math.Exp(float64(spaceValue(player, space))*0.2))*2
	if (player == 1 && (space > 6 || space == SpaceBarPlayer)) || (player == 2 && (space < 19 || space == SpaceBarOpponent)) {
		v += 36
	}
	return v
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

func movesEqual(a [][]int, b [][]int) bool {
	l := len(a)
	if len(b) != l {
		return false
	}
	for _, m := range a {
		switch m[0] {
		case SpaceBarPlayer, SpaceBarOpponent:
			return false
		}
		switch m[1] {
		case SpaceHomePlayer, SpaceHomeOpponent:
			return false
		}
	}
	switch l {
	case 0:
		return true
	case 1:
		return a[0][0] == b[0][0] && a[0][1] == b[0][1]
	case 2:
		return (a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1]) || // 1, 2
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1]) // 2, 1
	case 3:
		return (a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1]) || // 1, 2, 3
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1]) || // 2, 3, 1
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1]) || // 3, 1, 2
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1]) || // 1, 3, 2
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1]) || // 2, 1, 3
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1]) // 3, 2, 1
	case 4:
		return (a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 1,2,3,4
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 2,1,3,4
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 3,1,2,4
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 1,3,2,4
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 2,3,1,4
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[3][0] && a[3][1] == b[3][1]) || // 3,2,1,4
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 3,2,4,1
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 2,3,4,1
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 4,3,2,1
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 3,4,2,1
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 2,4,3,1
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[0][0] && a[3][1] == b[0][1]) || // 4,2,3,1
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 4,1,3,2
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[2][0] && a[2][1] == b[2][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 1,4,3,2
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 3,4,1,2
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 4,3,1,2
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[2][0] && a[1][1] == b[2][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 1,3,4,2
			(a[0][0] == b[2][0] && a[0][1] == b[2][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[1][0] && a[3][1] == b[1][1]) || // 3,1,4,2
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 2,1,4,3
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[3][0] && a[2][1] == b[3][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 1,2,4,3
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[1][0] && a[1][1] == b[1][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 4,2,1,3
			(a[0][0] == b[1][0] && a[0][1] == b[1][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[0][0] && a[2][1] == b[0][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 2,4,1,3
			(a[0][0] == b[0][0] && a[0][1] == b[0][1] && a[1][0] == b[3][0] && a[1][1] == b[3][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) || // 1,4,2,3
			(a[0][0] == b[3][0] && a[0][1] == b[3][1] && a[1][0] == b[0][0] && a[1][1] == b[0][1] && a[2][0] == b[1][0] && a[2][1] == b[1][1] && a[3][0] == b[2][0] && a[3][1] == b[2][1]) // 4,1,2,3
	default:
		log.Panicf("more than 4 moves were provided: %+v %+v", a, b)
		return false
	}
}
