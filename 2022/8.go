// Day 8: Treetop Tree House
//
// The input is a grid of tree heights 0-9.
//
// Part 1 counts trees visible from outside: those taller than every tree
// between them and some edge.
// Part 2 is the highest scenic score: the product, over the four
// directions, of how many trees can be seen before one at least as tall.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 16
	maxSide       = 128
)

// look returns how many trees (x, y) sees in direction (dx, dy), and
// whether it sees past the edge.
func look(rows []string, x, y, dx, dy int) (int, bool, error) {
	if y < 0 || y >= len(rows) || x < 0 || x >= len(rows[y]) {
		return 0, false, fmt.Errorf("tree %d,%d is off the grid", x, y)
	}
	n := 0
	for cx, cy := x+dx, y+dy; cy >= 0 && cy < len(rows) && cx >= 0 && cx < len(rows[cy]); cx, cy = cx+dx, cy+dy {
		n++
		if rows[cy][cx] >= rows[y][x] {
			return n, false, nil
		}
	}
	return n, true, nil
}

func day8(path string) error {
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
	rows := strings.Fields(string(content))
	if len(rows) == 0 || len(rows) > maxSide {
		return fmt.Errorf("%s: grid is empty or larger than %d", path, maxSide)
	}
	visible, best := 0, 0
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, "0123456789") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		for x := range len(row) {
			seen, score := false, 1
			for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
				n, out, err := look(rows, x, y, d[0], d[1])
				if err != nil {
					return err
				}
				seen, score = seen || out, score*n
			}
			if seen {
				visible++
			}
			best = max(best, score)
		}
	}

	// Every tree on the edge is visible.
	if edge := 2*(len(rows)+len(rows[0])) - 4; visible < edge {
		return fmt.Errorf("%d visible trees, fewer than the %d on the edge", visible, edge)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", visible, best); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
