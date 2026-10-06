// Day 3: Crossed Wires
//
// The input is two wires, each a path of moves from a central port:
// "R8,U5,L5,D3". The wires cross at some points (the port itself does not
// count).
//
// Part 1 is the taxicab distance from the port to the nearest crossing.
// Part 2 is the fewest combined steps both wires take to reach a crossing.
//
// Each wire is a list of axis-aligned segments. Every horizontal segment of
// one wire is checked against every vertical segment of the other; a few
// hundred segments each makes that cheap.
package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 20
	maxMoves      = 2_000
	maxLength     = 1_000_000
)

// segment is one straight piece of a wire, from (x1,y1) to (x2,y2), with the
// wire length walked before it starts.
type segment struct {
	x1, y1, x2, y2, before int
}

// parseWire turns "R8,U5" into segments.
func parseWire(line string) ([]segment, error) {
	moves := strings.Split(strings.TrimSpace(line), ",")
	if len(moves) == 0 || len(moves) > maxMoves {
		return nil, fmt.Errorf("wire of %d moves is outside 1..%d", len(moves), maxMoves)
	}
	segs := make([]segment, len(moves))
	x, y, walked := 0, 0, 0
	for i, m := range moves {
		if len(m) < 2 {
			return nil, fmt.Errorf("malformed move %q", m)
		}
		n, err := strconv.Atoi(m[1:])
		if err != nil || n < 1 || n > maxLength {
			return nil, fmt.Errorf("invalid move %q", m)
		}
		dx, dy := 0, 0
		switch m[0] {
		case 'R':
			dx = n
		case 'L':
			dx = -n
		case 'U':
			dy = n
		case 'D':
			dy = -n
		default:
			return nil, fmt.Errorf("unknown direction in %q", m)
		}
		segs[i] = segment{x, y, x + dx, y + dy, walked}
		x, y, walked = x+dx, y+dy, walked+n
	}
	return segs, nil
}

// readWires returns the two wires.
func readWires(path string) ([]segment, []segment, error) {
	if path == "" {
		return nil, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Fields(string(content))
	if len(lines) != 2 {
		return nil, nil, fmt.Errorf("%s: %d wires, expected 2", path, len(lines))
	}
	a, err := parseWire(lines[0])
	if err != nil {
		return nil, nil, err
	}
	b, err := parseWire(lines[1])
	if err != nil {
		return nil, nil, err
	}
	return a, b, nil
}

// cross returns where horizontal h meets vertical v, if they do.
func cross(h, v segment) (x, y int, ok bool) {
	if h.y1 != h.y2 || v.x1 != v.x2 {
		return 0, 0, false
	}
	x, y = v.x1, h.y1
	ok = x >= min(h.x1, h.x2) && x <= max(h.x1, h.x2) && y >= min(v.y1, v.y2) && y <= max(v.y1, v.y2)
	return x, y, ok && (x != 0 || y != 0)
}

// nearest returns the closest crossing to the port, and the fewest combined
// steps to any crossing.
func nearest(a, b []segment) (int, int, error) {
	if len(a) == 0 || len(b) == 0 {
		return 0, 0, errors.New("a wire is empty")
	}
	if len(a) > maxMoves || len(b) > maxMoves {
		return 0, 0, fmt.Errorf("wires of %d and %d moves are too long", len(a), len(b))
	}
	best, fewest := -1, -1
	for _, pair := range [2][2][]segment{{a, b}, {b, a}} {
		for _, h := range pair[0] {
			for _, v := range pair[1] {
				if x, y, ok := cross(h, v); ok {
					d := max(x, -x) + max(y, -y)
					if best < 0 || d < best {
						best = d
					}
					steps := h.before + max(x-h.x1, h.x1-x) + v.before + max(y-v.y1, v.y1-y)
					if fewest < 0 || steps < fewest {
						fewest = steps
					}
				}
			}
		}
	}
	if best < 0 {
		return 0, 0, errors.New("the wires never cross")
	}
	return best, fewest, nil
}

func day3(path string) error {
	a, b, err := readWires(path)
	if err != nil {
		return err
	}
	d, steps, err := nearest(a, b)
	if err != nil {
		return err
	}

	if d < 1 {
		return fmt.Errorf("nearest crossing %d is at the port", d)
	}
	if steps < d {
		return fmt.Errorf("%d steps cannot reach a crossing %d away", steps, d)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", d, steps); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
