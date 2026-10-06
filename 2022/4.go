// Day 4: Camp Cleanup
//
// Each line of the input is a pair of section ranges, "2-4,6-8".
//
// Part 1 counts pairs where one range contains the other.
// Part 2 counts pairs that overlap at all.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 16
	maxPairs      = 10_000
)

func day4(path string) error {
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
	if len(lines) > maxPairs {
		return fmt.Errorf("%s: more than %d pairs", path, maxPairs)
	}
	contained, overlapping := 0, 0
	for _, l := range lines {
		var a1, a2, b1, b2 int
		n, err := fmt.Sscanf(l, "%d-%d,%d-%d", &a1, &a2, &b1, &b2)
		if err != nil || n != 4 || a1 > a2 || b1 > b2 {
			return fmt.Errorf("%s: malformed pair %q", path, l)
		}
		if (a1 <= b1 && b2 <= a2) || (b1 <= a1 && a2 <= b2) {
			contained++
		}
		if a1 <= b2 && b1 <= a2 {
			overlapping++
		}
	}
	if contained > overlapping {
		return fmt.Errorf("%d contained pairs but only %d overlap", contained, overlapping)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", contained, overlapping); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
