// Day 11: Cosmic Expansion
//
// The input is an image of galaxies ('#'). Every row and column with no
// galaxy is actually wider.
//
// Part 1 sums the distances between all pairs of galaxies (moving
// up/down/left/right) when empty rows and columns count twice.
// Part 2 does the same when they count a million times.
//
// Along each axis the pairwise distances add up separately; with the
// galaxies' expanded coordinates in sorted order, each gap contributes
// once per pair it separates.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/11.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
)

// axisSum sums |a-b| over pairs of galaxies along one axis, given how many
// galaxies sit at each position, with empty positions widened by factor.
func axisSum(count []int, total, factor int) (int, error) {
	if total < 0 || factor < 1 {
		return 0, errors.New("negative galaxy count or factor")
	}
	sum, left := 0, 0
	for i, c := range count {
		if i > 0 {
			step := 1
			if count[i-1] == 0 {
				step = factor
			}
			// Every pair split by this gap crosses it.
			sum += step * left * (total - left)
		}
		left += c
	}
	if left != total {
		return 0, fmt.Errorf("counted %d galaxies of %d", left, total)
	}
	return sum, nil
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
	rows := strings.Fields(string(content))
	if len(rows) == 0 || len(rows) > maxSide || len(rows[0]) > maxSide {
		return fmt.Errorf("%s: image is empty or larger than %d", path, maxSide)
	}
	rowCount, colCount := make([]int, len(rows)), make([]int, len(rows[0]))
	total := 0
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, ".#") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		for x := range len(row) {
			if row[x] == '#' {
				rowCount[y], colCount[x], total = rowCount[y]+1, colCount[x]+1, total+1
			}
		}
	}
	var answers [2]int
	for k, factor := range [2]int{2, 1_000_000} {
		a, err := axisSum(rowCount, total, factor)
		if err != nil {
			return err
		}
		b, err := axisSum(colCount, total, factor)
		if err != nil {
			return err
		}
		answers[k] = a + b
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
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
