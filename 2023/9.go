// Day 9: Mirage Maintenance
//
// Each line of the input is a sequence of numbers. Taking differences
// repeatedly eventually gives all zeros; working back up extends each
// sequence.
//
// Part 1 sums the next value of every sequence.
// Part 2 sums the value before the first.
//
// The next value is the sum of the last number of every difference row;
// the previous is the alternating sum of their first numbers. The rows
// are computed in place in a fixed array.
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
	maxLen        = 32
	maxLines      = 1_000
)

// extend returns the values after and before a sequence.
func extend(line string) (next, prev int, err error) {
	f := strings.Fields(line)
	if len(f) < 2 || len(f) > maxLen {
		return 0, 0, fmt.Errorf("sequence of %d values is out of range", len(f))
	}
	var row [maxLen]int
	for i, s := range f {
		if row[i], err = strconv.Atoi(s); err != nil {
			return 0, 0, fmt.Errorf("invalid value %q", s)
		}
		if row[i] > 1<<40 || row[i] < -(1<<40) {
			return 0, 0, fmt.Errorf("value %d is too large", row[i])
		}
	}
	n, sign := len(f), 1
	for depth := 0; depth < len(f); depth++ {
		next += row[n-1]
		prev += sign * row[0]
		sign = -sign
		zero := true
		for i := 0; i+1 < n; i++ {
			row[i] = row[i+1] - row[i]
			zero = zero && row[i] == 0
		}
		n--
		if zero {
			return next, prev, nil
		}
	}
	return 0, 0, fmt.Errorf("differences of %q never reach zero", line)
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
	if len(lines) == 0 || len(lines) > maxLines {
		return fmt.Errorf("%s: more than %d sequences", path, maxLines)
	}
	after, before := 0, 0
	for _, l := range lines {
		n, p, err := extend(l)
		if err != nil {
			return err
		}
		after, before = after+n, before+p
	}

	if after == 0 && before == 0 {
		return errors.New("every sequence extends with zeros")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", after, before); err != nil {
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
