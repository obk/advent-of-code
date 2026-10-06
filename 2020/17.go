// Day 17: Conway Cubes
//
// The input is a slice of active ('#') and inactive cubes. Each cycle, an
// active cube stays active with 2 or 3 active neighbours, and an inactive
// one becomes active with exactly 3.
//
// Part 1 counts active cubes after six cycles in three dimensions.
// Part 2 does the same in four.
//
// The space is a fixed array wide enough for six cycles of growth either
// side; in three dimensions only the middle w layer is used.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 10
	cycles        = 6
	maxStart      = 12
	side          = maxStart + 2*cycles + 4 // one spare layer each side
	cells         = side * side * side * side
)

// space holds two copies of the cubes, for the current and next cycle.
type space struct {
	cur, next []bool
	start     []string
}

// load reads the starting slice.
func load(path string) (*space, error) {
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
	if len(rows) == 0 || len(rows) > maxStart || !strings.Contains(string(content), "#") {
		return nil, fmt.Errorf("%s: %d rows, expected 1..%d", path, len(rows), maxStart)
	}
	for y, row := range rows {
		if len(row) != len(rows) || strings.Trim(row, ".#") != "" {
			return nil, fmt.Errorf("%s: row %d is not %d of '.' and '#'", path, y, len(rows))
		}
	}
	return &space{cur: make([]bool, cells), next: make([]bool, cells), start: rows}, nil
}

// neighbours counts active cubes around (x, y, z, w).
func (s *space) neighbours(x, y, z, w int, four bool) (int, error) {
	if x < 1 || y < 1 || z < 1 || w < 1 || x > side-2 || y > side-2 || z > side-2 || w > side-2 {
		return 0, fmt.Errorf("cube %d,%d,%d,%d is on the edge", x, y, z, w)
	}
	n, dwLo, dwHi := 0, 0, 0
	if four {
		dwLo, dwHi = -1, 1
	}
	for dw := dwLo; dw <= dwHi; dw++ {
		for dz := -1; dz <= 1; dz++ {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					i := (((w+dw)*side+z+dz)*side+y+dy)*side + x + dx
					if (dx != 0 || dy != 0 || dz != 0 || dw != 0) && s.cur[i] {
						n++
					}
				}
			}
		}
	}
	if n > 80 {
		return 0, fmt.Errorf("cube %d,%d,%d,%d has %d neighbours", x, y, z, w, n)
	}
	return n, nil
}

// run simulates the cycles and counts the active cubes.
func (s *space) run(four bool) (int, error) {
	if len(s.start) > maxStart || len(s.cur) != cells || len(s.next) != cells {
		return 0, errors.New("the space does not fit the starting slice")
	}
	clear(s.cur)
	off, mid := cycles+2, side/2
	for y, row := range s.start {
		if off+len(row) > side-cycles-2 || off+y > side-cycles-2 {
			return 0, fmt.Errorf("row %d does not leave room to grow", y)
		}
		for x := range len(row) {
			s.cur[((mid*side+mid)*side+off+y)*side+off+x] = row[x] == '#'
		}
	}
	wLo, wHi := mid, mid
	if four {
		wLo, wHi = 1, side-2
	}
	for range cycles {
		for w := wLo; w <= wHi; w++ {
			for z := 1; z < side-1; z++ {
				for y := 1; y < side-1; y++ {
					for x := 1; x < side-1; x++ {
						n, err := s.neighbours(x, y, z, w, four)
						if err != nil {
							return 0, err
						}
						i := ((w*side+z)*side+y)*side + x
						s.next[i] = n == 3 || (s.cur[i] && n == 2)
					}
				}
			}
		}
		s.cur, s.next = s.next, s.cur
	}
	active := 0
	for i, c := range s.cur {
		if !c {
			continue
		}
		active++
		for v, k := i, 0; k < 4; v, k = v/side, k+1 {
			if v%side <= 1 || v%side >= side-2 {
				return 0, errors.New("the cubes grew to the edge of the space")
			}
		}
	}
	return active, nil
}

func day17(path string) error {
	s, err := load(path)
	if err != nil {
		return err
	}
	three, err := s.run(false)
	if err != nil {
		return err
	}
	four, err := s.run(true)
	if err != nil {
		return err
	}

	if three < 1 || four < 1 {
		return errors.New("every cube went inactive")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", three, four); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
