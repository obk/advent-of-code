// Day 3: Gear Ratios
//
// The input is an engine schematic: numbers, symbols, and '.' for space.
//
// Part 1 sums the numbers next to a symbol (diagonals included).
// Part 2 finds the '*' symbols next to exactly two numbers, and sums the
// products of those pairs.
//
// Each number is found by scanning rows, and its border is searched for
// symbols. Every '*' it borders gets the number counted and multiplied in,
// in tables indexed by the star's position.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 16
	maxSide       = 160
)

// schematic is the grid and the gear tables.
type schematic struct {
	rows    []string
	count   [maxSide * maxSide]int
	product [maxSide * maxSide]int
}

// symbol reports whether (x, y) holds a symbol.
func (s *schematic) symbol(x, y int) bool {
	if y < 0 || y >= len(s.rows) || x < 0 || x >= len(s.rows[y]) {
		return false
	}
	c := s.rows[y][x]
	return c != '.' && (c < '0' || c > '9')
}

// number handles the number at row y, columns x1..x2, and reports whether
// it borders a symbol.
func (s *schematic) number(y, x1, x2, v int) (bool, error) {
	if y < 0 || y >= len(s.rows) || x1 > x2 {
		return false, fmt.Errorf("number at row %d, %d..%d is out of range", y, x1, x2)
	}
	if v < 0 {
		return false, fmt.Errorf("negative part number %d", v)
	}
	near := false
	for ny := y - 1; ny <= y+1; ny++ {
		for nx := x1 - 1; nx <= x2+1; nx++ {
			if !s.symbol(nx, ny) {
				continue
			}
			near = true
			if s.rows[ny][nx] == '*' {
				k := ny*maxSide + nx
				if s.count[k] == 0 {
					s.product[k] = 1
				}
				s.count[k]++
				s.product[k] *= v
			}
		}
	}
	return near, nil
}

func day3(path string) error {
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
	s := &schematic{rows: strings.Fields(string(content))}
	if len(s.rows) == 0 {
		return fmt.Errorf("%s: empty schematic", path)
	}
	if len(s.rows) > maxSide || len(s.rows[0]) > maxSide {
		return fmt.Errorf("%s: schematic is larger than %d", path, maxSide)
	}
	parts := 0
	for y, row := range s.rows {
		if len(row) != len(s.rows[0]) {
			return fmt.Errorf("%s: row %d is %d wide, not %d", path, y, len(row), len(s.rows[0]))
		}
		for x := 0; x < len(row); x++ {
			if row[x] < '0' || row[x] > '9' {
				continue
			}
			v, end := 0, x
			for k := x; k < len(row) && row[k] >= '0' && row[k] <= '9'; k++ {
				v, end = v*10+int(row[k]-'0'), k
			}
			near, err := s.number(y, x, end, v)
			if err != nil {
				return err
			}
			if near {
				parts += v
			}
			x = end
		}
	}
	ratios := 0
	for k, c := range s.count {
		if c == 2 {
			ratios += s.product[k]
		}
	}

	if parts < 1 {
		return errors.New("no part numbers")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", parts, ratios); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
