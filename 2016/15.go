// Day 15: Timing is Everything
//
// Each line of the input describes a spinning disc: "Disc #1 has 5 positions;
// at time=0, it is at position 4." Each disc advances one position a second.
// A capsule dropped at time t reaches disc k at time t+k, and passes only if
// that disc is then at position 0.
//
// Part 1 is the first time t at which the capsule passes every disc.
// Part 2 adds a disc of 11 positions, starting at 0, below the others.
//
// Disc k needs (start + t + k) mod positions = 0. The discs are taken one at
// a time: once t satisfies the first few, stepping t by the least common
// multiple of their sizes keeps them satisfied while searching for the next
// (a sieve form of the Chinese remainder theorem).
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
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 12
	maxLines      = 20
	maxPositions  = 1_000
)

// disc is one disc: its size and its position at time 0.
type disc struct {
	positions, start int
}

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

// parseDisc reads "Disc #1 has 5 positions; at time=0, it is at position 4.".
func parseDisc(line string) (disc, error) {
	w := strings.Fields(strings.TrimSuffix(line, "."))
	if len(w) != 12 || w[0] != "Disc" || w[4] != "positions;" {
		return disc{}, fmt.Errorf("malformed disc %q", line)
	}
	positions, err := strconv.Atoi(w[3])
	if err != nil {
		return disc{}, fmt.Errorf("invalid size in %q: %w", line, err)
	}
	start, err := strconv.Atoi(w[11])
	if err != nil {
		return disc{}, fmt.Errorf("invalid position in %q: %w", line, err)
	}
	if positions < 1 || positions > maxPositions || start < 0 || start >= positions {
		return disc{}, fmt.Errorf("disc %q is out of range", line)
	}
	return disc{positions: positions, start: start}, nil
}

func parseInput(path string) ([]disc, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	discs := make([]disc, 0, len(lines)+1) // room for part 2's extra disc
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		d, err := parseDisc(line)
		if err != nil {
			return nil, err
		}
		discs = append(discs, d)
	}
	if len(discs) == 0 {
		return nil, fmt.Errorf("%s: no discs found", path)
	}
	return discs, nil
}

// gcd is Euclid's algorithm; inputs below 2^63 need fewer than 100 steps.
func gcd(a, b int) (int, error) {
	if a < 1 || b < 1 {
		return 0, fmt.Errorf("gcd of %d and %d is not defined here", a, b)
	}
	for i := 0; i < 100 && b != 0; i++ {
		a, b = b, a%b
	}
	if b != 0 {
		return 0, errors.New("gcd did not converge")
	}
	return a, nil
}

// firstTime returns the earliest drop time that passes every disc.
func firstTime(discs []disc) (int, error) {
	if len(discs) == 0 || len(discs) > maxLines+1 {
		return 0, fmt.Errorf("%d discs is outside 1..%d", len(discs), maxLines)
	}

	t, step := 0, 1
	for k, d := range discs {
		// Within d.positions steps of size step, t mod d.positions cycles
		// through every value it can reach.
		found := false
		for range d.positions {
			if (d.start+t+k+1)%d.positions == 0 {
				found = true
				break
			}
			t += step
		}
		if !found {
			return 0, fmt.Errorf("disc %d can never line up", k+1)
		}
		g, err := gcd(step, d.positions)
		if err != nil {
			return 0, err
		}
		step = step / g * d.positions
		if step > 1<<50 {
			return 0, fmt.Errorf("disc cycle %d is too long", step)
		}
	}

	for k, d := range discs {
		if (d.start+t+k+1)%d.positions != 0 {
			return 0, fmt.Errorf("time %d misses disc %d", t, k+1)
		}
	}
	return t, nil
}

func day15(path string) error {
	discs, err := parseInput(path)
	if err != nil {
		return err
	}
	t, err := firstTime(discs)
	if err != nil {
		return err
	}
	if len(discs) >= cap(discs) {
		return errors.New("no room for the extra disc")
	}
	withExtra, err := firstTime(append(discs, disc{positions: 11, start: 0}))
	if err != nil {
		return err
	}

	if withExtra < t {
		return fmt.Errorf("an extra disc made the wait shorter (%d < %d)", withExtra, t)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", t, withExtra); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
