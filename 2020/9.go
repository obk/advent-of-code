// Day 9: Encoding Error
//
// Each line of the input is a number. After a preamble of 25, every number
// should be the sum of two different numbers among the 25 before it.
//
// Part 1 is the first number that is not.
// Part 2 finds a run of at least two consecutive numbers summing to it,
// and adds the run's smallest and largest numbers.
//
// The run is found with a sliding window: grow it on the right while the
// sum is too small, shrink it on the left while it is too big.
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
	inputPath     = "inputs/9.txt"
	maxInputBytes = 1 << 16
	maxNumbers    = 10_000
	maxValue      = 1 << 50
	preamble      = 25
)

// readNumbers parses one non-negative number per line.
func readNumbers(path string) ([]int, error) {
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
	fields := strings.Fields(string(content))
	if len(fields) <= preamble || len(fields) > maxNumbers {
		return nil, fmt.Errorf("%s: %d numbers, expected %d..%d", path, len(fields), preamble+1, maxNumbers)
	}
	numbers := make([]int, len(fields))
	for i, f := range fields {
		numbers[i], err = strconv.Atoi(f)
		if err != nil || numbers[i] < 0 || numbers[i] > maxValue {
			return nil, fmt.Errorf("%s: invalid number %q", path, f)
		}
	}
	return numbers, nil
}

// firstInvalid returns the first number that is not a sum of two of the
// preamble before it.
func firstInvalid(numbers []int) (int, error) {
	if len(numbers) <= preamble {
		return 0, errors.New("no numbers after the preamble")
	}
	if len(numbers) > maxNumbers {
		return 0, fmt.Errorf("more than %d numbers", maxNumbers)
	}
	for i := preamble; i < len(numbers); i++ {
		ok := false
		for a := i - preamble; a < i && !ok; a++ {
			for b := a + 1; b < i && !ok; b++ {
				ok = numbers[a] != numbers[b] && numbers[a]+numbers[b] == numbers[i]
			}
		}
		if !ok {
			return numbers[i], nil
		}
	}
	return 0, errors.New("every number is a valid sum")
}

// weakness finds a run of two or more summing to target and returns its
// smallest plus largest.
func weakness(numbers []int, target int) (int, error) {
	if target < 0 || len(numbers) < 2 {
		return 0, errors.New("negative target or too few numbers")
	}
	if target > maxValue {
		return 0, fmt.Errorf("target %d is too large", target)
	}
	lo, sum := 0, 0
	for hi, n := range numbers {
		sum += n
		for k := 0; k < len(numbers) && sum > target && lo < hi; k++ {
			sum, lo = sum-numbers[lo], lo+1
		}
		if sum == target && hi > lo {
			smallest, largest := numbers[lo], numbers[lo]
			for _, v := range numbers[lo : hi+1] {
				smallest, largest = min(smallest, v), max(largest, v)
			}
			return smallest + largest, nil
		}
	}
	return 0, fmt.Errorf("no run sums to %d", target)
}

func day9(path string) error {
	numbers, err := readNumbers(path)
	if err != nil {
		return err
	}
	bad, err := firstInvalid(numbers)
	if err != nil {
		return err
	}
	weak, err := weakness(numbers, bad)
	if err != nil {
		return err
	}

	if weak < 1 {
		return fmt.Errorf("weakness %d is not positive", weak)
	}
	// Both ends of the run are parts of the sum, so neither exceeds it.
	if weak > 2*bad {
		return fmt.Errorf("weakness %d is more than twice %d", weak, bad)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", bad, weak); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day9(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
