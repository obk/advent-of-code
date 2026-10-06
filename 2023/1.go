// Day 1: Trebuchet?!
//
// Each line of the input holds a calibration value: its first and last
// digit, read as a two-digit number.
//
// Part 1 sums the values using only the digits 1-9.
// Part 2 also counts spelled digits ("one" to "nine"), which may overlap
// ("eightwo" holds 8 and 2).
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 16
	maxLines      = 10_000
)

// digitAt returns the digit starting at position i of s (or -1).
func digitAt(s string, i int, words bool) (int, error) {
	if i < 0 || i >= len(s) {
		return 0, fmt.Errorf("position %d is outside %q", i, s)
	}
	if s[i] >= '0' && s[i] <= '9' {
		return int(s[i] - '0'), nil
	}
	if !words {
		return -1, nil
	}
	for d, w := range [9]string{"one", "two", "three", "four", "five", "six", "seven", "eight", "nine"} {
		if strings.HasPrefix(s[i:], w) {
			return d + 1, nil
		}
	}
	return -1, nil
}

// value returns a line's calibration value.
func value(line string, words bool) (int, error) {
	if line == "" {
		return 0, errors.New("empty line")
	}
	first, last := -1, -1
	for i := range len(line) {
		d, err := digitAt(line, i, words)
		if err != nil {
			return 0, err
		}
		if d >= 0 {
			if first < 0 {
				first = d
			}
			last = d
		}
	}
	if first < 0 {
		return 0, fmt.Errorf("no digit in %q", line)
	}
	if first > 9 || last > 9 {
		return 0, fmt.Errorf("digits %d and %d are not single digits", first, last)
	}
	return 10*first + last, nil
}

func day1(path string) error {
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
	lines := strings.Fields(string(content))
	if len(lines) == 0 || len(lines) > maxLines {
		return fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	plain, spelled := 0, 0
	for _, l := range lines {
		if a, err := value(l, false); err == nil {
			plain += a // the part 2 examples have lines without plain digits
		}
		b, err := value(l, true)
		if err != nil {
			return err
		}
		spelled += b
	}

	if spelled < 1 {
		return errors.New("no calibration values")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", plain, spelled); err != nil {
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
