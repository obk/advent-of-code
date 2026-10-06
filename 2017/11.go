// Day 11: Hex Ed
//
// The input is a path of steps on a hex grid: n, ne, se, s, sw, nw.
//
// Part 1 is the fewest steps from the start to where the path ends.
// Part 2 is the furthest the path ever gets from the start.
//
// Hexes use cube coordinates (x, y, z) with x + y + z = 0; each direction
// changes two of them by one, and the distance from the origin is the
// largest of |x|, |y| and |z|.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 20
)

// cube is a hex in cube coordinates.
type cube struct {
	x, y, z int
}

// readSteps returns the comma-separated steps.
func readSteps(path string) ([]string, error) {
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
	steps := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(steps) == 0 || steps[0] == "" {
		return nil, fmt.Errorf("%s: no steps", path)
	}
	return steps, nil
}

// distance is the number of steps from the origin to c.
func distance(c cube) (int, error) {
	if c.x+c.y+c.z != 0 {
		return 0, fmt.Errorf("%+v is not a valid hex", c)
	}
	return max(c.x, -c.x, c.y, -c.y, c.z, -c.z), nil
}

// walk follows the steps and returns the final and the furthest distance.
func walk(steps []string) (final, furthest int, err error) {
	if len(steps) == 0 {
		return 0, 0, errors.New("no steps")
	}
	if len(steps) > maxInputBytes {
		return 0, 0, fmt.Errorf("%d steps, more than the input can hold", len(steps))
	}
	var c cube
	for _, s := range steps {
		switch s {
		case "n":
			c.y, c.z = c.y+1, c.z-1
		case "s":
			c.y, c.z = c.y-1, c.z+1
		case "ne":
			c.x, c.z = c.x+1, c.z-1
		case "sw":
			c.x, c.z = c.x-1, c.z+1
		case "nw":
			c.x, c.y = c.x-1, c.y+1
		case "se":
			c.x, c.y = c.x+1, c.y-1
		default:
			return 0, 0, fmt.Errorf("unknown step %q", s)
		}
		d, err := distance(c)
		if err != nil {
			return 0, 0, err
		}
		final, furthest = d, max(furthest, d)
	}
	if furthest > len(steps) {
		return 0, 0, fmt.Errorf("got %d away in %d steps", furthest, len(steps))
	}
	return final, furthest, nil
}

func day11(path string) error {
	steps, err := readSteps(path)
	if err != nil {
		return err
	}
	d, furthest, err := walk(steps)
	if err != nil {
		return err
	}

	if furthest < 1 {
		return fmt.Errorf("the path never left the start (furthest %d)", furthest)
	}
	if d > furthest {
		return fmt.Errorf("final distance %d beats the furthest %d", d, furthest)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", d, furthest); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day11(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
