// Day 7: The Treachery of Whales
//
// The input is the horizontal positions of crabs.
//
// Part 1 is the least fuel to line them all up, a step costing 1 fuel.
// Part 2 is the same when the n-th step of a move costs n fuel, so moving
// d costs d(d+1)/2.
//
// Positions are few enough to try every one between the lowest and
// highest crab.
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 14
	maxCrabs      = 2_000
	maxPosition   = 10_000
)

// readCrabs parses the positions.
func readCrabs(path string) ([]int, error) {
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
	if len(fields) > maxCrabs {
		return nil, fmt.Errorf("%s: more than %d crabs", path, maxCrabs)
	}
	crabs := make([]int, len(fields))
	for i, f := range fields {
		if crabs[i], err = strconv.Atoi(f); err != nil || crabs[i] < 0 || crabs[i] > maxPosition {
			return nil, fmt.Errorf("%s: invalid position %q", path, f)
		}
	}
	return crabs, nil
}

// cheapest returns the least fuel over every target position.
func cheapest(crabs []int, growing bool) (int, error) {
	if len(crabs) == 0 {
		return 0, errors.New("no crabs")
	}
	if len(crabs) > maxCrabs {
		return 0, fmt.Errorf("more than %d crabs", maxCrabs)
	}
	best := math.MaxInt
	for target := 0; target <= maxPosition; target++ {
		fuel := 0
		for _, c := range crabs {
			d := max(c-target, target-c)
			if growing {
				d = d * (d + 1) / 2
			}
			fuel += d
		}
		best = min(best, fuel)
	}
	if best == math.MaxInt {
		return 0, errors.New("no position was tried")
	}
	return best, nil
}

func day7(path string) error {
	crabs, err := readCrabs(path)
	if err != nil {
		return err
	}
	linear, err := cheapest(crabs, false)
	if err != nil {
		return err
	}
	growing, err := cheapest(crabs, true)
	if err != nil {
		return err
	}

	// Each move costs at least as much as it did, so the best does too.
	if growing < linear {
		return fmt.Errorf("growing cost %d is below the linear %d", growing, linear)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", linear, growing); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
