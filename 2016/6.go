// Day 6: Signals and Noise
//
// Each line of the input is one copy of a repeated message, garbled. Every
// column has to be recovered on its own from its letter counts.
//
// Part 1 takes the most common letter in each column.
// Part 2 takes the least common letter that appears at all.
//
// The letters are counted once into a table of columns x 26.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/6.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxWidth      = 100
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

// countLetters tallies each column's letters. All lines must be the same
// width.
func countLetters(lines []string) (counts [maxWidth][26]int, width int, err error) {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if width == 0 {
			width = len(line)
		}
		if len(line) != width || width > maxWidth {
			return counts, 0, fmt.Errorf("line %q is not %d letters wide (at most %d)", line, width, maxWidth)
		}
		for col := range width {
			c := line[col]
			if c < 'a' || c > 'z' {
				return counts, 0, fmt.Errorf("line %q holds %q", line, c)
			}
			counts[col][c-'a']++
		}
	}
	if width == 0 {
		return counts, 0, errors.New("no messages")
	}
	return counts, width, nil
}

// decode picks the most common (or least common) letter of each column into
// out; ties go to the earlier letter, and letters never seen are ignored.
func decode(counts *[maxWidth][26]int, width int, leastCommon bool, out []byte) ([]byte, error) {
	if width < 1 || width > maxWidth {
		return nil, fmt.Errorf("width %d is outside 1..%d", width, maxWidth)
	}
	if cap(out) < width {
		return nil, fmt.Errorf("room for %d letters, need %d", cap(out), width)
	}

	out = out[:width]
	for col := range width {
		best := -1
		for letter := range 26 {
			n := counts[col][letter]
			if n == 0 {
				continue
			}
			if best < 0 || (!leastCommon && n > counts[col][best]) || (leastCommon && n < counts[col][best]) {
				best = letter
			}
		}
		if best < 0 {
			return nil, fmt.Errorf("column %d holds no letters", col)
		}
		out[col] = byte('a' + best)
	}
	return out, nil
}

func day6(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
	counts, width, err := countLetters(lines)
	if err != nil {
		return err
	}
	message, err := decode(&counts, width, false, make([]byte, 0, width))
	if err != nil {
		return err
	}
	original, err := decode(&counts, width, true, make([]byte, 0, width))
	if err != nil {
		return err
	}

	if len(message) != len(original) {
		return fmt.Errorf("messages %s and %s differ in length", message, original)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", message, original); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day6(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
