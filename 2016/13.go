// Day 13: A Maze of Twisty Little Cubicles
//
// The input is a favourite number. Location (x, y), for non-negative x and y,
// is open when x*x + 3*x + 2*x*y + y + y*y + favourite has an even number of
// one bits, and a wall otherwise. You start at 1,1.
//
// Part 1 is the fewest steps to reach 31,39.
// Part 2 counts the locations reachable in at most 50 steps.
//
// A breadth-first search over a bounded square of the maze, with the queue
// and distances in fixed arrays.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 10
	side          = 200 // the search stays inside 0..side-1 on both axes
	maxFavourite  = 1_000_000
)

// maze holds the BFS's distances and queue.
type maze struct {
	favourite int
	dist      [side * side]int // -1 until reached
	queue     [side * side]int
}

// readFavourite returns the designer's favourite number.
func readFavourite(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid number: %w", path, err)
	}
	if n < 0 || n > maxFavourite {
		return 0, fmt.Errorf("%s: %d is outside 0..%d", path, n, maxFavourite)
	}
	return n, nil
}

// open reports whether (x, y) is open space inside the searched square.
func (m *maze) open(x, y int) bool {
	if m.favourite < 0 {
		return false
	}
	if x < 0 || y < 0 || x >= side || y >= side {
		return false
	}
	return bits.OnesCount(uint(x*x+3*x+2*x*y+y+y*y+m.favourite))%2 == 0
}

// fill runs the BFS from (sx, sy), recording every distance.
func (m *maze) fill(sx, sy int) error {
	if !m.open(sx, sy) {
		return fmt.Errorf("start %d,%d is a wall", sx, sy)
	}
	if m.favourite < 0 {
		return errors.New("maze has no favourite number")
	}
	for i := range m.dist {
		m.dist[i] = -1
	}

	m.dist[sy*side+sx] = 0
	m.queue[0] = sy*side + sx
	tail := 1
	for head := 0; head < tail; head++ {
		cell := m.queue[head]
		x, y := cell%side, cell/side
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			nx, ny := x+d[0], y+d[1]
			if !m.open(nx, ny) || m.dist[ny*side+nx] >= 0 {
				continue
			}
			if tail >= len(m.queue) {
				return errors.New("queue overflowed")
			}
			m.dist[ny*side+nx] = m.dist[cell] + 1
			m.queue[tail] = ny*side + nx
			tail++
		}
	}
	if m.dist[sy*side+sx] != 0 {
		return errors.New("start lost its distance of 0")
	}
	return nil
}

func day13(path string) error {
	favourite, err := readFavourite(path)
	if err != nil {
		return err
	}

	if favourite > maxFavourite {
		return fmt.Errorf("favourite number %d is too large", favourite)
	}
	m := &maze{favourite: favourite}
	if err := m.fill(1, 1); err != nil {
		return err
	}
	steps := m.dist[39*side+31]
	if steps < 0 {
		return errors.New("part 1: 31,39 cannot be reached")
	}
	near := 0
	for _, d := range m.dist {
		if d >= 0 && d <= 50 {
			near++
		}
	}

	// Fifty steps cannot reach past 51,51, well inside the searched square.
	if near < 1 || near > 51*51 {
		return fmt.Errorf("part 2: %d locations within 50 steps", near)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, near); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
