// Day 2: Corruption Checksum
//
// The input is a spreadsheet: rows of whitespace-separated numbers.
//
// Part 1 is the sum over the rows of each row's largest value minus its
// smallest.
// Part 2 is the sum over the rows of the quotient of the one pair of values
// where one divides the other evenly.
//
// The rows are parsed once into a slice of rows.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxValue      = 1_000_000_000
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

func parseInput(path string) ([][]int, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	rows := make([][]int, 0, len(lines))
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		row := make([]int, len(fields))
		for i, f := range fields {
			if row[i], err = strconv.Atoi(f); err != nil {
				return nil, fmt.Errorf("invalid number in %q: %w", line, err)
			}
			if row[i] < 0 || row[i] > maxValue {
				return nil, fmt.Errorf("number %d is outside 0..%d", row[i], maxValue)
			}
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%s: empty spreadsheet", path)
	}
	return rows, nil
}

// spread returns a row's largest value minus its smallest.
func spread(row []int) (int, error) {
	if len(row) == 0 {
		return 0, errors.New("empty row")
	}
	lo, hi := row[0], row[0]
	for _, v := range row {
		lo, hi = min(lo, v), max(hi, v)
	}
	if hi < lo {
		return 0, fmt.Errorf("row %v has max %d below min %d", row, hi, lo)
	}
	return hi - lo, nil
}

// evenQuotient finds the pair where one value divides the other and returns
// the quotient; exactly one such pair must exist.
func evenQuotient(row []int) (int, error) {
	if len(row) < 2 {
		return 0, fmt.Errorf("row %v has fewer than two values", row)
	}
	quotient, found := 0, 0
	for i, a := range row {
		for j, b := range row {
			if i != j && b > 0 && a%b == 0 {
				quotient, found = a/b, found+1
			}
		}
	}
	if found != 1 {
		return 0, fmt.Errorf("row %v has %d evenly dividing pairs, expected 1", row, found)
	}
	if quotient < 1 {
		return 0, fmt.Errorf("row %v gives quotient %d", row, quotient)
	}
	return quotient, nil
}

func day2(path string) error {
	rows, err := parseInput(path)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return errors.New("empty spreadsheet")
	}
	checksum, quotients := 0, 0
	for _, row := range rows {
		s, err := spread(row)
		if err != nil {
			return err
		}
		q, err := evenQuotient(row)
		if err != nil {
			return err
		}
		checksum, quotients = checksum+s, quotients+q
	}

	if checksum < 0 || quotients < 0 {
		return fmt.Errorf("sums %d and %d should not be negative", checksum, quotients)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", checksum, quotients); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
