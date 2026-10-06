// Day 24: Planet of Discord
//
// The input is a 5x5 grid of bugs ('#') and empty tiles. Each minute a bug
// survives only with exactly one bug next to it, and an empty tile gets a
// bug with one or two next to it.
//
// Part 1 is the biodiversity of the first layout seen twice: the sum of
// 2^i over bug tiles i, numbering tiles across rows.
//
// The biodiversity is also the layout as a 25-bit number, so a bitset over
// all 2^25 layouts records which have been seen.
//
// Part 2 makes the grid recursive: the middle tile is a whole grid one
// level in, and the grid itself is the middle tile of one a level out.
// Edge tiles border the outer level's tiles around its middle, and the
// tiles around the middle border the inner level's whole edge. The answer
// is the number of bugs after 200 minutes, starting from the scan alone.
//
// Each level is a 25-bit number in a fixed array. Bugs spread at most one
// level a minute, so 200 levels either side is enough; the outermost
// levels must stay empty, which is checked.
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
	maxInputBytes = 1 << 8
	side          = 5
	tiles         = side * side
	middle        = tiles / 2
	minutes       = 200
	levels        = 2*minutes + 3
)

// readLayout returns the grid as a 25-bit number.
func readLayout(path string) (int, error) {
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
	rows := strings.Fields(string(content))
	if len(rows) != side {
		return 0, fmt.Errorf("%s: %d rows, expected %d", path, len(rows), side)
	}
	layout := 0
	for y, row := range rows {
		if len(row) != side || strings.Trim(row, ".#") != "" {
			return 0, fmt.Errorf("%s: row %d is not %d of '.' and '#'", path, y, side)
		}
		for x := range side {
			if row[x] == '#' {
				layout |= 1 << (y*side + x)
			}
		}
	}
	return layout, nil
}

// next returns the layout a minute later.
func next(layout int) (int, error) {
	if layout < 0 || layout >= 1<<tiles {
		return 0, fmt.Errorf("layout %d is not 25 bits", layout)
	}
	result := 0
	for i := range tiles {
		x, y := i%side, i/side
		n := 0
		if x > 0 {
			n += layout >> (i - 1) & 1
		}
		if x < side-1 {
			n += layout >> (i + 1) & 1
		}
		if y > 0 {
			n += layout >> (i - side) & 1
		}
		if y < side-1 {
			n += layout >> (i + side) & 1
		}
		bug := layout>>i&1 == 1
		if n == 1 || (!bug && n == 2) {
			result |= 1 << i
		}
	}
	return result, nil
}

// firstRepeat returns the first layout seen twice.
func firstRepeat(layout int, seen []uint64) (int, error) {
	if layout < 0 || layout >= 1<<tiles {
		return 0, fmt.Errorf("layout %d is not 25 bits", layout)
	}
	if len(seen) != 1<<tiles/64 {
		return 0, errors.New("seen set is the wrong size")
	}
	for range 1<<tiles + 1 {
		if seen[layout/64]>>(layout%64)&1 == 1 {
			return layout, nil
		}
		seen[layout/64] |= 1 << (layout % 64)
		var err error
		if layout, err = next(layout); err != nil {
			return 0, err
		}
	}
	return 0, errors.New("no layout repeated")
}

// neighbours counts the bugs next to tile i on level l of grids.
func neighbours(grids *[levels]int, l, i int) (int, error) {
	if l < 1 || l >= levels-1 || i < 0 || i >= tiles || i == middle {
		return 0, fmt.Errorf("tile %d on level %d has no neighbours", i, l)
	}
	x, y, n := i%side, i/side, 0
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		switch {
		case nx < 0 || ny < 0 || nx >= side || ny >= side:
			// Out of this grid: the tile beside the middle on the level out.
			n += grids[l-1] >> ((2+d[1])*side + 2 + d[0]) & 1
		case nx == 2 && ny == 2:
			// Into the middle: the whole facing edge of the level in.
			for k := range side {
				ex, ey := k, 0
				switch d {
				case [2]int{0, -1}:
					ex, ey = k, side-1
				case [2]int{-1, 0}:
					ex, ey = side-1, k
				case [2]int{1, 0}:
					ex, ey = 0, k
				}
				n += grids[l+1] >> (ey*side + ex) & 1
			}
		default:
			n += grids[l] >> (ny*side + nx) & 1
		}
	}
	if n > 8 {
		return 0, fmt.Errorf("tile %d on level %d has %d neighbours", i, l, n)
	}
	return n, nil
}

// recursive runs the recursive grids for some minutes and counts the bugs.
func recursive(layout, mins int, grids, nextGrids *[levels]int) (int, error) {
	if mins < 0 || mins > minutes || layout&(1<<middle) != 0 {
		return 0, fmt.Errorf("%d minutes, or a bug in the middle tile", mins)
	}
	if grids == nil || nextGrids == nil {
		return 0, errors.New("no grids to run")
	}
	*grids, *nextGrids = [levels]int{}, [levels]int{}
	grids[levels/2] = layout
	for range mins {
		for l := 1; l < levels-1; l++ {
			nextGrids[l] = 0
			for i := range tiles {
				if i == middle {
					continue
				}
				n, err := neighbours(grids, l, i)
				if err != nil {
					return 0, err
				}
				if n == 1 || (grids[l]>>i&1 == 0 && n == 2) {
					nextGrids[l] |= 1 << i
				}
			}
		}
		*grids, *nextGrids = *nextGrids, *grids
		if grids[1] != 0 || grids[levels-2] != 0 {
			return 0, errors.New("bugs reached the outermost levels")
		}
	}
	bugs := 0
	for l, g := range grids {
		if g&(1<<middle) != 0 {
			return 0, fmt.Errorf("level %d has a bug in its middle tile", l)
		}
		for i := range tiles {
			bugs += g >> i & 1
		}
	}
	return bugs, nil
}

func day24(path string) error {
	layout, err := readLayout(path)
	if err != nil {
		return err
	}
	repeat, err := firstRepeat(layout, make([]uint64, 1<<tiles/64))
	if err != nil {
		return err
	}

	var grids, nextGrids [levels]int
	bugs, err := recursive(layout, minutes, &grids, &nextGrids)
	if err != nil {
		return err
	}
	if repeat < 0 || repeat >= 1<<tiles {
		return fmt.Errorf("repeated layout %d is not 25 bits", repeat)
	}
	if bugs < 1 {
		return errors.New("the bugs all died out")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", repeat, bugs); err != nil {
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
