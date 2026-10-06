// Day 2: Red-Nosed Reports
//
// Each line of the input is a report of levels. A report is safe if the
// levels all rise or all fall, by 1 to 3 each step.
//
// Part 1 counts the safe reports.
// Part 2 also counts reports made safe by removing any one level.
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
	inputPath     = "inputs/2.txt"
	maxInputBytes = 1 << 16
	maxLevels     = 16
)

// safe reports whether the levels, skipping index skip (-1 for none),
// are safe.
func safe(levels []int, skip int) (bool, error) {
	if skip < -1 || skip >= len(levels) {
		return false, fmt.Errorf("level %d does not exist", skip)
	}
	prev, dir := -1, 0
	for i, v := range levels {
		if i == skip {
			continue
		}
		if prev >= 0 {
			d := v - levels[prev]
			step := max(d, -d)
			if step < 1 || step > 3 || (dir != 0 && (d > 0) != (dir > 0)) {
				return false, nil
			}
			dir = d
		}
		prev = i
	}
	return true, nil
}

func day2(path string) error {
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
	var levels [maxLevels]int
	strict, damped := 0, 0
	for _, line := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || len(f) > maxLevels {
			return fmt.Errorf("%s: report %q has %d levels", path, line, len(f))
		}
		for i, s := range f {
			if levels[i], err = strconv.Atoi(s); err != nil || levels[i] < 0 {
				return fmt.Errorf("%s: invalid level %q", path, s)
			}
		}
		for skip := -1; skip < len(f); skip++ {
			ok, err := safe(levels[:len(f)], skip)
			if err != nil {
				return err
			}
			if ok && skip < 0 {
				strict++
			}
			if ok {
				damped++
				break
			}
		}
	}

	if strict < 1 {
		return errors.New("no report is safe")
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", strict, damped); err != nil {
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
