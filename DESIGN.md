# Design

[tabula](https://code.rocket9labs.com/tslocum/tabula) is a multi-threaded backgammon analysis engine.
To find the best move, the engine performs a series of simulations and scores the resulting game states.
The result with the lowest score is the best move.

## Scoring

### Space value

Space value is defined as what the term 'pips' would normally refer to. The
value of a space is the same as the space number as it would appear to the
player being scored.

### Pseudopips

All scoring calculations use pseudopips.

A pseudopip value is assigned to each board space as follows:

- Each space is worth 6 pseudopips plus **double the space value**.
- Spaces outside of the player's home board are worth an additional 6 pseudopips.

Space 2 (from the perspective of the player being scored) is therefore worth 10
pseudopips, space 6 is worth 18, space 7 is worth 26 and the bar space is worth 62.

## Analysis

### Step 1: Simulate all legal move available to the player

### Step 2: Simulate all opponent dice rolls and opponent moves that may follow the above simulations

### Step 3: Sort simulation results by score
