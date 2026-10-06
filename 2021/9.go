// Day 9: Smoke Basin
//
// The input is a heightmap of digits.
//
// Part 1 sums 1 + height over the low points (lower than every neighbour
// up, down, left and right).
// Part 2 multiplies the sizes of the three largest basins: areas bounded by
// height 9, found by breadth-first flood fill.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 16
	maxSide       = 128
)

// heightmap is the grid and the fill's buffers.
type heightmap struct {
	rows  []string
	w, h  int
	seen  [maxSide * maxSide]bool
	queue [maxSide * maxSide]int
}

// load reads the map.
func load(path string) (*heightmap, error) {
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
	if len(rows) == 0 || len(rows) > maxSide || len(rows[0]) > maxSide {
		return nil, fmt.Errorf("%s: map is empty or larger than %d", path, maxSide)
	}
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, "0123456789") != "" {
			return nil, fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	return &heightmap{rows: rows, w: len(rows[0]), h: len(rows)}, nil
}

// at returns the height at (x, y), or 9 off the map.
func (m *heightmap) at(x, y int) byte {
	if x < 0 || y < 0 || x >= m.w || y >= m.h {
		return '9'
	}
	return m.rows[y][x]
}

// basin returns the size of the basin containing (x, y).
func (m *heightmap) basin(x, y int) (int, error) {
	if m.at(x, y) == '9' {
		return 0, fmt.Errorf("%d,%d is not in a basin", x, y)
	}
	if m.seen[y*m.w+x] {
		return 0, fmt.Errorf("the basin at %d,%d was already filled", x, y)
	}
	m.queue[0], m.seen[y*m.w+x] = y*m.w+x, true
	tail := 1
	for head := 0; head < tail; head++ {
		cx, cy := m.queue[head]%m.w, m.queue[head]/m.w
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := cx+d[0], cy+d[1]
			if m.at(nx, ny) == '9' || m.seen[ny*m.w+nx] {
				continue
			}
			if tail == len(m.queue) {
				return 0, errors.New("fill queue is full")
			}
			m.seen[ny*m.w+nx] = true
			m.queue[tail], tail = ny*m.w+nx, tail+1
		}
	}
	if tail >= m.w*m.h {
		return 0, errors.New("a basin covers the whole map")
	}
	return tail, nil
}

func day9(path string) error {
	m, err := load(path)
	if err != nil {
		return err
	}
	risk := 0
	var top [3]int
	for y := range m.h {
		for x := range m.w {
			c := m.at(x, y)
			if c >= m.at(x-1, y) || c >= m.at(x+1, y) || c >= m.at(x, y-1) || c >= m.at(x, y+1) {
				continue
			}
			risk += int(c-'0') + 1
			size, err := m.basin(x, y)
			if err != nil {
				return err
			}
			// Keep the three largest, largest first.
			for k := range top {
				if size > top[k] {
					size, top[k] = top[k], size
				}
			}
		}
	}
	if risk < 1 {
		return errors.New("no low points")
	}
	if top[2] == 0 {
		return errors.New("fewer than three basins")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", risk, top[0]*top[1]*top[2]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
