// Day 5: Binary Boarding
//
// Each line of the input is a boarding pass of ten letters: F/B pick the
// row half, L/R the column half. Read as binary (B and R are 1), the pass
// is the seat ID, row*8 + column.
//
// Part 1 is the highest seat ID.
// Part 2 is the one missing ID whose neighbours (ID-1 and ID+1) are both
// taken.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 16
	seats         = 1 << 10
)

func day5(path string) error {
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
	var taken [seats]bool
	highest := -1
	for _, pass := range strings.Fields(string(content)) {
		if len(pass) != 10 || strings.Trim(pass[:7], "FB") != "" || strings.Trim(pass[7:], "LR") != "" {
			return fmt.Errorf("%s: malformed pass %q", path, pass)
		}
		id := 0
		for k := range len(pass) {
			id <<= 1
			if pass[k] == 'B' || pass[k] == 'R' {
				id |= 1
			}
		}
		if taken[id] {
			return fmt.Errorf("%s: seat %d appears twice", path, id)
		}
		taken[id], highest = true, max(highest, id)
	}
	mine := -1
	for id := 1; id < seats-1; id++ {
		if !taken[id] && taken[id-1] && taken[id+1] {
			if mine >= 0 {
				return fmt.Errorf("both %d and %d could be my seat", mine, id)
			}
			mine = id
		}
	}
	if highest < 0 || mine < 0 {
		return errors.New("no passes, or no free seat between two taken ones")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", highest, mine); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
