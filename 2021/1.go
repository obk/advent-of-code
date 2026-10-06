// Day 1: Sonar Sweep
//
// Each line of the input is a depth measurement.
//
// Part 1 counts measurements larger than the one before.
// Part 2 counts three-measurement windows larger than the window before.
// Two neighbouring windows share two measurements, so the second window is
// larger exactly when its new measurement beats the one the first dropped.
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
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 16
	maxDepths     = 10_000
)

// readDepths parses one depth per line.
func readDepths(path string) ([]int, error) {
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
	if len(fields) < 4 || len(fields) > maxDepths {
		return nil, fmt.Errorf("%s: %d depths, expected 4..%d", path, len(fields), maxDepths)
	}
	depths := make([]int, len(fields))
	for i, f := range fields {
		if depths[i], err = strconv.Atoi(f); err != nil {
			return nil, fmt.Errorf("%s: invalid depth %q", path, f)
		}
		if depths[i] < 0 || depths[i] > 1_000_000 {
			return nil, fmt.Errorf("%s: depth %d is out of range", path, depths[i])
		}
	}
	return depths, nil
}

// increases counts depths larger than the one gap places before.
func increases(depths []int, gap int) (int, error) {
	if gap < 1 || gap >= len(depths) {
		return 0, fmt.Errorf("gap %d does not fit %d depths", gap, len(depths))
	}
	n := 0
	for i := gap; i < len(depths); i++ {
		if depths[i] > depths[i-gap] {
			n++
		}
	}
	if n > len(depths)-gap {
		return 0, fmt.Errorf("%d increases from %d comparisons", n, len(depths)-gap)
	}
	return n, nil
}

func day1(path string) error {
	depths, err := readDepths(path)
	if err != nil {
		return err
	}
	single, err := increases(depths, 1)
	if err != nil {
		return err
	}
	windows, err := increases(depths, 3)
	if err != nil {
		return err
	}

	if single >= len(depths) {
		return fmt.Errorf("%d increases among %d depths", single, len(depths))
	}
	if windows > len(depths)-3 {
		return fmt.Errorf("%d window increases among %d windows", windows, len(depths)-2)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", single, windows); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
