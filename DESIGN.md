# Design

[tabula](https://code.rocket9labs.com/tslocum/tabula) is a multi-threaded
backgammon analysis engine. To find the best move, the engine performs a series
of simulations and scores the resulting game states. The combination that
results in the lowest scoring game state is the best move.

## Scoring

The score of each game state is comprised of weighted calculations.

Each game state is initially scored as follows:

```
score = pips*pipsWeight + blots*blotsWeight + hits*hitsWeight
```

The pips and blots weights are positive. The hits weight is negative.

When past the opponent (there is no longer any chance of hitting the opponent)
the game state is scored as follows:

```
score = pips*pipsWeight
```

All scoring calculations use pseudopips.

Space value is defined as what the term 'pips' would normally refer to for a
given space. The value of a space is the same as the space number as it appears
to the player being scored.

Base value is 6 for spaces within the home board of the player being scored.
All other spaces have a base value of 42. The base values incentivize
prioritizing moving all checkers into the player's home board, and subsequently
bearing checkers off instead of moving them.

The pseudopip value of each space is calculated as follows:

```
pseudoPips = baseValue(space) + spaceValue(space) + exp(spaceValue(space)*0.2)*2
```

### Pips

Pips are any player checkers that are on the board.

### Blots

Blots are single checkers that could be hit by the opponent during this turn or
any other turn.

### Hits

Hits are single checkers that may be hit by the player during this turn using
the available dice rolls.

## Analysis

Analysis is performed in parallel, utilizing all available CPU cores.

### Step 1: Simulate all legal moves available to the player

Copy the current gamestate and simulate each available combination of legal moves.
Combinations that are logically equal are skipped.

Score the resulting gamestate after making each combination of legal moves.
This is the player score.

### Step 2: Simulate all possible opponent dice rolls and opponent moves following the above simulations

Copy each resulting gamestate from step one and simulate all 21 possible dice
roll combinations and resulting legal moves.

Score the resulting gamestate after making each combination of legal moves.
Average out all of the scores. This is the opponent score.

### Step 3: Sort simulation results by score

For each legal player move combination, the overall score is calculated as follows:

```
score = playerScore + opponentScore*opponentScoreWeight
```

The opponent score weight is negative.

When past the opponent (there is no longer any chance of hitting the opponent)
the overall score is calculated as follows:

```
score = playerScore
```

Each combination is sorted by its overall score. The combination with the
lowest overall score is the best move.

## Pseudopip values

The following table lists the pseudopip value of each space. Space 25 is the bar.

| Space | Pseudopips |
| --- | --- |
| 1 | 9 |
| 2 | 10 |
| 3 | 11 |
| 4 | 14 |
| 5 | 15 |
| 6 | 18 |
| 7 | 57 |
| 8 | 58 |
| 9 | 63 |
| 10 | 66 |
| 11 | 71 |
| 12 | 76 |
| 13 | 81 |
| 14 | 88 |
| 15 | 97 |
| 16 | 106 |
| 17 | 117 |
| 18 | 132 |
| 19 | 149 |
| 20 | 170 |
| 21 | 195 |
| 22 | 226 |
| 23 | 263 |
| 24 | 308 |
| 25 | 363 |
