// Day 3: Toboggan Trajectory
//
// The input is a map of open squares ('.') and trees ('#'), repeating
// endlessly to the right.
//
// Part 1 counts the trees met going right 3, down 1 from the top left.
// Part 2 multiplies the trees met on slopes (1,1), (3,1), (5,1), (7,1) and
// (1,2).
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
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 16
	maxRows       = 1_000
)

// readMap returns the rows, which must all be the same width.
func readMap(path string) ([]string, error) {
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
	rows := strings.Fields(string(content))
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s: empty map", path)
	}
	if len(rows) > maxRows {
		return nil, fmt.Errorf("%s: more than %d rows", path, maxRows)
	}
	for i, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, ".#") != "" {
			return nil, fmt.Errorf("%s: row %d is not %d of '.' and '#'", path, i, len(rows[0]))
		}
	}
	return rows, nil
}

// trees counts the trees on one slope.
func trees(rows []string, right, down int) (int, error) {
	if right < 0 || down < 1 || len(rows) == 0 {
		return 0, fmt.Errorf("slope %d,%d or empty map", right, down)
	}
	n, x := 0, 0
	for y := 0; y < len(rows); y += down {
		if rows[y][x%len(rows[y])] == '#' {
			n++
		}
		x += right
	}
	return n, nil
}

func day3(path string) error {
	rows, err := readMap(path)
	if err != nil {
		return err
	}
	first, err := trees(rows, 3, 1)
	if err != nil {
		return err
	}
	product := 1
	for _, s := range [5][2]int{{1, 1}, {3, 1}, {5, 1}, {7, 1}, {1, 2}} {
		n, err := trees(rows, s[0], s[1])
		if err != nil {
			return err
		}
		if n > 0 && product > math.MaxInt64/n {
			return errors.New("the product overflows")
		}
		product *= n
	}
	if first > len(rows) {
		return fmt.Errorf("met %d trees in %d rows", first, len(rows))
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, product); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
