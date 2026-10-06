// Day 4: Secure Container
//
// The input is a range "lo-hi" of six-digit numbers. A password has two
// equal adjacent digits and digits that never decrease from left to right.
//
// Part 1 counts the passwords in the range.
// Part 2 also needs some pair of equal digits that is not part of a longer
// run of that digit.
//
// Each number's digits are taken from the right; checking them as they come
// needs no strings.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 10
	lowest        = 100_000
	highest       = 999_999
)

// readRange returns the range's ends.
func readRange(path string) (lo, hi int, err error) {
	if path == "" {
		return 0, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	a, b, found := strings.Cut(strings.TrimSpace(string(content)), "-")
	lo, errA := strconv.Atoi(a)
	hi, errB := strconv.Atoi(b)
	if !found || errA != nil || errB != nil {
		return 0, 0, fmt.Errorf("%s: malformed range", path)
	}
	if lo < lowest || hi > highest || lo > hi {
		return 0, 0, fmt.Errorf("%s: range %d-%d is not six-digit", path, lo, hi)
	}
	return lo, hi, nil
}

// valid reports whether n has non-decreasing digits and a doubled digit,
// and whether it has a run of exactly two equal digits.
func valid(n int) (double, exactPair bool, err error) {
	if n < lowest || n > highest {
		return false, false, fmt.Errorf("%d is not six digits", n)
	}
	right, run := n%10, 1
	n /= 10
	for range 5 {
		d := n % 10
		if d > right {
			return false, false, nil
		}
		if d == right {
			run++
		} else {
			exactPair = exactPair || run == 2
			run = 1
		}
		double = double || d == right
		right, n = d, n/10
	}
	exactPair = exactPair || run == 2
	if exactPair && !double {
		return false, false, fmt.Errorf("found an exact pair but no double in %d", n)
	}
	return double, exactPair, nil
}

func day4(path string) error {
	lo, hi, err := readRange(path)
	if err != nil {
		return err
	}
	count, strict := 0, 0
	for n := lo; n <= hi; n++ {
		ok, pair, err := valid(n)
		if err != nil {
			return err
		}
		if ok {
			count++
		}
		if ok && pair {
			strict++
		}
	}

	if count > hi-lo+1 {
		return fmt.Errorf("%d passwords in a range of %d", count, hi-lo+1)
	}
	if strict > count {
		return fmt.Errorf("%d strict passwords but only %d plain ones", strict, count)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", count, strict); err != nil {
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
