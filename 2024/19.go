// Day 19: Linen Layout
//
// The input is towel patterns (strings of colours), then designs. A design
// can be made by placing towels end to end.
//
// Part 1 counts the designs that can be made.
// Part 2 sums the number of ways to make each design.
//
// For each design, ways[i] counts the ways to make its first i colours,
// filled forwards from ways[0] = 1.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/19.txt"
	maxInputBytes = 1 << 16
	maxDesign     = 64
)

// arrangements counts the ways to make a design from the towels.
func arrangements(design string, towels []string, ways []int) (int, error) {
	if len(design) == 0 || len(design) >= len(ways) {
		return 0, fmt.Errorf("design %q is empty or too long", design)
	}
	clear(ways[:len(design)+1])
	ways[0] = 1
	for i := range len(design) {
		if ways[i] == 0 {
			continue
		}
		for _, t := range towels {
			if strings.HasPrefix(design[i:], t) {
				ways[i+len(t)] += ways[i]
			}
		}
	}
	if ways[len(design)] < 0 {
		return 0, errors.New("the count overflowed")
	}
	return ways[len(design)], nil
}

func day19(path string) error {
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
	head, designs, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	towels := strings.Split(head, ", ")
	if !found || len(towels) == 0 {
		return fmt.Errorf("%s: no towels or no designs", path)
	}
	for _, t := range towels {
		if t == "" {
			return fmt.Errorf("%s: an empty towel", path)
		}
	}
	ways := make([]int, maxDesign+1)
	possible, total := 0, 0
	for _, d := range strings.Fields(designs) {
		n, err := arrangements(d, towels, ways)
		if err != nil {
			return err
		}
		if n > 0 {
			possible++
		}
		total += n
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", possible, total); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day19(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
