// Day 18: Snailfish
//
// Each line of the input is a snailfish number: a pair [a,b] whose
// elements are regular numbers or pairs. Adding two makes a pair of them,
// then reduces it: a pair nested inside four pairs explodes (its left value
// is added to the first regular number on its left, its right to the
// first on its right, and it becomes 0), else a number of 10 or more splits
// into [n/2 rounded down, n/2 rounded up]; repeat until neither applies.
// The magnitude of a pair is 3*left + 2*right.
//
// Part 1 is the magnitude of the sum of all the numbers in order.
// Part 2 is the largest magnitude of any two different numbers added.
//
// A number is kept as a flat list of its regular numbers with their
// depths, which is all that exploding and splitting need. For the
// magnitude, the two leftmost numbers at the greatest depth always form a
// pair, so pairs are folded deepest first.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/18.txt"
	maxInputBytes = 1 << 14
	maxNumbers    = 200
	maxTokens     = 256
	maxReduce     = 10_000
)

// number is a snailfish number as regular numbers with their depths.
type number struct {
	val, depth [maxTokens]int
	n          int
}

// parse reads a number.
func parse(text string) (number, error) {
	var x number
	depth := 0
	for k := range len(text) {
		switch c := text[k]; {
		case c == '[':
			depth++
		case c == ']':
			depth--
		case c == ',':
		case c >= '0' && c <= '9' && x.n < maxTokens:
			x.val[x.n], x.depth[x.n], x.n = int(c-'0'), depth, x.n+1
		default:
			return x, fmt.Errorf("unexpected %q in %q", c, text)
		}
		if depth < 0 || depth > 4 {
			return x, fmt.Errorf("%q nests too deeply", text)
		}
	}
	if depth != 0 || x.n < 2 {
		return x, fmt.Errorf("%q is not a pair", text)
	}
	return x, nil
}

// explode explodes the first pair at depth 5, reporting whether there was
// one.
func (x *number) explode() (bool, error) {
	for i := 0; i+1 < x.n; i++ {
		if x.depth[i] < 5 {
			continue
		}
		if x.depth[i] > 5 || x.depth[i+1] != 5 {
			return false, fmt.Errorf("number at depth %d is not half of an exploding pair", x.depth[i])
		}
		if i > 0 {
			x.val[i-1] += x.val[i]
		}
		if i+2 < x.n {
			x.val[i+2] += x.val[i+1]
		}
		x.val[i], x.depth[i] = 0, 4
		copy(x.val[i+1:x.n-1], x.val[i+2:x.n])
		copy(x.depth[i+1:x.n-1], x.depth[i+2:x.n])
		x.n--
		return true, nil
	}
	return false, nil
}

// split splits the first number of 10 or more, reporting whether there
// was one.
func (x *number) split() (bool, error) {
	for i := range x.n {
		if x.val[i] < 10 {
			continue
		}
		if x.n == maxTokens {
			return false, errors.New("number grew too long")
		}
		copy(x.val[i+2:x.n+1], x.val[i+1:x.n])
		copy(x.depth[i+2:x.n+1], x.depth[i+1:x.n])
		v, d := x.val[i], x.depth[i]+1
		if d > 5 {
			return false, fmt.Errorf("split at depth %d before exploding", d)
		}
		x.val[i], x.depth[i], x.val[i+1], x.depth[i+1] = v/2, d, (v+1)/2, d
		x.n++
		return true, nil
	}
	return false, nil
}

// add returns the reduced sum of a and b.
func add(a, b *number) (number, error) {
	var s number
	if a.n < 2 || b.n < 2 {
		return s, errors.New("an addend is not a pair")
	}
	if a.n+b.n > maxTokens {
		return s, errors.New("sum is too long")
	}
	for i := range a.n {
		s.val[i], s.depth[i] = a.val[i], a.depth[i]+1
	}
	for i := range b.n {
		s.val[a.n+i], s.depth[a.n+i] = b.val[i], b.depth[i]+1
	}
	s.n = a.n + b.n
	for range maxReduce {
		exploded, err := s.explode()
		if err != nil {
			return s, err
		}
		if exploded {
			continue
		}
		did, err := s.split()
		if err != nil {
			return s, err
		}
		if !did {
			return s, nil
		}
	}
	return s, fmt.Errorf("reduction took more than %d steps", maxReduce)
}

// magnitude folds the number's pairs, deepest first.
func magnitude(x number) (int, error) {
	if x.n < 1 {
		return 0, errors.New("empty number")
	}
	for range maxTokens {
		if x.n == 1 {
			return x.val[0], nil
		}
		deepest := 0
		for i := range x.n {
			deepest = max(deepest, x.depth[i])
		}
		i := x.n
		for k := x.n - 1; k >= 0; k-- {
			if x.depth[k] == deepest {
				i = k
			}
		}
		if i+1 >= x.n || x.depth[i+1] != deepest {
			return 0, errors.New("the deepest number has no partner")
		}
		x.val[i], x.depth[i] = 3*x.val[i]+2*x.val[i+1], deepest-1
		copy(x.val[i+1:x.n-1], x.val[i+2:x.n])
		copy(x.depth[i+1:x.n-1], x.depth[i+2:x.n])
		x.n--
	}
	return 0, errors.New("magnitude did not fold to one number")
}

func day18(path string) error {
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
	lines := strings.Fields(string(content))
	if len(lines) < 2 || len(lines) > maxNumbers {
		return fmt.Errorf("%s: %d numbers, expected 2..%d", path, len(lines), maxNumbers)
	}
	nums := make([]number, len(lines))
	for i, l := range lines {
		if nums[i], err = parse(l); err != nil {
			return err
		}
	}
	sum := nums[0]
	for i := 1; i < len(nums); i++ {
		if sum, err = add(&sum, &nums[i]); err != nil {
			return err
		}
	}
	total, err := magnitude(sum)
	if err != nil {
		return err
	}
	best := 0
	for i := range nums {
		for j := range nums {
			if i == j {
				continue
			}
			s, err := add(&nums[i], &nums[j])
			if err != nil {
				return err
			}
			m, err := magnitude(s)
			if err != nil {
				return err
			}
			best = max(best, m)
		}
	}

	if total < 1 || best < total/len(nums) {
		return fmt.Errorf("magnitudes %d and %d are implausible", total, best)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", total, best); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day18(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
