// Day 23: Unstable Diffusion
//
// The input is a map of elves ('#'). Each round, every elf with another
// elf among its eight neighbours proposes a step: north, south, west or
// east (trying the directions in an order that rotates each round), to the
// first whose three cells on that side are empty. Elves whose target no
// other elf proposed then move.
//
// Part 1 counts empty ground in the smallest rectangle holding every elf
// after 10 rounds.
// Part 2 is the first round in which no elf moves.
//
// Elves are a list of positions over a fixed grid; proposals are counted
// per target cell, and the counts cleared after each round.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/23.txt"
	maxInputBytes = 1 << 14
	side          = 512
	maxElves      = 4_096
	maxRounds     = 2_000
)

// grove is the elves and the round's buffers.
type grove struct {
	elf    [side][side]bool
	claims [side][side]uint8
	pos    [][2]int
	target [][2]int
}

// propose returns elf i's target (its own place if it stays), trying the
// directions from first.
func (g *grove) propose(i, first int) ([2]int, error) {
	x, y := g.pos[i][0], g.pos[i][1]
	if x < 1 || y < 1 || x >= side-1 || y >= side-1 {
		return [2]int{}, errors.New("an elf reached the edge of the grid")
	}
	var near [3][3]bool
	alone := true
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			near[dy+1][dx+1] = g.elf[y+dy][x+dx]
			alone = alone && (near[dy+1][dx+1] == (dx == 0 && dy == 0))
		}
	}
	if alone {
		return g.pos[i], nil
	}
	// North, south, west, east: the three cells on that side, and the step.
	sides := [4][4][2]int{
		{{-1, -1}, {0, -1}, {1, -1}, {0, -1}},
		{{-1, 1}, {0, 1}, {1, 1}, {0, 1}},
		{{-1, -1}, {-1, 0}, {-1, 1}, {-1, 0}},
		{{1, -1}, {1, 0}, {1, 1}, {1, 0}},
	}
	for k := range 4 {
		s := sides[(first+k)%4]
		if !near[s[0][1]+1][s[0][0]+1] && !near[s[1][1]+1][s[1][0]+1] && !near[s[2][1]+1][s[2][0]+1] {
			return [2]int{x + s[3][0], y + s[3][1]}, nil
		}
	}
	return g.pos[i], nil
}

// round runs one round and returns how many elves moved.
func (g *grove) round(first int) (int, error) {
	if len(g.target) != len(g.pos) {
		return 0, errors.New("target list does not match the elves")
	}
	for i := range g.pos {
		t, err := g.propose(i, first)
		if err != nil {
			return 0, err
		}
		g.target[i] = t
		if t != g.pos[i] && g.claims[t[1]][t[0]] == 4 {
			return 0, errors.New("five elves claimed one cell")
		}
		if t != g.pos[i] {
			g.claims[t[1]][t[0]]++
		}
	}
	moved := 0
	for i, t := range g.target {
		if t == g.pos[i] {
			continue
		}
		if g.claims[t[1]][t[0]] == 1 {
			p := g.pos[i]
			g.elf[p[1]][p[0]], g.elf[t[1]][t[0]] = false, true
			g.pos[i], moved = t, moved+1
		}
	}
	for _, t := range g.target {
		g.claims[t[1]][t[0]] = 0
	}
	return moved, nil
}

// empty counts empty ground in the elves' bounding rectangle.
func (g *grove) empty() (int, error) {
	if len(g.pos) == 0 {
		return 0, errors.New("no elves")
	}
	x1, y1, x2, y2 := side, side, 0, 0
	for _, p := range g.pos {
		x1, y1, x2, y2 = min(x1, p[0]), min(y1, p[1]), max(x2, p[0]), max(y2, p[1])
	}
	return (x2-x1+1)*(y2-y1+1) - len(g.pos), nil
}

func day23(path string) error {
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
	rows := strings.Fields(string(content))
	if len(rows) > side/4 {
		return fmt.Errorf("%s: map is too large", path)
	}
	g := &grove{pos: make([][2]int, 0, maxElves)}
	for y, row := range rows {
		if len(row) > side/4 || strings.Trim(row, ".#") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		for x := range len(row) {
			if row[x] == '#' && len(g.pos) < maxElves {
				p := [2]int{x + 3*side/8, y + 3*side/8}
				g.elf[p[1]][p[0]], g.pos = true, append(g.pos, p)
			}
		}
	}
	g.target = make([][2]int, len(g.pos))
	ground, still := 0, 0
	for r := 0; r < maxRounds && still == 0; r++ {
		moved, err := g.round(r % 4)
		if err != nil {
			return err
		}
		if r == 9 {
			if ground, err = g.empty(); err != nil {
				return err
			}
		}
		if moved == 0 {
			still = r + 1
		}
	}
	if ground < 0 {
		return fmt.Errorf("negative empty ground %d", ground)
	}
	if still < 10 {
		return fmt.Errorf("the elves stopped at round %d (needed past 10)", still)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", ground, still); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day23(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
