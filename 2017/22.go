// Day 22: Sporifica Virus
//
// The input is the infected ('#') and clean ('.') nodes at the middle of an
// infinite grid. A virus carrier starts in the centre facing up. Each burst
// it turns right on an infected node and left on a clean one, flips the
// node, and moves forward one node.
//
// Part 1 counts the bursts, out of 10000, that infect a node.
// Part 2 has the virus evolve: nodes go clean -> weakened -> infected ->
// flagged -> clean; the carrier turns left on clean, keeps straight on
// weakened, turns right on infected and reverses on flagged. It counts the
// infections over 10000000 bursts.
//
// The grid is a fixed square much larger than the carrier can wander in the
// bursts allowed, with the input in its middle; leaving it is an error.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 16
	maxMap        = 101
	side          = 2001
	part1Bursts   = 10_000
	part2Bursts   = 10_000_000
)

// Node states, in the order an evolved virus would cycle through them.
const (
	clean byte = iota
	weakened
	infected
	flagged
)

// cluster is the grid of nodes.
type cluster [side * side]byte

// load places the map in the middle of the grid and returns the centre.
func (g *cluster) load(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	rows := strings.Split(strings.TrimSpace(string(content)), "\n")
	n := len(rows)
	if n%2 == 0 || n > maxMap {
		return 0, fmt.Errorf("%s: map of %d rows has no single centre", path, n)
	}
	off := side/2 - n/2
	for r, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) != n || strings.Trim(row, ".#") != "" {
			return 0, fmt.Errorf("%s: row %d is not %d of '.' and '#'", path, r, n)
		}
		for c := range n {
			if row[c] == '#' {
				g[(off+r)*side+off+c] = infected
			}
		}
	}
	return (side/2)*side + side/2, nil
}

// bursts runs the virus and counts the bursts that infect a node.
func (g *cluster) bursts(start, count int) (int, error) {
	if start < 0 || start >= side*side || count < 0 {
		return 0, fmt.Errorf("start %d or count %d is invalid", start, count)
	}
	// Up, right, down, left.
	dr, dc := [4]int{-1, 0, 1, 0}, [4]int{0, 1, 0, -1}
	r, c, dir, infections := start/side, start%side, 0, 0
	for range count {
		if g[r*side+c] == infected {
			dir = (dir + 1) % 4
			g[r*side+c] = clean
		} else {
			dir = (dir + 3) % 4
			g[r*side+c] = infected
			infections++
		}
		r, c = r+dr[dir], c+dc[dir]
		if r < 0 || r >= side || c < 0 || c >= side {
			return 0, errors.New("the carrier left the grid")
		}
	}
	return infections, nil
}

// evolvedBursts runs the evolved virus and counts the infections.
func (g *cluster) evolvedBursts(start, count int) (int, error) {
	if start < 0 || start >= side*side || count < 0 {
		return 0, fmt.Errorf("start %d or count %d is invalid", start, count)
	}
	// Up, right, down, left; a state's turn is how many right turns it makes.
	dr, dc := [4]int{-1, 0, 1, 0}, [4]int{0, 1, 0, -1}
	turn := [4]int{clean: 3, weakened: 0, infected: 1, flagged: 2}
	r, c, dir, infections := start/side, start%side, 0, 0
	for range count {
		node := &g[r*side+c]
		if *node > flagged {
			return 0, fmt.Errorf("node %d,%d is in unknown state %d", r, c, *node)
		}
		dir = (dir + turn[*node]) % 4
		*node = (*node + 1) % 4
		if *node == infected {
			infections++
		}
		r, c = r+dr[dir], c+dc[dir]
		if r < 0 || r >= side || c < 0 || c >= side {
			return 0, errors.New("the carrier left the grid")
		}
	}
	return infections, nil
}

func day22(path string) error {
	g := new(cluster)
	start, err := g.load(path)
	if err != nil {
		return err
	}
	infections, err := g.bursts(start, part1Bursts)
	if err != nil {
		return err
	}
	g = new(cluster)
	if start, err = g.load(path); err != nil {
		return err
	}
	evolved, err := g.evolvedBursts(start, part2Bursts)
	if err != nil {
		return err
	}

	if infections > part1Bursts || evolved > part2Bursts {
		return fmt.Errorf("more infections (%d, %d) than bursts", infections, evolved)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", infections, evolved); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
