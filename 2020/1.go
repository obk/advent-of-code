// Day 1: Report Repair
//
// Each line of the input is an expense entry.
//
// Part 1 multiplies the two entries that sum to 2020.
// Part 2 multiplies the three entries that sum to 2020.
//
// The entries are small, so a table indexed by value says whether a value
// is present: a pair is then one pass, and a triple one pass per entry.
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
	maxInputBytes = 1 << 16
	maxEntries    = 1_000
	target        = 2020
)

// readEntries parses one entry per line; each is between 0 and 2020.
func readEntries(path string) ([]int, error) {
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
	if len(fields) < 3 || len(fields) > maxEntries {
		return nil, fmt.Errorf("%s: %d entries, expected 3..%d", path, len(fields), maxEntries)
	}
	entries := make([]int, len(fields))
	for i, f := range fields {
		entries[i], err = strconv.Atoi(f)
		if err != nil || entries[i] < 0 || entries[i] > target {
			return nil, fmt.Errorf("%s: invalid entry %q", path, f)
		}
	}
	return entries, nil
}

// pair returns the product of two entries (other than skip) summing to sum,
// using count, the number of times each value occurs.
func pair(entries []int, count *[target + 1]int, sum, skip int) (int, bool, error) {
	if sum < 0 || sum > target || count == nil {
		return 0, false, fmt.Errorf("sum %d is out of range", sum)
	}
	if skip < -1 || skip >= len(entries) {
		return 0, false, fmt.Errorf("entry %d to skip does not exist", skip)
	}
	for i, e := range entries {
		other := sum - e
		if i == skip || other < 0 {
			continue
		}
		// The other entry must be a different one: if it has the same
		// value, that value must occur twice (three times if skip has it).
		need := 1
		if other == e {
			need++
		}
		if skip >= 0 && entries[skip] == other {
			need++
		}
		if count[other] >= need {
			return e * other, true, nil
		}
	}
	return 0, false, nil
}

func day1(path string) error {
	entries, err := readEntries(path)
	if err != nil {
		return err
	}
	var count [target + 1]int
	for _, e := range entries {
		count[e]++
	}
	two, ok, err := pair(entries, &count, target, -1)
	if err != nil || !ok {
		return fmt.Errorf("no two entries sum to %d (%v)", target, err)
	}
	three := -1
	for i, e := range entries {
		p, ok, err := pair(entries, &count, target-e, i)
		if err != nil {
			return err
		}
		if ok {
			three = e * p
			break
		}
	}
	if three < 0 {
		return fmt.Errorf("no three entries sum to %d", target)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", two, three); err != nil {
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
