// Day 4: Scratchcards
//
// Each line of the input is a card: winning numbers, '|', and the numbers
// you have.
//
// Part 1 scores each card 1 for its first match, doubling for each more,
// and sums the scores.
// Part 2: a card with n matches wins one copy of each of the next n cards
// (copies win too). The answer is the total number of cards.
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
	inputPath     = "inputs/4.txt"
	maxInputBytes = 1 << 16
	maxCards      = 1_000
	maxNumber     = 100
)

// matches counts a card's numbers that are winning numbers.
func matches(line string) (int, error) {
	_, body, found := strings.Cut(line, ": ")
	win, have, bar := strings.Cut(body, " | ")
	if !found || !bar {
		return 0, fmt.Errorf("malformed card %q", line)
	}
	var winning [maxNumber]bool
	for _, f := range strings.Fields(win) {
		v, err := strconv.Atoi(f)
		if err != nil || v < 0 || v >= maxNumber {
			return 0, fmt.Errorf("invalid number %q", f)
		}
		winning[v] = true
	}
	n := 0
	for _, f := range strings.Fields(have) {
		v, err := strconv.Atoi(f)
		if err != nil || v < 0 || v >= maxNumber {
			return 0, fmt.Errorf("invalid number %q", f)
		}
		if winning[v] {
			n++
		}
	}
	return n, nil
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxCards {
		return fmt.Errorf("%s: more than %d cards", path, maxCards)
	}
	copies := make([]int, len(lines))
	points, cards := 0, 0
	for i, l := range lines {
		m, err := matches(l)
		if err != nil {
			return err
		}
		if m > 0 {
			points += 1 << (m - 1)
		}
		copies[i]++
		cards += copies[i]
		for k := i + 1; k <= i+m && k < len(lines); k++ {
			copies[k] += copies[i]
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", points, cards); err != nil {
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
