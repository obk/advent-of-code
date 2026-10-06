// Day 17: No Such Thing as Too Much
//
// Each line of the input is the capacity of one container. Exactly 150 litres
// of eggnog must be stored, with every container used filled to the brim.
//
// Part 1 counts the combinations of containers that hold exactly 150 litres.
// Part 2 counts those combinations that use the fewest containers possible.
//
// With about twenty containers there are only about a million subsets, so
// every subset (as a bitmask) is tried.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/17.txt"
	maxInputBytes = 1 << 20
	maxLines      = 100
	maxContainers = 24 // 16 million subsets
	maxCapacity   = 1_000
	litres        = 150
)

// readLines returns the lines of the input file, which must hold between 1 and
// maxInputBytes bytes and at most maxLines lines.
func readLines(path string) ([]string, error) {
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
	lines := strings.Split(strings.TrimRight(string(content), "\r\n"), "\n")
	if len(lines) > maxLines {
		return nil, fmt.Errorf("%s: %d lines, more than the %d allowed", path, len(lines), maxLines)
	}
	return lines, nil
}

func parseInput(path string) ([]int, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	containers := make([]int, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		c, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid capacity %q: %w", line, err)
		}
		if c < 1 || c > maxCapacity {
			return nil, fmt.Errorf("capacity %d is outside 1..%d", c, maxCapacity)
		}
		containers = append(containers, c)
	}
	if len(containers) == 0 || len(containers) > maxContainers {
		return nil, fmt.Errorf("%s: %d containers, expected 1..%d", path, len(containers), maxContainers)
	}
	return containers, nil
}

// countCombinations returns how many subsets hold exactly the target, and how
// many of those use the fewest containers.
func countCombinations(containers []int, target int) (all, fewest int, err error) {
	if len(containers) > maxContainers {
		return 0, 0, fmt.Errorf("%d containers, more than %d", len(containers), maxContainers)
	}
	if target < 1 {
		return 0, 0, fmt.Errorf("target %d must be positive", target)
	}
	if len(containers) == 0 {
		return 0, 0, errors.New("no containers")
	}

	minUsed := len(containers) + 1
	for subset := range 1 << len(containers) {
		total := 0
		for i, c := range containers {
			total += (subset >> i & 1) * c
		}
		if total != target {
			continue
		}
		all++
		switch used := bits.OnesCount(uint(subset)); {
		case used < minUsed:
			minUsed, fewest = used, 1
		case used == minUsed:
			fewest++
		}
	}

	if all == 0 {
		return 0, 0, fmt.Errorf("no combination holds %d litres", target)
	}
	if fewest > all {
		return 0, 0, fmt.Errorf("%d smallest combinations out of %d", fewest, all)
	}
	return all, fewest, nil
}

func day17(path string) error {
	containers, err := parseInput(path)
	if err != nil {
		return err
	}
	all, fewest, err := countCombinations(containers, litres)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", all, fewest); err != nil {
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
