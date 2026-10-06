// Day 5: Print Queue
//
// The input is ordering rules "47|53" (47 must be printed before 53 when
// both are in an update), then updates: comma-separated page numbers.
//
// Part 1 sums the middle page of the updates already in order.
// Part 2 puts each other update in order and sums their middle pages.
//
// The rules go in a table indexed by page pair; out-of-order updates are
// sorted by insertion, using the rules as the comparison.
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
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 16
	pages         = 100
	maxUpdate     = 32
)

// readUpdate parses one update into buf.
func readUpdate(line string, buf []int) ([]int, error) {
	f := strings.Split(line, ",")
	if len(f)%2 == 0 || len(f) > cap(buf) {
		return nil, fmt.Errorf("update %q has no single middle page", line)
	}
	buf = buf[:len(f)]
	for i, s := range f {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 || v >= pages {
			return nil, fmt.Errorf("invalid page %q", s)
		}
		buf[i] = v
	}
	return buf, nil
}

// order sorts the update by the rules and reports whether it was already
// in order.
func order(u []int, before *[pages][pages]bool) (bool, error) {
	if before == nil || len(u) == 0 {
		return false, errors.New("no rules or an empty update")
	}
	sorted := true
	for i := 1; i < len(u); i++ {
		for k := i; k > 0 && before[u[k]][u[k-1]]; k-- {
			u[k-1], u[k] = u[k], u[k-1]
			sorted = false
		}
	}
	return sorted, nil
}

func day5(path string) error {
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
	rules, updates, found := strings.Cut(strings.TrimSpace(string(content)), "\n\n")
	if !found {
		return fmt.Errorf("%s: no blank line between rules and updates", path)
	}
	var before [pages][pages]bool
	for _, r := range strings.Fields(rules) {
		var a, b int
		if n, err := fmt.Sscanf(r, "%d|%d", &a, &b); err != nil || n != 2 || a < 0 || b < 0 || a >= pages || b >= pages {
			return fmt.Errorf("%s: invalid rule %q", path, r)
		}
		before[a][b] = true
	}
	buf := make([]int, 0, maxUpdate)
	good, fixed := 0, 0
	for _, line := range strings.Fields(updates) {
		u, err := readUpdate(line, buf)
		if err != nil {
			return err
		}
		ok, err := order(u, &before)
		if err != nil {
			return err
		}
		if ok {
			good += u[len(u)/2]
		} else {
			fixed += u[len(u)/2]
		}
	}

	if good < 1 {
		return errors.New("no update was already in order")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", good, fixed); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
