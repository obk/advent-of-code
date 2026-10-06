// Day 3: Mull It Over
//
// The input is corrupted memory containing instructions mul(X,Y), with X
// and Y of 1-3 digits, among junk; also do() and don't().
//
// Part 1 sums the products of every mul.
// Part 2 skips muls after a don't() until the next do().
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 16
)

// mul reads "mul(X,Y)" at s[i:], with X and Y of 1-3 digits, and returns
// the product.
func mul(s string, i int) (int, bool, error) {
	if i < 0 || i > len(s) {
		return 0, false, fmt.Errorf("position %d is outside the memory", i)
	}
	if !strings.HasPrefix(s[i:], "mul(") {
		return 0, false, nil
	}
	k := i + 4
	var v [2]int
	for n, sep := range [2]byte{',', ')'} {
		start := k
		for j := k; j < len(s) && j < start+3 && s[j] >= '0' && s[j] <= '9'; j++ {
			v[n], k = v[n]*10+int(s[j]-'0'), j+1
		}
		if k == start || k >= len(s) || s[k] != sep {
			return 0, false, nil
		}
		k++
	}
	if v[0] > 999 || v[1] > 999 {
		return 0, false, fmt.Errorf("operands %v have more than three digits", v)
	}
	return v[0] * v[1], true, nil
}

func day3(path string) error {
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
	s := string(content)
	all, enabled, on := 0, 0, true
	for i := range len(s) {
		switch {
		case strings.HasPrefix(s[i:], "do()"):
			on = true
		case strings.HasPrefix(s[i:], "don't()"):
			on = false
		}
		p, ok, err := mul(s, i)
		if err != nil {
			return err
		}
		if ok {
			all += p
			if on {
				enabled += p
			}
		}
	}
	if all < 1 {
		return errors.New("no mul instructions")
	}
	if enabled > all {
		return fmt.Errorf("enabled sum %d exceeds the total %d", enabled, all)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", all, enabled); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day3(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
