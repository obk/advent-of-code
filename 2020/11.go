// Day 11: Seating System
//
// The input is a seat layout: floor '.', empty seats 'L', occupied '#'.
// Each round, an empty seat with no occupied seat in view fills, and an
// occupied seat with too many occupied seats in view empties.
//
// Part 1 looks at the eight adjacent seats and empties at 4 or more.
// Part 2 looks along the eight directions to the first seat and empties at
// 5 or more.
// Each answer is the number of occupied seats once nothing changes.
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
	maxInputBytes = 1 << 16
	maxSide       = 128
	maxRounds     = 1_000
)

// hall is the layout and a buffer for the next round.
type hall struct {
	cur, next [maxSide][maxSide]byte
	w, h      int
	start     []string
}

// load reads the layout.
func load(path string) (*hall, error) {
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
	rows := strings.Fields(string(content))
	if len(rows) == 0 || !strings.Contains(string(content), "L") {
		return nil, fmt.Errorf("%s: no seats", path)
	}
	if len(rows) > maxSide || len(rows[0]) > maxSide {
		return nil, fmt.Errorf("%s: layout is larger than %d", path, maxSide)
	}
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, ".L#") != "" {
			return nil, fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	return &hall{w: len(rows[0]), h: len(rows), start: rows}, nil
}

// seen counts occupied seats in view of (x, y): adjacent ones, or the first
// seat along each direction when far is set.
func (hl *hall) seen(x, y int, far bool) (int, error) {
	if x < 0 || y < 0 || x >= hl.w || y >= hl.h {
		return 0, fmt.Errorf("seat %d,%d is outside the hall", x, y)
	}
	n := 0
	for _, d := range [8][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		cx, cy := x+d[0], y+d[1]
		for k := 0; k < maxSide && cx >= 0 && cy >= 0 && cx < hl.w && cy < hl.h; k++ {
			c := hl.cur[cy][cx]
			if c == '#' {
				n++
			}
			if c != '.' || !far {
				break
			}
			cx, cy = cx+d[0], cy+d[1]
		}
	}
	if n > 8 {
		return 0, fmt.Errorf("seat %d,%d sees %d seats", x, y, n)
	}
	return n, nil
}

// settle runs rounds until nothing changes and counts the occupied seats.
func (hl *hall) settle(far bool, tolerance int) (int, error) {
	if tolerance < 1 || tolerance > 8 {
		return 0, fmt.Errorf("tolerance %d is outside 1..8", tolerance)
	}
	if len(hl.start) != hl.h {
		return 0, errors.New("the starting layout does not match the hall")
	}
	for y, row := range hl.start {
		if len(row) != hl.w {
			return 0, fmt.Errorf("row %d is %d wide, not %d", y, len(row), hl.w)
		}
		copy(hl.cur[y][:], row)
	}
	for range maxRounds {
		changed := false
		for y := range hl.h {
			for x := range hl.w {
				c := hl.cur[y][x]
				hl.next[y][x] = c
				if c == '.' {
					continue
				}
				n, err := hl.seen(x, y, far)
				if err != nil {
					return 0, err
				}
				if c == 'L' && n == 0 {
					hl.next[y][x], changed = '#', true
				} else if c == '#' && n >= tolerance {
					hl.next[y][x], changed = 'L', true
				}
			}
		}
		hl.cur, hl.next = hl.next, hl.cur
		if !changed {
			return hl.occupied()
		}
	}
	return 0, fmt.Errorf("the seats did not settle in %d rounds", maxRounds)
}

// occupied counts the occupied seats.
func (hl *hall) occupied() (int, error) {
	if hl.w > maxSide || hl.h > maxSide {
		return 0, fmt.Errorf("hall of %dx%d is too large", hl.w, hl.h)
	}
	n := 0
	for y := range hl.h {
		for x := range hl.w {
			if hl.cur[y][x] == '#' {
				n++
			}
		}
	}
	return n, nil
}

func day11(path string) error {
	hl, err := load(path)
	if err != nil {
		return err
	}
	near, err := hl.settle(false, 4)
	if err != nil {
		return err
	}
	far, err := hl.settle(true, 5)
	if err != nil {
		return err
	}

	if near < 1 || far < 1 {
		return errors.New("nobody sat down")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", near, far); err != nil {
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
