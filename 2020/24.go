// Day 24: Lobby Layout
//
// Each line of the input is a path of hex steps (e, se, sw, w, nw, ne)
// from a reference tile; the tile at its end is flipped. All tiles start
// white.
//
// Part 1 counts black tiles after every path.
// Part 2 then runs 100 days: a black tile with 0 or more than 2 black
// neighbours turns white, a white tile with exactly 2 turns black.
//
// Tiles use axial coordinates (q, r) in a fixed grid with room for 100
// days of growth.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 16
	maxPath       = 100
	maxLines      = 1_000
	days          = 100
	side          = 2*(maxPath+days) + 3
)

// The six neighbours in axial coordinates: e, w, ne, sw, nw, se.
const steps = "e w ne sw nw se"

// floor holds two copies of the tiles.
type floor struct {
	cur, next [side][side]bool
}

// stepOffset returns the axial offset for a step name.
func stepOffset(name string) (dq, dr int, err error) {
	switch name {
	case "e":
		return 1, 0, nil
	case "w":
		return -1, 0, nil
	case "ne":
		return 1, -1, nil
	case "sw":
		return -1, 1, nil
	case "nw":
		return 0, -1, nil
	case "se":
		return 0, 1, nil
	}
	return 0, 0, fmt.Errorf("unknown step %q", name)
}

// flip follows one path from the centre and flips its tile.
func (f *floor) flip(path string) error {
	if len(path) == 0 || len(path) > 2*maxPath {
		return fmt.Errorf("path of %d letters is out of range", len(path))
	}
	if strings.Trim(path, "nsew") != "" {
		return fmt.Errorf("path %q has letters other than n, s, e and w", path)
	}
	q, r := side/2, side/2
	for i := 0; i < len(path); i++ {
		n := 1
		if path[i] == 'n' || path[i] == 's' {
			n = 2
		}
		if i+n > len(path) {
			return fmt.Errorf("path %q ends mid-step", path)
		}
		dq, dr, err := stepOffset(path[i : i+n])
		if err != nil {
			return err
		}
		q, r, i = q+dq, r+dr, i+n-1
	}
	if q < 0 || r < 0 || q >= side || r >= side {
		return fmt.Errorf("path %q leaves the floor", path)
	}
	f.cur[r][q] = !f.cur[r][q]
	return nil
}

// black counts the black tiles.
func (f *floor) black() int {
	n := 0
	for r := range side {
		for q := range side {
			if f.cur[r][q] {
				n++
			}
		}
	}
	return n
}

// day runs one day of the art exhibit.
func (f *floor) day() error {
	offsets := strings.Fields(steps)
	if len(offsets) != 6 {
		return fmt.Errorf("%d neighbour directions, expected 6", len(offsets))
	}
	for r := 1; r < side-1; r++ {
		for q := 1; q < side-1; q++ {
			n := 0
			for _, s := range offsets {
				dq, dr, err := stepOffset(s)
				if err != nil {
					return err
				}
				if f.cur[r+dr][q+dq] {
					n++
				}
			}
			if n > 6 {
				return fmt.Errorf("tile %d,%d has %d neighbours", q, r, n)
			}
			f.next[r][q] = n == 2 || (f.cur[r][q] && n == 1)
			if f.next[r][q] && (r == 1 || q == 1 || r == side-2 || q == side-2) {
				return errors.New("the tiles grew to the edge of the floor")
			}
		}
	}
	f.cur, f.next = f.next, f.cur
	return nil
}

func day24(path string) error {
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
	lines := strings.Fields(string(content))
	if len(lines) == 0 || len(lines) > maxLines {
		return fmt.Errorf("%s: %d paths, expected 1..%d", path, len(lines), maxLines)
	}
	f := new(floor)
	for _, line := range lines {
		if err := f.flip(line); err != nil {
			return err
		}
	}
	first := f.black()
	if first < 1 || first > len(lines) {
		return fmt.Errorf("%d black tiles from %d paths", first, len(lines))
	}
	for range days {
		if err := f.day(); err != nil {
			return err
		}
	}
	last := f.black()
	if last < 1 {
		return errors.New("every tile turned white")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, last); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
