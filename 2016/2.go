// Day 2: Bathroom Security
//
// Each line of the input is a run of moves (U, D, L, R) on a keypad, starting
// from where the previous line ended (the "5" for the first line). A move that
// would leave the keypad is ignored. The button reached at the end of each
// line is the next digit of the code.
//
// Part 1 uses the ordinary 3x3 keypad 1-9.
// Part 2 uses a diamond-shaped keypad of 1-9 and A-D, still starting on "5".
//
// The keypad is a small grid of characters, with ' ' marking the gaps, so a
// move is allowed exactly when it lands on a non-blank cell.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 20
	maxLines      = 100
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

// code follows the lines on a keypad, starting on the given row and column,
// and writes the buttons pressed into out, which must have room for one per
// line.
func code(keypad []string, row, col int, lines []string, out []byte) ([]byte, error) {
	if len(keypad) == 0 || row < 0 || row >= len(keypad) || col < 0 || col >= len(keypad[row]) {
		return nil, fmt.Errorf("start %d,%d is not on the keypad", row, col)
	}
	if cap(out) < len(lines) {
		return nil, fmt.Errorf("room for %d buttons, need %d", cap(out), len(lines))
	}

	out = out[:0]
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		for i := range len(line) {
			r, c := row, col
			switch line[i] {
			case 'U':
				r--
			case 'D':
				r++
			case 'L':
				c--
			case 'R':
				c++
			default:
				return nil, fmt.Errorf("unknown move %q", line[i])
			}
			if r >= 0 && r < len(keypad) && c >= 0 && c < len(keypad[r]) && keypad[r][c] != ' ' {
				row, col = r, c
			}
		}
		out = append(out, keypad[row][col])
	}
	if len(out) == 0 {
		return nil, errors.New("no instructions")
	}
	if out[len(out)-1] == ' ' {
		return nil, errors.New("code ended on a gap in the keypad")
	}
	return out, nil
}

func day2(path string) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	square := []string{"123", "456", "789"}
	diamond := []string{"  1  ", " 234 ", "56789", " ABC ", "  D  "}
	first, err := code(square, 1, 1, lines, make([]byte, 0, len(lines)))
	if err != nil {
		return err
	}
	second, err := code(diamond, 2, 0, lines, make([]byte, 0, len(lines)))
	if err != nil {
		return err
	}

	if len(first) != len(second) {
		return fmt.Errorf("codes %s and %s differ in length", first, second)
	}
	if _, err := fmt.Printf("Part 1: %s\nPart 2: %s\n", first, second); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day2(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
