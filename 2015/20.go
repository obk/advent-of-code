// Day 20: Infinite Elves and Infinite Houses
//
// Elf number e visits houses e, 2e, 3e, ... and leaves 10e presents at each.
// The input is a target number of presents.
//
// Part 1 is the lowest house that receives at least the target.
// Part 2 has each elf stop after 50 houses but leave 11e presents each time.
//
// House n always gets at least 10n presents from elf n alone (11n in part 2),
// so the answer is no higher than target/10. Every house up to that bound is
// tallied with a sieve: each elf adds to the houses it visits. One array of
// that size is allocated once and reused for both parts.
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
	maxInputBytes = 1 << 10
	maxTarget     = 1_000_000_000
)

// readTarget returns the target number of presents.
func readTarget(path string) (int, error) {
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
	target, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil {
		return 0, fmt.Errorf("%s: invalid target: %w", path, err)
	}
	if target < 10 || target > maxTarget {
		return 0, fmt.Errorf("%s: target %d is outside 10..%d", path, target, maxTarget)
	}
	return target, nil
}

// lowestHouse tallies the presents for houses 1..len(presents)-1, with each
// elf leaving perVisit times its number at up to visitLimit houses (0 for no
// limit), and returns the first house reaching the target.
func lowestHouse(presents []int, target, perVisit, visitLimit int) (int, error) {
	if len(presents) < 2 || perVisit < 1 || visitLimit < 0 {
		return 0, errors.New("invalid sieve parameters")
	}
	if target < 1 {
		return 0, fmt.Errorf("target %d must be positive", target)
	}
	if len(presents) > maxTarget/10+1 {
		return 0, fmt.Errorf("%d houses, more than %d", len(presents), maxTarget/10+1)
	}

	clear(presents)
	n := len(presents)
	for elf := 1; elf < n; elf++ {
		visits := 0
		for house := elf; house < n; house += elf {
			if visitLimit > 0 && visits == visitLimit {
				break
			}
			presents[house] += perVisit * elf
			visits++
		}
	}

	for house := 1; house < n; house++ {
		if presents[house] < perVisit*house {
			return 0, fmt.Errorf("house %d got %d presents, less than its own elf leaves", house, presents[house])
		}
		if presents[house] >= target {
			return house, nil
		}
	}
	return 0, fmt.Errorf("no house below %d reaches %d presents", n, target)
}

func day20(path string) error {
	target, err := readTarget(path)
	if err != nil {
		return err
	}

	// House target/10 gets at least target presents in either part.
	presents := make([]int, target/10+1)
	first, err := lowestHouse(presents, target, 10, 0)
	if err != nil {
		return err
	}
	second, err := lowestHouse(presents, target, 11, 50)
	if err != nil {
		return err
	}

	if first > target/10 || second > target/10 {
		return fmt.Errorf("houses %d and %d lie past the bound %d", first, second, target/10)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", first, second); err != nil {
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
