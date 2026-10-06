// Day 2: Password Philosophy
//
// Each line of the input is a policy and a password, "1-3 a: abcde".
//
// Part 1 counts passwords where the letter occurs between 1 and 3 times.
// Part 2 counts passwords where exactly one of positions 1 and 3 (counting
// from 1) holds the letter.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 16
	maxLines      = 10_000
)

// entry is one policy and password.
type entry struct {
	lo, hi   int
	letter   byte
	password string
}

// readEntries parses the input.
func readEntries(path string) ([]entry, error) {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: more than %d lines", path, maxLines)
	}
	entries := make([]entry, len(lines))
	for i, line := range lines {
		f := strings.Fields(strings.NewReplacer("-", " ", ":", "").Replace(line))
		if len(f) != 4 || len(f[2]) != 1 {
			return nil, fmt.Errorf("%s: malformed line %q", path, line)
		}
		lo, errLo := strconv.Atoi(f[0])
		hi, errHi := strconv.Atoi(f[1])
		if errLo != nil || errHi != nil || lo < 1 || hi < lo {
			return nil, fmt.Errorf("%s: invalid range in %q", path, line)
		}
		if f[3] == "" {
			return nil, fmt.Errorf("%s: empty password in %q", path, line)
		}
		entries[i] = entry{lo, hi, f[2][0], f[3]}
	}
	return entries, nil
}

// valid checks an entry under both policies.
func valid(e entry) (byCount, byPosition bool, err error) {
	if e.lo < 1 || e.hi < e.lo {
		return false, false, fmt.Errorf("invalid range %d-%d", e.lo, e.hi)
	}
	if e.letter < 'a' || e.letter > 'z' {
		return false, false, fmt.Errorf("policy letter %q is not a-z", e.letter)
	}
	n := strings.Count(e.password, string(e.letter))
	byCount = n >= e.lo && n <= e.hi
	// Positions count from 1; one past the end holds no letter.
	first := e.lo <= len(e.password) && e.password[e.lo-1] == e.letter
	second := e.hi <= len(e.password) && e.password[e.hi-1] == e.letter
	return byCount, first != second, nil
}

func day2(path string) error {
	entries, err := readEntries(path)
	if err != nil {
		return err
	}
	count, position := 0, 0
	for _, e := range entries {
		a, b, err := valid(e)
		if err != nil {
			return err
		}
		if a {
			count++
		}
		if b {
			position++
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, position); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
