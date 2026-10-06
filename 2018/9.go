// Day 9: Marble Mania
//
// The input is "N players; last marble is worth M points". Marbles 1..M are
// placed in a circle in turn: normally between the marbles one and two
// places clockwise of the current one. A marble that is a multiple of 23 is
// kept instead, together with the marble seven places counter-clockwise,
// which is removed; both score for the player.
//
// Part 1 is the winning score.
// Part 2 is the winning score with the last marble 100 times larger.
//
// The circle is a doubly linked list held in two arrays, next and prev,
// indexed by marble number, so every move is a few array writes.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 10
	maxPlayers    = 1_000
	maxMarble     = 100_000_000
)

// readGame returns the number of players and the last marble.
func readGame(path string) (players, last int, err error) {
	if path == "" {
		return 0, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	w := strings.Fields(string(content))
	if len(w) != 8 {
		return 0, 0, fmt.Errorf("%s: malformed game description", path)
	}
	if players, err = strconv.Atoi(w[0]); err != nil || players < 1 || players > maxPlayers {
		return 0, 0, fmt.Errorf("%s: invalid player count %q", path, w[0])
	}
	if last, err = strconv.Atoi(w[6]); err != nil || last < 1 || last > maxMarble {
		return 0, 0, fmt.Errorf("%s: invalid last marble %q", path, w[6])
	}
	return players, last, nil
}

// play runs the game and returns the winning score. next, prev and scores
// are scratch space.
func play(players, last int, next, prev []int32, scores []int) (int, error) {
	if len(next) <= last || len(prev) <= last || len(scores) < players || players < 1 {
		return 0, errors.New("scratch space is too small")
	}
	clear(scores[:players])
	next[0], prev[0] = 0, 0
	current := int32(0)
	for m := 1; m <= last; m++ {
		if m%23 == 0 {
			for range 7 {
				current = prev[current]
			}
			scores[m%players] += m + int(current)
			n, p := next[current], prev[current]
			next[p], prev[n] = n, p
			current = n
			continue
		}
		a := next[current] // the new marble goes between a and next[a]
		b := next[a]
		next[a], prev[b] = int32(m), int32(m)
		next[m], prev[m] = b, a
		current = int32(m)
	}
	best := 0
	for _, s := range scores[:players] {
		best = max(best, s)
	}
	if best < 0 {
		return 0, fmt.Errorf("winning score %d is negative", best)
	}
	return best, nil
}

func day9(path string) error {
	players, last, err := readGame(path)
	if err != nil {
		return err
	}
	if 100*last > maxMarble {
		return fmt.Errorf("last marble %d is too large for part 2", last)
	}
	next, prev := make([]int32, 100*last+1), make([]int32, 100*last+1)
	scores := make([]int, players)
	score, err := play(players, last, next, prev, scores)
	if err != nil {
		return err
	}
	longer, err := play(players, 100*last, next, prev, scores)
	if err != nil {
		return err
	}

	if longer < score {
		return fmt.Errorf("the longer game scored less (%d < %d)", longer, score)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", score, longer); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
