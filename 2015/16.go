// Day 16: Aunt Sue
//
// Each line of the input lists what you remember about one of 500 Aunts Sue,
// such as "Sue 1: goldfish: 9, cars: 0, samoyeds: 9". Anything not listed is
// unknown, not zero. The gift's analysis gave a fixed reading for ten
// compounds.
//
// Part 1 is the Sue whose remembered values all equal the reading.
// Part 2 reads the cats and trees values as "more than" the reading, and the
// pomeranians and goldfish values as "fewer than"; the rest must still match.
//
// Each Sue is simply checked against the reading under both rules.
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
	inputPath     = "inputs/16.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxCount      = 1_000
)

// reading is one compound in the analysis machine's output.
type reading struct {
	compound string
	value    int
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

// matches reports whether one remembered compound agrees with the reading,
// under part 1's exact rule and under part 2's ranges.
func matches(compound string, value int) (exact, ranged bool, err error) {
	if compound == "" {
		return false, false, errors.New("empty compound name")
	}
	if value < 0 || value > maxCount {
		return false, false, fmt.Errorf("%s count %d is outside 0..%d", compound, value, maxCount)
	}
	readings := [10]reading{
		{"children", 3}, {"cats", 7}, {"samoyeds", 2}, {"pomeranians", 3}, {"akitas", 0},
		{"vizslas", 0}, {"goldfish", 5}, {"trees", 3}, {"cars", 2}, {"perfumes", 1},
	}
	for _, r := range readings {
		if r.compound != compound {
			continue
		}
		exact = value == r.value
		switch compound {
		case "cats", "trees":
			ranged = value > r.value
		case "pomeranians", "goldfish":
			ranged = value < r.value
		default:
			ranged = exact
		}
		return exact, ranged, nil
	}
	return false, false, fmt.Errorf("unknown compound %q", compound)
}

// checkSue parses one line and reports its number and whether it fits each
// part's rules.
func checkSue(line string) (number int, exact, ranged bool, err error) {
	name, list, found := strings.Cut(line, ": ")
	if !found || !strings.HasPrefix(name, "Sue ") {
		return 0, false, false, fmt.Errorf("malformed line %q", line)
	}
	if number, err = strconv.Atoi(strings.TrimPrefix(name, "Sue ")); err != nil {
		return 0, false, false, fmt.Errorf("invalid Sue number in %q: %w", line, err)
	}
	if number < 1 || number > maxLines {
		return 0, false, false, fmt.Errorf("Sue number in %q is outside 1..%d", line, maxLines)
	}
	if list == "" {
		return 0, false, false, fmt.Errorf("nothing remembered about %q", name)
	}

	exact, ranged = true, true
	for _, item := range strings.Split(list, ", ") {
		compound, valueText, found := strings.Cut(item, ": ")
		if !found {
			return 0, false, false, fmt.Errorf("malformed compound %q in %q", item, line)
		}
		value, err := strconv.Atoi(valueText)
		if err != nil {
			return 0, false, false, fmt.Errorf("invalid count in %q: %w", line, err)
		}
		e, r, err := matches(compound, value)
		if err != nil {
			return 0, false, false, err
		}
		exact, ranged = exact && e, ranged && r
	}
	return number, exact, ranged, nil
}

func day16(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	exactSue, rangedSue, exactCount, rangedCount := 0, 0, 0, 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		number, exact, ranged, err := checkSue(line)
		if err != nil {
			return err
		}
		if exact {
			exactSue, exactCount = number, exactCount+1
		}
		if ranged {
			rangedSue, rangedCount = number, rangedCount+1
		}
	}

	if exactCount != 1 {
		return fmt.Errorf("part 1: %d Sues match exactly, expected 1", exactCount)
	}
	if rangedCount != 1 {
		return fmt.Errorf("part 2: %d Sues match the ranges, expected 1", rangedCount)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", exactSue, rangedSue); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day16(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
