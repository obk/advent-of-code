// Day 1: Historian Hysteria
//
// Each line of the input holds a number from the left list and one from
// the right.
//
// Part 1 pairs the smallest with the smallest, and so on, and sums the
// differences.
// Part 2 sums each left number times how often it appears on the right.
//
// The lists are sorted by insertion, and the right list's counts kept in a
// table indexed by value.
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
	maxLines      = 2_000
	maxValue      = 100_000
)

// insert puts v into sorted list (which has room for it).
func insert(list []int, v int) ([]int, error) {
	if len(list) == cap(list) {
		return nil, errors.New("list is full")
	}
	list = append(list, v)
	for k := len(list) - 1; k > 0 && list[k-1] > list[k]; k-- {
		list[k-1], list[k] = list[k], list[k-1]
	}
	return list, nil
}

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
	f := strings.Fields(string(content))
	if len(f)%2 != 0 || len(f) > 2*maxLines {
		return fmt.Errorf("%s: numbers do not pair up", path)
	}
	left, right := make([]int, 0, len(f)/2), make([]int, 0, len(f)/2)
	count := make([]int, maxValue)
	for i := 0; i < len(f); i += 2 {
		a, errA := strconv.Atoi(f[i])
		b, errB := strconv.Atoi(f[i+1])
		if errA != nil || errB != nil || a < 0 || b < 0 || a >= maxValue || b >= maxValue {
			return fmt.Errorf("%s: invalid pair %q %q", path, f[i], f[i+1])
		}
		if left, err = insert(left, a); err != nil {
			return err
		}
		if right, err = insert(right, b); err != nil {
			return err
		}
		count[b]++
	}
	if len(left) == 0 {
		return fmt.Errorf("%s: no numbers", path)
	}
	distance, similarity := 0, 0
	for i := range left {
		distance += max(left[i]-right[i], right[i]-left[i])
		similarity += left[i] * count[left[i]]
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", distance, similarity); err != nil {
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
