// Day 6: Lanternfish
//
// The input is the timers of lanternfish. Each day a timer counts down;
// a fish at 0 resets to 6 and makes a new fish with timer 8.
//
// Part 1 counts the fish after 80 days.
// Part 2 counts them after 256 days.
//
// Fish with the same timer behave alike, so only the count at each timer
// is kept.
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
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 12
	timers        = 9
)

// grow returns the number of fish after the given days.
func grow(start [timers]int, days int) (int, error) {
	if days < 0 || days > 1_000 {
		return 0, fmt.Errorf("%d days is out of range", days)
	}
	counts := start
	for range days {
		born := counts[0]
		copy(counts[:], counts[1:])
		counts[6] += born
		counts[8] = born
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	if total < 0 {
		return 0, errors.New("the count overflowed")
	}
	return total, nil
}

func day6(path string) error {
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
	var start [timers]int
	for _, f := range strings.Split(strings.TrimSpace(string(content)), ",") {
		t, err := strconv.Atoi(f)
		if err != nil || t < 0 || t >= timers {
			return fmt.Errorf("%s: invalid timer %q", path, f)
		}
		start[t]++
	}
	short, err := grow(start, 80)
	if err != nil {
		return err
	}
	long, err := grow(start, 256)
	if err != nil {
		return err
	}

	if long < short {
		return fmt.Errorf("the school shrank from %d to %d", short, long)
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
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
