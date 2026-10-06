// Day 13: Shuttle Search
//
// The input is the earliest time you can leave, then the bus list: bus
// IDs (each leaves at every multiple of its ID) and 'x' for no bus.
//
// Part 1 multiplies the first bus you can catch by the minutes you wait.
// Part 2 finds the first time t at which the bus at list position i leaves
// at t+i, for every bus.
//
// Part 2 is solved one bus at a time: once t fits the first buses, adding
// their combined period keeps them fitting, so step by it until the next
// bus fits too. The IDs are prime, so the combined period is their product.
package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/13.txt"
	maxInputBytes = 1 << 10
	maxBuses      = 200
	maxID         = 1_000_000
)

// readNotes returns the departure time and the buses (0 for 'x').
func readNotes(path string) (int, []int, error) {
	if path == "" {
		return 0, nil, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, nil, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	lines := strings.Fields(string(content))
	if len(lines) != 2 {
		return 0, nil, fmt.Errorf("%s: %d lines, expected 2", path, len(lines))
	}
	depart, err := strconv.Atoi(lines[0])
	if err != nil || depart < 0 {
		return 0, nil, fmt.Errorf("%s: invalid time %q", path, lines[0])
	}
	list := strings.Split(lines[1], ",")
	if len(list) > maxBuses {
		return 0, nil, fmt.Errorf("%s: more than %d buses", path, maxBuses)
	}
	buses := make([]int, len(list))
	for i, b := range list {
		if b == "x" {
			continue
		}
		buses[i], err = strconv.Atoi(b)
		if err != nil || buses[i] < 2 || buses[i] > maxID {
			return 0, nil, fmt.Errorf("%s: invalid bus %q", path, b)
		}
	}
	return depart, buses, nil
}

// earliest returns the first bus to leave at or after depart, times the
// wait.
func earliest(depart int, buses []int) (int, error) {
	if depart < 0 {
		return 0, fmt.Errorf("negative departure time %d", depart)
	}
	best, wait := 0, math.MaxInt
	for _, b := range buses {
		if b == 0 {
			continue
		}
		if w := (b - depart%b) % b; w < wait {
			best, wait = b, w
		}
	}
	if best == 0 {
		return 0, errors.New("no buses")
	}
	return best * wait, nil
}

// contest returns the first time each bus leaves its position after it.
func contest(buses []int) (int, error) {
	t, step := 0, 1
	for i, b := range buses {
		if b == 0 {
			continue
		}
		found := false
		for k := 0; k < b && !found; k++ {
			if (t+i)%b == 0 {
				found = true
			} else {
				t += step
			}
		}
		if !found {
			return 0, fmt.Errorf("bus %d never fits; the IDs are not coprime", b)
		}
		if step > math.MaxInt64/b {
			return 0, errors.New("the combined period overflows")
		}
		step *= b
	}
	return t, nil
}

func day13(path string) error {
	depart, buses, err := readNotes(path)
	if err != nil {
		return err
	}
	first, err := earliest(depart, buses)
	if err != nil {
		return err
	}
	t, err := contest(buses)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, t); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day13(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
