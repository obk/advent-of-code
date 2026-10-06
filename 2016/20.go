// Day 20: Firewall Rules
//
// Each line of the input is an inclusive range of blocked IP addresses,
// "5-8", among the 32-bit addresses 0..4294967295.
//
// Part 1 is the lowest address not blocked.
// Part 2 is how many addresses are not blocked.
//
// The ranges are sorted by their start with an insertion sort (there are
// about a thousand), then walked in order while tracking the lowest address
// not yet covered.
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
	inputPath     = "inputs/20.txt"
	maxInputBytes = 1 << 20
	maxLines      = 5_000
	maxAddress    = 1<<32 - 1
)

// block is an inclusive range of blocked addresses.
type block struct {
	lo, hi int
}

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

func parseInput(path string) ([]block, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}
	blocks := make([]block, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		loText, hiText, found := strings.Cut(line, "-")
		if !found {
			return nil, fmt.Errorf("malformed range %q", line)
		}
		lo, err := strconv.Atoi(loText)
		if err != nil {
			return nil, fmt.Errorf("invalid range %q: %w", line, err)
		}
		hi, err := strconv.Atoi(hiText)
		if err != nil {
			return nil, fmt.Errorf("invalid range %q: %w", line, err)
		}
		if lo < 0 || hi > maxAddress || lo > hi {
			return nil, fmt.Errorf("range %q is outside 0..%d or backwards", line, maxAddress)
		}
		blocks = append(blocks, block{lo, hi})
	}
	if len(blocks) == 0 {
		return nil, fmt.Errorf("%s: no ranges found", path)
	}
	return blocks, nil
}

// sortByStart sorts the ranges in place with an insertion sort.
func sortByStart(blocks []block) error {
	if len(blocks) > maxLines {
		return fmt.Errorf("%d ranges, more than %d", len(blocks), maxLines)
	}
	for i := 1; i < len(blocks); i++ {
		for j := i; j > 0 && blocks[j].lo < blocks[j-1].lo; j-- {
			blocks[j], blocks[j-1] = blocks[j-1], blocks[j]
		}
	}
	for i := 1; i < len(blocks); i++ {
		if blocks[i].lo < blocks[i-1].lo {
			return fmt.Errorf("ranges out of order at %d", i)
		}
	}
	return nil
}

// lowestAllowed returns the lowest address no sorted range covers.
func lowestAllowed(blocks []block) (int, error) {
	if len(blocks) == 0 {
		return 0, errors.New("no ranges")
	}
	candidate := 0
	for _, b := range blocks {
		if b.lo > candidate {
			return candidate, nil
		}
		candidate = max(candidate, b.hi+1)
	}
	if candidate > maxAddress {
		return 0, errors.New("every address is blocked")
	}
	return candidate, nil
}

// countAllowed adds up the gaps between the sorted ranges and after the last.
func countAllowed(blocks []block) (int, error) {
	if len(blocks) == 0 {
		return 0, errors.New("no ranges")
	}
	allowed, candidate := 0, 0
	for _, b := range blocks {
		if b.lo > candidate {
			allowed += b.lo - candidate
		}
		candidate = max(candidate, b.hi+1)
	}
	allowed += max(maxAddress+1-candidate, 0)
	if allowed < 0 || allowed > maxAddress+1 {
		return 0, fmt.Errorf("%d allowed addresses is out of range", allowed)
	}
	return allowed, nil
}

func day20(path string) error {
	blocks, err := parseInput(path)
	if err != nil {
		return err
	}
	if err := sortByStart(blocks); err != nil {
		return err
	}
	lowest, err := lowestAllowed(blocks)
	if err != nil {
		return err
	}
	allowed, err := countAllowed(blocks)
	if err != nil {
		return err
	}

	if allowed < 1 {
		return errors.New("part 2: no address is allowed, yet part 1 found one")
	}
	if lowest > maxAddress {
		return fmt.Errorf("lowest allowed address %d is not a 32-bit address", lowest)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", lowest, allowed); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day20(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
