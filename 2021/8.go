// Day 8: Seven Segment Search
//
// Each line of the input is ten scrambled digit patterns, '|', then four
// output digits; segments a-g are wired to the wrong segments.
//
// Part 1 counts output digits that are 1, 4, 7 or 8 (the only digits with
// 2, 4, 3 and 7 segments).
// Part 2 decodes every display and sums the four-digit outputs.
//
// Patterns are 7-bit sets. 1, 4, 7 and 8 are known by size. Of the six-
// segment digits, 9 contains 4, 0 contains 1, and 6 is the other. Of the
// five-segment digits, 3 contains 1, 5 is inside 6, and 2 is the other.
package main

import (
	"errors"
	"fmt"
	"math/bits"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/8.txt"
	maxInputBytes = 1 << 16
	maxLines      = 1_000
)

// pattern turns "acf" into a bit set.
func pattern(text string) (uint8, error) {
	if text == "" || strings.Trim(text, "abcdefg") != "" {
		return 0, fmt.Errorf("invalid pattern %q", text)
	}
	var p uint8
	for k := range len(text) {
		p |= 1 << (text[k] - 'a')
	}
	return p, nil
}

// solve works out which pattern is which digit.
func solve(patterns [10]uint8) ([10]uint8, error) {
	var digit [10]uint8
	sizes := [8]int{0, 0, 1, 7, 4, 0, 0, 8} // segments -> unique digit
	for _, p := range patterns {
		if d := sizes[bits.OnesCount8(p)]; d != 0 {
			digit[d] = p
		}
	}
	if digit[1] == 0 || digit[4] == 0 || digit[7] == 0 || digit[8] == 0 {
		return digit, errors.New("missing a pattern of unique size")
	}
	for _, p := range patterns {
		if bits.OnesCount8(p) == 6 {
			switch {
			case p&digit[4] == digit[4]:
				digit[9] = p
			case p&digit[1] == digit[1]:
				digit[0] = p
			default:
				digit[6] = p
			}
		}
	}
	for _, p := range patterns {
		if bits.OnesCount8(p) == 5 {
			switch {
			case p&digit[1] == digit[1]:
				digit[3] = p
			case p&digit[6] == p:
				digit[5] = p
			default:
				digit[2] = p
			}
		}
	}
	for d, p := range digit {
		if p == 0 {
			return digit, fmt.Errorf("could not find digit %d", d)
		}
	}
	return digit, nil
}

// decode returns a display's output and how many of its digits are 1, 4,
// 7 or 8.
func decode(line string) (value, easy int, err error) {
	left, right, found := strings.Cut(line, " | ")
	in, out := strings.Fields(left), strings.Fields(right)
	if !found || len(in) != 10 || len(out) != 4 {
		return 0, 0, fmt.Errorf("malformed display %q", line)
	}
	var patterns [10]uint8
	for i, t := range in {
		if patterns[i], err = pattern(t); err != nil {
			return 0, 0, err
		}
	}
	digit, err := solve(patterns)
	if err != nil {
		return 0, 0, err
	}
	for _, t := range out {
		p, err := pattern(t)
		if err != nil {
			return 0, 0, err
		}
		d := -1
		for k, q := range digit {
			if q == p {
				d = k
			}
		}
		if d < 0 {
			return 0, 0, fmt.Errorf("output %q is no digit", t)
		}
		if d == 1 || d == 4 || d == 7 || d == 8 {
			easy++
		}
		value = value*10 + d
	}
	if value > 9999 {
		return 0, 0, fmt.Errorf("output %d has more than four digits", value)
	}
	return value, easy, nil
}

func day8(path string) error {
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
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) > maxLines {
		return fmt.Errorf("%s: more than %d displays", path, maxLines)
	}
	easy, sum := 0, 0
	for _, line := range lines {
		v, e, err := decode(line)
		if err != nil {
			return err
		}
		easy, sum = easy+e, sum+v
	}

	if easy > 4*len(lines) {
		return fmt.Errorf("%d easy digits among %d displays", easy, len(lines))
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", easy, sum); err != nil {
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
