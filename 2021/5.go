// Day 5: Hydrothermal Venture
//
// Each line of the input is a line of vents, "x1,y1 -> x2,y2": horizontal,
// vertical or at 45 degrees.
//
// Part 1 counts points covered by at least two horizontal or vertical
// lines.
// Part 2 counts the same including the diagonal lines.
//
// Each line is drawn into a fixed grid of counts, one cell at a time.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
	side          = 1_000
)

// segment is one line of vents.
type segment struct {
	x1, y1, x2, y2 int
}

// readSegments parses the lines.
func readSegments(path string) ([]segment, error) {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	segs := make([]segment, len(lines))
	for i, l := range lines {
		s := &segs[i]
		n, err := fmt.Sscanf(l, "%d,%d -> %d,%d", &s.x1, &s.y1, &s.x2, &s.y2)
		if err != nil || n != 4 {
			return nil, fmt.Errorf("%s: malformed line %q", path, l)
		}
		for _, v := range [4]int{s.x1, s.y1, s.x2, s.y2} {
			if v < 0 || v >= side {
				return nil, fmt.Errorf("%s: %q leaves the grid", path, l)
			}
		}
		dx, dy := s.x2-s.x1, s.y2-s.y1
		if dx != 0 && dy != 0 && dx != dy && dx != -dy {
			return nil, fmt.Errorf("%s: %q is not at a multiple of 45 degrees", path, l)
		}
	}
	return segs, nil
}

// overlaps draws the lines (diagonals too, if asked) and counts points
// covered at least twice.
func overlaps(segs []segment, diagonal bool, grid *[side][side]uint16) (int, error) {
	if grid == nil {
		return 0, errors.New("no grid")
	}
	*grid = [side][side]uint16{}
	count := 0
	for _, s := range segs {
		dx, dy := max(-1, min(1, s.x2-s.x1)), max(-1, min(1, s.y2-s.y1))
		if dx != 0 && dy != 0 && !diagonal {
			continue
		}
		steps := max(s.x2-s.x1, s.x1-s.x2, s.y2-s.y1, s.y1-s.y2)
		for k := 0; k <= steps; k++ {
			c := &grid[s.y1+k*dy][s.x1+k*dx]
			*c++
			if *c == 2 {
				count++
			}
		}
	}
	return count, nil
}

func day5(path string) error {
	segs, err := readSegments(path)
	if err != nil {
		return err
	}
	grid := new([side][side]uint16)
	straight, err := overlaps(segs, false, grid)
	if err != nil {
		return err
	}
	all, err := overlaps(segs, true, grid)
	if err != nil {
		return err
	}

	if straight < 1 {
		return errors.New("no straight lines overlap")
	}
	if all < straight {
		return fmt.Errorf("adding diagonals lowered the overlaps from %d to %d", straight, all)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", straight, all); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
