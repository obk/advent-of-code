// Day 21: Step Counter
//
// The input is a garden: plots '.', rocks '#', and the start S. Each step
// moves to an adjacent plot.
//
// Part 1 counts the plots reachable in exactly 64 steps.
// Part 2 tiles the garden endlessly and counts the plots reachable in
// exactly 26501365 steps.
//
// A plot is reachable in exactly n steps if its shortest distance is at
// most n with the same parity. In these inputs S is in the middle of a
// square garden of side 131 with clear rows and columns through it, and
// 26501365 = 65 + 202300*131. The count grows quadratically in the number
// of whole gardens crossed, so it is fitted from the counts at 65, 196
// and 327 steps (found on a 5x5 tiling) and extended.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 16
	maxSide       = 131
	tiles         = 5
	big           = maxSide * tiles
	target        = 26501365
)

// garden is the map and the search's buffers over the tiling.
type garden struct {
	rows  []string
	n     int
	dist  [big * big]int
	queue [big * big]int
}

// distances fills dist from the start, in the middle of the tiling.
func (g *garden) distances() error {
	side := g.n * tiles
	if g.n < 1 || side > big {
		return errors.New("garden size is out of range")
	}
	for i := range g.dist {
		g.dist[i] = -1
	}
	start := (side/2)*side + side/2
	g.dist[start], g.queue[0] = 0, start
	tail := 1
	for head := 0; head < tail; head++ {
		x, y := g.queue[head]%side, g.queue[head]/side
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := x+d[0], y+d[1]
			if nx < 0 || ny < 0 || nx >= side || ny >= side || g.dist[ny*side+nx] >= 0 || g.rows[ny%g.n][nx%g.n] == '#' {
				continue
			}
			g.dist[ny*side+nx] = g.dist[y*side+x] + 1
			g.queue[tail], tail = ny*side+nx, tail+1
		}
	}
	return nil
}

// reachable counts plots reachable in exactly steps.
func (g *garden) reachable(steps int) (int, error) {
	if steps < 0 || steps > g.n*tiles/2 {
		return 0, fmt.Errorf("%d steps leave the tiling", steps)
	}
	n := 0
	for _, d := range g.dist[:g.n*tiles*g.n*tiles] {
		if d >= 0 && d <= steps && d%2 == steps%2 {
			n++
		}
	}
	return n, nil
}

func day21(path string) error {
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
	g.n = len(g.rows)
	if g.n != maxSide || g.rows[g.n/2][g.n/2] != 'S' {
		return fmt.Errorf("%s: expected a %dx%d garden with S in the middle", path, maxSide, maxSide)
	}
	for y, r := range g.rows {
		if len(r) != g.n || strings.Trim(r, ".#S") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
	}
	if err := g.distances(); err != nil {
		return err
	}
	near, err := g.reachable(64)
	if err != nil {
		return err
	}
	half := g.n / 2
	var f [3]int
	for k := range 3 {
		if f[k], err = g.reachable(half + k*g.n); err != nil {
			return err
		}
	}
	// f(k) = a k^2 + b k + c through the three counts; k = whole gardens.
	k := (target - half) / g.n
	if half+k*g.n != target {
		return errors.New("the step count does not end on a garden edge")
	}
	if (f[2]-2*f[1]+f[0])%2 != 0 {
		return fmt.Errorf("counts %v do not fit a whole-number quadratic", f)
	}
	c, b, a := f[0], (4*f[1]-3*f[0]-f[2])/2, (f[2]-2*f[1]+f[0])/2
	far := a*k*k + b*k + c

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", near, far); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
