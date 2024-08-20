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
	SpaceHomePlayer      int8 = 0
	SpaceHomeOpponent    int8 = 25
	SpaceBarPlayer       int8 = 26
	SpaceBarOpponent     int8 = 27
	SpaceRoll1           int8 = 28
	SpaceRoll2           int8 = 29
	SpaceRoll3           int8 = 30
	SpaceRoll4           int8 = 31
	SpaceEnteredPlayer   int8 = 32 // Whether the player has fully entered the board. Only used in acey-deucey games.
	SpaceEnteredOpponent int8 = 33 // Whether the opponent has fully entered the board. Only used in acey-deucey games.
	SpaceVariant         int8 = 34 // 0 - Backgammon, 1 - Acey-deucey, 2 - Tabula.
)

const (
	boardSpaces = 35
)

const (
	VariantBackgammon int8 = 0
	VariantAceyDeucey int8 = 1
	VariantTabula     int8 = 2
)

// Board represents the state of a game. It contains spaces for the checkers,
// as well as four "spaces" which contain the available die rolls.
type Board [boardSpaces]int8

// NewBoard returns a new board with checkers placed in their starting positions.
func NewBoard(variant int8) Board {
	if variant != VariantBackgammon {
		return Board{15, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, -15, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	}
	return Board{0, -2, 0, 0, 0, 0, 5, 0, 3, 0, 0, 0, -5, 5, 0, 0, 0, -3, 0, -5, 0, 0, 0, 0, 2, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0}
}

func (b Board) String() string {
	var board []byte
	for i, v := range b {
		if i != 0 {
			board = append(board, ',')
		}
		board = append(board, []byte(fmt.Sprintf("%d", v))...)
	}
	variant := "Backgammon"
	switch b[SpaceVariant] {
	case VariantAceyDeucey:
		variant = "Acey-deucey"
	case VariantTabula:
		variant = "Tabula"
	}
	entered1, entered2 := "Y", "Y"
	if b[SpaceEnteredPlayer] == 0 {
		entered1 = "N"
	}
	if b[SpaceEnteredOpponent] == 0 {
		entered2 = "N"
	}
	off1 := b[SpaceHomePlayer]
	off2 := b[SpaceHomeOpponent] * -1
	var rolls []byte
	if b[SpaceRoll1] != 0 {
		rolls = append(rolls, []byte(fmt.Sprintf("%d", b[SpaceRoll1]))...)
	}
	if b[SpaceRoll2] != 0 {
		if len(rolls) != 0 {
			rolls = append(rolls, ' ')
		}
		rolls = append(rolls, []byte(fmt.Sprintf("%d", b[SpaceRoll2]))...)
	}
	if b[SpaceRoll3] != 0 {
		if len(rolls) != 0 {
			rolls = append(rolls, ' ')
		}
		rolls = append(rolls, []byte(fmt.Sprintf("%d", b[SpaceRoll3]))...)
	}
	if b[SpaceRoll4] != 0 {
		if len(rolls) != 0 {
			rolls = append(rolls, ' ')
		}
		rolls = append(rolls, []byte(fmt.Sprintf("%d", b[SpaceRoll4]))...)
	}
	return fmt.Sprintf("Board: %s\nVariant: %s\nEntered: %s / %s\nOff: %d / %d\nRolls: %s", board, variant, entered1, entered2, off1, off2, rolls)
}

func (b Board) SetValue(space int, value int8) Board {
	b[space] = value
	return b
}

// Move moves a checker on the board.
func (b Board) Move(from int8, to int8, player int8) Board {
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
		delta = -1
	}
	b[from], b[to] = b[from]-delta, b[to]+delta
	if (player == 1 && from == SpaceHomePlayer && b[SpaceEnteredPlayer] == 0 && b[SpaceHomePlayer] == 0) || (player == 2 && from == SpaceHomeOpponent && b[SpaceEnteredOpponent] == 0 && b[SpaceHomeOpponent] == 0) {
		if player == 1 {
			b[SpaceEnteredPlayer] = 1
		} else {
			b[SpaceEnteredOpponent] = 1
		}
	}
	return b
}

// checkers returns the number of checkers that belong to the spcified player at the provided space.
func checkers(player int8, v int8) int8 {
	if player == 1 && v > 0 {
		return v
	} else if player == 2 && v < 0 {
		return v * -1
	}
	return 0
}

func (b Board) MayBearOff(player int8) bool {
	if b[SpaceVariant] != VariantBackgammon && ((player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0)) {
		return false
	} else if b[SpaceVariant] == VariantTabula && !b.SecondHalf(player) {
		return false
	}
	barSpace := SpaceBarPlayer
	if player == 2 {
		barSpace = SpaceBarOpponent
	}
	if checkers(player, b[barSpace]) != 0 {
		return false
	}
	if b[SpaceVariant] != VariantTabula {
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
	}
	return true
}

func (b Board) spaceDiff(player int8, from int8, to int8) int8 {
	switch {
	case from < 0 || from > 27 || to < 0 || to > 27:
		return 0
	case to == SpaceBarPlayer || to == SpaceBarOpponent:
		return 0
	case (from == SpaceHomePlayer || from == SpaceHomeOpponent || from == SpaceBarPlayer || from == SpaceBarOpponent) && (to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomePlayer || to == SpaceHomeOpponent):
		return 0
	case to == SpaceHomePlayer:
		if player == 2 {
			return 0
		}
		if b[SpaceVariant] == VariantTabula {
			if (player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0) || !b.SecondHalf(player) {
				return 0
			}
			return 25 - from
		}
		return from
	case to == SpaceHomeOpponent:
		if player == 1 {
			return 0
		}
		return 25 - from
	case from == SpaceHomePlayer || from == SpaceHomeOpponent:
		if (player == 1 && from == SpaceHomeOpponent) || (player == 2 && from == SpaceHomePlayer) {
			return 0
		}
		switch b[SpaceVariant] {
		case VariantAceyDeucey:
			if player == 1 && from == SpaceHomePlayer && b[SpaceEnteredPlayer] == 0 {
				return 25 - to
			} else if player == 2 && from == SpaceHomeOpponent && b[SpaceEnteredOpponent] == 0 {
				return to
			}
		case VariantTabula:
			if (player == 1 && from != SpaceHomePlayer && b[SpaceEnteredPlayer] == 0) || (player == 2 && from != SpaceHomeOpponent && b[SpaceEnteredOpponent] == 0) {
				return 0
			}
			return to
		}
		return 0
	case from == SpaceBarPlayer:
		if player == 2 {
			return 0
		}
		if b[SpaceVariant] == VariantTabula {
			return to
		}
		return 25 - to
	case from == SpaceBarOpponent:
		if player == 1 {
			return 0
		}
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
func (b Board) HaveRoll(from int8, to int8, player int8) bool {
	barSpace := SpaceBarPlayer
	if player == 2 {
		barSpace = SpaceBarOpponent
	}
	if b[barSpace] != 0 && from != barSpace {
		return false
	}

	if b[SpaceVariant] == VariantTabula && to > 12 && to < 25 && ((player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0)) {
		return false
	}

	delta := b.spaceDiff(player, from, to)
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
	if b.MayBearOff(player) && b[SpaceVariant] == VariantBackgammon {
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

// UseRoll uses a die roll. UseRoll must be called before making a move.
func (b Board) UseRoll(from int8, to int8, player int8) Board {
	delta := b.spaceDiff(player, from, to)
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
	if b.MayBearOff(player) && b[SpaceVariant] == VariantBackgammon {
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

func (b Board) _available(player int8) [][2]int8 {
	var bearOff int
	mayBearOff := func() bool {
		if bearOff != 0 {
			return bearOff == 1
		}
		if b.MayBearOff(player) {
			bearOff = 1
			return true
		}
		bearOff = 2
		return false
	}

	homeSpace := SpaceHomePlayer
	barSpace := SpaceBarPlayer
	opponentBarSpace := SpaceBarOpponent
	if player == 2 {
		homeSpace = SpaceHomeOpponent
		barSpace = SpaceBarOpponent
		opponentBarSpace = SpaceBarPlayer
	}
	onBar := b[barSpace] != 0

	var moves [][2]int8

	// Enter board from home space.
	if b[SpaceVariant] != VariantBackgammon && ((player == 1 && b[SpaceEnteredPlayer] == 0) || (player == 2 && b[SpaceEnteredOpponent] == 0)) && b[homeSpace] != 0 {
		for space := int8(1); space < 25; space++ {
			v := b[space]
			if ((player == 1 && v >= -1) || (player == 2 && v <= 1)) && b.HaveRoll(homeSpace, space, player) {
				moves = append(moves, [2]int8{homeSpace, space})
			}
		}
	}

	for from := int8(0); from < 28; from++ {
		// Skip invalid spaces, spaces without player checkers and non-bar spaces when a checker is on the bar.
		if from == SpaceHomePlayer || from == SpaceHomeOpponent || from == opponentBarSpace || checkers(player, b[from]) == 0 || (onBar && from != barSpace) {
			continue
		}

		// Iterate over destination spaces to determine available moves.
		if player == 1 && b[SpaceVariant] != VariantTabula {
			for to := int8(0); to < from; to++ {
				if to == SpaceBarPlayer || to == SpaceBarOpponent || to == SpaceHomeOpponent || (to == SpaceHomePlayer && !mayBearOff()) {
					continue
				}
				v := b[to]
				if (player == 1 && v < -1) || (player == 2 && v > 1) || !b.HaveRoll(from, to, player) {
					continue
				}
				moves = append(moves, [2]int8{from, to})
			}
		} else { // TODO clean up
			start := from + 1
			if from == SpaceBarPlayer || from == SpaceBarOpponent {
				start = 1
			}
			for i := start; i <= 25; i++ {
				to := i
				if player == 1 && to == SpaceHomeOpponent {
					to = SpaceHomePlayer
				} else if player == 2 && to == SpaceHomePlayer {
					to = SpaceHomeOpponent
				}
				if to == SpaceBarPlayer || to == SpaceBarOpponent || (((player == 1 && to == SpaceHomePlayer) || (player == 2 && to == SpaceHomeOpponent)) && !mayBearOff()) {
					continue
				}
				v := b[to]
				if (player == 1 && v < -1) || (player == 2 && v > 1) || !b.HaveRoll(from, to, player) {
					continue
				}
				moves = append(moves, [2]int8{from, to})
			}
		}
	}

	return moves
}

// Available returns legal moves available.
func (b Board) Available(player int8) ([][4][2]int8, []Board) {
	var allMoves [][4][2]int8

	movesFound := func(moves [4][2]int8) bool {
		for i := range allMoves {
			if movesEqual(allMoves[i], moves) {
				return true
			}
		}
		return false
	}

	var boards []Board
	a := b._available(player)
	mayBearOff := b.MayBearOff(player)
	maxLen := 1
	for _, move := range a {
		if (move[1] == SpaceHomePlayer || move[1] == SpaceHomeOpponent) && !mayBearOff {
			continue
		}
		newBoard := b.UseRoll(move[0], move[1], player).Move(move[0], move[1], player)
		newAvailable := newBoard._available(player)
		if len(newAvailable) == 0 {
			moves := [4][2]int8{move}
			if !movesFound(moves) {
				allMoves = append(allMoves, moves)
				boards = append(boards, newBoard)
			}
			continue
		}
		newBearOff := mayBearOff || newBoard.MayBearOff(player)
		for _, move2 := range newAvailable {
			if (move2[1] == SpaceHomePlayer || move2[1] == SpaceHomeOpponent) && !newBearOff {
				continue
			}
			newBoard2 := newBoard.UseRoll(move2[0], move2[1], player).Move(move2[0], move2[1], player)
			newAvailable2 := newBoard2._available(player)
			if len(newAvailable2) == 0 {
				moves := [4][2]int8{move, move2}
				if !movesFound(moves) {
					allMoves = append(allMoves, moves)
					boards = append(boards, newBoard2)
					if maxLen <= 2 {
						maxLen = 2
					}
				}
				continue
			}
			newBearOff2 := newBearOff || newBoard2.MayBearOff(player)
			for _, move3 := range newAvailable2 {
				if (move3[1] == SpaceHomePlayer || move3[1] == SpaceHomeOpponent) && !newBearOff2 {
					continue
				}
				newBoard3 := newBoard2.UseRoll(move3[0], move3[1], player).Move(move3[0], move3[1], player)
				newAvailable3 := newBoard3._available(player)
				if len(newAvailable3) == 0 {
					moves := [4][2]int8{move, move2, move3}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard3)
						if maxLen <= 2 {
							maxLen = 3
						}
					}
					continue
				}
				newBearOff3 := newBearOff2 || newBoard3.MayBearOff(player)
				for _, move4 := range newAvailable3 {
					if (move4[1] == SpaceHomePlayer || move4[1] == SpaceHomeOpponent) && !newBearOff3 {
						continue
					}
					newBoard4 := newBoard3.UseRoll(move4[0], move4[1], player).Move(move4[0], move4[1], player)
					moves := [4][2]int8{move, move2, move3, move4}
					if !movesFound(moves) {
						allMoves = append(allMoves, moves)
						boards = append(boards, newBoard4)
						maxLen = 4
					}
				}
			}
		}
	}
	var newMoves [][4][2]int8
	for i := 0; i < len(allMoves); i++ {
		l := 0
		if (allMoves[i][3][0] != 0 || allMoves[i][3][1] != 0) && allMoves[i][2][0] == 0 && allMoves[i][2][1] == 0 {
			allMoves[i][2][0], allMoves[i][2][1] = allMoves[i][3][0], allMoves[i][3][1]
			allMoves[i][2][0], allMoves[i][2][1] = 0, 0
		}
		if (allMoves[i][2][0] != 0 || allMoves[i][2][1] != 0) && allMoves[i][1][0] == 0 && allMoves[i][1][1] == 0 {
			allMoves[i][1][0], allMoves[i][1][1] = allMoves[i][2][0], allMoves[i][2][1]
			allMoves[i][2][0], allMoves[i][2][1] = 0, 0
		}
		if (allMoves[i][1][0] != 0 || allMoves[i][1][1] != 0) && allMoves[i][0][0] == 0 && allMoves[i][0][1] == 0 {
			allMoves[i][0][0], allMoves[i][0][1] = allMoves[i][1][0], allMoves[i][1][1]
			allMoves[i][1][0], allMoves[i][1][1] = 0, 0
		}
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
	if maxLen == 1 && len(newMoves) > 1 {
		moved := b[SpaceRoll1] == 0
		if !moved {
			moved = b[SpaceRoll2] == 0
		}
		if !moved && (b[SpaceVariant] == VariantTabula || b[SpaceRoll1] == b[SpaceRoll2]) {
			moved = b[SpaceRoll3] == 0
		}
		if !moved && b[SpaceVariant] != VariantTabula && b[SpaceRoll1] == b[SpaceRoll2] {
			moved = b[SpaceRoll4] == 0
		}
		if !moved {
			var highestRoll int8
			for _, move := range newMoves {
				roll := b.spaceDiff(player, move[0][0], move[0][1])
				if roll > highestRoll {
					highestRoll = roll
				}
			}
			var highRollMoves [][4][2]int8
			for _, move := range newMoves {
				roll := b.spaceDiff(player, move[0][0], move[0][1])
				if roll != highestRoll {
					continue
				}
				highRollMoves = append(highRollMoves, move)
			}
			newMoves = highRollMoves
		}
	}
	return newMoves, boards
}

func (b Board) FirstLast(player int8) (playerFirst int8, opponentLast int8) {
	playerFirst, opponentLast = -1, -1
	if b[SpaceBarPlayer] != 0 || b[SpaceBarOpponent] != 0 || b[SpaceVariant] == VariantTabula {
		return playerFirst, opponentLast
	} else if b[SpaceVariant] == VariantAceyDeucey && ((b[SpaceEnteredPlayer] == 0 && b[SpaceHomePlayer] != 0) || (b[SpaceEnteredOpponent] == 0 && b[SpaceHomeOpponent] != 0)) {
		return playerFirst, opponentLast
	}
	for space := int8(1); space < 25; space++ {
		v := b[space]
		if v == 0 {
			continue
		} else if v > 0 {
			if space > playerFirst {
				playerFirst = space
			}
		} else {
			if opponentLast == -1 {
				opponentLast = space
			}
		}
	}
	if player == 2 {
		return opponentLast, playerFirst
	}
	return playerFirst, opponentLast
}

func (b Board) Past() bool {
	playerFirst, opponentLast := b.FirstLast(1)
	if playerFirst == -1 || opponentLast == -1 {
		return false
	}
	return playerFirst < opponentLast
}

func (b Board) SecondHalf(player int8) bool {
	if b[SpaceVariant] != VariantTabula {
		return false
	}

	switch player {
	case 1:
		if b[SpaceBarPlayer] != 0 {
			return false
		} else if b[SpaceEnteredPlayer] == 0 && b[SpaceHomePlayer] != 0 {
			return false
		}
	case 2:
		if b[SpaceBarOpponent] != 0 {
			return false
		} else if b[SpaceEnteredOpponent] == 0 && b[SpaceHomeOpponent] != 0 {
			return false
		}
	default:
		log.Panicf("unknown player: %d", player)
	}

	for space := 1; space < 13; space++ {
		v := b[space]
		if (player == 1 && v > 0) || (player == 2 && v < 0) {
			return false
		}
	}

	return true
}

func (b Board) Pips(player int8) int {
	var pips int
	if b[SpaceVariant] != VariantBackgammon {
		if player == 1 && b[SpaceEnteredPlayer] == 0 {
			pips += int(checkers(player, b[SpaceHomePlayer])) * PseudoPips(player, SpaceHomePlayer, b[SpaceVariant])
		} else if player == 2 && b[SpaceEnteredOpponent] == 0 {
			pips += int(checkers(player, b[SpaceHomeOpponent])) * PseudoPips(player, SpaceHomeOpponent, b[SpaceVariant])
		}
	}
	if player == 1 {
		pips += int(checkers(player, b[SpaceBarPlayer])) * PseudoPips(player, SpaceBarPlayer, b[SpaceVariant])
	} else {
		pips += int(checkers(player, b[SpaceBarOpponent])) * PseudoPips(player, SpaceBarOpponent, b[SpaceVariant])
	}
	for space := int8(1); space < 25; space++ {
		pips += int(checkers(player, b[space])) * PseudoPips(player, space, b[SpaceVariant])
	}
	return pips
}

func (b Board) Blots(player int8) int {
	_, last := b.FirstLast(player)
	o := opponent(player)
	var pips int
	var pastBlots int
	var div int
	for space := int8(1); space < 25; space++ {
		if checkers(player, b[space]) == 1 {
			if last != -1 && ((player == 1 && space < last) || (player == 2 && space > last)) && pastBlots == 0 {
				pastBlots++
				div = 4
			} else {
				div = 1
			}
			v := PseudoPips(o, space, b[SpaceVariant]) / div
			if v < 1 {
				v = 1
			}
			pips += v
		}
	}
	return pips
}

func (b Board) evaluate(player int8, hitScore int, a *Analysis) {
	pips := b.Pips(player)
	score := float64(pips)
	blotWeight := WeightBlot
	if player == 1 {
		var blocks int8
		for space := 19; space <= 24; space++ {
			if checkers(2, b[space]) > 1 {
				blocks++
			}
		}
		switch blocks {
		case 6:
			blotWeight *= 1.5
		case 5:
			blotWeight *= 1.25
		case 4:
			blotWeight *= 1.1
		}
	}
	var blots int
	if !a.Past {
		blots = b.Blots(player)
		score += float64(blots)*blotWeight + float64(hitScore)*WeightHit
	}
	a.Pips = pips
	a.Blots = blots
	a.Hits = hitScore
	a.PlayerScore = score
	a.hitScore = hitScore
}

func (b Board) Evaluation(player int8, hitScore int, moves [4][2]int8) *Analysis {
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

func (b Board) Analyze(available [][4][2]int8, result *[]*Analysis, skipOpponent bool) {
	if len(available) == 0 {
		*result = (*result)[:0]
		return
	}
	const priorityScore = -1000000

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
			skipOpp:     skipOpponent,
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
		if a.player == 1 && !past && a.Past {
			a.Score += priorityScore
		}
	}

	if b[SpaceVariant] != VariantTabula && b.StartingPosition(1) {
		r1, r2 := b[SpaceRoll1], b[SpaceRoll2]
		if r2 > r1 {
			r1, r2 = r2, r1
		}
		var opening [4][2]int8
		if r1 == r2 {
			switch r1 {
			case 1:
				opening = [4][2]int8{{24, 23}, {24, 23}, {6, 5}, {6, 5}}
			case 2:
				opening = [4][2]int8{{13, 11}, {13, 11}, {11, 9}, {11, 9}}
			case 3:
				opening = [4][2]int8{{13, 10}, {13, 10}, {10, 7}, {10, 7}}
			case 4:
				opening = [4][2]int8{{13, 9}, {13, 9}, {6, 2}, {6, 2}}
			case 5:
				opening = [4][2]int8{{13, 8}, {13, 8}, {8, 3}, {8, 3}}
			case 6:
				opening = [4][2]int8{{24, 18}, {24, 18}, {13, 7}, {13, 7}}
			}
		} else {
			switch r1 {
			case 2:
				opening = [4][2]int8{{13, 11}, {6, 5}}
			case 3:
				switch r2 {
				case 1:
					opening = [4][2]int8{{8, 5}, {6, 5}}
				case 2:
					opening = [4][2]int8{{13, 11}, {13, 10}}
				}
			case 4:
				switch r2 {
				case 1:
					opening = [4][2]int8{{24, 23}, {13, 9}}
				case 2:
					opening = [4][2]int8{{8, 4}, {6, 4}}
				case 3:
					opening = [4][2]int8{{13, 10}, {13, 9}}
				}
			case 5:
				switch r2 {
				case 1:
					opening = [4][2]int8{{24, 23}, {13, 8}}
				case 2:
					opening = [4][2]int8{{24, 22}, {13, 8}}
				case 3:
					opening = [4][2]int8{{8, 3}, {6, 3}}
				case 4:
					opening = [4][2]int8{{24, 20}, {13, 8}}
				}
			case 6:
				switch r2 {
				case 1:
					opening = [4][2]int8{{13, 7}, {8, 7}}
				case 2:
					opening = [4][2]int8{{24, 18}, {13, 11}}
				case 3:
					opening = [4][2]int8{{24, 18}, {13, 10}}
				case 4:
					opening = [4][2]int8{{8, 2}, {6, 2}}
				case 5:
					opening = [4][2]int8{{24, 18}, {18, 13}}
				}
			}
		}
		for _, a := range *result {
			if movesEqual(a.Moves, opening) {
				a.Score = priorityScore
				break
			}
		}
	}

	sort.Slice(*result, func(i, j int) bool {
		return (*result)[i].Score < (*result)[j].Score
	})
}

func (b Board) StartingPosition(player int8) bool {
	if player == 1 {
		return b[6] == 5 && b[8] == 3 && b[13] == 5 && b[24] == 2
	}
	return b[1] == -2 && b[12] == -5 && b[17] == -3 && b[19] == -5
}

func (b Board) ChooseDoubles(result *[]*Analysis) int {
	if b[SpaceVariant] != VariantAceyDeucey {
		return 0
	}

	bestDoubles := 6
	bestScore := math.MaxFloat64

	var available [][4][2]int8
	for i := 0; i < 6; i++ {
		doubles := int8(i + 1)
		bc := b
		bc[SpaceRoll1], bc[SpaceRoll2], bc[SpaceRoll3], bc[SpaceRoll4] = doubles, doubles, doubles, doubles

		available, _ = bc.Available(1)
		bc.Analyze(available, result, true)
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

func opponent(player int8) int8 {
	if player == 1 {
		return 2
	}
	return 1
}

func spaceValue(player int8, space int8, variant int8) int {
	if space == SpaceHomePlayer || space == SpaceHomeOpponent || space == SpaceBarPlayer || space == SpaceBarOpponent {
		return 25
	} else if player == 1 || variant == VariantTabula {
		return int(space)
	} else {
		return int(25 - space)
	}
}

func PseudoPips(player int8, space int8, variant int8) int {
	v := 6 + spaceValue(player, space, variant) + int(math.Exp(float64(spaceValue(player, space, variant))*0.2))*2
	if space == SpaceHomePlayer || space == SpaceHomeOpponent || (variant == VariantTabula && space < 13) || (variant != VariantTabula && ((player == 1 && (space > 6 || space == SpaceBarPlayer)) || (player == 2 && (space < 19 || space == SpaceBarOpponent)))) {
		v += 24
	}
	return v
}

func movesEqual(a [4][2]int8, b [4][2]int8) bool {
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
