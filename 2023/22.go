// Day 22: Sand Slabs
//
// Each line of the input is a brick, "1,0,1~1,2,1": a line of cubes
// between two corners. Bricks fall straight down until they rest on the
// ground (z = 1) or another brick.
//
// Part 1 counts the bricks that could be removed without any other brick
// falling.
// Part 2 sums, over every brick, how many others would fall if it alone
// were removed.
//
// Bricks are settled lowest first over a height map that records the top
// brick at each column; each brick keeps the list of bricks it rests on.
// A brick falls in part 2 if everything it rests on falls; going upwards
// in settling order decides that for each brick in turn.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/22.txt"
	maxInputBytes = 1 << 16
	maxBricks     = 2_000
	maxSupports   = 8
	side          = 10
)

// brick is one brick, and what it rests on once settled.
type brick struct {
	lo, hi [3]int
	under  [maxSupports]int
	nUnder int
}

// readBricks parses the bricks and sorts them by their lowest z.
func readBricks(path string) ([]brick, error) {
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
	lines := strings.Fields(string(content))
	if len(lines) > maxBricks {
		return nil, fmt.Errorf("%s: more than %d bricks", path, maxBricks)
	}
	bricks := make([]brick, 0, len(lines))
	for _, l := range lines {
		var b brick
		n, err := fmt.Sscanf(l, "%d,%d,%d~%d,%d,%d", &b.lo[0], &b.lo[1], &b.lo[2], &b.hi[0], &b.hi[1], &b.hi[2])
		if err != nil || n != 6 || b.lo[2] < 1 || b.hi[0] >= side || b.hi[1] >= side || b.lo[0] < 0 || b.lo[1] < 0 {
			return nil, fmt.Errorf("%s: invalid brick %q", path, l)
		}
		for k := range 3 {
			if b.lo[k] > b.hi[k] {
				return nil, fmt.Errorf("%s: brick %q is backwards", path, l)
			}
		}
		bricks = append(bricks, b)
		for k := len(bricks) - 1; k > 0 && bricks[k-1].lo[2] > bricks[k].lo[2]; k-- {
			bricks[k-1], bricks[k] = bricks[k], bricks[k-1]
		}
	}
	return bricks, nil
}

// settle drops every brick (lowest first) and records what each rests on.
func settle(bricks []brick) error {
	if len(bricks) == 0 || len(bricks) > maxBricks {
		return fmt.Errorf("%d bricks is out of range", len(bricks))
	}
	var height, top [side][side]int // top: brick index + 1 at that height
	for i := range bricks {
		b := &bricks[i]
		rest := 0
		for x := b.lo[0]; x <= b.hi[0]; x++ {
			for y := b.lo[1]; y <= b.hi[1]; y++ {
				rest = max(rest, height[x][y])
			}
		}
		drop := b.lo[2] - rest - 1
		b.lo[2], b.hi[2] = b.lo[2]-drop, b.hi[2]-drop
		for x := b.lo[0]; x <= b.hi[0]; x++ {
			for y := b.lo[1]; y <= b.hi[1]; y++ {
				s, known := top[x][y]-1, false
				for _, u := range b.under[:b.nUnder] {
					known = known || u == s
				}
				if height[x][y] == rest && rest > 0 && !known {
					if b.nUnder == maxSupports {
						return errors.New("a brick rests on too many others")
					}
					b.under[b.nUnder], b.nUnder = s, b.nUnder+1
				}
				height[x][y], top[x][y] = b.hi[2], i+1
			}
		}
	}
	return nil
}

// chain counts the bricks that fall if brick i is removed.
func chain(bricks []brick, i int, fell []bool) (int, error) {
	if i < 0 || i >= len(bricks) || len(fell) != len(bricks) {
		return 0, errors.New("brick or scratch out of range")
	}
	clear(fell)
	fell[i] = true
	count := 0
	for j := i + 1; j < len(bricks); j++ {
		b := &bricks[j]
		falls := b.nUnder > 0
		for _, s := range b.under[:b.nUnder] {
			falls = falls && fell[s]
		}
		if falls {
			fell[j], count = true, count+1
		}
	}
	if count >= len(bricks) {
		return 0, fmt.Errorf("%d bricks fell of %d", count, len(bricks))
	}
	return count, nil
}

func day22(path string) error {
	bricks, err := readBricks(path)
	if err != nil {
		return err
	}
	if err := settle(bricks); err != nil {
		return err
	}
	// A brick only rests on bricks settled before it, so in settling order
	// every brick's supporters come first.
	safe, total := 0, 0
	fell := make([]bool, len(bricks))
	for i := range bricks {
		n, err := chain(bricks, i, fell)
		if err != nil {
			return err
		}
		if n == 0 {
			safe++
		}
		total += n
	}

	if safe < 1 || safe > len(bricks) {
		return fmt.Errorf("%d safe bricks of %d", safe, len(bricks))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", safe, total); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day22(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
