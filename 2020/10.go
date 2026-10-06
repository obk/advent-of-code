// Day 10: Adapter Array
//
// Each line of the input is an adapter's joltage. An adapter takes 1 to 3
// jolts less than its own; the outlet is 0 and the device is 3 above the
// highest adapter.
//
// Part 1 chains every adapter in order and multiplies the count of 1-jolt
// gaps by the count of 3-jolt gaps.
// Part 2 counts the ways to chain some of the adapters from outlet to
// device.
//
// The adapters are sorted with an insertion sort; part 2 counts ways to
// reach each joltage from the three below it.
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
	inputPath     = "inputs/10.txt"
	maxInputBytes = 1 << 12
	maxAdapters   = 200
	maxJolts      = 1_000
)

// readSorted parses the adapters and sorts them, with the outlet (0) first.
func readSorted(path string) ([]int, error) {
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
	fields := strings.Fields(string(content))
	if len(fields) == 0 || len(fields) > maxAdapters {
		return nil, fmt.Errorf("%s: more than %d adapters", path, maxAdapters)
	}
	jolts := make([]int, len(fields)+1)
	for i, f := range fields {
		v, err := strconv.Atoi(f)
		if err != nil || v < 1 || v > maxJolts {
			return nil, fmt.Errorf("%s: invalid adapter %q", path, f)
		}
		// Insertion sort: shift larger values right.
		k := i + 1
		for shift := 0; shift < i && jolts[k-1] > v; shift++ {
			jolts[k], k = jolts[k-1], k-1
		}
		if k > 1 && jolts[k-1] == v {
			return nil, fmt.Errorf("%s: two adapters of %d jolts", path, v)
		}
		jolts[k] = v
	}
	return jolts, nil
}

// gaps returns ones * threes for the full chain.
func gaps(jolts []int) (int, error) {
	if len(jolts) < 2 {
		return 0, errors.New("no adapters")
	}
	ones, threes := 0, 1 // the device is always 3 above
	for i := 1; i < len(jolts); i++ {
		switch jolts[i] - jolts[i-1] {
		case 1:
			ones++
		case 2:
		case 3:
			threes++
		default:
			return 0, fmt.Errorf("gap of %d before %d", jolts[i]-jolts[i-1], jolts[i])
		}
	}
	return ones * threes, nil
}

// arrangements counts the chains from the outlet to the highest adapter.
func arrangements(jolts []int, ways []int) (int, error) {
	if len(ways) != len(jolts) || len(jolts) < 2 {
		return 0, errors.New("wrong scratch size or no adapters")
	}
	ways[0] = 1
	for i := 1; i < len(jolts); i++ {
		ways[i] = 0
		for k := max(0, i-3); k < i; k++ {
			if jolts[i]-jolts[k] <= 3 {
				ways[i] += ways[k]
			}
		}
		if ways[i] < 0 {
			return 0, errors.New("the count overflowed")
		}
	}
	if ways[len(ways)-1] < 1 {
		return 0, errors.New("no arrangement reaches the device")
	}
	return ways[len(ways)-1], nil
}

func day10(path string) error {
	jolts, err := readSorted(path)
	if err != nil {
		return err
	}
	product, err := gaps(jolts)
	if err != nil {
		return err
	}
	ways, err := arrangements(jolts, make([]int, len(jolts)))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", product, ways); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day10(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
