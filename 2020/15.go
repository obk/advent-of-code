// Day 15: Rambunctious Recitation
//
// The input is the starting numbers of a game. After them, each number is
// 0 if the previous number was new, or else how many turns apart its last
// two appearances were.
//
// Part 1 is the 2020th number spoken.
// Part 2 is the 30000000th.
//
// No number can exceed the turn count, so an array indexed by number holds
// the turn each was last spoken (0 for never).
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
	inputPath     = "inputs/15.txt"
	maxInputBytes = 1 << 8
	maxStart      = 32
	lastTurn      = 30_000_000
)

// readStart parses the starting numbers.
func readStart(path string) ([]int, error) {
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
	fields := strings.Split(strings.TrimSpace(string(content)), ",")
	if len(fields) == 0 || len(fields) > maxStart {
		return nil, fmt.Errorf("%s: more than %d numbers", path, maxStart)
	}
	start := make([]int, len(fields))
	for i, f := range fields {
		start[i], err = strconv.Atoi(f)
		if err != nil || start[i] < 0 || start[i] >= lastTurn {
			return nil, fmt.Errorf("%s: invalid number %q", path, f)
		}
		for _, earlier := range start[:i] {
			if earlier == start[i] {
				return nil, fmt.Errorf("%s: %d starts the game twice", path, earlier)
			}
		}
	}
	return start, nil
}

// play returns the number spoken on the given turn, using last (one entry
// per possible number) as scratch.
func play(start []int, turns int, last []uint32) (int, error) {
	if len(start) == 0 || turns < len(start) || turns > len(last) {
		return 0, fmt.Errorf("cannot play %d turns", turns)
	}
	clear(last)
	for i, n := range start[:len(start)-1] {
		last[n] = uint32(i + 1)
	}
	spoken := start[len(start)-1]
	for turn := len(start); turn < turns; turn++ {
		// spoken was said on this turn; was it said before?
		next := 0
		if last[spoken] != 0 {
			next = turn - int(last[spoken])
		}
		last[spoken] = uint32(turn)
		if next >= len(last) {
			return 0, fmt.Errorf("number %d is past the table", next)
		}
		spoken = next
	}
	if spoken >= turns {
		return 0, fmt.Errorf("spoke %d after only %d turns", spoken, turns)
	}
	return spoken, nil
}

func day15(path string) error {
	start, err := readStart(path)
	if err != nil {
		return err
	}
	last := make([]uint32, lastTurn)
	short, err := play(start, 2020, last)
	if err != nil {
		return err
	}
	long, err := play(start, lastTurn, last)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", short, long); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day15(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
