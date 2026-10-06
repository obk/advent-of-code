// Day 3: Perfectly Spherical Houses in a Vacuum
//
// The input is one line of moves on an infinite grid of houses: '^', 'v',
// '>' and '<' step north, south, east and west. A present is delivered at the
// start and after every move.
//
// Part 1 counts the houses Santa alone delivers to.
// Part 2 has Santa and Robo-Santa take turns with the moves, both starting at
// the same house, and counts the houses either of them visits.
//
// The visited houses are kept in a set. Each move visits at most one new
// house, so the set is created big enough for every move up front and never
// has to grow.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 20
)

// house is a position on the grid, start at the origin.
type house struct {
	x, y int
}

// readMoves returns the single line of moves.
func readMoves(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	moves := strings.TrimSpace(string(content))
	if moves == "" || strings.Trim(moves, "^v<>") != "" {
		return "", fmt.Errorf("%s: holds something other than moves", path)
	}
	return moves, nil
}

// countHouses returns how many houses get a present when the moves are dealt
// out in turn to the given number of deliverers. visited must be empty and
// is filled in.
func countHouses(moves string, deliverers int, visited map[house]bool) (int, error) {
	if deliverers < 1 || deliverers > 2 {
		return 0, fmt.Errorf("%d deliverers is outside 1..2", deliverers)
	}
	if len(visited) != 0 {
		return 0, errors.New("visited set is not empty")
	}
	if moves == "" {
		return 0, errors.New("no moves")
	}

	var positions [2]house
	visited[house{}] = true
	for i := range len(moves) {
		p := &positions[i%deliverers]
		switch moves[i] {
		case '^':
			p.y++
		case 'v':
			p.y--
		case '>':
			p.x++
		case '<':
			p.x--
		default:
			return 0, fmt.Errorf("unknown move %q at %d", moves[i], i+1)
		}
		visited[*p] = true
	}

	if len(visited) > len(moves)+1 {
		return 0, fmt.Errorf("%d moves visited %d houses", len(moves), len(visited))
	}
	return len(visited), nil
}

func day3(path string) error {
	moves, err := readMoves(path)
	if err != nil {
		return err
	}

	// At most one new house per move, plus the start.
	visited := make(map[house]bool, len(moves)+1)
	alone, err := countHouses(moves, 1, visited)
	if err != nil {
		return err
	}
	clear(visited)
	withRobot, err := countHouses(moves, 2, visited)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", alone, withRobot); err != nil {
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
