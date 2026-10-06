// Day 3: Rucksack Reorganization
//
// Each line of the input is a rucksack's items (a-z, A-Z), split evenly
// into two compartments. Item priorities are 1-26 for a-z, 27-52 for A-Z.
//
// Part 1 sums the priority of the one item type in both compartments.
// Part 2 sums the priority of the one item type in all three rucksacks of
// each group of three lines.
//
// Item sets are 64-bit masks indexed by priority.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 16
	maxSacks      = 1_000
)

// items returns the set of item priorities in s.
func items(s string) (uint64, error) {
	var set uint64
	for k := range len(s) {
		c := s[k]
		switch {
		case c >= 'a' && c <= 'z':
			set |= 1 << (c - 'a' + 1)
		case c >= 'A' && c <= 'Z':
			set |= 1 << (c - 'A' + 27)
		default:
			return 0, fmt.Errorf("invalid item %q", c)
		}
	}
	return set, nil
}

// only returns the priority of the single item in set.
func only(set uint64) (int, error) {
	if bits.OnesCount64(set) != 1 {
		return 0, fmt.Errorf("%d items in common, expected 1", bits.OnesCount64(set))
	}
	return bits.TrailingZeros64(set), nil
}

func day3(path string) error {
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
	sacks := strings.Fields(string(content))
	if len(sacks)%3 != 0 || len(sacks) > maxSacks {
		return fmt.Errorf("%s: %d rucksacks is not whole groups of three", path, len(sacks))
	}
	shared, badges := 0, 0
	group := ^uint64(0)
	for i, s := range sacks {
		if len(s)%2 != 0 {
			return fmt.Errorf("%s: rucksack %q is not two equal halves", path, s)
		}
		a, errA := items(s[:len(s)/2])
		b, errB := items(s[len(s)/2:])
		if errA != nil || errB != nil {
			return fmt.Errorf("%s: invalid rucksack %q", path, s)
		}
		p, err := only(a & b)
		if err != nil {
			return err
		}
		shared += p
		group &= a | b
		if i%3 == 2 {
			if p, err = only(group); err != nil {
				return err
			}
			badges, group = badges+p, ^uint64(0)
		}
	}

	if shared < len(sacks) || badges < len(sacks)/3 {
		return errors.New("an item has priority below 1")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", shared, badges); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
