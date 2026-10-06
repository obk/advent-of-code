// Day 25: Let It Snow
//
// The codes are written on an infinite grid, filling each diagonal from the
// bottom left up to the top right: (1,1), then (2,1) and (1,2), then (3,1),
// (2,2) and (1,3), and so on. The first code is 20151125 and each next one is
// the previous times 252533, modulo 33554393. The input names the row and
// column of the code wanted.
//
// Part 1 is that code. (Day 25 has no second puzzle.)
//
// The cell's position in the fill order is worked out directly: (row, col)
// lies on diagonal row+col-1, after all the cells of the diagonals before it.
// The code is then 20151125 * 252533^(position-1) mod 33554393, found by
// square-and-multiply in at most 64 steps.
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
	inputPath     = "inputs/25.txt"
	maxInputBytes = 1 << 12
	maxCoordinate = 1_000_000
)

const (
	firstCode  = 20151125
	multiplier = 252533
	modulus    = 33554393
)

// readTarget finds the row and column in the console message, which ends
// "Enter the code at row R, column C."
func readTarget(path string) (row, col int, err error) {
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

	_, rest, found := strings.Cut(string(content), "row ")
	rowText, colText, found2 := strings.Cut(rest, ", column ")
	if !found || !found2 {
		return 0, 0, fmt.Errorf("%s: no row and column found", path)
	}
	if row, err = strconv.Atoi(rowText); err != nil {
		return 0, 0, fmt.Errorf("%s: invalid row: %w", path, err)
	}
	if col, err = strconv.Atoi(strings.TrimRight(strings.TrimSpace(colText), ".")); err != nil {
		return 0, 0, fmt.Errorf("%s: invalid column: %w", path, err)
	}
	if row < 1 || col < 1 || row > maxCoordinate || col > maxCoordinate {
		return 0, 0, fmt.Errorf("%s: cell %d,%d is outside 1..%d", path, row, col, maxCoordinate)
	}
	return row, col, nil
}

// codeAt returns the code in the given cell.
func codeAt(row, col int) (int, error) {
	if row < 1 || col < 1 || row > maxCoordinate || col > maxCoordinate {
		return 0, fmt.Errorf("cell %d,%d is outside 1..%d", row, col, maxCoordinate)
	}

	diagonal := row + col - 1
	position := diagonal*(diagonal-1)/2 + col // 1-based place in the fill order
	if position < 1 {
		return 0, fmt.Errorf("cell %d,%d has position %d", row, col, position)
	}

	code, base := firstCode, multiplier
	for exp := position - 1; exp > 0; exp >>= 1 {
		if exp&1 == 1 {
			code = code * base % modulus
		}
		base = base * base % modulus
	}
	if code < 0 || code >= modulus {
		return 0, fmt.Errorf("code %d is outside 0..%d", code, modulus-1)
	}
	return code, nil
}

func day25(path string) error {
	row, col, err := readTarget(path)
	if err != nil {
		return err
	}
	code, err := codeAt(row, col)
	if err != nil {
		return err
	}

	// The fixed codes given in the puzzle text double as a check.
	if check, err := codeAt(6, 6); err != nil || check != 27995004 {
		return fmt.Errorf("code at 6,6 came out as %d, not 27995004 (%v)", check, err)
	}
	if _, err := fmt.Printf("Part 1: %d\n", code); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day25(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
