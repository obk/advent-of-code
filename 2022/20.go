// Day 20: Grove Positioning System
//
// The input is a circular list of numbers. Mixing moves each number, in
// its original order, forward or back by its own value.
//
// Part 1 mixes once, and sums the numbers 1000, 2000 and 3000 places after
// the 0.
// Part 2 multiplies every number by 811589153 and mixes ten times.
//
// The list is an array of original indices; moving a number shifts the
// ones between its old and new places. A number moved round the circle
// passes the other n-1, so its distance is taken modulo n-1.
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
	maxInputBytes = 1 << 16
	maxNumbers    = 10_000
	decryptionKey = 811589153
)

// mix mixes the numbers the given times and returns the grove sum.
func mix(values []int, times int, order []int) (int, error) {
	n := len(values)
	if n < 2 || len(order) != n || times < 1 {
		return 0, errors.New("too few numbers, wrong scratch size, or no mixing")
	}
	for i := range order {
		order[i] = i
	}
	for range times {
		for i, v := range values {
			pos := 0
			for p, o := range order {
				if o == i {
					pos = p
				}
			}
			to := ((pos+v)%(n-1) + (n - 1)) % (n - 1)
			if to > pos {
				copy(order[pos:to], order[pos+1:to+1])
			} else {
				copy(order[to+1:pos+1], order[to:pos])
			}
			order[to] = i
		}
	}
	zero, zeros := -1, 0
	for p, o := range order {
		if values[o] == 0 {
			zero, zeros = p, zeros+1
		}
	}
	if zero < 0 {
		return 0, errors.New("no zero in the list")
	}
	if zeros != 1 {
		return 0, fmt.Errorf("%d zeros in the list", zeros)
	}
	sum := 0
	for _, k := range [3]int{1000, 2000, 3000} {
		sum += values[order[(zero+k)%n]]
	}
	return sum, nil
}

func day20(path string) error {
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
	fields := strings.Fields(string(content))
	if len(fields) > maxNumbers {
		return fmt.Errorf("%s: more than %d numbers", path, maxNumbers)
	}
	values, scaled := make([]int, len(fields)), make([]int, len(fields))
	for i, f := range fields {
		if values[i], err = strconv.Atoi(f); err != nil {
			return fmt.Errorf("%s: invalid number %q", path, f)
		}
		scaled[i] = values[i] * decryptionKey
	}
	order := make([]int, len(values))
	once, err := mix(values, 1, order)
	if err != nil {
		return err
	}
	ten, err := mix(scaled, 10, order)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", once, ten); err != nil {
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
