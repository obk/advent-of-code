// Day 17: Pyroclastic Flow
//
// The input is a pattern of jets ('<' and '>'). Five rock shapes fall in
// turn into a chamber 7 wide, appearing two from the left wall and three
// above the tower. Each is pushed by the next jet (if it can move), then
// falls one unit, until it cannot fall and comes to rest.
//
// Part 1 is the tower's height after 2022 rocks.
// Part 2 is its height after 1000000000000 rocks.
//
// Rows are 7-bit masks, bit 6 being the leftmost column. Part 2 finds a
// cycle: once a rock starts with the same shape, the same jet and the same
// top 32 rows as an earlier rock, everything repeats from there, adding
// the same height per cycle.
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
	maxInputBytes = 1 << 14
	maxRocks      = 20_000
	maxRows       = 4*maxRocks + 8
	topRows       = 32
	width         = 7
)

// rock is a shape, bottom row first, aligned to the left wall.
type rock struct {
	rows [4]uint8
	h, w int
}

// key identifies a moment for cycle finding.
type key struct {
	rock, jet int
	top       [topRows / 8]uint64
}

// chamber is the tower and the history of its height.
type chamber struct {
	rows   [maxRows]uint8
	top    int   // the tower's height
	height []int // tower height after n rocks
	jets   string
	jet    int
}

// fits reports whether r fits with its left edge at x and bottom at y.
func (c *chamber) fits(r rock, x, y int) bool {
	if x < 0 || x+r.w > width || y < 0 || y+r.h > maxRows {
		return false
	}
	for k := range r.h {
		if c.rows[y+k]&(r.rows[k]>>x) != 0 {
			return false
		}
	}
	return true
}

// drop lets one rock fall until it rests.
func (c *chamber) drop(r rock) error {
	if len(c.jets) == 0 {
		return errors.New("no jets")
	}
	x, y := 2, c.top+3
	if !c.fits(r, x, y) {
		return errors.New("no room for a new rock")
	}
	for range maxRows {
		dx := 1
		if c.jets[c.jet] == '<' {
			dx = -1
		}
		c.jet = (c.jet + 1) % len(c.jets)
		if c.fits(r, x+dx, y) {
			x += dx
		}
		if !c.fits(r, x, y-1) {
			for k := range r.h {
				c.rows[y+k] |= r.rows[k] >> x
			}
			c.top = max(c.top, y+r.h)
			if c.top > maxRows-8 {
				return errors.New("the tower outgrew the chamber")
			}
			return nil
		}
		y--
	}
	return errors.New("the rock never came to rest")
}

// state returns the cycle key before dropping rock n.
func (c *chamber) state(n int) (key, error) {
	if n < 0 {
		return key{}, fmt.Errorf("rock %d does not exist", n)
	}
	if c.top < 0 || c.top > maxRows {
		return key{}, fmt.Errorf("tower height %d is out of range", c.top)
	}
	k := key{rock: n % 5, jet: c.jet}
	for i := range topRows {
		row := uint64(0)
		if c.top-1-i >= 0 {
			row = uint64(c.rows[c.top-1-i])
		}
		k.top[i/8] |= row << (8 * (i % 8))
	}
	return k, nil
}

// heightAfter returns the tower's height after n rocks, finding a cycle
// if n is beyond what is simulated.
func (c *chamber) heightAfter(n int, seen map[key]int) (int, error) {
	if seen == nil || n < 0 {
		return 0, errors.New("no cycle table, or a negative rock count")
	}
	rocks := [5]rock{
		{rows: [4]uint8{0b1111000}, h: 1, w: 4},
		{rows: [4]uint8{0b0100000, 0b1110000, 0b0100000}, h: 3, w: 3},
		{rows: [4]uint8{0b1110000, 0b0010000, 0b0010000}, h: 3, w: 3},
		{rows: [4]uint8{0b1000000, 0b1000000, 0b1000000, 0b1000000}, h: 4, w: 1},
		{rows: [4]uint8{0b1100000, 0b1100000}, h: 2, w: 2},
	}
	clear(seen)
	c.height = c.height[:1]
	for i := 0; i < maxRocks; i++ {
		if i == n {
			return c.height[n], nil
		}
		k, err := c.state(i)
		if err != nil {
			return 0, err
		}
		if start, ok := seen[k]; ok {
			cycle, gain := i-start, c.height[i]-c.height[start]
			if cycle < 1 || gain < 0 {
				return 0, fmt.Errorf("cycle of %d rocks gaining %d", cycle, gain)
			}
			q, r := (n-start)/cycle, (n-start)%cycle
			if start+r >= len(c.height) {
				return 0, errors.New("the cycle's remainder was not simulated")
			}
			if gain > 0 && q > (1<<62)/gain {
				return 0, errors.New("the height overflows")
			}
			return c.height[start+r] + q*gain, nil
		}
		seen[k] = i
		if err := c.drop(rocks[i%5]); err != nil {
			return 0, err
		}
		c.height = append(c.height, c.top)
	}
	return 0, fmt.Errorf("no cycle within %d rocks", maxRocks)
}

func day17(path string) error {
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
	jets := strings.TrimSpace(string(content))
	if jets == "" || strings.Trim(jets, "<>") != "" {
		return fmt.Errorf("%s: jets must be '<' and '>'", path)
	}
	seen := make(map[key]int, maxRocks)
	var answers [2]int
	for k, n := range [2]int{2022, 1_000_000_000_000} {
		c := &chamber{jets: jets, height: make([]int, 1, maxRocks+1)}
		if answers[k], err = c.heightAfter(n, seen); err != nil {
			return err
		}
	}

	if answers[1] < answers[0] {
		return fmt.Errorf("more rocks made a lower tower: %d < %d", answers[1], answers[0])
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
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
