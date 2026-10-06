// Day 11: Dumbo Octopus
//
// The input is a 10x10 grid of energy levels. Each step every level rises
// by one; an octopus above 9 flashes, raising its eight neighbours, which
// may flash in turn (once each per step); flashed octopuses reset to 0.
//
// Part 1 counts the flashes in 100 steps.
// Part 2 is the first step on which every octopus flashes.
//
// Octopuses due to flash wait on a stack.
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
	maxInputBytes = 1 << 8
	side          = 10
	cells         = side * side
	maxStepsTried = 10_000
)

// step advances the grid one step and returns how many flashed.
func step(grid *[cells]int) (int, error) {
	var stack [cells * 9]int
	n := 0
	for i := range grid {
		grid[i]++
		if grid[i] == 10 {
			stack[n], n = i, n+1
		}
	}
	flashes := 0
	for k := 0; k <= cells && n > 0; k++ {
		n, flashes = n-1, flashes+1
		i := stack[n]
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				x, y := i%side+dx, i/side+dy
				if (dx == 0 && dy == 0) || x < 0 || y < 0 || x >= side || y >= side {
					continue
				}
				grid[y*side+x]++
				if grid[y*side+x] == 10 && n == len(stack) {
					return 0, errors.New("flash stack is full")
				}
				if grid[y*side+x] == 10 {
					stack[n], n = y*side+x, n+1
				}
			}
		}
	}
	if flashes > cells {
		return 0, errors.New("an octopus flashed twice in one step")
	}
	for i := range grid {
		if grid[i] > 9 {
			grid[i] = 0
		}
	}
	return flashes, nil
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
	text := strings.Join(strings.Fields(string(content)), "")
	if len(text) != cells || strings.Trim(text, "0123456789") != "" {
		return fmt.Errorf("%s: not a %dx%d grid of digits", path, side, side)
	}
	var grid [cells]int
	for i := range cells {
		grid[i] = int(text[i] - '0')
	}
	total, sync := 0, 0
	for s := 1; s <= maxStepsTried && (sync == 0 || s <= 100); s++ {
		f, err := step(&grid)
		if err != nil {
			return err
		}
		if s <= 100 {
			total += f
		}
		if f == cells && sync == 0 {
			sync = s
		}
	}
	if total < 1 {
		return errors.New("no octopus flashed in 100 steps")
	}
	if sync == 0 {
		return fmt.Errorf("the octopuses never all flashed in %d steps", maxStepsTried)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, sync); err != nil {
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
