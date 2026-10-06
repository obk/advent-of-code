// Day 19: An Elephant Named Joseph
//
// The input is the number of elves sitting in a circle, numbered from 1. In
// turn, each elf still holding presents takes all the presents of the elf to
// its left, who then leaves the circle.
//
// Part 1 is the number of the elf that ends up with everything.
// Part 2 has each elf rob the one directly across the circle instead (the
// nearer-left one when two are equally across).
//
// The circle is a linked list in one array, next[i] being the elf to i's
// left. Each turn removes one elf, so there are exactly n-1 turns.
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
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 10
	maxElves      = 100_000_000
)

// readElves returns the number of elves.
func readElves(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid number: %w", path, err)
	}
	if n < 1 || n > maxElves {
		return 0, fmt.Errorf("%s: %d elves is outside 1..%d", path, n, maxElves)
	}
	return n, nil
}

// stealLeft plays the game where each elf robs its left neighbour and returns
// the winner's number. next is scratch space for n elves.
func stealLeft(n int, next []int32) (int, error) {
	if n < 1 || len(next) < n {
		return 0, fmt.Errorf("cannot seat %d elves in %d places", n, len(next))
	}
	for i := range n {
		next[i] = int32((i + 1) % n)
	}

	elf := 0
	for range n - 1 {
		victim := next[elf]
		if int(victim) == elf {
			return 0, errors.New("an elf tried to rob itself")
		}
		next[elf] = next[victim]
		elf = int(next[elf])
	}
	if int(next[elf]) != elf {
		return 0, errors.New("more than one elf is left")
	}
	return elf + 1, nil
}

// stealAcross plays the game where each elf robs the one across the circle.
// before always points at the elf just before the next victim. Removing an
// elf from a circle of odd size moves "across" on by one place; from an even
// circle it stays put, because the stealer moves on as well.
func stealAcross(n int, next []int32) (int, error) {
	if n < 2 || len(next) < n {
		return 0, fmt.Errorf("cannot seat %d elves in %d places", n, len(next))
	}
	for i := range n {
		next[i] = int32((i + 1) % n)
	}

	before := n/2 - 1
	for count := n; count > 1; count-- {
		victim := next[before]
		if int(victim) == before {
			return 0, errors.New("an elf tried to rob itself")
		}
		next[before] = next[victim]
		if count%2 == 1 {
			before = int(next[before])
		}
	}
	if int(next[before]) != before {
		return 0, errors.New("more than one elf is left")
	}
	return before + 1, nil
}

func day19(path string) error {
	n, err := readElves(path)
	if err != nil {
		return err
	}

	next := make([]int32, n)
	winner, err := stealLeft(n, next)
	if err != nil {
		return err
	}
	acrossWinner, err := stealAcross(n, next)
	if err != nil {
		return err
	}

	if winner < 1 || winner > n || acrossWinner < 1 || acrossWinner > n {
		return fmt.Errorf("winners %d and %d are not among the %d elves", winner, acrossWinner, n)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", winner, acrossWinner); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
