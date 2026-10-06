// Day 10: Elves Look, Elves Say
//
// The input is a starting string of digits for the look-and-say sequence:
// each step replaces every run of a digit with the run's length followed by
// the digit, so 1211 becomes 111221.
//
// Part 1 is the length of the result after 40 steps, part 2 after 50.
//
// The sequence grows by about 30% a step, to a few million digits by step 50.
// Two buffers big enough for that are allocated up front and the steps
// alternate between them. A step at most doubles the length (a run of one
// becomes two digits), which is checked before every step.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 10
	maxDigits     = 16 << 20 // room for the sequence at every step up to 50
)

const (
	part1Steps = 40
	part2Steps = 50
)

// readSeed returns the starting digits.
func readSeed(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	seed := strings.TrimSpace(string(content))
	if seed == "" || strings.Trim(seed, "123456789") != "" {
		return "", fmt.Errorf("%s: seed %q must be non-zero digits", path, seed)
	}
	return seed, nil
}

// step writes the look-and-say successor of in into out's backing array and
// returns it.
func step(in, out []byte) ([]byte, error) {
	if len(in) == 0 {
		return nil, errors.New("empty sequence")
	}
	if 2*len(in) > cap(out) {
		return nil, fmt.Errorf("sequence of %d digits could outgrow the %d byte buffer", len(in), cap(out))
	}

	out = out[:0]
	run := 1
	for i := 1; i <= len(in); i++ {
		if i < len(in) && in[i] == in[i-1] {
			run++
			continue
		}
		// Runs never reach 10 in look-and-say, so the count is one digit.
		if run > 9 {
			return nil, fmt.Errorf("run of %d digits cannot be written as one digit", run)
		}
		out = append(out, byte('0'+run), in[i-1])
		run = 1
	}
	return out, nil
}

func day10(path string) error {
	seed, err := readSeed(path)
	if err != nil {
		return err
	}

	current := make([]byte, 0, maxDigits)
	next := make([]byte, 0, maxDigits)
	current = append(current, seed...)

	var after40, after50 int
	for i := 1; i <= part2Steps; i++ {
		if next, err = step(current, next); err != nil {
			return err
		}
		current, next = next, current
		if i == part1Steps {
			after40 = len(current)
		}
	}
	after50 = len(current)

	if after40 <= len(seed) {
		return fmt.Errorf("sequence did not grow past %d digits", after40)
	}
	if after50 < after40 {
		return fmt.Errorf("sequence shrank from %d to %d digits", after40, after50)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", after40, after50); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
