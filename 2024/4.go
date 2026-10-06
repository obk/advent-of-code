// Day 4: Ceres Search
//
// The input is a word search of the letters X, M, A and S.
//
// Part 1 counts XMAS in any of the eight directions.
// Part 2 counts X-shaped pairs of MAS: an A whose two diagonals each read
// MAS or SAM.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
)

// at returns the letter at (x, y), or 0 off the grid.
func at(rows []string, x, y int) byte {
	if y < 0 || y >= len(rows) || x < 0 || x >= len(rows[y]) {
		return 0
	}
	return rows[y][x]
}

func day4(path string) error {
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
	if len(rows) == 0 || len(rows) > maxSide {
		return fmt.Errorf("%s: more than %d rows", path, maxSide)
	}
	xmas, crosses := 0, 0
	for y, row := range rows {
		if len(row) != len(rows[0]) || strings.Trim(row, "XMAS") != "" {
			return fmt.Errorf("%s: row %d is malformed", path, y)
		}
		if len(row) > maxSide {
			return fmt.Errorf("%s: row %d is wider than %d", path, y, maxSide)
		}
		for x := range len(row) {
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					k := 0
					for j := 0; j < 4 && at(rows, x+j*dx, y+j*dy) == "XMAS"[j]; j++ {
						k++
					}
					if k == 4 && (dx != 0 || dy != 0) {
						xmas++
					}
				}
			}
			if row[x] != 'A' {
				continue
			}
			d1 := string([]byte{at(rows, x-1, y-1), at(rows, x+1, y+1)})
			d2 := string([]byte{at(rows, x+1, y-1), at(rows, x-1, y+1)})
			if (d1 == "MS" || d1 == "SM") && (d2 == "MS" || d2 == "SM") {
				crosses++
			}
		}
	}

	if xmas < 1 {
		return errors.New("no XMAS found")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", xmas, crosses); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day4(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
