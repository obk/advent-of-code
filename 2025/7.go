// Day 7: Laboratories
//
// The input is a grid: 'S' is where the beam enters, '^' is a splitter, and
// the beam otherwise falls straight down. Hitting a splitter ends that beam
// and starts two new ones, to its left and to its right.
//
// Part 1 counts the splitters the beam reaches. Beams that arrive at a cell
// another beam already lit up merge into it, so all that matters is which
// cells are lit.
//
// Part 2 counts the distinct timelines instead, and this time the branches
// never merge: each splitter turns one timeline into two, and a particle that
// falls out of the grid has completed one.
//
// Beams only ever move down, so both parts are one sweep over the rows,
// carrying a count per column of how many timelines are falling through it.
// A column is lit for part 1 exactly when its count is non-zero, so part 1
// deduplicates where part 2 adds up - which is why part 2's answer is so much
// larger than part 1's.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
)

const (
	startCell    = 'S' // where the beam enters the manifold
	splitterCell = '^' // splits a beam into a left and a right one
	emptyCell    = '.'
)

// manifold is the grid the beam falls through; every row has the same length.
type manifold [][]byte

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// parseInput reads the grid. Splitters never sit side by side, so a beam
// leaving a splitter always starts falling from an empty cell.
func parseInput(path string) (manifold, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	if len(lines) < 2 {
		return nil, fmt.Errorf("%s: a manifold needs at least two rows", path)
	}
	m := make(manifold, len(lines))
	for r, line := range lines {
		if len(line) == 0 || len(line) != len(lines[0]) {
			return nil, fmt.Errorf("row %d is %d wide, expected %d", r, len(line), len(lines[0]))
		}
		if strings.Trim(line, string([]byte{startCell, splitterCell, emptyCell})) != "" {
			return nil, fmt.Errorf("row %d holds an unknown cell: %q", r, line)
		}
		if strings.Contains(line, string([]byte{splitterCell, splitterCell})) {
			return nil, fmt.Errorf("row %d has two splitters side by side", r)
		}
		m[r] = []byte(line)
	}
	return m, nil
}

// findStart locates the single 'S' cell.
func (m manifold) findStart() (row, col int, err error) {
	if len(m) == 0 {
		return 0, 0, errors.New("manifold is empty")
	}
	found := 0
	for r, cells := range m {
		for c, cell := range cells {
			if cell == startCell {
				row, col = r, c
				found++
			}
		}
	}
	if found != 1 {
		return 0, 0, fmt.Errorf("found %d start cells %q, expected 1", found, startCell)
	}
	return row, col, nil
}

// add adds n timelines to a running total, guarding against overflow.
func add(total, n int64) (int64, error) {
	if n < 0 || total < 0 {
		return 0, fmt.Errorf("negative timeline count %d + %d", total, n)
	}
	if total > math.MaxInt64-n {
		return 0, fmt.Errorf("timeline count %d + %d overflows int64", total, n)
	}
	return total + n, nil
}

// stepRow carries the timelines falling through one row into next, which
// holds what arrives at the row below. It returns how many splitters on this
// row were reached and how many timelines left the grid sideways.
func (m manifold) stepRow(row int, current, next []int64) (splits int, exited int64, err error) {
	if row < 0 || row >= len(m) {
		return 0, 0, fmt.Errorf("row %d is off the grid", row)
	}
	if len(current) != len(m[row]) || len(next) != len(m[row]) {
		return 0, 0, fmt.Errorf("column buffers do not match the grid width %d", len(m[row]))
	}

	clear(next)
	for col, count := range current {
		if count == 0 {
			continue
		}
		if m[row][col] != splitterCell {
			next[col], err = add(next[col], count)
			if err != nil {
				return 0, 0, err
			}
			continue
		}

		splits++
		for _, side := range [2]int{col - 1, col + 1} {
			if side < 0 || side >= len(next) {
				exited, err = add(exited, count)
			} else {
				next[side], err = add(next[side], count)
			}
			if err != nil {
				return 0, 0, err
			}
		}
	}
	// Splitters never touch, so at most every other cell is one.
	if splits > (len(current)+1)/2 {
		return 0, 0, fmt.Errorf("row %d reached %d splitters", row, splits)
	}
	return splits, exited, nil
}

// simulate drops the beam from the start and returns how many splitters it
// reaches (part 1) and how many timelines it ends up in (part 2).
func simulate(m manifold) (splits int, timelines int64, err error) {
	startRow, startCol, err := m.findStart()
	if err != nil {
		return 0, 0, err
	}

	if len(m[0]) == 0 {
		return 0, 0, errors.New("manifold has no columns")
	}

	// The two column buffers are swapped row by row, never reallocated.
	current := make([]int64, len(m[0]))
	next := make([]int64, len(m[0]))
	current[startCol] = 1

	for row := startRow; row < len(m); row++ {
		rowSplits, exited, err := m.stepRow(row, current, next)
		if err != nil {
			return 0, 0, err
		}
		splits += rowSplits
		if timelines, err = add(timelines, exited); err != nil {
			return 0, 0, err
		}
		current, next = next, current
	}

	// Whatever is still falling leaves through the bottom of the grid.
	for _, count := range current {
		if timelines, err = add(timelines, count); err != nil {
			return 0, 0, err
		}
	}
	if timelines < int64(splits)+1 {
		return 0, 0, fmt.Errorf("%d splits left only %d timelines", splits, timelines)
	}
	return splits, timelines, nil
}

func day7(path string) error {
	m, err := parseInput(path)
	if err != nil {
		return err
	}
	splits, timelines, err := simulate(m)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", splits, timelines); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day7(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
