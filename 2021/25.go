// Day 25: Sea Cucumber
//
// The input is a map of sea cucumbers: '>' move east, 'v' move south,
// each into an empty cell ('.'), wrapping round the edges. Each step all
// east-movers move at once, then all south-movers.
//
// Part 1 is the first step on which nothing moves. (Day 25 has no second
// puzzle.)
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	maxStepsTried = 10_000
)

// floor is the map and a buffer for the next half-step.
type floor struct {
	cur, next [maxSide][maxSide]byte
	w, h      int
}

// move moves one herd and reports how many moved.
func (f *floor) move(herd byte, dx, dy int) (int, error) {
	if herd != '>' && herd != 'v' {
		return 0, fmt.Errorf("unknown herd %q", herd)
	}
	f.next = f.cur
	moved := 0
	for y := range f.h {
		for x := range f.w {
			if f.cur[y][x] != herd {
				continue
			}
			nx, ny := (x+dx)%f.w, (y+dy)%f.h
			if f.cur[ny][nx] == '.' {
				f.next[y][x], f.next[ny][nx], moved = '.', herd, moved+1
			}
		}
	}
	f.cur = f.next
	return moved, nil
}

func day25(path string) error {
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
	if len(rows) == 0 || len(rows) > maxSide || len(rows[0]) > maxSide {
		return fmt.Errorf("%s: map is empty or larger than %d", path, maxSide)
	}
	f := &floor{w: len(rows[0]), h: len(rows)}
	for y, row := range rows {
		if len(row) != f.w || strings.Trim(row, ".>v") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		copy(f.cur[y][:], row)
	}
	still := 0
	for step := 1; step <= maxStepsTried && still == 0; step++ {
		east, err := f.move('>', 1, 0)
		if err != nil {
			return err
		}
		south, err := f.move('v', 0, 1)
		if err != nil {
			return err
		}
		if east+south == 0 {
			still = step
		}
	}
	if still == 0 {
		return fmt.Errorf("still moving after %d steps", maxStepsTried)
	}

	if _, err := fmt.Printf("Part 1: %d\n", still); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
