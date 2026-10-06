// Day 7: Bridge Repair
//
// Each line of the input is a test value and numbers, "190: 10 19". The
// numbers combine left to right with operators placed between them.
//
// Part 1 sums the test values reachable with + and *.
// Part 2 sums those reachable with +, * and || (concatenation of digits).
//
// The search works backwards from the test value: the last number must
// have been added (so subtract it), multiplied (so it must divide), or
// concatenated (so the value must end in its digits). An explicit stack
// holds (numbers left, value still to make).
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
	inputPath     = "inputs/7.txt"
	maxInputBytes = 1 << 16
	maxNumbers    = 16
	maxStack      = 3 * maxNumbers
)

// solvable reports whether target can be made from nums.
func solvable(target int, nums []int, concat bool) (bool, error) {
	if len(nums) == 0 || len(nums) > maxNumbers {
		return false, fmt.Errorf("%d numbers is out of range", len(nums))
	}
	var stack [maxStack][2]int
	stack[0] = [2]int{len(nums), target}
	n := 1
	for steps := 0; n > 0 && steps < 1<<20; steps++ {
		n--
		k, v := stack[n][0], stack[n][1]
		last := nums[k-1]
		if k == 1 {
			if v == last {
				return true, nil
			}
			continue
		}
		if n+3 > maxStack {
			return false, errors.New("search stack is full")
		}
		if v-last >= 0 {
			stack[n], n = [2]int{k - 1, v - last}, n+1
		}
		if last != 0 && v%last == 0 {
			stack[n], n = [2]int{k - 1, v / last}, n+1
		}
		if concat {
			pow := 10
			for p := 0; p < 18 && pow <= last; p++ {
				pow *= 10
			}
			if v > last && (v-last)%pow == 0 {
				stack[n], n = [2]int{k - 1, (v - last) / pow}, n+1
			}
		}
	}
	return false, nil
}

func day7(path string) error {
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
	var nums [maxNumbers]int
	two, three := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		head, rest, ok := strings.Cut(line, ": ")
		target, err := strconv.Atoi(head)
		f := strings.Fields(rest)
		if !ok || err != nil || target < 0 || len(f) == 0 || len(f) > maxNumbers {
			return fmt.Errorf("%s: malformed equation %q", path, line)
		}
		for i, s := range f {
			if nums[i], err = strconv.Atoi(s); err != nil || nums[i] < 0 {
				return fmt.Errorf("%s: invalid number %q", path, s)
			}
		}
		for k, concat := range [2]bool{false, true} {
			ok, err := solvable(target, nums[:len(f)], concat)
			if err != nil {
				return err
			}
			if ok && k == 0 {
				two += target
			}
			if ok && k == 1 {
				three += target
			}
		}
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", two, three); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day7(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
