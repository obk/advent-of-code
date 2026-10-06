// Day 8: Memory Maneuver
//
// The input is a list of numbers encoding a tree. Each node is a header (the
// number of children and the number of metadata entries), then its
// children, then its metadata.
//
// Part 1 is the sum of all metadata entries.
// Part 2 is the root's value: a leaf is worth its metadata sum; any other
// node is worth the sum of the children its metadata entries name (1-based,
// missing children worth 0).
//
// The tree is read with an explicit stack of the nodes still open: each frame
// counts down the children it still has to read before its metadata. Node
// values go on a second stack; when a node closes, its children's values are
// the top ones, and they are replaced by the node's own.
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
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 20
	maxDepth      = 10_000
)

// frame is one node being read.
type frame struct {
	children     int
	childrenLeft int
	metadata     int
	start        int // where this node's children's values begin
}

// readNumbers parses the list.
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
	nums := make([]int, len(fields))
	for i, f := range fields {
		if nums[i], err = strconv.Atoi(f); err != nil {
			return nil, fmt.Errorf("%s: invalid number %q: %w", path, f, err)
		}
		if nums[i] < 0 {
			return nil, fmt.Errorf("%s: negative number %d", path, nums[i])
		}
	}
	return nums, nil
}

// walk reads the tree, returning the metadata sum and the root's value.
// stack and values are scratch space.
func walk(nums []int, stack []frame, values []int) (sum, root int, err error) {
	if len(nums) < 2 || len(stack) == 0 || len(values) == 0 {
		return 0, 0, errors.New("no tree or no scratch space")
	}
	pos, top, nv := 2, 0, 0
	stack[0] = frame{children: nums[0], childrenLeft: nums[0], metadata: nums[1]}
	for range len(nums) {
		if top < 0 {
			break
		}
		f := &stack[top]
		if f.childrenLeft > 0 {
			f.childrenLeft--
			if pos+2 > len(nums) || top+1 >= len(stack) {
				return 0, 0, fmt.Errorf("tree runs past the input or deeper than %d", len(stack))
			}
			stack[top+1] = frame{children: nums[pos], childrenLeft: nums[pos], metadata: nums[pos+1], start: nv}
			pos, top = pos+2, top+1
			continue
		}
		if pos+f.metadata > len(nums) || f.start+f.children != nv {
			return 0, 0, errors.New("metadata runs past the input or child values are missing")
		}
		value := 0
		for _, m := range nums[pos : pos+f.metadata] {
			sum += m
			switch {
			case f.children == 0:
				value += m
			case m >= 1 && m <= f.children:
				value += values[f.start+m-1]
			}
		}
		if f.start >= len(values) {
			return 0, 0, errors.New("value stack overflowed")
		}
		values[f.start], nv = value, f.start+1
		pos, top = pos+f.metadata, top-1
	}
	if top >= 0 || pos != len(nums) || nv != 1 {
		return 0, 0, fmt.Errorf("tree ends at %d of %d numbers", pos, len(nums))
	}
	return sum, values[0], nil
}

func day8(path string) error {
	nums, err := readNumbers(path)
	if err != nil {
		return err
	}
	sum, root, err := walk(nums, make([]frame, maxDepth), make([]int, len(nums)/2+1))
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", sum, root); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day8(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
