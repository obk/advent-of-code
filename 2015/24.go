// Day 24: It Hangs in the Balance
//
// Each line of the input is a package weight. The packages must be split into
// groups of equal weight; the first group should hold as few packages as
// possible, and among those the smallest "quantum entanglement" (the product
// of its weights).
//
// Part 1 splits into three groups, part 2 into four. Both want the
// entanglement of the best first group.
//
// Subsets are bitmasks over the packages. For each group size k from 1 up,
// every k-package subset is visited in increasing order using Gosper's hack;
// one weighing exactly total/groups is a candidate. A candidate only counts if
// the remaining packages really can be split into the other groups, and that
// is checked only for candidates that would improve on the best so far. The
// first size with a valid candidate is the answer.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/24.txt"
	maxInputBytes = 1 << 20
	maxLines      = 100
	maxPackages   = 32
	maxWeight     = 10_000
)

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

func parseInput(path string) ([]int, error) {
	lines, err := readLines(path)
	if err != nil {
		return nil, err
	}

	weights := make([]int, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		w, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid weight %q: %w", line, err)
		}
		if w < 1 || w > maxWeight {
			return nil, fmt.Errorf("weight %d is outside 1..%d", w, maxWeight)
		}
		weights = append(weights, w)
	}
	if len(weights) == 0 || len(weights) > maxPackages {
		return nil, fmt.Errorf("%s: %d packages, expected 1..%d", path, len(weights), maxPackages)
	}
	return weights, nil
}

// weigh returns the total weight of the packages in mask.
func weigh(weights []int, mask uint64) (int, error) {
	if mask>>len(weights) != 0 {
		return 0, fmt.Errorf("mask %b names packages beyond the %d there are", mask, len(weights))
	}
	sum := 0
	for i, w := range weights {
		sum += int(mask>>i&1) * w
	}
	if sum > maxPackages*maxWeight {
		return 0, fmt.Errorf("packages in %b weigh %d", mask, sum)
	}
	return sum, nil
}

// entanglement returns the product of the weights in mask.
func entanglement(weights []int, mask uint64) (int, error) {
	if mask == 0 || mask>>len(weights) != 0 {
		return 0, fmt.Errorf("mask %b is empty or names missing packages", mask)
	}
	product := 1
	for i, w := range weights {
		if mask>>i&1 == 0 {
			continue
		}
		if product > (1<<62)/w {
			return 0, fmt.Errorf("entanglement of %b overflows", mask)
		}
		product *= w
	}
	return product, nil
}

// canSplit reports whether the packages in mask can be split into the given
// number of groups that each weigh target. Every submask of mask is tried as
// the first group, which is affordable because it only runs on a few
// candidates and with at most three groups to make.
func canSplit(weights []int, mask uint64, target, groups int) (bool, error) {
	if groups < 1 || groups > 3 {
		return false, fmt.Errorf("cannot check a split into %d groups", groups)
	}
	if len(weights) == 0 {
		return false, errors.New("no packages")
	}
	sum, err := weigh(weights, mask)
	if err != nil {
		return false, err
	}
	if sum != target*groups {
		return false, nil
	}
	if groups == 1 {
		return true, nil
	}

	// Walk the non-empty submasks of mask from the largest down.
	for sub := mask; sub != 0; sub = (sub - 1) & mask {
		subSum, err := weigh(weights, sub)
		if err != nil {
			return false, err
		}
		if subSum != target {
			continue
		}
		if groups == 2 {
			return true, nil
		}
		// Three groups: the rest must split into two.
		ok, err := canSplitTwo(weights, mask&^sub, target)
		if err != nil || ok {
			return ok, err
		}
	}
	return false, nil
}

// canSplitTwo is canSplit for two groups.
func canSplitTwo(weights []int, mask uint64, target int) (bool, error) {
	if target < 1 {
		return false, fmt.Errorf("target %d must be positive", target)
	}
	if mask>>len(weights) != 0 {
		return false, fmt.Errorf("mask %b names packages beyond the %d there are", mask, len(weights))
	}
	for sub := mask; sub != 0; sub = (sub - 1) & mask {
		subSum, err := weigh(weights, sub)
		if err != nil {
			return false, err
		}
		if subSum == target {
			sum, err := weigh(weights, mask)
			return err == nil && sum == 2*target, err
		}
	}
	return false, nil
}

// bestEntanglement returns the entanglement of the ideal first group when the
// packages are split into the given number of groups.
func bestEntanglement(weights []int, groups int) (int, error) {
	n := len(weights)
	if n < groups || n > maxPackages || groups < 2 || groups > 4 {
		return 0, fmt.Errorf("cannot split %d packages into %d groups", n, groups)
	}
	total, err := weigh(weights, 1<<n-1)
	if err != nil {
		return 0, err
	}
	if total%groups != 0 || total < groups {
		return 0, fmt.Errorf("total weight %d does not divide into %d groups", total, groups)
	}
	target := total / groups

	for k := 1; k <= n; k++ {
		best := -1
		// Gosper's hack steps to the next larger mask with k bits set.
		for mask, next := uint64(1)<<k-1, uint64(0); mask < 1<<n; mask = next {
			low := mask & -mask
			ripple := mask + low
			next = (((ripple ^ mask) >> 2) / low) | ripple
			if next <= mask || bits.OnesCount64(next) != k {
				return 0, fmt.Errorf("subset walk lost track at size %d", k)
			}

			sum, err := weigh(weights, mask)
			if err != nil {
				return 0, err
			}
			if sum != target {
				continue
			}
			product, err := entanglement(weights, mask)
			if err != nil {
				return 0, err
			}
			if best >= 0 && product >= best {
				continue
			}
			ok, err := canSplit(weights, (1<<n-1)&^mask, target, groups-1)
			if err != nil {
				return 0, err
			}
			if ok {
				best = product
			}
		}
		if best >= 0 {
			return best, nil
		}
	}
	return 0, fmt.Errorf("no way to split into %d equal groups", groups)
}

func day24(path string) error {
	weights, err := parseInput(path)
	if err != nil {
		return err
	}
	three, err := bestEntanglement(weights, 3)
	if err != nil {
		return err
	}
	four, err := bestEntanglement(weights, 4)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", three, four); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day24(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
