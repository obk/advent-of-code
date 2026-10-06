// Day 21: Fractal Art
//
// An image starts as the 3x3 glider ".#./..#/###". Each iteration splits it
// into 2x2 squares (if its size is even) or 3x3 squares, and replaces each
// square by the output of the matching rule; a rule matches a square in any
// rotation or reflection. The input is the rules, "../.# => ##./#../...".
//
// Part 1 counts the pixels on after 5 iterations.
// Part 2 counts them after 18 iterations.
//
// A square is encoded as a bitmask (bit r*n+c for pixel (r, c)), so a 2x2
// square has 16 possible codes and a 3x3 one 512. Each rule is entered into
// a lookup table under all eight of its orientations, and then every square
// is replaced with one array lookup.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/21.txt"
	maxInputBytes = 1 << 20
	maxLines      = 1_000
	maxIterations = 18
	maxSide       = 2187 // the side after 18 iterations from 3
	part1Rounds   = 5
	part2Rounds   = 18
)

// rules maps every 2x2 and 3x3 square code to its output square.
type rules struct {
	two   [16]uint16  // output 3x3 code, +1 so that 0 means "no rule"
	three [512]uint16 // output 4x4 code, +1 likewise
}

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

// encode turns "#./.." into a code and its side.
func encode(pattern string) (code, n int, err error) {
	rows := strings.Split(pattern, "/")
	n = len(rows)
	if n < 2 || n > 4 {
		return 0, 0, fmt.Errorf("pattern %q is not 2, 3 or 4 rows", pattern)
	}
	for r, row := range rows {
		if len(row) != n || strings.Trim(row, ".#") != "" {
			return 0, 0, fmt.Errorf("pattern %q is not a square of '.' and '#'", pattern)
		}
		for c := range n {
			if row[c] == '#' {
				code |= 1 << (r*n + c)
			}
		}
	}
	return code, n, nil
}

// orient returns the code rotated a quarter turn (turn) or mirrored.
func orient(code, n int, turn bool) (int, error) {
	if n < 2 || n > 4 || code < 0 || code >= 1<<(n*n) {
		return 0, fmt.Errorf("cannot orient code %b of side %d", code, n)
	}
	out := 0
	for r := range n {
		for c := range n {
			if code>>(r*n+c)&1 == 0 {
				continue
			}
			if turn {
				out |= 1 << (c*n + n - 1 - r)
			} else {
				out |= 1 << (r*n + n - 1 - c)
			}
		}
	}
	return out, nil
}

// load enters every rule under all eight orientations of its input.
func (rs *rules) load(lines []string) error {
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		from, to, found := strings.Cut(strings.TrimSpace(line), " => ")
		if !found {
			return fmt.Errorf("malformed rule %q", line)
		}
		in, n, err := encode(from)
		if err != nil {
			return err
		}
		out, m, err := encode(to)
		if err != nil {
			return err
		}
		if (n != 2 || m != 3) && (n != 3 || m != 4) {
			return fmt.Errorf("rule %q maps %dx%d to %dx%d", line, n, n, m, m)
		}
		for range 4 {
			mirrored, err := orient(in, n, false)
			if err != nil {
				return err
			}
			for _, code := range [2]int{in, mirrored} {
				if n == 2 {
					rs.two[code] = uint16(out + 1)
				} else {
					rs.three[code] = uint16(out + 1)
				}
			}
			if in, err = orient(in, n, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// enhance turns the size x size image src into dst and returns the new size.
func (rs *rules) enhance(src, dst []bool, size int) (int, error) {
	n := 3
	if size%2 == 0 {
		n = 2
	}
	if size < 2 || size%n != 0 {
		return 0, fmt.Errorf("size %d does not split into squares of %d", size, n)
	}
	newSize := size / n * (n + 1)
	if newSize > maxSide || len(src) < size*size || len(dst) < newSize*newSize {
		return 0, fmt.Errorf("image of side %d does not fit", newSize)
	}
	for br := range size / n {
		for bc := range size / n {
			code := 0
			for r := range n {
				for c := range n {
					if src[(br*n+r)*size+bc*n+c] {
						code |= 1 << (r*n + c)
					}
				}
			}
			if code >= 1<<(n*n) {
				return 0, fmt.Errorf("square code %b is too big for side %d", code, n)
			}
			out := int(rs.three[code&511])
			if n == 2 {
				out = int(rs.two[code&15])
			}
			if out == 0 {
				return 0, fmt.Errorf("no rule matches square %b of side %d", code, n)
			}
			out--
			m := n + 1
			for r := range m {
				for c := range m {
					dst[(br*m+r)*newSize+bc*m+c] = out>>(r*m+c)&1 == 1
				}
			}
		}
	}
	return newSize, nil
}

// countOn runs the given number of iterations and counts the pixels on.
func (rs *rules) countOn(rounds int, a, b []bool) (int, error) {
	if rounds < 0 || rounds > maxIterations {
		return 0, fmt.Errorf("%d iterations is outside 0..%d", rounds, maxIterations)
	}
	if len(a) < 9 || len(b) < 9 {
		return 0, errors.New("image buffers are too small")
	}
	start := ".#...####"
	for i := range 9 {
		a[i] = start[i] == '#'
	}
	size := 3
	for range rounds {
		var err error
		if size, err = rs.enhance(a, b, size); err != nil {
			return 0, err
		}
		a, b = b, a
	}
	on := 0
	for i := range size * size {
		if a[i] {
			on++
		}
	}
	return on, nil
}

func day21(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	if len(lines) == 0 {
		return fmt.Errorf("%s: no rules", path)
	}
	rs := new(rules)
	if err := rs.load(lines); err != nil {
		return err
	}
	side := 3
	for range part2Rounds {
		side = side / (2 + side%2) * (3 + side%2)
	}
	if side > maxSide {
		return fmt.Errorf("the image would grow to side %d", side)
	}
	a, b := make([]bool, side*side), make([]bool, side*side)
	on, err := rs.countOn(part1Rounds, a, b)
	if err != nil {
		return err
	}
	clear(a)
	clear(b)
	more, err := rs.countOn(part2Rounds, a, b)
	if err != nil {
		return err
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", on, more); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day21(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
