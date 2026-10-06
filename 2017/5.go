// Day 5: A Maze of Twisty Trampolines, All Alike
//
// Each line of the input is a jump offset. Starting at the first, you jump by
// the current offset and then increase the offset you just used by one,
// until a jump lands outside the list.
//
// Part 1 is the number of steps to escape.
// Part 2 changes the rule: an offset of three or more is decreased by one
// after its jump instead of increased.
//
// The offsets are parsed once and copied into a work array for each run.
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
	inputPath     = "inputs/5.txt"
	maxInputBytes = 1 << 20
	maxLines      = 10_000
	maxOffset     = 1_000_000
	maxSteps      = 1_000_000_000
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
	offsets := make([]int, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		v, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid offset %q: %w", line, err)
		}
		if v < -maxOffset || v > maxOffset {
			return nil, fmt.Errorf("offset %d is outside ±%d", v, maxOffset)
		}
		offsets = append(offsets, v)
	}
	if len(offsets) == 0 {
		return nil, fmt.Errorf("%s: no offsets", path)
	}
	if len(offsets) > maxLines {
		return nil, fmt.Errorf("%s: %d offsets, more than %d", path, len(offsets), maxLines)
	}
	return offsets, nil
}

// escape runs the jumps on a copy of the offsets (in work) and returns the
// number of steps taken. With strange set, offsets of three or more shrink.
func escape(offsets, work []int, strange bool) (int, error) {
	if len(work) != len(offsets) || len(offsets) == 0 {
		return 0, errors.New("work space does not match the offsets")
	}
	if len(offsets) > maxLines {
		return 0, fmt.Errorf("%d offsets, more than %d", len(offsets), maxLines)
	}
	copy(work, offsets)

	pc := 0
	for steps := 0; steps < maxSteps; steps++ {
		if pc < 0 || pc >= len(work) {
			return steps, nil
		}
		jump := work[pc]
		if strange && jump >= 3 {
			work[pc]--
		} else {
			work[pc]++
		}
		pc += jump
	}
	return 0, fmt.Errorf("no escape within %d steps", maxSteps)
}

func day5(path string) error {
	offsets, err := parseInput(path)
	if err != nil {
		return err
	}
	work := make([]int, len(offsets))
	if len(work) == 0 {
		return errors.New("no offsets to run")
	}
	steps, err := escape(offsets, work, false)
	if err != nil {
		return err
	}
	strangeSteps, err := escape(offsets, work, true)
	if err != nil {
		return err
	}

	if steps < 1 || strangeSteps < 1 {
		return fmt.Errorf("escapes took %d and %d steps", steps, strangeSteps)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", steps, strangeSteps); err != nil {
		return err
	}
	return nil
}

func main() {
	path := inputPath
	if len(os.Args) > 1 {
		path = os.Args[1] // another input, such as an example from the puzzle
	}
	if err := day5(path); err != nil {
		// Nothing more can be done if writing to stderr fails too.
		_, _ = fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
