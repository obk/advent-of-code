// Day 13: Point of Incidence
//
// The input is patterns of ash '.' and rock '#', separated by blank lines.
// Each has a line of reflection between two rows or two columns.
//
// Part 1 sums the number of columns left of each vertical line, plus 100
// times the rows above each horizontal one.
// Part 2 does the same for the line that reflects with exactly one cell
// different (the smudge).
//
// For each candidate line, the cells that differ from their reflection are
// counted, stopping past one.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 16
	maxSide       = 32
)

// mismatches counts cells differing across a line after `at` rows (or
// columns, if vertical), stopping once past one.
func mismatches(rows []string, at int, vertical bool) (int, error) {
	if len(rows) == 0 {
		return 0, errors.New("empty pattern")
	}
	h, w := len(rows), len(rows[0])
	if at < 1 || (vertical && at >= w) || (!vertical && at >= h) {
		return 0, fmt.Errorf("line after %d is out of range", at)
	}
	n := 0
	for d := 0; d < max(h, w) && n < 2; d++ {
		a, b := at-1-d, at+d
		if a < 0 || (vertical && b >= w) || (!vertical && b >= h) {
			break
		}
		for k := 0; k < max(h, w) && n < 2; k++ {
			if vertical && k < h && rows[k][a] != rows[k][b] {
				n++
			}
			if !vertical && k < w && rows[a][k] != rows[b][k] {
				n++
			}
		}
	}
	return n, nil
}

// summarise returns a pattern's score for lines with the given number of
// differences.
func summarise(rows []string, smudges int) (int, error) {
	if len(rows) == 0 || len(rows) > maxSide || len(rows[0]) > maxSide {
		return 0, errors.New("pattern is empty or too large")
	}
	if smudges < 0 || smudges > 1 {
		return 0, fmt.Errorf("%d smudges is not 0 or 1", smudges)
	}
	for _, vertical := range [2]bool{true, false} {
		size, weight := len(rows), 100
		if vertical {
			size, weight = len(rows[0]), 1
		}
		for at := 1; at < size; at++ {
			n, err := mismatches(rows, at, vertical)
			if err != nil {
				return 0, err
			}
			if n == smudges {
				return weight * at, nil
			}
		}
	}
	return 0, fmt.Errorf("no line with %d differences", smudges)
}

func day13(path string) error {
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
	var answers [2]int
	for _, block := range strings.Split(strings.TrimSpace(string(content)), "\n\n") {
		rows := strings.Fields(block)
		for y, r := range rows {
			if len(r) != len(rows[0]) || strings.Trim(r, ".#") != "" {
				return fmt.Errorf("%s: pattern row %d is malformed", path, y)
			}
		}
		for k := range 2 {
			s, err := summarise(rows, k)
			if err != nil {
				return err
			}
			answers[k] += s
		}
	}

	if answers[0] < 1 || answers[1] < 1 {
		return errors.New("no reflections found")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
