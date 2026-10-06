// Day 18: Like a Rogue
//
// The input is the first row of a room's floor, '.' safe and '^' a trap. Each
// later tile is a trap exactly when the tiles to its upper left and upper
// right differ (beyond the edges counts as safe) - the puzzle's four rules
// all come down to that.
//
// Part 1 counts the safe tiles in the first 40 rows.
// Part 2 counts them in the first 400000 rows.
//
// Only two rows are kept, alternating as the current and the next one.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 12
	maxWidth      = 1_000
	maxRows       = 10_000_000
	firstRows     = 40
	secondRows    = 400_000
)

// readRow returns the first row as booleans: true for a trap.
func readRow(path string) ([]bool, error) {
	if path == "" {
		return nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	line := strings.TrimSpace(string(content))
	if line == "" || len(line) > maxWidth || strings.Trim(line, ".^") != "" {
		return nil, fmt.Errorf("%s: %q is not a row of '.' and '^'", path, line)
	}
	if len(line) < 2 {
		return nil, fmt.Errorf("%s: a row needs at least two tiles", path)
	}
	row := make([]bool, len(line))
	for i := range len(line) {
		row[i] = line[i] == '^'
	}
	return row, nil
}

// countSafe counts the safe tiles in the first rows rows, starting from
// first. a and b are scratch rows as wide as first.
func countSafe(first, a, b []bool, rows int) (int, error) {
	if rows < 1 || rows > maxRows {
		return 0, fmt.Errorf("%d rows is outside 1..%d", rows, maxRows)
	}
	if len(first) == 0 || len(a) != len(first) || len(b) != len(first) {
		return 0, errors.New("scratch rows do not match the first row")
	}

	current, next := a, b
	copy(current, first)
	safe := 0
	w := len(first)
	for range rows {
		for i := range w {
			if !current[i] {
				safe++
			}
			left := i > 0 && current[i-1]
			right := i < w-1 && current[i+1]
			next[i] = left != right
		}
		current, next = next, current
	}

	if safe > rows*w {
		return 0, fmt.Errorf("%d safe tiles in %d rows of %d", safe, rows, w)
	}
	return safe, nil
}

func day18(path string) error {
	first, err := readRow(path)
	if err != nil {
		return err
	}

	a, b := make([]bool, len(first)), make([]bool, len(first))
	safe, err := countSafe(first, a, b, firstRows)
	if err != nil {
		return err
	}
	more, err := countSafe(first, a, b, secondRows)
	if err != nil {
		return err
	}

	if more < safe {
		return fmt.Errorf("more rows held fewer safe tiles (%d < %d)", more, safe)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", safe, more); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
