// Day 10: Pipe Maze
//
// The input is a grid of pipes (| - L J 7 F), ground '.', and the start S,
// which sits on one big loop.
//
// Part 1 is the number of steps to the farthest point of the loop: half
// its length.
// Part 2 counts the tiles enclosed by the loop.
//
// The loop is walked from S. For part 2, each row is scanned left to
// right: crossing a loop tile that connects north flips inside/outside
// (corners L and J count, 7 and F don't, so a horizontal run is crossed
// once or not at all), and tiles off the loop are counted while inside.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
)

// Connection bits.
const (
	north = 1 << iota
	east
	south
	west
)

// maze is the grid, the start's shape, and the loop.
type maze struct {
	rows   []string
	sx, sy int
	sMask  int
	onLoop [maxSide][maxSide]bool
}

// at returns the connections of (x, y), with S's worked-out shape.
func (m *maze) at(x, y int) int {
	if y < 0 || y >= len(m.rows) || x < 0 || x >= len(m.rows[y]) {
		return 0
	}
	if x == m.sx && y == m.sy {
		return m.sMask
	}
	switch m.rows[y][x] {
	case '|':
		return north | south
	case '-':
		return east | west
	case 'L':
		return north | east
	case 'J':
		return north | west
	case '7':
		return south | west
	case 'F':
		return south | east
	}
	return 0
}

// walk finds S's shape, marks the loop and returns its length.
func (m *maze) walk() (int, error) {
	if m.sy < 0 || m.sy >= len(m.rows) || m.sx < 0 || m.sx >= len(m.rows[m.sy]) {
		return 0, errors.New("S is off the grid")
	}
	m.sMask = 0
	dirs := [4][3]int{{0, -1, south}, {1, 0, west}, {0, 1, north}, {-1, 0, east}}
	for k, d := range dirs {
		if m.at(m.sx+d[0], m.sy+d[1])&d[2] != 0 {
			m.sMask |= 1 << k
		}
	}
	if bitsSet := m.sMask&1 + m.sMask>>1&1 + m.sMask>>2&1 + m.sMask>>3&1; bitsSet != 2 {
		return 0, fmt.Errorf("S connects to %d pipes, not 2", bitsSet)
	}
	x, y, from := m.sx, m.sy, -1
	for length := 1; length <= maxSide*maxSide; length++ {
		m.onLoop[y][x] = true
		k := 4
		for c := 3; c >= 0; c-- {
			if m.at(x, y)&(1<<c) != 0 && c != from {
				k = c
			}
		}
		if k == 4 {
			return 0, fmt.Errorf("the loop breaks at %d,%d", x, y)
		}
		x, y, from = x+dirs[k][0], y+dirs[k][1], (k+2)%4
		if x == m.sx && y == m.sy {
			return length, nil
		}
	}
	return 0, errors.New("the loop never closes")
}

// enclosed counts tiles inside the loop.
func (m *maze) enclosed() (int, error) {
	n, cells := 0, 0
	for y, row := range m.rows {
		inside := false
		cells += len(row)
		for x := range len(row) {
			if m.onLoop[y][x] {
				if m.at(x, y)&north != 0 {
					inside = !inside
				}
			} else if inside {
				n++
			}
		}
	}
	if n >= cells {
		return 0, fmt.Errorf("%d tiles enclosed of %d", n, cells)
	}
	return n, nil
}

func day10(path string) error {
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
	m := &maze{rows: strings.Fields(string(content)), sx: -1}
	if len(m.rows) > maxSide {
		return fmt.Errorf("%s: more than %d rows", path, maxSide)
	}
	for y, row := range m.rows {
		if len(row) > maxSide || strings.Trim(row, "|-LJ7F.S") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if x := strings.IndexByte(row, 'S'); x >= 0 {
			m.sx, m.sy = x, y
		}
	}
	if m.sx < 0 {
		return fmt.Errorf("%s: no S", path)
	}
	length, err := m.walk()
	if err != nil {
		return err
	}

	// A loop on a square grid always has even length.
	if length%2 != 0 {
		return fmt.Errorf("loop of odd length %d", length)
	}
	inside, err := m.enclosed()
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", length/2, inside); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
