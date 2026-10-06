// Day 4: Giant Squid
//
// The input is the numbers to draw, then 5x5 bingo boards. A board wins
// when a whole row or column has been drawn.
//
// Part 1 is the first winning board's score: the sum of its undrawn
// numbers times the number just drawn.
// Part 2 is the score of the board that wins last.
//
// The turn each number is drawn on is looked up in a table; a line wins
// on the latest turn among its numbers, and a board on its earliest line.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 16
	maxNumber     = 1_000
	size          = 5
	never         = maxNumber + 1
)

// board is a bingo card.
type board [size][size]int

// winTurn returns the turn the board wins on (never if it doesn't).
func (b *board) winTurn(turn *[maxNumber]int) (int, error) {
	if turn == nil {
		return 0, errors.New("no draw table")
	}
	best := never
	for k := range size {
		row, col := 0, 0
		for j := range size {
			row, col = max(row, turn[b[k][j]]), max(col, turn[b[j][k]])
		}
		best = min(best, row, col)
	}
	return best, nil
}

// score sums the numbers not drawn by the given turn, times the last draw.
func (b *board) score(turn *[maxNumber]int, upTo, last int) (int, error) {
	if upTo < 0 || upTo >= never || last < 0 {
		return 0, fmt.Errorf("board never won (turn %d)", upTo)
	}
	sum := 0
	for _, row := range b {
		for _, v := range row {
			if turn[v] > upTo {
				sum += v
			}
		}
	}
	return sum * last, nil
}

// load parses the draws (into a turn table) and the boards.
func load(path string, turn *[maxNumber]int) ([]int, []board, error) {
	if path == "" {
		return nil, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	parts := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	draws := strings.Split(parts[0], ",")
	if len(parts) < 2 || len(draws) > maxNumber {
		return nil, nil, fmt.Errorf("%s: no boards, or too many draws", path)
	}
	for i := range turn {
		turn[i] = never
	}
	order := make([]int, len(draws))
	for i, d := range draws {
		v, err := strconv.Atoi(d)
		if err != nil || v < 0 || v >= maxNumber || turn[v] != never {
			return nil, nil, fmt.Errorf("%s: invalid or repeated draw %q", path, d)
		}
		turn[v], order[i] = i, v
	}
	boards := make([]board, len(parts)-1)
	for b, text := range parts[1:] {
		f := strings.Fields(text)
		if len(f) != size*size {
			return nil, nil, fmt.Errorf("%s: board %d has %d numbers", path, b, len(f))
		}
		for k, s := range f {
			v, err := strconv.Atoi(s)
			if err != nil || v < 0 || v >= maxNumber {
				return nil, nil, fmt.Errorf("%s: invalid number %q", path, s)
			}
			boards[b][k/size][k%size] = v
		}
	}
	return order, boards, nil
}

func day4(path string) error {
	var turn [maxNumber]int
	order, boards, err := load(path, &turn)
	if err != nil {
		return err
	}
	first, last := -1, -1
	firstTurn, lastTurn := never, -1
	for i := range boards {
		t, err := boards[i].winTurn(&turn)
		if err != nil {
			return err
		}
		if t < firstTurn {
			first, firstTurn = i, t
		}
		if t != never && t > lastTurn {
			last, lastTurn = i, t
		}
	}
	if first < 0 || last < 0 {
		return errors.New("no board ever wins")
	}
	if firstTurn > lastTurn {
		return fmt.Errorf("first win on turn %d is after the last on %d", firstTurn, lastTurn)
	}
	s1, err := boards[first].score(&turn, firstTurn, order[firstTurn])
	if err != nil {
		return err
	}
	s2, err := boards[last].score(&turn, lastTurn, order[lastTurn])
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", s1, s2); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
