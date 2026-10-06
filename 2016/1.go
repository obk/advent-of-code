// Day 1: No Time for a Taxicab
//
// The input is one line of turns and walks, such as "R2, L3": turn right or
// left by 90 degrees, then walk that many blocks. You start facing north.
//
// Part 1 is how many blocks away (by taxicab distance) the walk ends.
// Part 2 is how far away the first location visited twice is.
//
// Headings are numbered 0..3 clockwise from north, so a right turn adds one
// and a left turn adds three, modulo four. For part 2 every block walked
// through goes into a set, which is created big enough for the whole walk.
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
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
	maxBlocks     = 1_000_000
)

// point is a street corner, with the start at the origin.
type point struct {
	x, y int
}

// move is one instruction: which way to turn, and how far to walk.
type move struct {
	right  bool
	blocks int
}

// readMoves parses the comma-separated instructions.
func readMoves(path string) ([]move, error) {
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

	fields := strings.Split(strings.TrimSpace(string(content)), ",")
	moves := make([]move, len(fields))
	for i, field := range fields {
		field = strings.TrimSpace(field)
		if len(field) < 2 || (field[0] != 'R' && field[0] != 'L') {
			return nil, fmt.Errorf("malformed move %q", field)
		}
		blocks, err := strconv.Atoi(field[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid move %q: %w", field, err)
		}
		if blocks < 0 || blocks > maxBlocks {
			return nil, fmt.Errorf("move %q is outside 0..%d blocks", field, maxBlocks)
		}
		moves[i] = move{right: field[0] == 'R', blocks: blocks}
	}
	return moves, nil
}

// taxicab is the distance from the start to p along the street grid.
func taxicab(p point) int {
	return max(p.x, -p.x) + max(p.y, -p.y)
}

// walk follows the moves block by block and returns the distance to the end
// point and to the first point visited twice (-1 if none is). visited must be
// empty and is filled in.
func walk(moves []move, visited map[point]bool) (end, twice int, err error) {
	if len(moves) == 0 {
		return 0, 0, errors.New("no moves")
	}
	if len(visited) != 0 {
		return 0, 0, errors.New("visited set is not empty")
	}
	if len(moves) > maxInputBytes {
		return 0, 0, fmt.Errorf("%d moves, more than the input can hold", len(moves))
	}

	// North, east, south, west.
	steps := [4]point{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	heading, p, twice := 0, point{}, -1
	visited[p] = true
	for _, m := range moves {
		if m.blocks < 0 || m.blocks > maxBlocks {
			return 0, 0, fmt.Errorf("move of %d blocks is out of range", m.blocks)
		}
		if m.right {
			heading = (heading + 1) % 4
		} else {
			heading = (heading + 3) % 4
		}
		for range m.blocks {
			p.x, p.y = p.x+steps[heading].x, p.y+steps[heading].y
			if visited[p] && twice < 0 {
				twice = taxicab(p)
			}
			visited[p] = true
		}
	}

	if taxicab(p) > len(moves)*maxBlocks {
		return 0, 0, fmt.Errorf("ended %d blocks away, further than the moves allow", taxicab(p))
	}
	return taxicab(p), twice, nil
}

func day1(path string) error {
	moves, err := readMoves(path)
	if err != nil {
		return err
	}
	total := 0
	for _, m := range moves {
		total += m.blocks
	}
	// One entry per block walked, plus the start: the set never grows.
	visited := make(map[point]bool, total+1)
	distance, twice, err := walk(moves, visited)
	if err != nil {
		return err
	}

	if twice < 0 {
		return errors.New("part 2: no location is visited twice")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", distance, twice); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
