// Day 1: Chronal Calibration
//
// Each line of the input is a frequency change such as "+6" or "-3".
//
// Part 1 is the frequency after every change, starting from zero.
// Part 2 repeats the list over and over and asks for the first frequency
// reached twice.
//
// The frequencies seen are marked in a fixed array offset so that 0 sits in
// the middle; a frequency outside it is an error rather than a guess.
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
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
	maxLines      = 100_000
	maxChange     = 1_000_000
	maxFrequency  = 10_000_000 // part 2 tracks frequencies within ± this
	maxPasses     = 100_000
)

// readChanges parses one change per line.
func readChanges(path string) ([]int, error) {
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
	fields := strings.Fields(strings.ReplaceAll(string(content), ",", " "))
	if len(fields) == 0 || len(fields) > maxLines {
		return nil, fmt.Errorf("%s: %d changes, expected 1..%d", path, len(fields), maxLines)
	}
	changes := make([]int, len(fields))
	for i, f := range fields {
		if changes[i], err = strconv.Atoi(f); err != nil {
			return nil, fmt.Errorf("%s: invalid change %q: %w", path, f, err)
		}
		if changes[i] < -maxChange || changes[i] > maxChange {
			return nil, fmt.Errorf("%s: change %d is outside ±%d", path, changes[i], maxChange)
		}
	}
	return changes, nil
}

// firstTwice returns the first frequency reached twice. seen is scratch
// space for every frequency in ±maxFrequency.
func firstTwice(changes []int, seen []bool) (int, error) {
	if len(seen) != 2*maxFrequency+1 || len(changes) == 0 {
		return 0, errors.New("scratch space or changes are missing")
	}
	f := 0
	seen[maxFrequency] = true
	for range maxPasses {
		for _, c := range changes {
			f += c
			if f < -maxFrequency || f > maxFrequency {
				return 0, fmt.Errorf("frequency %d is beyond ±%d", f, maxFrequency)
			}
			if seen[f+maxFrequency] {
				return f, nil
			}
			seen[f+maxFrequency] = true
		}
	}
	return 0, fmt.Errorf("no repeat within %d passes", maxPasses)
}

func day1(path string) error {
	changes, err := readChanges(path)
	if err != nil {
		return err
	}
	if len(changes) == 0 {
		return errors.New("no changes")
	}
	sum := 0
	for _, c := range changes {
		sum += c
	}

	if sum < -len(changes)*maxChange || sum > len(changes)*maxChange {
		return fmt.Errorf("frequency %d is out of reach", sum)
	}
	twice, err := firstTwice(changes, make([]bool, 2*maxFrequency+1))
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, twice); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
