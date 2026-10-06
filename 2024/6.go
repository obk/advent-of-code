// Day 6: Guard Gallivant
//
// The input is a lab map: obstructions '#', and the guard '^' facing up.
// The guard walks forward, turning right before any obstruction, until
// leaving the map.
//
// Part 1 counts the distinct positions the guard visits.
// Part 2 counts the places where one new obstruction would trap the guard
// in a loop.
//
// A new obstruction only matters on the guard's original path, so each of
// those places is tried. A walk loops if it returns to a position facing
// the same way; visits are stamped with the try number, so the table
// never needs clearing.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
)

// lab is the map and the walk's buffers.
type lab struct {
	rows   []string
	w, h   int
	sx, sy int
	stamp  [maxSide][maxSide][4]int32
	onPath [maxSide][maxSide]bool
}

// walk runs the guard with an extra obstruction at (ox, oy) (or none if
// ox < 0) and reports whether it loops; try stamps this walk's visits.
func (l *lab) walk(ox, oy int, try int32, mark bool) (bool, error) {
	if try <= 0 {
		return false, errors.New("try numbers start at 1")
	}
	x, y, d := l.sx, l.sy, 0
	dx, dy := [4]int{0, 1, 0, -1}, [4]int{-1, 0, 1, 0}
	for step := 0; step <= 4*l.w*l.h; step++ {
		if l.stamp[y][x][d] == try {
			return true, nil
		}
		l.stamp[y][x][d] = try
		if mark {
			l.onPath[y][x] = true
		}
		nx, ny := x+dx[d], y+dy[d]
		if nx < 0 || ny < 0 || nx >= l.w || ny >= l.h {
			return false, nil
		}
		if l.rows[ny][nx] == '#' || (nx == ox && ny == oy) {
			d = (d + 1) % 4
			continue
		}
		x, y = nx, ny
	}
	return false, errors.New("the walk ran past every state")
}

func day6(path string) error {
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
	l := &lab{rows: strings.Fields(string(content)), sx: -1}
	l.h = len(l.rows)
	if l.h == 0 || l.h > maxSide || len(l.rows[0]) > maxSide {
		return fmt.Errorf("%s: map is empty or larger than %d", path, maxSide)
	}
	l.w = len(l.rows[0])
	for y, r := range l.rows {
		if len(r) != l.w || strings.Trim(r, ".#^") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if x := strings.IndexByte(r, '^'); x >= 0 {
			l.sx, l.sy = x, y
		}
	}
	if l.sx < 0 {
		return fmt.Errorf("%s: no guard", path)
	}
	if loops, err := l.walk(-1, -1, 1, true); err != nil || loops {
		return fmt.Errorf("the guard's own path loops (%v)", err)
	}
	visited, traps := 0, 0
	try := int32(1)
	for y := range l.h {
		for x := range l.w {
			if !l.onPath[y][x] {
				continue
			}
			visited++
			if x == l.sx && y == l.sy {
				continue
			}
			try++
			loops, err := l.walk(x, y, try, false)
			if err != nil {
				return err
			}
			if loops {
				traps++
			}
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", visited, traps); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
