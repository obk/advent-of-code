// Day 20: Trench Map
//
// The input is a 512-character enhancement rule, then an image of lit
// ('#') and dark pixels on an infinite dark background. Enhancing replaces
// each pixel by the rule entry indexed by the 3x3 square around it, read as
// a 9-bit number.
//
// Part 1 counts lit pixels after enhancing twice.
// Part 2 counts them after 50 times.
//
// The image grows by one pixel per side each time, so it lives in a fixed
// grid with room for 50 steps. The infinite background is one value that
// is itself enhanced (rule entry 0 when dark, 511 when lit).
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 16
	maxImage      = 128
	steps         = 50
	side          = maxImage + 2*steps + 2
)

// picture is the image, the background, and a buffer for the next step.
type picture struct {
	rule       string
	cur, next  [side][side]bool
	background bool
}

// at returns pixel (x, y), or the background off the grid.
func (p *picture) at(x, y int) bool {
	if x < 0 || y < 0 || x >= side || y >= side {
		return p.background
	}
	return p.cur[y][x]
}

// enhance runs one step.
func (p *picture) enhance() error {
	if len(p.rule) != 512 {
		return errors.New("rule is not 512 characters")
	}
	for y := range side {
		for x := range side {
			idx := 0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					idx <<= 1
					if p.at(x+dx, y+dy) {
						idx |= 1
					}
				}
			}
			p.next[y][x] = p.rule[idx] == '#'
		}
	}
	if p.background {
		p.background = p.rule[511] == '#'
	} else {
		p.background = p.rule[0] == '#'
	}
	// The edges must look like the background, or the image outgrew the grid.
	for k := range side {
		if p.next[0][k] != p.background || p.next[side-1][k] != p.background || p.next[k][0] != p.background || p.next[k][side-1] != p.background {
			return errors.New("the image grew to the edge of the grid")
		}
	}
	p.cur, p.next = p.next, p.cur
	return nil
}

// lit counts lit pixels; it is an error if the background is lit.
func (p *picture) lit() (int, error) {
	if p.background {
		return 0, errors.New("infinitely many pixels are lit")
	}
	n := 0
	for y := range side {
		for x := range side {
			if p.cur[y][x] {
				n++
			}
		}
	}
	if n == side*side {
		return 0, errors.New("the whole grid is lit")
	}
	return n, nil
}

func day20(path string) error {
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
	rule, image, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	rows := strings.Fields(image)
	if !found || len(rule) != 512 || strings.Trim(rule, ".#") != "" || len(rows) == 0 || len(rows) > maxImage {
		return fmt.Errorf("%s: malformed rule or image", path)
	}
	if rule[0] == '#' && rule[511] == '#' {
		return fmt.Errorf("%s: the background would stay lit forever", path)
	}
	p := &picture{rule: rule}
	for y, row := range rows {
		if len(row) != len(rows[0]) || len(row) > maxImage || strings.Trim(row, ".#") != "" {
			return fmt.Errorf("%s: image row %d is malformed", path, y)
		}
		for x := range len(row) {
			p.cur[steps+1+y][steps+1+x] = row[x] == '#'
		}
	}
	var after [2]int
	for s := 1; s <= steps; s++ {
		if err := p.enhance(); err != nil {
			return err
		}
		if s == 2 || s == steps {
			if after[s/steps], err = p.lit(); err != nil {
				return err
			}
		}
	}

	if after[0] < 1 {
		return errors.New("nothing is lit after two steps")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", after[0], after[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
