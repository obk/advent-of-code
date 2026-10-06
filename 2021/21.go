// Day 21: Dirac Dice
//
// The input is both players' starting spaces on a ring of 1-10. On a turn
// a player rolls three times, moves round the ring by the total, and adds
// the space landed on to their score.
//
// Part 1 uses a die rolling 1, 2, ... 100 in turn; the game ends at 1000
// points, and the answer is the loser's score times the number of rolls.
// Part 2 uses a three-sided die that splits the universe on every roll;
// the game ends at 21. The answer is the number of universes the more
// successful player wins in.
//
// For part 2, a state is (mover's space and score, other's space and
// score). Every move raises the mover's score, so states are filled from
// the highest combined score down, each from states already done. The 27
// outcomes of three rolls give totals 3-9 with weights 1, 3, 6, 7, 6, 3, 1.
package main

import (
	"errors"
	"fmt"
	"os"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 7
	spaces        = 10
	goal          = 21
)

// practice plays the deterministic game.
func practice(p1, p2 int) (int, error) {
	if p1 < 1 || p1 > spaces || p2 < 1 || p2 > spaces {
		return 0, fmt.Errorf("starting spaces %d and %d are not 1-10", p1, p2)
	}
	pos, score := [2]int{p1, p2}, [2]int{}
	rolls := 0
	for turn := 0; turn < 2_000; turn++ {
		p := turn % 2
		move := (rolls%100 + 1) + ((rolls+1)%100 + 1) + ((rolls+2)%100 + 1) // the next three rolls
		rolls += 3
		pos[p] = (pos[p]+move-1)%spaces + 1
		score[p] += pos[p]
		if score[p] >= 1000 {
			return score[1-p] * rolls, nil
		}
	}
	return 0, errors.New("nobody reached 1000")
}

// dirac counts the universes each player wins in.
func dirac(p1, p2 int) (int, error) {
	if p1 < 1 || p1 > spaces || p2 < 1 || p2 > spaces {
		return 0, fmt.Errorf("starting spaces %d and %d are not 1-10", p1, p2)
	}
	weights := [10]int{0, 0, 0, 1, 3, 6, 7, 6, 3, 1}
	// wins[p][s][q][t] = universes won by (mover, other) from that state.
	var wins [spaces + 1][goal][spaces + 1][goal][2]int
	for total := 2*goal - 2; total >= 0; total-- {
		for s := max(0, total-goal+1); s <= min(total, goal-1); s++ {
			t := total - s
			for p := 1; p <= spaces; p++ {
				for q := 1; q <= spaces; q++ {
					var w [2]int
					for roll := 3; roll <= 9; roll++ {
						np := (p+roll-1)%spaces + 1
						if s+np >= goal {
							w[0] += weights[roll]
							continue
						}
						// The other player moves next: their wins are ours swapped.
						after := wins[q][t][np][s+np]
						w[0] += weights[roll] * after[1]
						w[1] += weights[roll] * after[0]
					}
					wins[p][s][q][t] = w
				}
			}
		}
	}
	w := wins[p1][0][p2][0]
	if w[0]+w[1] < 1 {
		return 0, errors.New("nobody wins in any universe")
	}
	if w[0] < 0 || w[1] < 0 {
		return 0, errors.New("the universe count overflowed")
	}
	return max(w[0], w[1]), nil
}

func day21(path string) error {
	if path == "" {
		return errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	var p1, p2 int
	n, err := fmt.Sscanf(string(content), "Player 1 starting position: %d\nPlayer 2 starting position: %d", &p1, &p2)
	if err != nil || n != 2 {
		return fmt.Errorf("%s: malformed starting positions", path)
	}
	first, err := practice(p1, p2)
	if err != nil {
		return err
	}
	most, err := dirac(p1, p2)
	if err != nil {
		return err
	}

	if first < 1 {
		return fmt.Errorf("practice result %d is not positive", first)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, most); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
