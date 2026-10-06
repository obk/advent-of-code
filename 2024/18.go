// Day 18: RAM Run
//
// Each line of the input is a byte "x,y" that will fall into a 71x71
// memory grid, corrupting that cell.
//
// Part 1 is the fewest steps from the top left to the bottom right after
// the first 1024 bytes have fallen.
// Part 2 is the first byte after which no path remains, as "x,y".
//
// A breadth-first search answers whether (and how far) the exit is
// reachable after the first n bytes; part 2 binary-searches n.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 16
	side          = 71
	cells         = side * side
	first         = 1024
)

// memory holds when each cell is corrupted and the search's buffers.
type memory struct {
	fall  [cells]int // index of the byte that lands here, or a large number
	bytes [][2]int
	dist  [cells]int
	queue [cells]int
}

// steps returns the shortest path once n bytes have fallen (-1 if none).
func (m *memory) steps(n int) (int, error) {
	if n < 0 || n > len(m.bytes) {
		return 0, fmt.Errorf("%d bytes is out of range", n)
	}
	for i := range m.dist {
		m.dist[i] = -1
	}
	m.dist[0], m.queue[0] = 0, 0
	tail := 1
	for head := 0; head < tail; head++ {
		c := m.queue[head]
		if c == cells-1 {
			return m.dist[c], nil
		}
		x, y := c%side, c/side
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= side || ny >= side || m.dist[ny*side+nx] >= 0 || m.fall[ny*side+nx] < n {
				continue
			}
			m.dist[ny*side+nx] = m.dist[c] + 1
			m.queue[tail], tail = ny*side+nx, tail+1
		}
	}
	return -1, nil
}

func day18(path string) error {
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
	if len(lines) < first || len(lines) > cells {
		return fmt.Errorf("%s: %d bytes, expected %d..%d", path, len(lines), first, cells)
	}
	m := &memory{bytes: make([][2]int, len(lines))}
	for i := range m.fall {
		m.fall[i] = cells + 1
	}
	for i, l := range lines {
		var x, y int
		if n, err := fmt.Sscanf(l, "%d,%d", &x, &y); err != nil || n != 2 || x < 0 || y < 0 || x >= side || y >= side {
			return fmt.Errorf("%s: invalid byte %q", path, l)
		}
		m.bytes[i] = [2]int{x, y}
		m.fall[y*side+x] = min(m.fall[y*side+x], i)
	}
	shortest, err := m.steps(first)
	if err != nil {
		return err
	}
	lo, hi := first, len(lines) // a path exists after lo bytes, none after hi
	for k := 0; k < 32 && hi-lo > 1; k++ {
		mid := (lo + hi) / 2
		s, err := m.steps(mid)
		if err != nil {
			return err
		}
		if s >= 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	if s, err := m.steps(hi); err != nil || s >= 0 || shortest < 0 {
		return errors.New("the bytes never cut the exit off")
	}
	b := m.bytes[hi-1]

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d,%d\n", shortest, b[0], b[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
