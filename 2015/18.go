// Day 18: Like a GIF For Your Yard
//
// The input is a 100x100 grid of lights, '#' on and '.' off, animated as
// Conway's Game of Life: an on light stays on with 2 or 3 neighbours on, an
// off light turns on with exactly 3, and lights beyond the edge count as off.
//
// Part 1 is how many lights are on after 100 steps.
// Part 2 does the same with the four corner lights stuck on.
//
// Two grids are allocated once and alternate as the current and next state.
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
	maxInputBytes = 1 << 20
	side          = 100
	steps         = 100
)

type grid [side][side]bool

// readGrid parses the starting lights, which must be exactly side x side.
func readGrid(path string) (grid, error) {
	if path == "" {
		return grid{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return grid{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return grid{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != side {
		return grid{}, fmt.Errorf("%s: %d rows, expected %d", path, len(lines), side)
	}
	var g grid
	for r, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) != side || strings.Trim(line, "#.") != "" {
			return grid{}, fmt.Errorf("%s: row %d is not %d of '#' and '.'", path, r, side)
		}
		for c := range side {
			g[r][c] = line[c] == '#'
		}
	}
	return g, nil
}

// step writes the next state of current into next.
func step(current, next *grid, stuckCorners bool) error {
	if current == nil || next == nil {
		return errors.New("missing grid")
	}
	if current == next {
		return errors.New("current and next must be different grids")
	}
	for r := range side {
		for c := range side {
			on := 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					nr, nc := r+dr, c+dc
					if (dr != 0 || dc != 0) && nr >= 0 && nr < side && nc >= 0 && nc < side && current[nr][nc] {
						on++
					}
				}
			}
			if on > 8 {
				return fmt.Errorf("light %d,%d has %d neighbours on", r, c, on)
			}
			next[r][c] = on == 3 || (on == 2 && current[r][c])
		}
	}
	if stuckCorners {
		next[0][0], next[0][side-1], next[side-1][0], next[side-1][side-1] = true, true, true, true
	}
	return nil
}

// animate runs the steps from start and returns how many lights end up on.
// The two grids are scratch space.
func animate(start grid, stuckCorners bool, a, b *grid) (int, error) {
	if a == b {
		return 0, errors.New("animation needs two distinct grids")
	}
	*a = start
	if stuckCorners {
		a[0][0], a[0][side-1], a[side-1][0], a[side-1][side-1] = true, true, true, true
	}
	for range steps {
		if err := step(a, b, stuckCorners); err != nil {
			return 0, err
		}
		a, b = b, a
	}

	on := 0
	for r := range side {
		for c := range side {
			if a[r][c] {
				on++
			}
		}
	}
	if on > side*side {
		return 0, fmt.Errorf("%d lights on in a grid of %d", on, side*side)
	}
	if stuckCorners && on < 4 {
		return 0, fmt.Errorf("only %d lights on with four stuck on", on)
	}
	return on, nil
}

func day18(path string) error {
	start, err := readGrid(path)
	if err != nil {
		return err
	}

	a, b := new(grid), new(grid)
	plain, err := animate(start, false, a, b)
	if err != nil {
		return err
	}
	stuck, err := animate(start, true, a, b)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", plain, stuck); err != nil {
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
