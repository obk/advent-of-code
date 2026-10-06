// Day 18: Settlers of The North Pole
//
// The input is a square map of acres: '.' open, '|' trees, '#' lumberyard.
// Each minute every acre changes at once by its eight neighbours: open
// becomes trees with three or more trees around, trees become a lumberyard
// with three or more lumberyards around, and a lumberyard stays only next to
// at least one lumberyard and one acre of trees.
//
// Part 1 is the resource value (trees times lumberyards) after 10 minutes.
// Part 2 is the resource value after a billion minutes.
//
// The map is a fixed array, copied by value; each minute builds the next map
// from the current one. The maps eventually repeat: mu minutes lead into a
// loop of lambda minutes. Brent's cycle-finding algorithm finds both while
// keeping only two maps, after which a billion minutes come down to
// mu + (billion - mu) mod lambda.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 16
	maxSide       = 50
	maxCycle      = 1_000_000
	billion       = 1_000_000_000
)

// area is the map; only the top-left side x side acres are used.
type area struct {
	acres [maxSide][maxSide]byte
	side  int
}

// load parses the map.
func load(path string) (area, error) {
	if path == "" {
		return area{}, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return area{}, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return area{}, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	rows := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(rows) == 0 || rows[0] == "" {
		return area{}, fmt.Errorf("%s: empty map", path)
	}
	if len(rows) > maxSide {
		return area{}, fmt.Errorf("%s: %d rows, more than %d", path, len(rows), maxSide)
	}
	a := area{side: len(rows)}
	for r, row := range rows {
		row = strings.TrimSpace(row)
		if len(row) != a.side || strings.Trim(row, ".|#") != "" {
			return area{}, fmt.Errorf("%s: row %d is not %d of '.', '|' and '#'", path, r, a.side)
		}
		copy(a.acres[r][:], row)
	}
	return a, nil
}

// next returns the map one minute later.
func (a area) next() (area, error) {
	if a.side < 1 || a.side > maxSide {
		return area{}, fmt.Errorf("side %d is outside 1..%d", a.side, maxSide)
	}
	out := area{side: a.side}
	for r := range a.side {
		for c := range a.side {
			trees, yards := 0, 0
			for dr := -1; dr <= 1; dr++ {
				for dc := -1; dc <= 1; dc++ {
					nr, nc := r+dr, c+dc
					if (dr == 0 && dc == 0) || nr < 0 || nc < 0 || nr >= a.side || nc >= a.side {
						continue
					}
					switch a.acres[nr][nc] {
					case '|':
						trees++
					case '#':
						yards++
					}
				}
			}
			cur := a.acres[r][c]
			switch {
			case cur == '.' && trees >= 3:
				cur = '|'
			case cur == '|' && yards >= 3:
				cur = '#'
			case cur == '#' && (yards == 0 || trees == 0):
				cur = '.'
			}
			out.acres[r][c] = cur
		}
	}
	return out, nil
}

// value is trees times lumberyards.
func (a area) value() (int, error) {
	if a.side < 1 || a.side > maxSide {
		return 0, fmt.Errorf("side %d is outside 1..%d", a.side, maxSide)
	}
	trees, yards := 0, 0
	for r := range a.side {
		for c := range a.side {
			switch a.acres[r][c] {
			case '|':
				trees++
			case '#':
				yards++
			}
		}
	}
	if trees+yards > a.side*a.side {
		return 0, fmt.Errorf("%d trees and %d lumberyards on %d acres", trees, yards, a.side*a.side)
	}
	return trees * yards, nil
}

// findCycle returns the lead-in mu and loop length lambda of the maps.
func findCycle(start area) (mu, lambda int, err error) {
	if start.side < 1 {
		return 0, 0, errors.New("empty map")
	}
	power, lambda := 1, 1
	tortoise := start
	hare, err := start.next()
	if err != nil {
		return 0, 0, err
	}
	for steps := 0; tortoise != hare; steps++ {
		if steps > maxCycle {
			return 0, 0, fmt.Errorf("no repeat within %d minutes", maxCycle)
		}
		if power == lambda {
			tortoise, power, lambda = hare, power*2, 0
		}
		if hare, err = hare.next(); err != nil {
			return 0, 0, err
		}
		lambda++
	}
	tortoise, hare = start, start
	for range lambda {
		if hare, err = hare.next(); err != nil {
			return 0, 0, err
		}
	}
	for mu = 0; tortoise != hare; mu++ {
		if mu > maxCycle {
			return 0, 0, errors.New("lead-in did not end")
		}
		if tortoise, err = tortoise.next(); err != nil {
			return 0, 0, err
		}
		if hare, err = hare.next(); err != nil {
			return 0, 0, err
		}
	}
	return mu, lambda, nil
}

func day18(path string) error {
	start, err := load(path)
	if err != nil {
		return err
	}
	a := start
	for range 10 {
		if a, err = a.next(); err != nil {
			return err
		}
	}
	mu, lambda, err := findCycle(start)
	if err != nil {
		return err
	}
	if lambda < 1 || mu > billion {
		return fmt.Errorf("lead-in %d and loop %d cannot reach a billion minutes", mu, lambda)
	}
	b := start
	for range mu + (billion-mu)%lambda {
		if b, err = b.next(); err != nil {
			return err
		}
	}

	v10, err := a.value()
	if err != nil {
		return err
	}
	vBillion, err := b.value()
	if err != nil {
		return err
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", v10, vBillion); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
