// Day 9: Rope Bridge
//
// Each line of the input moves the head of a rope: "R 4". Each knot
// follows the one ahead: if they are no longer touching (including
// diagonally), it steps one towards it on each axis that differs.
//
// Part 1 counts the positions the tail visits with two knots.
// Part 2 counts them with ten knots.
//
// Visited positions are marked in a fixed grid around the start.
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
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 16
	maxMoves      = 5_000
	side          = 2_048
	knots         = 10
)

// simulate pulls a rope of n knots and counts the tail's positions.
func simulate(lines []string, n int, seen *[side][side]bool) (int, error) {
	if n < 2 || n > knots {
		return 0, fmt.Errorf("rope of %d knots", n)
	}
	*seen = [side][side]bool{}
	var rope [knots][2]int
	for k := range rope {
		rope[k] = [2]int{side / 2, side / 2}
	}
	seen[side/2][side/2] = true
	count := 1
	for _, l := range lines {
		dir, steps, ok := strings.Cut(l, " ")
		s, err := strconv.Atoi(steps)
		d := strings.Index("RLUD", dir)
		if !ok || err != nil || d < 0 || len(dir) != 1 || s < 1 || s > side/4 {
			return 0, fmt.Errorf("invalid move %q", l)
		}
		for range s {
			rope[0][0] += [4]int{1, -1, 0, 0}[d]
			rope[0][1] += [4]int{0, 0, -1, 1}[d]
			for k := 1; k < n; k++ {
				dx, dy := rope[k-1][0]-rope[k][0], rope[k-1][1]-rope[k][1]
				if max(dx, -dx) <= 1 && max(dy, -dy) <= 1 {
					break
				}
				rope[k][0] += max(-1, min(1, dx))
				rope[k][1] += max(-1, min(1, dy))
			}
			t := rope[n-1]
			if t[0] < 0 || t[1] < 0 || t[0] >= side || t[1] >= side {
				return 0, errors.New("the tail left the grid")
			}
			if !seen[t[1]][t[0]] {
				seen[t[1]][t[0]], count = true, count+1
			}
		}
	}
	return count, nil
}

func day9(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxMoves {
		return fmt.Errorf("%s: more than %d moves", path, maxMoves)
	}
	seen := new([side][side]bool)
	short, err := simulate(lines, 2, seen)
	if err != nil {
		return err
	}
	long, err := simulate(lines, knots, seen)
	if err != nil {
		return err
	}
	if long > short {
		return fmt.Errorf("the long tail visited %d places, more than the short tail's %d", long, short)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", short, long); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
