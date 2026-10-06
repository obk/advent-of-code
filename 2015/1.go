// Day 1: Not Quite Lisp
//
// The input is one line of parentheses: '(' sends Santa up a floor and ')'
// down one, starting from floor 0.
//
// Part 1 is the floor he ends up on.
// Part 2 is the 1-based position of the first instruction that takes him into
// the basement (floor -1).
//
// Both come out of a single pass over the instructions.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Limits on the puzzle input, so that every loop over it has a fixed bound.
const (
	inputPath     = "inputs/1.txt"
	maxInputBytes = 1 << 20
)

// readInstructions returns the single line of parentheses.
func readInstructions(path string) (string, error) {
	if path == "" {
		return "", errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return "", fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	line := strings.TrimSpace(string(content))
	if strings.Trim(line, "()") != "" {
		return "", fmt.Errorf("%s: holds something other than parentheses", path)
	}
	return line, nil
}

// followInstructions returns the final floor and the position of the first
// step into the basement, or 0 if Santa never goes there.
func followInstructions(line string) (floor, basementAt int, err error) {
	if line == "" {
		return 0, 0, errors.New("no instructions")
	}
	if len(line) > maxInputBytes {
		return 0, 0, fmt.Errorf("%d instructions, more than %d", len(line), maxInputBytes)
	}

	for i := range len(line) {
		switch line[i] {
		case '(':
			floor++
		case ')':
			floor--
		default:
			return 0, 0, fmt.Errorf("unknown instruction %q at %d", line[i], i+1)
		}
		if floor == -1 && basementAt == 0 {
			basementAt = i + 1
		}
	}

	if floor < -len(line) || floor > len(line) {
		return 0, 0, fmt.Errorf("floor %d is unreachable in %d steps", floor, len(line))
	}
	return floor, basementAt, nil
}

func day1(path string) error {
	line, err := readInstructions(path)
	if err != nil {
		return err
	}
	floor, basementAt, err := followInstructions(line)
	if err != nil {
		return err
	}
	if basementAt == 0 {
		return errors.New("part 2: Santa never enters the basement")
	}
	if basementAt > len(line) {
		return fmt.Errorf("basement entered at %d, after the last instruction", basementAt)
	}

	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", floor, basementAt); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day1(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
