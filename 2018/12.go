// Day 12: Subterranean Sustainability
//
// The input is a row of pots ('#' has a plant, '.' not), starting at pot 0,
// followed by rules: each pot's next state depends on itself and the two
// pots either side, "..#.# => #".
//
// Part 1 is the sum of the numbers of the pots with plants after 20
// generations.
// Part 2 asks the same after fifty billion generations. The plants settle
// into a pattern that just drifts along, so the sum then changes by the same
// amount every generation; once it has done so for a long run, the rest is
// extrapolated.
//
// The rules are a 32-entry table indexed by the five pots as bits. The row
// is a fixed buffer wide enough for the plants to spread a pot each way per
// generation, with pot 0 at an offset in the middle.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 12
	maxPots       = 200
	maxGens       = 1_000
	width         = maxPots + 2*maxGens + 8
	offset        = maxGens + 4 // pot 0 sits here in the buffer
	longRun       = 100         // generations of equal change taken as settled
	target        = 50_000_000_000
)

// farm holds the rules and two rows of pots.
type farm struct {
	rules [32]bool
	a, b  [width]bool
}

// load parses the input into the farm.
func (f *farm) load(path string) error {
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
	if len(lines) < 2 {
		return fmt.Errorf("%s: need an initial state and rules", path)
	}
	initial, found := strings.CutPrefix(lines[0], "initial state: ")
	if !found || len(initial) > maxPots || strings.Trim(initial, ".#") != "" {
		return fmt.Errorf("%s: malformed initial state", path)
	}
	for i := range len(initial) {
		f.a[offset+i] = initial[i] == '#'
	}
	rules := 0
	for _, line := range lines[1:] {
		from, to, found := strings.Cut(strings.TrimSpace(line), " => ")
		if !found {
			continue
		}
		if len(from) != 5 || strings.Trim(from, ".#") != "" || (to != "#" && to != ".") {
			return fmt.Errorf("%s: malformed rule %q", path, line)
		}
		code := 0
		for i := range 5 {
			if from[i] == '#' {
				code |= 1 << (4 - i)
			}
		}
		f.rules[code] = to == "#"
		rules++
	}
	if rules == 0 {
		return fmt.Errorf("%s: no rules", path)
	}
	if f.rules[0] {
		return errors.New("rule ..... => # would fill the infinite row")
	}
	return nil
}

// grow advances the row in a by one generation.
func (f *farm) grow() error {
	if f.a[0] || f.a[1] || f.a[width-1] || f.a[width-2] {
		return errors.New("plants reached the edge of the buffer")
	}
	for i := 2; i < width-2; i++ {
		code := 0
		for k := -2; k <= 2; k++ {
			code <<= 1
			if f.a[i+k] {
				code |= 1
			}
		}
		f.b[i] = f.rules[code]
	}
	f.a = f.b
	return nil
}

// potSum adds up the numbers of the pots with plants.
func (f *farm) potSum() int {
	sum := 0
	for i, p := range f.a {
		if p {
			sum += i - offset
		}
	}
	return sum
}

func day12(path string) error {
	f := new(farm)
	if err := f.load(path); err != nil {
		return err
	}
	part1, prevSum, diff, same := 0, f.potSum(), 0, 0
	gen := 0
	for gen = 1; gen <= maxGens && same < longRun; gen++ {
		if err := f.grow(); err != nil {
			return err
		}
		s := f.potSum()
		if gen == 20 {
			part1 = s
		}
		if s-prevSum == diff {
			same++
		} else {
			diff, same = s-prevSum, 0
		}
		prevSum = s
	}
	gen--
	if same < longRun || gen < 20 {
		return fmt.Errorf("the plants did not settle within %d generations", maxGens)
	}
	if diff < -width*width || diff > width*width {
		return fmt.Errorf("the sum changes by %d a generation, more than the row allows", diff)
	}
	part2 := prevSum + (target-gen)*diff

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", part1, part2); err != nil {
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
