// Day 24: Blizzard Basin
//
// The input is a walled valley with blizzards (^ v < >) that move one
// step a minute in their direction, wrapping at the walls. Each minute you
// may move one step or wait, but never share a cell with a blizzard. You
// start in the gap in the top wall and leave by the gap in the bottom.
//
// Part 1 is the fewest minutes to reach the exit.
// Part 2 goes to the exit, back to the start, and to the exit again.
//
// A blizzard at a cell at minute t started t steps behind it, so the
// original map answers whether a cell is free at any minute. The search is
// breadth-first in time: the set of reachable cells, minute by minute.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 14
	maxSide       = 160
	maxMinutes    = 10_000
)

// valley is the map (without walls) and the search's two frontiers.
type valley struct {
	rows      []string // interior rows, walls stripped
	w, h      int
	cur, next [maxSide + 2][maxSide]bool // row 0 is the start, h+1 the exit
}

// open reports whether search cell (x, r) can be stood on at t: the start
// and exit gaps, or a free interior cell.
func (v *valley) open(x, r, t int) bool {
	if x < 0 || x >= v.w || r < 0 || r > v.h+1 {
		return false
	}
	switch r {
	case 0:
		return x == 0
	case v.h + 1:
		return x == v.w-1
	}
	// A blizzard here at minute t started t steps behind, in its direction.
	w, h, y := v.w, v.h, r-1
	return v.rows[y][((x-t)%w+w)%w] != '>' && v.rows[y][(x+t)%w] != '<' &&
		v.rows[((y-t)%h+h)%h][x] != 'v' && v.rows[(y+t)%h][x] != '^'
}

// trip returns the minute at which the goal row is first reached, leaving
// from the given row's gap at minute start.
func (v *valley) trip(fromRow, toRow, start int) (int, error) {
	if fromRow != 0 && fromRow != v.h+1 || toRow != 0 && toRow != v.h+1 {
		return 0, errors.New("trips go between the two gaps")
	}
	if start < 0 {
		return 0, fmt.Errorf("negative start minute %d", start)
	}
	v.cur = [maxSide + 2][maxSide]bool{}
	v.cur[fromRow][min(fromRow, 1)*(v.w-1)] = true
	for t := start + 1; t < start+maxMinutes; t++ {
		v.next = [maxSide + 2][maxSide]bool{}
		for r := 0; r <= v.h+1; r++ {
			for x := range v.w {
				if !v.cur[r][x] {
					continue
				}
				for _, d := range [5][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					if nx, nr := x+d[0], r+d[1]; v.open(nx, nr, t) {
						v.next[nr][nx] = true
					}
				}
			}
		}
		v.cur, v.next = v.next, v.cur
		if v.cur[toRow][min(toRow, 1)*(v.w-1)] {
			return t, nil
		}
	}
	return 0, fmt.Errorf("no way across within %d minutes", maxMinutes)
}

func day24(path string) error {
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
	if len(lines) < 3 || len(lines) > maxSide+2 || len(lines[0]) > maxSide+2 {
		return fmt.Errorf("%s: valley is too small or too large", path)
	}
	v := &valley{h: len(lines) - 2, w: len(lines[0]) - 2, rows: make([]string, len(lines)-2)}
	if !strings.HasPrefix(lines[0], "#.#") || !strings.HasSuffix(lines[len(lines)-1], "#.#") {
		return fmt.Errorf("%s: the gaps are not at the expected corners", path)
	}
	for i, l := range lines[1 : len(lines)-1] {
		if len(l) != v.w+2 || l[0] != '#' || l[len(l)-1] != '#' || strings.Trim(l[1:len(l)-1], ".<>^v") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, i+1)
		}
		v.rows[i] = l[1 : len(l)-1]
	}
	there, err := v.trip(0, v.h+1, 0)
	if err != nil {
		return err
	}
	back, err := v.trip(v.h+1, 0, there)
	if err != nil {
		return err
	}
	again, err := v.trip(0, v.h+1, back)
	if err != nil {
		return err
	}

	if !(there < back && back < again) {
		return fmt.Errorf("trip times %d, %d, %d are out of order", there, back, again)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", there, again); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
