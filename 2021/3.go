// Day 3: Binary Diagnostic
//
// Each line of the input is a binary number of the same width.
//
// Part 1 builds gamma from the most common bit in each position and
// epsilon from the least common, and multiplies them.
// Part 2 finds the oxygen rating (repeatedly keep the numbers with the
// most common bit in the next position, 1 on a tie, until one is left) and
// the CO2 rating (least common, 0 on a tie), and multiplies them.
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
	inputPath     = "inputs/3.txt"
	maxInputBytes = 1 << 16
	maxNumbers    = 10_000
	maxWidth      = 32
)

// readReport parses the numbers and returns their width.
func readReport(path string) ([]int, int, error) {
	if path == "" {
		return nil, 0, errors.New("no input path given")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, 0, err
	}
	if len(content) == 0 || len(content) > maxInputBytes {
		return nil, 0, fmt.Errorf("%s: size %d is outside 1..%d bytes", path, len(content), maxInputBytes)
	}
	fields := strings.Fields(string(content))
	if len(fields) == 0 {
		return nil, 0, fmt.Errorf("%s: no numbers", path)
	}
	if len(fields) > maxNumbers || len(fields[0]) > maxWidth {
		return nil, 0, fmt.Errorf("%s: too many or too wide numbers", path)
	}
	numbers := make([]int, len(fields))
	for i, f := range fields {
		v, err := strconv.ParseInt(f, 2, 64)
		if err != nil || len(f) != len(fields[0]) {
			return nil, 0, fmt.Errorf("%s: invalid number %q", path, f)
		}
		numbers[i] = int(v)
	}
	return numbers, len(fields[0]), nil
}

// ones counts the numbers still kept with bit b set, and how many are kept.
func ones(numbers []int, keep []bool, b int) (set, kept int, err error) {
	if len(keep) != len(numbers) {
		return 0, 0, errors.New("scratch size does not match the numbers")
	}
	if b < 0 || b >= maxWidth {
		return 0, 0, fmt.Errorf("bit %d is out of range", b)
	}
	for i, v := range numbers {
		if keep[i] {
			kept++
			set += v >> b & 1
		}
	}
	return set, kept, nil
}

// rating filters by the most common bit (or least, if least is set).
func rating(numbers []int, width int, keep []bool, least bool) (int, error) {
	if len(keep) != len(numbers) || width < 1 || width > maxWidth {
		return 0, errors.New("wrong scratch size or width")
	}
	for i := range keep {
		keep[i] = true
	}
	for b := width - 1; b >= 0; b-- {
		set, kept, err := ones(numbers, keep, b)
		if err != nil {
			return 0, err
		}
		if kept == 1 {
			break
		}
		want := 0
		if 2*set >= kept {
			want = 1
		}
		if least {
			want = 1 - want
		}
		for i, v := range numbers {
			keep[i] = keep[i] && v>>b&1 == want
		}
	}
	found, value := 0, 0
	for i, k := range keep {
		if k {
			found, value = found+1, numbers[i]
		}
	}
	if found != 1 {
		return 0, fmt.Errorf("%d numbers survived the filter", found)
	}
	return value, nil
}

func day3(path string) error {
	numbers, width, err := readReport(path)
	if err != nil {
		return err
	}
	keep := make([]bool, len(numbers))
	for i := range keep {
		keep[i] = true
	}
	gamma := 0
	for b := width - 1; b >= 0; b-- {
		set, kept, err := ones(numbers, keep, b)
		if err != nil {
			return err
		}
		gamma <<= 1
		if 2*set > kept {
			gamma |= 1
		}
	}
	epsilon := ^gamma & (1<<width - 1)
	oxygen, err := rating(numbers, width, keep, false)
	if err != nil {
		return err
	}
	co2, err := rating(numbers, width, keep, true)
	if err != nil {
		return err
	}

	if oxygen == co2 {
		return fmt.Errorf("both ratings are %d", oxygen)
	}
	if _, err := fmt.Printf("Part 1: %d\nPart 2: %d\n", gamma*epsilon, oxygen*co2); err != nil {
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
