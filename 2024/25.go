// Day 25: Code Chronicle
//
// The input is schematics of locks (top row filled) and keys (bottom row
// filled), each 5 wide and 7 tall, separated by blank lines.
//
// Part 1 counts lock/key pairs that fit: in no column do the lock's pins
// and the key's teeth overlap. (Day 25 has no second puzzle.)
//
// Each schematic becomes five column heights; a pair fits if every
// column's heights add up to at most 5.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 16
	maxEach       = 512
)

// heights returns a schematic's column heights and whether it is a lock.
func heights(block string) ([5]int, bool, error) {
	var h [5]int
	rows := strings.Fields(block)
	if len(rows) != 7 {
		return h, false, fmt.Errorf("schematic has %d rows, not 7", len(rows))
	}
	for _, r := range rows {
		if len(r) != 5 || strings.Trim(r, ".#") != "" {
			return h, false, fmt.Errorf("malformed row %q", r)
		}
		for c := range 5 {
			if r[c] == '#' {
				h[c]++
			}
		}
	}
	for c := range 5 {
		h[c]-- // the full top or bottom row is not a pin
		if h[c] < 0 || h[c] > 5 {
			return h, false, fmt.Errorf("column %d has height %d", c, h[c])
		}
	}
	return h, rows[0] == "#####", nil
}

func day25(path string) error {
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
	locks, keys := make([][5]int, 0, maxEach), make([][5]int, 0, maxEach)
	for _, b := range strings.Split(strings.TrimSpace(string(content)), "\n\n") {
		h, lock, err := heights(b)
		if err != nil {
			return err
		}
		if len(locks) == maxEach || len(keys) == maxEach {
			return fmt.Errorf("%s: more than %d locks or keys", path, maxEach)
		}
		if lock {
			locks = append(locks, h)
		} else {
			keys = append(keys, h)
		}
	}
	fits := 0
	for _, l := range locks {
		for _, k := range keys {
			ok := true
			for c := range 5 {
				ok = ok && l[c]+k[c] <= 5
			}
			if ok {
				fits++
			}
		}
	}

	if _, err := fmt.Printf("Part 1: %d\n", fits); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
