// Day 4: Printing Department
//
// The input is a grid of shelves: '@' is a paper roll, '.' an empty spot. A
// roll can be reached only if fewer than four of its eight surrounding cells
// still hold a roll; any more and it is boxed in.
//
// Part 1 counts the rolls reachable to begin with.
// Part 2 keeps going: taking rolls frees the ones that were behind them, so
// the grid is swept again and again until a sweep finds nothing, and the
// answer is how many rolls came off in total.
//
// Each sweep finds every reachable roll before removing any of them. That
// keeps the removals within a sweep simultaneous, so the result does not
// depend on the order the grid happens to be scanned in. Every sweep but the
// last removes at least one roll, so there is at most one more sweep than
// there are rolls.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
)

const (
	roll  = '@' // a paper roll still on the shelf
	empty = '.' // a spot one has been taken from

	// A roll is reachable only if fewer than this many neighbours box it in.
	crowdedNeighbours = 4
	// Each cell has eight neighbours (Moore neighbourhood).
	neighbourCount = 8
)

// position is a row/column pair in the shelf grid.
type position struct {
	row, col int
}

// shelf is the warehouse grid; every row has the same length.
type shelf [][]byte

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

// parseInput reads the grid, which must be a non-empty rectangle of '@' and '.'.
func parseInput(path string) (shelf, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	s := make(shelf, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(s) > 0 && len(line) != len(s[0]) {
			return nil, fmt.Errorf("row %d is %d wide, expected %d", len(s), len(line), len(s[0]))
		}
		if strings.Trim(line, string([]byte{roll, empty})) != "" {
			return nil, fmt.Errorf("row %q holds something other than %q and %q", line, roll, empty)
		}
		s = append(s, []byte(line))
	}
	if len(s) == 0 {
		return nil, errors.New("grid is empty")
	}
	return s, nil
}

// countNeighbourRolls counts how many of the eight surrounding cells still
// hold a roll. Cells off the edge of the grid count as empty.
func (s shelf) countNeighbourRolls(p position) (int, error) {
	if p.row < 0 || p.row >= len(s) || p.col < 0 || p.col >= len(s[p.row]) {
		return 0, fmt.Errorf("position %v is off the grid", p)
	}

	count := 0
	for dr := -1; dr <= 1; dr++ {
		for dc := -1; dc <= 1; dc++ {
			row, col := p.row+dr, p.col+dc
			if (dr == 0 && dc == 0) || row < 0 || row >= len(s) || col < 0 || col >= len(s[row]) {
				continue
			}
			if s[row][col] == roll {
				count++
			}
		}
	}

	if count > neighbourCount {
		return 0, fmt.Errorf("position %v has %d neighbouring rolls", p, count)
	}
	return count, nil
}

// accessibleRolls lists every roll that can be reached right now, i.e. that
// has fewer than crowdedNeighbours rolls around it. The list is written into
// buf, which must be able to hold every cell so it never has to grow.
func (s shelf) accessibleRolls(buf []position) ([]position, error) {
	if len(s) == 0 {
		return nil, errors.New("grid is empty")
	}
	if cap(buf) < len(s)*len(s[0]) {
		return nil, fmt.Errorf("buffer holds %d positions, grid has %d cells", cap(buf), len(s)*len(s[0]))
	}

	accessible := buf[:0]
	for r := range s {
		for c := range s[r] {
			if s[r][c] != roll {
				continue
			}
			neighbours, err := s.countNeighbourRolls(position{r, c})
			if err != nil {
				return nil, err
			}
			if neighbours < crowdedNeighbours {
				accessible = append(accessible, position{r, c})
			}
		}
	}
	return accessible, nil
}

// remove takes the rolls at the given positions off the shelf.
func (s shelf) remove(positions []position) error {
	for _, p := range positions {
		if p.row < 0 || p.row >= len(s) || p.col < 0 || p.col >= len(s[p.row]) {
			return fmt.Errorf("position %v is off the grid", p)
		}
		if s[p.row][p.col] != roll {
			return fmt.Errorf("no roll to remove at %v", p)
		}
		s[p.row][p.col] = empty
	}
	return nil
}

func day4(path string) error {
	s, err := parseInput(path)
	if err != nil {
		return err
	}

	cells := len(s) * len(s[0])
	accessible := make([]position, 0, cells)
	firstPass := 0
	totalRemoved := 0

	// Take every reachable roll, then look again: removing rolls frees up the
	// ones behind them. Stop once a pass finds nothing. Every pass but the
	// last removes a roll, so cells+1 passes always suffice.
	for pass := 0; pass <= cells; pass++ {
		accessible, err = s.accessibleRolls(accessible)
		if err != nil {
			return err
		}
		if len(accessible) == 0 {
			break
		}
		if pass == 0 {
			firstPass = len(accessible)
		}
		if err := s.remove(accessible); err != nil {
			return err
		}
		totalRemoved += len(accessible)
	}

	if len(accessible) != 0 {
		return fmt.Errorf("rolls were still being removed after %d passes", cells+1)
	}
	if firstPass > totalRemoved || totalRemoved > cells {
		return fmt.Errorf("removed %d rolls in all but %d in the first pass", totalRemoved, firstPass)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", firstPass, totalRemoved); err != nil {
		return err
	}
	return nil
}

func main() {
	if err := day4(inputPath); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
