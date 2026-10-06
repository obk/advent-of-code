// Day 8: Resonant Collinearity
//
// The input is a map of antennas, each marked by its frequency (a letter
// or digit). Two antennas of the same frequency make antinodes.
//
// Part 1 counts map positions holding an antinode: a point in line with
// the pair, twice as far from one as from the other.
// Part 2 counts positions in line with any pair at all (the antennas
// included).
//
// Antennas are grouped by frequency. For part 2 each pair's step is
// reduced by its gcd so every grid point on the line is found.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 14
	maxSide       = 64
	maxPerFreq    = 16
)

// city is the antennas by frequency and the antinode maps.
type city struct {
	antennas   [128][maxPerFreq][2]int
	count      [128]int
	w, h       int
	near, line [maxSide][maxSide]bool
}

// load reads the map.
func load(path string) (*city, error) {
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
	c := &city{h: len(rows), w: len(rows[0])}
	for y, r := range rows {
		if len(r) != c.w {
			return nil, fmt.Errorf("%s: row %d is %d wide, not %d", path, y, len(r), c.w)
		}
		for x := range c.w {
			f := r[x]
			if f == '.' {
				continue
			}
			if f >= 128 || c.count[f] == maxPerFreq {
				return nil, fmt.Errorf("%s: too many antennas of %q", path, f)
			}
			c.antennas[f][c.count[f]], c.count[f] = [2]int{x, y}, c.count[f]+1
		}
	}
	return c, nil
}

// mark records the antinodes of every ordered pair of like antennas.
func (c *city) mark() error {
	if c.w < 1 || c.h < 1 || c.w > maxSide || c.h > maxSide {
		return errors.New("map size is out of range")
	}
	for f := range c.antennas {
		for i := range c.count[f] {
			for j := range c.count[f] {
				if i == j {
					continue
				}
				a, b := c.antennas[f][i], c.antennas[f][j]
				dx, dy := b[0]-a[0], b[1]-a[1]
				if x, y := b[0]+dx, b[1]+dy; x >= 0 && y >= 0 && x < c.w && y < c.h {
					c.near[y][x] = true
				}
				g, r := max(dx, -dx), max(dy, -dy) // gcd by Euclid's algorithm
				for k := 0; k < 2*maxSide && r != 0; k++ {
					g, r = r, g%r
				}
				if g == 0 {
					return fmt.Errorf("two %c antennas share a place", rune(f))
				}
				for k, x, y := 0, b[0], b[1]; k < maxSide && x >= 0 && y >= 0 && x < c.w && y < c.h; k, x, y = k+1, x+dx/g, y+dy/g {
					c.line[y][x] = true
				}
			}
		}
	}
	return nil
}

func day8(path string) error {
	c, err := load(path)
	if err != nil {
		return err
	}
	if err := c.mark(); err != nil {
		return err
	}
	first, second := 0, 0
	for y := range c.h {
		for x := range c.w {
			if c.near[y][x] {
				first++
			}
			if c.line[y][x] {
				second++
			}
		}
	}
	if first < 1 || second < first {
		return fmt.Errorf("%d antinodes, then %d in line", first, second)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
