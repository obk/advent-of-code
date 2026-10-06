// Day 11: Chronal Charge
//
// A 300x300 grid of fuel cells, numbered from 1, has power levels set by a
// serial number (the input): with rack ID x+10, the level is the hundreds
// digit of ((rack*y + serial) * rack), minus 5.
//
// Part 1 is the top-left corner "x,y" of the 3x3 square with the most power.
// Part 2 allows any square size and asks for "x,y,size".
//
// A summed-area table gives the total of any square in four lookups.
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
	maxInputBytes = 1 << 10
	side          = 300
	maxSerial     = 1_000_000
)

// table is the summed-area table: sum[y][x] is the total of cells 1..x by
// 1..y, with row and column 0 left at zero.
type table [side + 1][side + 1]int

// readSerial returns the grid serial number.
func readSerial(path string) (int, error) {
	if path == "" {
		return 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid serial number: %w", path, err)
	}
	if n < 0 || n > maxSerial {
		return 0, fmt.Errorf("%s: serial %d is outside 0..%d", path, n, maxSerial)
	}
	return n, nil
}

// fill builds the summed-area table for the serial number.
func (t *table) fill(serial int) error {
	if t == nil {
		return errors.New("no table")
	}
	if serial < 0 || serial > maxSerial {
		return fmt.Errorf("serial %d is out of range", serial)
	}
	for y := 1; y <= side; y++ {
		for x := 1; x <= side; x++ {
			rack := x + 10
			power := (rack*y+serial)*rack/100%10 - 5
			t[y][x] = power + t[y-1][x] + t[y][x-1] - t[y-1][x-1]
		}
	}
	if t[side][side] < -5*side*side || t[side][side] > 4*side*side {
		return fmt.Errorf("total power %d is out of range", t[side][side])
	}
	return nil
}

// best returns the top-left corner and total of the best square of size n.
func (t *table) best(n int) (bx, by, total int, err error) {
	if t == nil {
		return 0, 0, 0, errors.New("no table")
	}
	if n < 1 || n > side {
		return 0, 0, 0, fmt.Errorf("square size %d is outside 1..%d", n, side)
	}
	total = -1 << 40
	for y := n; y <= side; y++ {
		for x := n; x <= side; x++ {
			s := t[y][x] - t[y-n][x] - t[y][x-n] + t[y-n][x-n]
			if s > total {
				bx, by, total = x-n+1, y-n+1, s
			}
		}
	}
	if bx < 1 || by < 1 || bx+n-1 > side || by+n-1 > side {
		return 0, 0, 0, fmt.Errorf("square %d,%d of size %d is off the grid", bx, by, n)
	}
	return bx, by, total, nil
}

func day11(path string) error {
	serial, err := readSerial(path)
	if err != nil {
		return err
	}
	t := new(table)
	if err := t.fill(serial); err != nil {
		return err
	}
	x, y, _, err := t.best(3)
	if err != nil {
		return err
	}
	bx, by, bn, bestTotal := 0, 0, 0, -1<<40
	for n := 1; n <= side; n++ {
		sx, sy, total, err := t.best(n)
		if err != nil {
			return err
		}
		if total > bestTotal {
			bx, by, bn, bestTotal = sx, sy, n, total
		}
	}

	if bn < 1 {
		return errors.New("part 2: no square was found")
	}
	if _, err := fmt.Printf("Part 1: %d,%d\nPart 2: %d,%d,%d\n", x, y, bx, by, bn); err != nil {
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
