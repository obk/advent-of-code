// Day 12: Garden Groups
//
// The input is a map of garden plots, each labelled with its plant. A
// region is a connected group of plots with the same plant.
//
// Part 1 sums area times perimeter over regions.
// Part 2 sums area times the number of sides (straight runs of fence).
//
// Regions are flood-filled with a stack. A region has as many sides as
// corners, and each plot contributes its outer and inner corners.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
	cells         = maxSide * maxSide
)

// garden is the map and the fill's buffers.
type garden struct {
	rows  []string
	w, h  int
	done  [cells]bool
	stack [cells]int
}

// same reports whether (x, y) is on the map and holds plant p.
func (g *garden) same(x, y int, p byte) bool {
	return x >= 0 && y >= 0 && x < g.w && y < g.h && g.rows[y][x] == p
}

// corners counts the corners at plot (x, y) of plant p.
func (g *garden) corners(x, y int, p byte) (int, error) {
	if !g.same(x, y, p) {
		return 0, fmt.Errorf("plot %d,%d does not hold %q", x, y, p)
	}
	n := 0
	for _, d := range [4][2]int{{1, 1}, {1, -1}, {-1, 1}, {-1, -1}} {
		a, b, c := g.same(x+d[0], y, p), g.same(x, y+d[1], p), g.same(x+d[0], y+d[1], p)
		if (!a && !b) || (a && b && !c) {
			n++
		}
	}
	if n > 4 {
		return 0, fmt.Errorf("plot %d,%d has %d corners", x, y, n)
	}
	return n, nil
}

// region fills the region at (x, y) and returns area, perimeter, sides.
func (g *garden) region(x, y int) (area, perimeter, sides int, err error) {
	if x < 0 || y < 0 || x >= g.w || y >= g.h || g.done[y*g.w+x] {
		return 0, 0, 0, fmt.Errorf("plot %d,%d is not a new region", x, y)
	}
	p := g.rows[y][x]
	g.stack[0], g.done[y*g.w+x] = y*g.w+x, true
	n := 1
	for steps := 0; n > 0 && steps < cells; steps++ {
		n--
		cx, cy := g.stack[n]%g.w, g.stack[n]/g.w
		area++
		c, err := g.corners(cx, cy, p)
		if err != nil {
			return 0, 0, 0, err
		}
		sides += c
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := cx+d[0], cy+d[1]
			if !g.same(nx, ny, p) {
				perimeter++
				continue
			}
			if !g.done[ny*g.w+nx] && n == len(g.stack) {
				return 0, 0, 0, errors.New("fill stack is full")
			}
			if !g.done[ny*g.w+nx] {
				g.done[ny*g.w+nx] = true
				g.stack[n], n = ny*g.w+nx, n+1
			}
		}
	}
	if sides < 4 || perimeter < 4 || sides > perimeter {
		return 0, 0, 0, fmt.Errorf("region of %d plots has %d sides and %d fence", area, sides, perimeter)
	}
	return area, perimeter, sides, nil
}

func day12(path string) error {
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
	g := &garden{rows: strings.Fields(string(content))}
	g.h = len(g.rows)
	if g.h == 0 || g.h > maxSide || len(g.rows[0]) > maxSide {
		return fmt.Errorf("%s: map is empty or larger than %d", path, maxSide)
	}
	g.w = len(g.rows[0])
	fence, bulk := 0, 0
	for y, r := range g.rows {
		if len(r) != g.w {
			return fmt.Errorf("%s: row %d is %d wide, not %d", path, y, len(r), g.w)
		}
		for x := range g.w {
			if g.done[y*g.w+x] {
				continue
			}
			a, p, s, err := g.region(x, y)
			if err != nil {
				return err
			}
			fence, bulk = fence+a*p, bulk+a*s
		}
	}

	if fence < 1 || bulk > fence {
		return fmt.Errorf("prices %d and %d are implausible", fence, bulk)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", fence, bulk); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
