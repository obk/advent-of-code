// Day 11: Plutonian Pebbles
//
// The input is numbers engraved on stones. Each blink, every stone
// changes: 0 becomes 1; a number with an even count of digits splits into
// its two halves; any other is multiplied by 2024.
//
// Part 1 counts the stones after 25 blinks.
// Part 2 counts them after 75.
//
// Stones with the same number behave alike, so only how many carry each
// number is kept, in two maps allocated once and swapped each blink.
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
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 8
	maxDistinct   = 1 << 13
)

// digits returns the number of decimal digits of v, and 10^(half of them).
func digits(v int) (int, int, error) {
	if v < 0 {
		return 0, 0, fmt.Errorf("negative stone %d", v)
	}
	n, half := 1, 1
	for p := 10; p <= v && n < 19; p *= 10 {
		n++
		if n%2 == 0 {
			half *= 10
		}
	}
	return n, half, nil
}

// blink runs the blinks and returns the number of stones.
func blink(start []int, times int, cur, next map[int]int) (int, error) {
	if times < 0 || cur == nil || next == nil {
		return 0, errors.New("no maps, or negative blinks")
	}
	clear(cur)
	for _, v := range start {
		cur[v]++
	}
	for range times {
		clear(next)
		for v, c := range cur {
			n, half, err := digits(v)
			if err != nil {
				return 0, err
			}
			if v > (1<<62)/2024 {
				return 0, fmt.Errorf("stone %d grows too large", v)
			}
			if v == 0 {
				next[1] += c
			} else if n%2 == 0 {
				next[v/half] += c
				next[v%half] += c
			} else {
				next[v*2024] += c
			}
		}
		if len(next) > maxDistinct {
			return 0, fmt.Errorf("more than %d distinct numbers", maxDistinct)
		}
		cur, next = next, cur
	}
	total := 0
	for _, c := range cur {
		total += c
	}
	return total, nil
}

func day11(path string) error {
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
	f := strings.Fields(string(content))
	if len(f) == 0 {
		return fmt.Errorf("%s: no stones", path)
	}
	start := make([]int, len(f))
	for i, s := range f {
		if start[i], err = strconv.Atoi(s); err != nil || start[i] < 0 {
			return fmt.Errorf("%s: invalid stone %q", path, s)
		}
	}
	cur, next := make(map[int]int, 2*maxDistinct), make(map[int]int, 2*maxDistinct)
	short, err := blink(start, 25, cur, next)
	if err != nil {
		return err
	}
	long, err := blink(start, 75, cur, next)
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
	if err := day11(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
