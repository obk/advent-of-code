// Day 14: Regolith Reservoir
//
// Each line of the input is a path of rock, "498,4 -> 498,6 -> 496,6".
// Sand falls from 500,0 one unit at a time: down if it can, else down-left,
// else down-right, else it rests.
//
// Part 1 counts the units that rest before sand starts falling into the
// abyss below the lowest rock.
// Part 2 adds an endless floor two below the lowest rock and counts the
// units until one rests at the source.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/14.txt"
	maxInputBytes = 1 << 16
	width         = 1_000
	maxHeight     = 200
	sourceX       = 500
)

// cave is the rock and sand.
type cave struct {
	solid  [maxHeight + 3][width]bool
	bottom int
}

// load draws the rock paths.
func load(path string) (*cave, error) {
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
	c := new(cave)
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		px, py := -1, -1
		for _, pt := range strings.Split(line, " -> ") {
			var x, y int
			if n, err := fmt.Sscanf(pt, "%d,%d", &x, &y); err != nil || n != 2 || x < 0 || x >= width || y < 0 || y > maxHeight {
				return nil, fmt.Errorf("%s: invalid point %q", path, pt)
			}
			if px >= 0 && px != x && py != y {
				return nil, fmt.Errorf("%s: diagonal rock in %q", path, line)
			}
			for k := 0; px >= 0 && k <= max(x-px, px-x, y-py, py-y); k++ {
				c.solid[py+k*max(-1, min(1, y-py))][px+k*max(-1, min(1, x-px))] = true
			}
			px, py = x, y
			c.bottom = max(c.bottom, y)
		}
	}
	if c.bottom == 0 {
		return nil, fmt.Errorf("%s: no rock below the source", path)
	}
	return c, nil
}

// fall drops one unit from the source and returns where it rests, or
// reports that it fell past the lowest rock (only possible with no floor).
func (c *cave) fall(floor bool) (x, y int, fell bool, err error) {
	if c.solid[0][sourceX] {
		return 0, 0, false, errors.New("the source is blocked")
	}
	if c.bottom < 1 || c.bottom > maxHeight {
		return 0, 0, false, fmt.Errorf("lowest rock %d is out of range", c.bottom)
	}
	x = sourceX
	for y = 0; y <= c.bottom; y++ {
		switch {
		case !c.solid[y+1][x]:
		case x > 0 && !c.solid[y+1][x-1]:
			x--
		case x < width-1 && !c.solid[y+1][x+1]:
			x++
		default:
			return x, y, false, nil
		}
	}
	// Now at the row just above the floor (or below every rock).
	if x < 1 || x >= width-1 {
		return 0, 0, false, errors.New("sand reached the side of the cave")
	}
	return x, y, !floor, nil
}

// pour drops sand until it falls past the rock (no floor) or blocks the
// source (floor), and returns how many units rested.
func (c *cave) pour(floor bool) (int, error) {
	if c.bottom < 1 || c.bottom > maxHeight {
		return 0, fmt.Errorf("lowest rock %d is out of range", c.bottom)
	}
	for units := 0; units < width*(maxHeight+3); units++ {
		if c.solid[0][sourceX] {
			return units, nil
		}
		x, y, fell, err := c.fall(floor)
		if err != nil {
			return 0, err
		}
		if fell {
			return units, nil
		}
		c.solid[y][x] = true
	}
	return 0, errors.New("the sand never stopped")
}

func day14(path string) error {
	c, err := load(path)
	if err != nil {
		return err
	}
	abyss, err := c.pour(false)
	if err != nil {
		return err
	}
	// The sand already resting stays; keep pouring onto the floor.
	more, err := c.pour(true)
	if err != nil {
		return err
	}

	if abyss < 1 || more < 1 {
		return fmt.Errorf("only %d and %d units of sand rested", abyss, more)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", abyss, abyss+more); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day14(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
