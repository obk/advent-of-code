// Day 12: Hot Springs
//
// Each line of the input is a row of springs, operational '.', damaged '#'
// or unknown '?', and the sizes of the contiguous damaged groups in order.
//
// Part 1 sums, over rows, the number of ways to fill the unknowns.
// Part 2 unfolds each row first: the springs five times joined by '?',
// the groups five times.
//
// A table counts the ways to place the first j groups in the first i
// springs, filled forwards: from (i, j) a '.' or '?' can be operational,
// and a group can start at i if its springs can all be damaged and the
// one after can be operational.
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
	inputPath     = "inputs/12.txt"
	maxInputBytes = 1 << 16
	maxSprings    = 5*32 + 4
	maxGroups     = 5 * 16
)

// arrangements counts the ways to fill a row.
func arrangements(springs string, groups []int, ways *[maxSprings + 2][maxGroups + 1]int) (int, error) {
	n, g := len(springs), len(groups)
	if n > maxSprings || g > maxGroups {
		return 0, fmt.Errorf("row of %d springs and %d groups is too long", n, g)
	}
	for i := range n + 2 {
		for j := range g + 1 {
			ways[i][j] = 0
		}
	}
	ways[0][0] = 1
	for i := 0; i <= n; i++ {
		for j := 0; j <= g; j++ {
			w := ways[i][j]
			if w == 0 || i == n {
				continue
			}
			if springs[i] != '#' {
				ways[i+1][j] += w
			}
			if j == g {
				continue
			}
			size := groups[j]
			if i+size > n || strings.IndexByte(springs[i:i+size], '.') >= 0 {
				continue
			}
			switch {
			case i+size == n:
				ways[n][j+1] += w
			case springs[i+size] != '#':
				ways[i+size+1][j+1] += w
			}
		}
	}
	if ways[n][g] < 0 {
		return 0, errors.New("the count overflowed")
	}
	return ways[n][g], nil
}

// row parses one line, unfolded the given number of times, into buffers.
func row(line string, times int, springs []byte, groups []int) (string, []int, error) {
	s, nums, ok := strings.Cut(line, " ")
	if !ok || times < 1 || strings.Trim(s, ".#?") != "" {
		return "", nil, fmt.Errorf("malformed row %q", line)
	}
	springs, groups = springs[:0], groups[:0]
	f := strings.Split(nums, ",")
	if times*(len(s)+1) > cap(springs) || times*len(f) > cap(groups) {
		return "", nil, fmt.Errorf("row %q is too long", line)
	}
	for k := range times {
		if k > 0 {
			springs = append(springs, '?')
		}
		springs = append(springs, s...)
		for _, x := range f {
			v, err := strconv.Atoi(x)
			if err != nil || v < 1 {
				return "", nil, fmt.Errorf("invalid group %q", x)
			}
			groups = append(groups, v)
		}
	}
	return string(springs), groups, nil
}

func day12(path string) error {
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
	ways := new([maxSprings + 2][maxGroups + 1]int)
	springBuf, groupBuf := make([]byte, 0, maxSprings+1), make([]int, 0, maxGroups)
	var answers [2]int
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		for k, times := range [2]int{1, 5} {
			s, g, err := row(line, times, springBuf, groupBuf)
			if err != nil {
				return err
			}
			w, err := arrangements(s, g, ways)
			if err != nil {
				return err
			}
			answers[k] += w
		}
	}

	// Any arrangement repeated five times (with '.' between) fits the
	// unfolded row, so there are at least as many.
	if answers[1] < answers[0] {
		return fmt.Errorf("unfolding lowered the count from %d to %d", answers[0], answers[1])
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", answers[0], answers[1]); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day12(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
