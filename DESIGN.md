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

The pips weight is positive. The blots and hits weights are negative.

When past the opponent (there is no longer any chance of hitting the opponent)
the game state is scored as follows:

```
score = pips*pipsWeight
```

All scoring calculations use pseudopips. Space value is defined as what the
term 'pips' would normally refer to. The value of a space is the same as the
space number as it would appear to the player being scored.

A pseudopip value is assigned to each board space as follows:

- Each space is worth 6 pseudopips plus double the space value.
- Spaces outside of the player's home board are worth an additional 6 pseudopips.

Space 2 (from the perspective of the player being scored) is therefore worth 10
pseudopips, space 6 is worth 18, space 7 is worth 26 and the bar space is worth 62.

The base value of 6 incentivizes bearing pieces off instead of moving them
when possible. Adding double the space value incentivizes prioritizing moving
checkers that are further away from home.

Adding an additional 6 pseudopips to checkers outside of the player's home board
incentivizes moving all checkers into the home board before moving any checkers
within the home board.

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

### Step 1: Simulate all legal move available to the player

Copy the current gamestate and simulate each available combination of legal moves.
Combinations that are logically equal are skipped.

Score the resulting gamestate after making each combination of legal moves.
This is the player score.

### Step 2: Simulate all opponent dice rolls and opponent moves that may follow the above simulations

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
