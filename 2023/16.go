// Day 16: The Floor Will Be Lava
//
// The input is a grid of empty space '.', mirrors '/' and '\', and
// splitters '|' and '-'. A beam passes through '.', turns at mirrors, and
// at a splitter it meets side-on it splits both ways.
//
// Part 1 counts the tiles energized by a beam entering the top-left tile
// heading right.
// Part 2 is the most energized tiles over every beam entering from any
// edge tile, heading inwards.
//
// Beams are followed with a stack; each tile records which directions a
// beam has crossed it in, so no beam is followed twice.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 16
	maxSide       = 128
	maxStack      = 4 * maxSide * maxSide
)

// contraption is the grid and the trace's buffers.
type contraption struct {
	rows  []string
	seen  [maxSide][maxSide]uint8 // bit d set: crossed heading d
	stack [maxStack][3]int
}

// energize traces a beam entering (x, y) heading d (0 right, 1 down,
// 2 left, 3 up) and counts the energized tiles.
func (c *contraption) energize(x, y, d int) (int, error) {
	if d < 0 || d > 3 || len(c.rows) == 0 {
		return 0, fmt.Errorf("direction %d does not exist, or no grid", d)
	}
	c.seen = [maxSide][maxSide]uint8{}
	h, w := len(c.rows), len(c.rows[0])
	c.stack[0] = [3]int{x, y, d}
	n, count := 1, 0
	for steps := 0; n > 0 && steps < maxStack*4; steps++ {
		n--
		x, y, d := c.stack[n][0], c.stack[n][1], c.stack[n][2]
		if x < 0 || y < 0 || x >= w || y >= h || c.seen[y][x]>>d&1 == 1 {
			continue
		}
		if c.seen[y][x] == 0 {
			count++
		}
		c.seen[y][x] |= 1 << d
		outs, k := [2]int{d, d}, 1 // straight on, unless turned or split
		switch tile := c.rows[y][x]; {
		case tile == '/':
			outs[0] = [4]int{3, 2, 1, 0}[d]
		case tile == '\\':
			outs[0] = [4]int{1, 0, 3, 2}[d]
		case tile == '|' && d%2 == 0: // met side-on: up and down
			outs, k = [2]int{1, 3}, 2
		case tile == '-' && d%2 == 1: // met side-on: left and right
			outs, k = [2]int{0, 2}, 2
		}
		for _, o := range outs[:k] {
			if n == maxStack {
				return 0, errors.New("beam stack is full")
			}
			c.stack[n], n = [3]int{x + [4]int{1, 0, -1, 0}[o], y + [4]int{0, 1, 0, -1}[o], o}, n+1
		}
	}
	return count, nil
}

func day16(path string) error {
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
	c := &contraption{rows: strings.Fields(string(content))}
	if len(c.rows) == 0 || len(c.rows) > maxSide || len(c.rows[0]) > maxSide {
		return fmt.Errorf("%s: grid is empty or larger than %d", path, maxSide)
	}
	for y, r := range c.rows {
		if len(r) != len(c.rows[0]) || strings.Trim(r, `./\|-`) != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	first, err := c.energize(0, 0, 0)
	if err != nil {
		return err
	}
	h, w := len(c.rows), len(c.rows[0])
	best := 0
	for k := range max(w, h) {
		for _, s := range [4][3]int{{0, k, 0}, {w - 1, k, 2}, {k, 0, 1}, {k, h - 1, 3}} {
			if s[0] >= w || s[1] >= h {
				continue
			}
			e, err := c.energize(s[0], s[1], s[2])
			if err != nil {
				return err
			}
			best = max(best, e)
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, best); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
