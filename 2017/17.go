// Day 17: Spinlock
//
// A circular buffer starts as [0]. For n = 1, 2, ..., the current position
// steps forward a fixed number of times (the input), n is inserted after it,
// and n becomes the current position.
//
// Part 1 is the value after 2017 once 2017 has been inserted.
// Part 2 is the value after 0 once 50 million values have been inserted.
//
// Part 2 needs no buffer: 0 never moves from the front, so only insertions
// landing at index 1 matter, and only the current index has to be tracked.
//
// The buffer is a linked list in one array: next[v] is the value after v.
// Inserting is two assignments, and stepping is following next.
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
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 10
	maxStep       = 100_000
	insertions    = 2017
	manyInserts   = 50_000_000
)

// readStep returns the step count.
func readStep(path string) (int, error) {
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
		return 0, fmt.Errorf("%s: invalid step: %w", path, err)
	}
	if n < 1 || n > maxStep {
		return 0, fmt.Errorf("%s: step %d is outside 1..%d", path, n, maxStep)
	}
	return n, nil
}

// afterLast fills the buffer up to the given number of insertions and
// returns the value after the last one inserted. next is scratch space.
func afterLast(step, count int, next []int) (int, error) {
	if step < 1 || count < 1 || len(next) < count+1 {
		return 0, fmt.Errorf("cannot insert %d values into %d slots", count, len(next))
	}
	next[0] = 0
	current := 0
	for v := 1; v <= count; v++ {
		// The buffer holds v values, so step mod v steps is enough.
		for range step % v {
			current = next[current]
		}
		next[v], next[current] = next[current], v
		current = v
	}
	if next[current] == current {
		return 0, errors.New("the buffer holds only one value")
	}
	if next[current] < 0 || next[current] > count {
		return 0, fmt.Errorf("value %d after the last is not in the buffer", next[current])
	}
	return next[current], nil
}

// afterZero tracks only where each value lands, returning the last value
// inserted straight after 0.
func afterZero(step, count int) (int, error) {
	if step < 1 || count < 1 || count > manyInserts {
		return 0, fmt.Errorf("cannot insert %d values with step %d", count, step)
	}
	pos, after := 0, -1
	for v := 1; v <= count; v++ {
		pos = (pos+step)%v + 1 // the buffer holds v values before inserting v
		if pos == 1 {
			after = v
		}
	}
	if after < 0 {
		return 0, errors.New("nothing was ever inserted after 0")
	}
	if pos < 1 || pos > count {
		return 0, fmt.Errorf("final position %d is outside the buffer of %d", pos, count+1)
	}
	return after, nil
}

func day17(path string) error {
	step, err := readStep(path)
	if err != nil {
		return err
	}
	after, err := afterLast(step, insertions, make([]int, insertions+1))
	if err != nil {
		return err
	}

	afterZeroValue, err := afterZero(step, manyInserts)
	if err != nil {
		return err
	}
	if afterZeroValue < 1 || afterZeroValue > manyInserts {
		return fmt.Errorf("value %d after 0 was never inserted", afterZeroValue)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", after, afterZeroValue); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day17(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
