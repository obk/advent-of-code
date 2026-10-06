// Day 13: Claw Contraption
//
// The input is claw machines: button A moves the claw by (ax, ay) for 3
// tokens, button B by (bx, by) for 1 token, and the prize is at (px, py).
//
// Part 1 sums the fewest tokens to win every prize that can be won.
// Part 2 adds 10000000000000 to each prize's coordinates first.
//
// a*ax + b*bx = px and a*ay + b*by = py is a 2x2 linear system; Cramer's
// rule gives its one solution, which counts if both presses are
// non-negative whole numbers.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 16
	offset        = 10_000_000_000_000
	maxMachines   = 1_000
)

// machine is one claw machine.
type machine struct {
	ax, ay, bx, by, px, py int
}

// tokens returns the tokens to win, or 0 if the prize cannot be won.
func tokens(m machine, extra int) (int, error) {
	if extra < 0 {
		return 0, errors.New("negative offset")
	}
	px, py := m.px+extra, m.py+extra
	det := m.ax*m.by - m.ay*m.bx
	if det == 0 {
		return 0, fmt.Errorf("machine %v has parallel buttons", m)
	}
	an, bn := px*m.by-py*m.bx, m.ax*py-m.ay*px
	if an%det != 0 || bn%det != 0 || an/det < 0 || bn/det < 0 {
		return 0, nil
	}
	a, b := an/det, bn/det
	if a*m.ax+b*m.bx != px || a*m.ay+b*m.by != py {
		return 0, errors.New("the solution does not reach the prize")
	}
	return 3*a + b, nil
}

func day13(path string) error {
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
	blocks := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(blocks) > maxMachines {
		return fmt.Errorf("%s: more than %d machines", path, maxMachines)
	}
	var answers [2]int
	for _, b := range blocks {
		var m machine
		n, err := fmt.Sscanf(b, "Button A: X+%d, Y+%d\nButton B: X+%d, Y+%d\nPrize: X=%d, Y=%d", &m.ax, &m.ay, &m.bx, &m.by, &m.px, &m.py)
		if err != nil || n != 6 {
			return fmt.Errorf("%s: malformed machine %q", path, b)
		}
		for k, extra := range [2]int{0, offset} {
			t, err := tokens(m, extra)
			if err != nil {
				return err
			}
			answers[k] += t
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
