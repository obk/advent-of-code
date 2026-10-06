// Day 1: Calorie Counting
//
// The input is each elf's food calories, one item per line, elves
// separated by blank lines.
//
// Part 1 is the most calories any elf carries.
// Part 2 is the total of the top three elves.
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
	maxElves      = 1_000
	maxCalories   = 1_000_000
)

func day1(path string) error {
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
	elves := strings.Split(strings.TrimSpace(string(content)), "\n\n")
	if len(elves) < 3 || len(elves) > maxElves {
		return fmt.Errorf("%s: %d elves, expected 3..%d", path, len(elves), maxElves)
	}
	var top [3]int // largest first
	for _, elf := range elves {
		total := 0
		for _, item := range strings.Fields(elf) {
			c, err := strconv.Atoi(item)
			if err != nil || c < 0 || c > maxCalories {
				return fmt.Errorf("%s: invalid calories %q", path, item)
			}
			total += c
		}
		for k := range top {
			if total > top[k] {
				total, top[k] = top[k], total
			}
		}
	}
	if top[0] < top[1] || top[1] < top[2] {
		return errors.New("the top three are out of order")
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", top[0], top[0]+top[1]+top[2]); err != nil {
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
