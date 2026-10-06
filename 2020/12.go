// Day 12: Rain Risk
//
// Each line of the input is a navigation step: N/S/E/W move, L/R turn by
// a multiple of 90 degrees, F moves forward.
//
// Part 1 moves the ship itself, starting facing east.
// Part 2 moves a waypoint instead (starting 10 east, 1 north of the ship):
// N/S/E/W move the waypoint, L/R rotate it about the ship, and F moves the
// ship to the waypoint that many times.
// Each answer is the ship's Manhattan distance from its start.
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
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 14
	maxSteps      = 10_000
	maxAmount     = 10_000
)

// step is one navigation instruction.
type step struct {
	action byte
	amount int
}

// readSteps parses the instructions; turns must be multiples of 90.
func readSteps(path string) ([]step, error) {
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
	fields := strings.Fields(string(content))
	if len(fields) > maxSteps {
		return nil, fmt.Errorf("%s: more than %d steps", path, maxSteps)
	}
	steps := make([]step, len(fields))
	for i, f := range fields {
		n, err := strconv.Atoi(f[1:])
		if err != nil || n < 0 || n > maxAmount || !strings.Contains("NSEWLRF", f[:1]) {
			return nil, fmt.Errorf("%s: invalid step %q", path, f)
		}
		if (f[0] == 'L' || f[0] == 'R') && n%90 != 0 {
			return nil, fmt.Errorf("%s: turn %q is not a multiple of 90", path, f)
		}
		steps[i] = step{f[0], n}
	}
	return steps, nil
}

// navigate follows the steps. With a waypoint, N/S/E/W move the waypoint;
// without, they move the ship and the "waypoint" is just the heading.
func navigate(steps []step, wx, wy int, waypoint bool) (int, error) {
	if len(steps) == 0 {
		return 0, errors.New("no steps")
	}
	x, y := 0, 0
	for _, s := range steps {
		if s.amount < 0 || s.amount > maxAmount {
			return 0, fmt.Errorf("step amount %d is out of range", s.amount)
		}
		dx, dy := 0, 0
		switch s.action {
		case 'N':
			dy = s.amount
		case 'S':
			dy = -s.amount
		case 'E':
			dx = s.amount
		case 'W':
			dx = -s.amount
		case 'L', 'R':
			turns := s.amount / 90 % 4
			if s.action == 'R' {
				turns = (4 - turns) % 4
			}
			for range turns { // a quarter turn left
				wx, wy = -wy, wx
			}
		case 'F':
			x, y = x+s.amount*wx, y+s.amount*wy
		default:
			return 0, fmt.Errorf("unknown action %q", s.action)
		}
		if waypoint {
			wx, wy = wx+dx, wy+dy
		} else {
			x, y = x+dx, y+dy
		}
	}
	return max(x, -x) + max(y, -y), nil
}

func day12(path string) error {
	steps, err := readSteps(path)
	if err != nil {
		return err
	}
	ship, err := navigate(steps, 1, 0, false)
	if err != nil {
		return err
	}
	way, err := navigate(steps, 10, 1, true)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", ship, way); err != nil {
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
