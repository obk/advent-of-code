// Day 14: Parabolic Reflector Dish
//
// The input is a platform of round rocks 'O' (which roll), cube rocks '#'
// (which don't) and empty space. A round rock's load is the number of rows
// from it to the south edge, inclusive.
//
// Part 1 tilts the platform north and gives the total load.
// Part 2 runs 1000000000 spin cycles (tilt north, west, south, east) and
// gives the load after them.
//
// The platform soon repeats a layout; each layout after a cycle is
// fingerprinted (FNV-1a), and the load at the billionth cycle is read from
// the repeating stretch.
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
	maxSide       = 128
	maxCycles     = 1_000
	spins         = 1_000_000_000
)

// platform is the grid.
type platform struct {
	g    [maxSide][maxSide]byte
	w, h int
}

// tilt rolls every round rock as far as it goes in direction d (0 north,
// 1 west, 2 south, 3 east).
func (p *platform) tilt(d int) error {
	if d < 0 || d > 3 {
		return fmt.Errorf("direction %d does not exist", d)
	}
	lines, length := p.w, p.h // north/south: one line per column
	if d%2 == 1 {
		lines, length = p.h, p.w
	}
	for a := range lines {
		free := 0 // the next free place, counted from the edge rolled towards
		for b := range length {
			// The cell b places from that edge on line a.
			x, y := a, b
			switch d {
			case 1:
				x, y = b, a
			case 2:
				y = p.h - 1 - b
			case 3:
				x, y = p.w-1-b, a
			}
			switch p.g[y][x] {
			case '#':
				free = b + 1
			case 'O':
				if free > b {
					return fmt.Errorf("a rock would roll past its own place on line %d", a)
				}
				p.g[y][x] = '.'
				fx, fy := a, free
				switch d {
				case 1:
					fx, fy = free, a
				case 2:
					fy = p.h - 1 - free
				case 3:
					fx, fy = p.w-1-free, a
				}
				p.g[fy][fx], free = 'O', free+1
			}
		}
	}
	return nil
}

// load is the total load on the north beams; hash fingerprints the layout.
func (p *platform) load() (int, uint64) {
	total, h := 0, uint64(14695981039346656037)
	for y := range p.h {
		for x := range p.w {
			h = (h ^ uint64(p.g[y][x])) * 1099511628211
			if p.g[y][x] == 'O' {
				total += p.h - y
			}
		}
	}
	return total, h
}

func day14(path string) error {
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
		return fmt.Errorf("%s: platform is empty or larger than %d", path, maxSide)
	}
	p := &platform{w: len(rows[0]), h: len(rows)}
	for y, r := range rows {
		if len(r) != p.w || strings.Trim(r, ".#O") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		copy(p.g[y][:], r)
	}
	if err := p.tilt(0); err != nil {
		return err
	}
	north, _ := p.load()
	seen := make(map[uint64]int, maxCycles)
	loads := make([]int, 0, maxCycles)
	for c := 0; c < maxCycles; c++ {
		for d := range 4 {
			if err := p.tilt(d); err != nil {
				return err
			}
		}
		l, h := p.load()
		if first, ok := seen[h]; ok {
			// Cycle c repeats cycle first; cycle numbers count from 0.
			period := c - first
			if period < 1 {
				return fmt.Errorf("a cycle of length %d", period)
			}
			at := first + (spins-1-first)%period
			if at < first || at >= len(loads) {
				return fmt.Errorf("cycle %d was not recorded", at)
			}

			if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", north, loads[at]); err != nil {
				return err
			}
			return nil
		}
		seen[h] = c
		loads = append(loads, l)
	}
	return fmt.Errorf("no repeat within %d cycles", maxCycles)
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
